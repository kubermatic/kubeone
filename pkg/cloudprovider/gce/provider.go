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

package gce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	gcetypes "k8c.io/machine-controller/sdk/cloudprovider/gce"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	apiServerPort = 6443

	// machineUIDLabel is set by the machine-controller GCE driver to the
	// Machine UID, it finds instances by it.
	machineUIDLabel = "machine_uid"
)

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
	_ cloudprovider.CleanupPreparer           = &Provider{}
)

func init() {
	cloudprovider.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "GCE" }

func (p *Provider) Enabled(s *state.State) bool {
	gce := s.Cluster.CloudProvider.GCE

	return gce != nil && gce.ControlPlane != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *Provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.GCE != nil
}

func (p *Provider) HasLoadBalancer(s *state.State) bool {
	return p.Enabled(s)
}

func (p *Provider) GenerateMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateGCEControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
}

func (p *Provider) EnsureVM(s *state.State, nodeSet kubeoneapi.NodeSet, capimachine clusterv1alpha1.Machine) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromMachines(provMachines, []clusterv1alpha1.Machine{capimachine}, []kubeoneapi.NodeSet{nodeSet})...)

	return ensureLoadBalancerEndpoint(s, capimachine)
}

func (p *Provider) LookupVMs(s *state.State) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	capimachines, err := p.GenerateMachines(
		s.Cluster.Name,
		s.Cluster.ControlPlane.NodeSets,
		s.Cluster.Versions.Kubernetes,
	)
	if err != nil {
		return err
	}

	if s.DestroyControlPlaneMachines {
		return lookupExistingVMs(s, capimachines)
	}

	provMachines, err := provisioner.FindMachines(s.Context, capimachines, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromMachines(provMachines, capimachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

// lookupExistingVMs adds the hosts of the machines whose instances exist and
// skips the others. It's used when the control plane is going to be
// destroyed, so that reset can clean up a partially created or already
// partially deleted control plane, where FindMachines would fail.
func lookupExistingVMs(s *state.State, capimachines []clusterv1alpha1.Machine) error {
	nodeSets, err := cloudprovider.MachineNodeSets(capimachines, s.Cluster.ControlPlane.NodeSets)
	if err != nil {
		return err
	}

	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	for i, machine := range capimachines {
		zone, err := machineZone(machine, cfg.ProjectID)
		if err != nil {
			return err
		}

		_, err = svc.Instances.Get(cfg.ProjectID, zone, machine.Name).Context(s.Context).Do()
		switch {
		case isNotFound(err):
			s.Logger.Warnf("control plane instance %q not found in zone %q, skipping it", machine.Name, zone)

			continue
		case err != nil:
			return fail.Cloud(err, "gce", "getting instance %q", machine.Name)
		}

		provMachines, err := provisioner.FindMachines(s.Context, []clusterv1alpha1.Machine{machine}, s.Logger)
		if err != nil {
			return err
		}

		s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, hostConfigsFromMachines(provMachines, []clusterv1alpha1.Machine{machine}, []kubeoneapi.NodeSet{nodeSets[i]})...)
	}

	return nil
}

// PrepareCleanup satisfies cloudprovider.CleanupPreparer. CleanupVMs deletes
// the instances through the machine-controller driver without going through
// EnsureVM or LookupVMs, so the credentials have to be exported here too.
//
// The driver deletes instances by name only, so before anything is deleted
// every existing instance with the name of a control plane machine must carry
// the labels KubeOne and machine-controller set when creating it.
func (p *Provider) PrepareCleanup(s *state.State) error {
	if err := exportCredentials(s); err != nil {
		return err
	}

	machines, err := p.GenerateMachines(s.Cluster.Name, s.Cluster.ControlPlane.NodeSets, s.Cluster.Versions.Kubernetes)
	if err != nil {
		return err
	}

	cfg, err := resolveLB(s)
	if err != nil {
		return err
	}

	svc, err := newComputeService(s)
	if err != nil {
		return err
	}

	return checkInstancesOwned(s.Context, svc, cfg.ProjectID, s.Cluster.Name, machines)
}

// checkInstancesOwned fails if an instance with the name of one of the
// machines exists, but wasn't created by KubeOne for the cluster.
func checkInstancesOwned(ctx context.Context, svc *compute.Service, projectID, clusterName string, machines []clusterv1alpha1.Machine) error {
	for _, machine := range machines {
		zone, err := machineZone(machine, projectID)
		if err != nil {
			return err
		}

		instance, err := svc.Instances.Get(projectID, zone, machine.Name).Context(ctx).Do()
		switch {
		case isNotFound(err):
			continue
		case err != nil:
			return fail.Cloud(err, "gce", "getting instance %q", machine.Name)
		}

		want := gceLabels(clusterName)
		want[machineUIDLabel] = string(machine.UID)
		for key, value := range want {
			if instance.Labels[key] != value {
				return fail.Cloud(
					fmt.Errorf("instance %q in zone %q was not created by KubeOne for cluster %q (label %s=%q, expected %q), refusing to delete any control plane instance",
						machine.Name, zone, clusterName, key, instance.Labels[key], value),
					"gce", "checking control plane instance %q", machine.Name,
				)
			}
		}
	}

	return nil
}

// hostConfigsFromMachines maps the found or created machines to hosts named
// after their instances. FindMachines and FindOrCreateMachines return the
// machines in the order of capimachines.
//
// The machine-controller GCE driver reports the instance name, its zonal and
// its global internal DNS name all as InternalDNS addresses, and which one
// becomes the hostname depends on map iteration order. The internal DNS names
// can exceed 63 characters, the limit of the kubernetes.io/hostname label the
// kubelet sets from the node name, so node registration fails. The instance
// name is validated to be at most 63 characters, it's the hostname of the VM,
// and the GCE cloud controller manager maps node names to instances by it.
func hostConfigsFromMachines(machines []provisioner.Machine, capimachines []clusterv1alpha1.Machine, nodeSets []kubeoneapi.NodeSet) []kubeoneapi.HostConfig {
	hosts := cloudprovider.HostConfigsFromMachines(machines, nodeSets)
	for i := range hosts {
		if i < len(capimachines) {
			hosts[i].Hostname = capimachines[i].Name
		}
	}

	return hosts
}

func newComputeService(s *state.State) (*compute.Service, error) {
	serviceAccount, err := serviceAccountJSON(s)
	if err != nil {
		return nil, err
	}

	svc, err := compute.NewService(s.Context, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(serviceAccount)))
	if err != nil {
		return nil, fail.Cloud(err, "gce", "creating compute client")
	}

	return svc, nil
}

func serviceAccountJSON(s *state.State) (string, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return "", err
	}

	// The credentials are keyed by their machine-controller names, the
	// service account JSON (GOOGLE_CREDENTIALS) is GOOGLE_SERVICE_ACCOUNT.
	return providerCreds[credentials.GoogleServiceAccountKeyMC], nil
}

// serviceAccountProjectID returns the project the service account belongs to.
func serviceAccountProjectID(serviceAccount string) (string, error) {
	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(serviceAccount), &sa); err != nil {
		return "", fail.Config(err, "parsing google service account")
	}

	return sa.ProjectID, nil
}

// exportCredentials publishes the service account for the machine-controller
// GCE driver, which KubeOne calls in-process. The driver reads it from the
// GOOGLE_SERVICE_ACCOUNT environment variable (base64 encoded JSON), there is
// no Kubernetes secret to read from at this point. The service account is
// deliberately not written into the Machine's providerSpec, which gets
// marshalled and logged.
func exportCredentials(s *state.State) error {
	creds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeMC)
	if err != nil {
		return err
	}

	if err := os.Setenv(credentials.GoogleServiceAccountKeyMC, creds[credentials.GoogleServiceAccountKeyMC]); err != nil {
		return fail.Runtime(err, "setting %s", credentials.GoogleServiceAccountKeyMC)
	}

	return nil
}

// waitForOperation waits until the given Compute Engine operation is DONE and
// returns its error, if any. The operations' Wait methods return once the
// operation is DONE or the request approaches its deadline (about 2 minutes),
// so they are called in a loop.
func waitForOperation(ctx context.Context, svc *compute.Service, projectID string, op *compute.Operation) error {
	name := op.Name

	for op.Status != "DONE" {
		var err error

		switch {
		case op.Zone != "":
			op, err = svc.ZoneOperations.Wait(projectID, path.Base(op.Zone), name).Context(ctx).Do()
		case op.Region != "":
			op, err = svc.RegionOperations.Wait(projectID, path.Base(op.Region), name).Context(ctx).Do()
		default:
			op, err = svc.GlobalOperations.Wait(projectID, name).Context(ctx).Do()
		}
		if err != nil {
			return fail.Cloud(err, "gce", "waiting for operation %q", name)
		}
	}

	if op.Error != nil && len(op.Error.Errors) > 0 {
		msgs := make([]string, 0, len(op.Error.Errors))
		for _, e := range op.Error.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Message))
		}

		return fail.Cloud(errors.New(strings.Join(msgs, "; ")), "gce", "operation %q (%s %s) failed", name, op.OperationType, op.TargetLink)
	}

	return nil
}

func controlPlaneTag(clusterName string) string {
	return fmt.Sprintf("%s-control-plane", clusterName)
}

func isNotFound(err error) bool {
	var gerr *googleapi.Error

	return errors.As(err, &gerr) && gerr.Code == http.StatusNotFound
}

func gceLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "control-plane",
	}
}

func generateGCEControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		nodeLabels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
		}
		maps.Copy(nodeLabels, gceLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, nodeLabels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			var gceConfig gcetypes.RawConfig
			if err = jsonutil.StrictUnmarshal(node.CloudProviderSpec, &gceConfig); err != nil {
				return nil, fail.Config(err, "decode gce config")
			}

			if gceConfig.Labels == nil {
				gceConfig.Labels = map[string]string{}
			}
			maps.Copy(gceConfig.Labels, nodeLabels)

			// Links the VM to the :6443 firewall rule created by EnsureLoadBalancer.
			gceConfig.Tags = append(gceConfig.Tags, controlPlaneTag(clusterName))

			gceSpec, err := json.Marshal(gceConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderGoogle,
				CloudProviderSpec: runtime.RawExtension{
					Raw: gceSpec,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "gce", "json marshaling provider config")
			}

			name := node.MachineName(clusterName, idx)
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
