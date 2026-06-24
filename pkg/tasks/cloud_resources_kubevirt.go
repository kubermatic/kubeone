/*
Copyright 2026 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strconv"
	"time"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	kubevirttypes "k8c.io/machine-controller/sdk/cloudprovider/kubevirt"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const kubevirtAPIServerPort = 6443

func generateKubevirtControlPlaneTasks(capimachines []clusterv1alpha1.Machine) Tasks {
	tasks := Tasks{}

	for _, machine := range capimachines {
		tasks = append(tasks,
			Task{
				Description: fmt.Sprintf("Ensure KubeVirt control-plane %q VM", machine.Name),
				Predicate:   isKubevirtControlPlaneEnabled,
				Fn: func(s *state.State) error {
					return ensureKubevirtControlPlaneVM(s, machine)
				},
			},
		)
	}

	return tasks
}

func isKubevirtControlPlaneEnabled(s *state.State) bool {
	return s.Cluster.CloudProvider.Kubevirt != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func isKubevirtLoadBalancerEnabled(s *state.State) bool {
	return isKubevirtControlPlaneEnabled(s) && s.Cluster.CloudProvider.Kubevirt.ControlPlane != nil
}

func kubevirtLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "api",
	}
}

// prepareKubevirtEnv ensures the environment variables consumed by the KubeVirt
// cloud provider are populated. The provider reads the infra cluster kubeconfig
// from KUBEVIRT_KUBECONFIG and creates resources in the namespace referenced by
// POD_NAMESPACE.
func prepareKubevirtEnv(s *state.State) error {
	if ns := s.Cluster.CloudProvider.Kubevirt.InfraNamespace; ns != "" {
		if err := os.Setenv("POD_NAMESPACE", ns); err != nil {
			return fail.Runtime(err, "setting POD_NAMESPACE")
		}
	}

	if os.Getenv("KUBEVIRT_KUBECONFIG") == "" {
		kubeconfig, err := kubevirtKubeconfig(s)
		if err != nil {
			return err
		}

		if err := os.Setenv("KUBEVIRT_KUBECONFIG", base64.StdEncoding.EncodeToString(kubeconfig)); err != nil {
			return fail.Runtime(err, "setting KUBEVIRT_KUBECONFIG")
		}
	}

	return nil
}

func kubevirtKubeconfig(s *state.State) ([]byte, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	raw := providerCreds[credentials.KubevirtKubeconfigKey]
	if raw == "" {
		return nil, fail.Config(fmt.Errorf("kubevirt kubeconfig is empty"), "reading kubevirt kubeconfig")
	}

	// The value can be either a base64-encoded kubeconfig or a plain one.
	if decoded, decErr := base64.StdEncoding.DecodeString(raw); decErr == nil {
		return decoded, nil
	}

	return []byte(raw), nil
}

func kubevirtInfraClient(s *state.State) (kubernetes.Interface, string, error) {
	kubeconfig, err := kubevirtKubeconfig(s)
	if err != nil {
		return nil, "", err
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, "", fail.Config(err, "parsing kubevirt kubeconfig")
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, "", fail.Runtime(err, "creating kubevirt infra cluster client")
	}

	return client, s.Cluster.CloudProvider.Kubevirt.InfraNamespace, nil
}

func ensureKubevirtLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, ns, err := kubevirtInfraClient(s)
	if err != nil {
		return err
	}

	lbSpec := s.Cluster.CloudProvider.Kubevirt.ControlPlane.LoadBalancer
	ctx := s.Context

	svc, err := client.CoreV1().Services(ns).Get(ctx, lbSpec.Name, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		s.Logger.Debugf("no existing apiserver service found, creating a new one")
		svc, err = createKubevirtLoadBalancer(ctx, client, s.Cluster, ns)
		if err != nil {
			return fail.KubeClient(err, "creating kubevirt apiserver service")
		}
	case err != nil:
		return fail.KubeClient(err, "getting kubevirt apiserver service")
	default:
		s.Logger.Debugf("apiserver service %q already exists", lbSpec.Name)
	}

	host, port, err := kubevirtServiceEndpoint(ctx, client, svc, ns)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = host
	s.Cluster.APIEndpoint.Port = port

	return nil
}

func createKubevirtLoadBalancer(
	ctx context.Context,
	client kubernetes.Interface,
	cluster *kubeoneapi.KubeOneCluster,
	namespace string,
) (*corev1.Service, error) {
	lbSpec := cluster.CloudProvider.Kubevirt.ControlPlane.LoadBalancer

	serviceType := corev1.ServiceTypeLoadBalancer
	if lbSpec.ServiceType != "" {
		serviceType = corev1.ServiceType(lbSpec.ServiceType)
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        lbSpec.Name,
			Namespace:   namespace,
			Labels:      kubevirtLabels(cluster.Name),
			Annotations: lbSpec.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Selector: kubevirtLabels(cluster.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "kube-apiserver",
					Protocol:   corev1.ProtocolTCP,
					Port:       kubevirtAPIServerPort,
					TargetPort: intstr.FromInt32(kubevirtAPIServerPort),
				},
			},
		},
	}

	return client.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
}

// kubevirtServiceEndpoint resolves the externally reachable host and port of the
// apiserver service. For LoadBalancer services it waits for the load balancer
// ingress to be assigned, for NodePort services it returns an infra cluster node
// address combined with the allocated node port.
func kubevirtServiceEndpoint(
	ctx context.Context,
	client kubernetes.Interface,
	svc *corev1.Service,
	namespace string,
) (string, int, error) {
	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		return kubevirtNodePortEndpoint(ctx, client, svc)
	}

	for range 60 {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				return ingress.IP, kubevirtAPIServerPort, nil
			}
			if ingress.Hostname != "" {
				return ingress.Hostname, kubevirtAPIServerPort, nil
			}
		}

		time.Sleep(5 * time.Second)

		updated, err := client.CoreV1().Services(namespace).Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			return "", 0, fail.KubeClient(err, "polling kubevirt apiserver service")
		}
		svc = updated
	}

	return "", 0, fail.Cloud(fmt.Errorf("load balancer ingress was not assigned for service %q", svc.Name), "kubevirt", "waiting for load balancer")
}

func kubevirtNodePortEndpoint(
	ctx context.Context,
	client kubernetes.Interface,
	svc *corev1.Service,
) (string, int, error) {
	var nodePort int32
	for _, port := range svc.Spec.Ports {
		if port.Port == kubevirtAPIServerPort {
			nodePort = port.NodePort

			break
		}
	}

	if nodePort == 0 {
		return "", 0, fail.Cloud(fmt.Errorf("no node port was allocated for service %q", svc.Name), "kubevirt", "looking up node port")
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", 0, fail.KubeClient(err, "listing kubevirt infra cluster nodes")
	}

	var internalIP string
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeExternalIP:
				return addr.Address, int(nodePort), nil
			case corev1.NodeInternalIP:
				if internalIP == "" {
					internalIP = addr.Address
				}
			case corev1.NodeHostName, corev1.NodeInternalDNS, corev1.NodeExternalDNS:
				// hostnames/DNS records are not used as apiserver endpoints
			}
		}
	}

	if internalIP == "" {
		return "", 0, fail.Cloud(fmt.Errorf("no reachable node address found for service %q", svc.Name), "kubevirt", "looking up node address")
	}

	return internalIP, int(nodePort), nil
}

func generateKubevirtControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		labels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
			"kubeone_role":                "control-plane",
		}
		maps.Copy(labels, kubevirtLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, labels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var kubevirtConfig kubevirttypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &kubevirtConfig); err != nil {
				return nil, fail.Config(err, "decode kubevirt config")
			}

			// ClusterName is required by the provider and is used to label the
			// created resources. The infra cluster kubeconfig is intentionally
			// left empty here so the provider resolves it from the
			// KUBEVIRT_KUBECONFIG environment variable.
			kubevirtConfig.ClusterName = providerconfig.ConfigVarString{Value: clusterName}

			kubevirtSpec, err := json.Marshal(kubevirtConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderKubeVirt,
				CloudProviderSpec: runtime.RawExtension{
					Raw: kubevirtSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "kubevirt", "json marshaling provider config")
			}

			name := fmt.Sprintf("%s-%s-%d", clusterName, node.Name, idx)
			machines = append(machines, clusterv1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					UID:  types.UID(name),
				},
				Spec: clusterv1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        name,
						Labels:      node.NodeSettings.Labels,
						Annotations: node.NodeSettings.Annotations,
					},
					Taints: node.NodeSettings.Taints,
					Versions: clusterv1alpha1.MachineVersionInfo{
						Kubelet: kubeletVersion,
					},
					ProviderSpec: clusterv1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: providerSpecRaw,
						},
					},
				},
			})
		}
	}

	return machines, nil
}

func ensureKubevirtControlPlaneVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
	if err := prepareKubevirtEnv(s); err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromKubevirtMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func lookupKubevirtVMs(s *state.State) error {
	if err := prepareKubevirtEnv(s); err != nil {
		return err
	}

	capimachines, err := generateKubevirtControlPlaneMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	provMachines, err := provisioner.FindMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromKubevirtMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func lookupKubevirtLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, ns, err := kubevirtInfraClient(s)
	if err != nil {
		return err
	}

	lbName := s.Cluster.CloudProvider.Kubevirt.ControlPlane.LoadBalancer.Name
	svc, err := client.CoreV1().Services(ns).Get(s.Context, lbName, metav1.GetOptions{})
	if err != nil {
		return fail.KubeClient(err, "getting kubevirt apiserver service")
	}

	host, port, err := kubevirtServiceEndpoint(s.Context, client, svc, ns)
	if err != nil {
		return err
	}

	s.Cluster.APIEndpoint.Host = host
	s.Cluster.APIEndpoint.Port = port

	return nil
}

// hostConfigsFromKubevirtMachines builds the control-plane host configs for
// KubeVirt machines. The KubeVirt cloud provider only ever reports the VM's
// in-cluster (internal) IP, so it is used as both the public and private address
// to make sure SSH (which connects to the public address) can reach the VM.
func hostConfigsFromKubevirtMachines(machines []provisioner.Machine, nodeSets []kubeoneapi.NodeSet) []kubeoneapi.HostConfig {
	var hosts []kubeoneapi.HostConfig
	idx := 0

	for _, nodeSet := range nodeSets {
		sshUsername := nodeSet.SSH.Username
		if sshUsername == "" {
			sshUsername = "root"
		}

		for range nodeSet.Replicas {
			if idx >= len(machines) {
				break
			}

			m := machines[idx]
			address := m.PublicAddress
			if address == "" {
				address = m.PrivateAddress
			}

			host := kubeoneapi.HostConfig{
				PublicAddress:        address,
				PrivateAddress:       m.PrivateAddress,
				Hostname:             m.Hostname,
				SSHUsername:          sshUsername,
				SSHPort:              nodeSet.SSH.Port,
				SSHPrivateKeyFile:    nodeSet.SSH.PrivateKeyFile,
				SSHCertFile:          nodeSet.SSH.CertFile,
				SSHHostPublicKey:     nodeSet.SSH.HostPublicKey,
				SSHAgentSocket:       nodeSet.SSH.AgentSocket,
				Bastion:              nodeSet.SSH.Bastion,
				BastionPort:          nodeSet.SSH.BastionPort,
				BastionUser:          nodeSet.SSH.BastionUser,
				BastionHostPublicKey: nodeSet.SSH.BastionHostPublicKey,
				OperatingSystem:      nodeSet.OperatingSystem,
				Labels:               nodeSet.NodeSettings.Labels,
				Annotations:          nodeSet.NodeSettings.Annotations,
				Taints:               nodeSet.NodeSettings.Taints,
			}

			hosts = append(hosts, host)
			idx++
		}
	}

	return hosts
}
