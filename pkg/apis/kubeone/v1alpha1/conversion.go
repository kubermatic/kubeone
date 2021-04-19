/*
Copyright 2020 The KubeOne Authors.

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

package v1alpha1

import (
	"errors"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
)

// TODO(xmudrii): Add conversion tests.
func Convert_v1alpha1_CNI_To_kubeone_CNI(in *CNI, out *kubeoneapi.CNI, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_CNI_To_kubeone_CNI(in, out, s); err != nil {
		return err
	}

	switch in.Provider {
	case CNIProviderCanal:
		out.Canal = &kubeoneapi.CanalSpec{
			MTU: 1450,
		}
	case CNIProviderWeaveNet:
		out.WeaveNet = &kubeoneapi.WeaveNetSpec{
			Encrypted: in.Encrypted,
		}
	case CNIProviderExternal:
		out.External = &kubeoneapi.ExternalCNISpec{}
	default:
		return errors.New("invalid input cni provider name")
	}

	return nil
}

func Convert_kubeone_CNI_To_v1alpha1_CNI(in *kubeoneapi.CNI, out *CNI, s conversion.Scope) error {
	if err := autoConvert_kubeone_CNI_To_v1alpha1_CNI(in, out, s); err != nil {
		return err
	}

	switch {
	case in.Canal != nil:
		out.Provider = CNIProviderCanal
	case in.WeaveNet != nil:
		out.Provider = CNIProviderWeaveNet
		out.Encrypted = in.WeaveNet.Encrypted
	case in.External != nil:
		out.Provider = CNIProviderExternal
	}

	return nil
}

func Convert_v1alpha1_CloudProviderSpec_To_kubeone_CloudProviderSpec(in *CloudProviderSpec, out *kubeoneapi.CloudProviderSpec, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_CloudProviderSpec_To_kubeone_CloudProviderSpec(in, out, s); err != nil {
		return err
	}

	switch in.Name {
	case CloudProviderNameAWS:
		out.AWS = &kubeoneapi.AWSSpec{}
	case CloudProviderNameAzure:
		out.Azure = &kubeoneapi.AzureSpec{}
	case CloudProviderNameDigitalOcean:
		out.DigitalOcean = &kubeoneapi.DigitalOceanSpec{}
	case CloudProviderNameGCE:
		out.GCE = &kubeoneapi.GCESpec{}
	case CloudProviderNameHetzner:
		// We can also set Hetzner cloud provider in the KubeOneCluster conversion function
		// We don't want to override it if it has been already set
		if out.Hetzner == nil {
			out.Hetzner = &kubeoneapi.HetznerSpec{}
		}
	case CloudProviderNameNone:
		out.None = &kubeoneapi.NoneSpec{}
	case CloudProviderNameOpenStack:
		out.Openstack = &kubeoneapi.OpenstackSpec{}
	case CloudProviderNamePacket:
		out.Packet = &kubeoneapi.PacketSpec{}
	case CloudProviderNameVSphere:
		out.Vsphere = &kubeoneapi.VsphereSpec{}
	default:
		return errors.New("invalid input cloud provider name")
	}

	return nil
}

func Convert_kubeone_CloudProviderSpec_To_v1alpha1_CloudProviderSpec(in *kubeoneapi.CloudProviderSpec, out *CloudProviderSpec, s conversion.Scope) error {
	if err := autoConvert_kubeone_CloudProviderSpec_To_v1alpha1_CloudProviderSpec(in, out, s); err != nil {
		return err
	}

	switch {
	case in.AWS != nil:
		out.Name = CloudProviderNameAWS
	case in.Azure != nil:
		out.Name = CloudProviderNameAzure
	case in.DigitalOcean != nil:
		out.Name = CloudProviderNameDigitalOcean
	case in.GCE != nil:
		out.Name = CloudProviderNameGCE
	case in.Hetzner != nil:
		out.Name = CloudProviderNameHetzner
	case in.Openstack != nil:
		out.Name = CloudProviderNameOpenStack
	case in.Packet != nil:
		out.Name = CloudProviderNamePacket
	case in.Vsphere != nil:
		out.Name = CloudProviderNameVSphere
	case in.None != nil:
		out.Name = CloudProviderNameNone
	}

	return nil
}

func Convert_v1alpha1_ClusterNetworkConfig_To_kubeone_ClusterNetworkConfig(in *ClusterNetworkConfig, out *kubeoneapi.ClusterNetworkConfig, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_ClusterNetworkConfig_To_kubeone_ClusterNetworkConfig(in, out, s); err != nil {
		return err
	}

	return nil
}

func Convert_v1alpha1_HostConfig_To_kubeone_HostConfig(in *HostConfig, out *kubeoneapi.HostConfig, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_HostConfig_To_kubeone_HostConfig(in, out, s); err != nil {
		return err
	}

	if in.Untaint {
		out.Taints = []corev1.Taint{}
	}

	return nil
}

func Convert_kubeone_HostConfig_To_v1alpha1_HostConfig(in *kubeoneapi.HostConfig, out *HostConfig, s conversion.Scope) error {
	if err := autoConvert_kubeone_HostConfig_To_v1alpha1_HostConfig(in, out, s); err != nil {
		return err
	}

	if in.Taints != nil {
		switch {
		// If there is no taint provided, that means the node should be untainted
		case len(in.Taints) == 0:
			out.Untaint = true
		// If there is only one taint provided, we want to check is it a default one
		// If it is default one, do nothing, but if it's a custom one, return error because
		// v1alpha1 API doesn't support custom taints
		case len(in.Taints) == 1:
			if in.Taints[0].Key != "node-role.kubernetes.io/master" || in.Taints[0].Effect != corev1.TaintEffectNoSchedule {
				return errors.New("v1alpha1 doesn't support custom taints")
			}
		// If there are more than one taints provided, return an error because
		// v1alpha1 API doesn't support custom and multiple taints
		default:
			return errors.New("v1alpha1 doesn't support multiple taints")
		}
	}

	return nil
}

func Convert_kubeone_ProviderSpec_To_v1alpha1_ProviderSpec(in *kubeoneapi.ProviderSpec, out *ProviderSpec, s conversion.Scope) error {
	if err := autoConvert_kubeone_ProviderSpec_To_v1alpha1_ProviderSpec(in, out, s); err != nil {
		return err
	}

	// The Annotations field is not available in the v1alpha1 API.

	return nil
}

func Convert_v1alpha1_KubeOneCluster_To_kubeone_KubeOneCluster(in *KubeOneCluster, out *kubeoneapi.KubeOneCluster, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_KubeOneCluster_To_kubeone_KubeOneCluster(in, out, s); err != nil {
		return err
	}

	// Control plane nodes
	for idx := range in.Hosts {
		outHost := kubeoneapi.HostConfig{}
		if err := Convert_v1alpha1_HostConfig_To_kubeone_HostConfig(&in.Hosts[idx], &outHost, s); err != nil {
			return err
		}
		out.ControlPlane.Hosts = append(out.ControlPlane.Hosts, outHost)
	}

	// Static worker nodes
	for idx := range in.StaticWorkers {
		outHost := kubeoneapi.HostConfig{}
		if err := Convert_v1alpha1_HostConfig_To_kubeone_HostConfig(&in.StaticWorkers[idx], &outHost, s); err != nil {
			return err
		}
		out.StaticWorkers.Hosts = append(out.StaticWorkers.Hosts, outHost)
	}

	// Dynamic worker nodes
	for idx, worker := range in.Workers {
		outWorker := kubeoneapi.DynamicWorkerConfig{}
		outWorker.Name = worker.Name
		outWorker.Replicas = worker.Replicas

		outProviderSpec := kubeoneapi.ProviderSpec{}
		if err := Convert_v1alpha1_ProviderSpec_To_kubeone_ProviderSpec(&in.Workers[idx].Config, &outProviderSpec, s); err != nil {
			return err
		}
		outWorker.Config = outProviderSpec

		out.DynamicWorkers = append(out.DynamicWorkers, outWorker)
	}

	// The NetworkID field has been moved from .ClusterNetwork.NetworkID to .CloudProvider.Hetzner.NetworkID
	if len(in.ClusterNetwork.NetworkID) != 0 && in.CloudProvider.Name == CloudProviderNameHetzner {
		if out.CloudProvider.Hetzner == nil {
			out.CloudProvider.Hetzner = &kubeoneapi.HetznerSpec{}
		}
		out.CloudProvider.Hetzner.NetworkID = in.ClusterNetwork.NetworkID
	}

	// Default to docker
	out.ContainerRuntime = kubeoneapi.ContainerRuntimeConfig{
		Docker: &kubeoneapi.ContainerRuntimeDocker{},
	}

	return nil
}

func Convert_kubeone_KubeOneCluster_To_v1alpha1_KubeOneCluster(in *kubeoneapi.KubeOneCluster, out *KubeOneCluster, s conversion.Scope) error {
	if err := autoConvert_kubeone_KubeOneCluster_To_v1alpha1_KubeOneCluster(in, out, s); err != nil {
		return err
	}

	// Control plane nodes
	for idx := range in.ControlPlane.Hosts {
		outHost := HostConfig{}
		if err := Convert_kubeone_HostConfig_To_v1alpha1_HostConfig(&in.ControlPlane.Hosts[idx], &outHost, s); err != nil {
			return err
		}
		out.Hosts = append(out.Hosts, outHost)
	}

	// Static worker nodes
	for idx := range in.StaticWorkers.Hosts {
		outHost := HostConfig{}
		if err := Convert_kubeone_HostConfig_To_v1alpha1_HostConfig(&in.StaticWorkers.Hosts[idx], &outHost, s); err != nil {
			return err
		}
		out.StaticWorkers = append(out.StaticWorkers, outHost)
	}

	// Dynamic worker nodes
	for idx, worker := range in.DynamicWorkers {
		outWorker := WorkerConfig{}
		outWorker.Name = worker.Name
		outWorker.Replicas = worker.Replicas

		outProviderSpec := ProviderSpec{}
		if err := autoConvert_kubeone_ProviderSpec_To_v1alpha1_ProviderSpec(&in.DynamicWorkers[idx].Config, &outProviderSpec, s); err != nil {
			return err
		}
		outWorker.Config = outProviderSpec

		out.Workers = append(out.Workers, outWorker)
	}

	// NetworkID
	if in.CloudProvider.Hetzner != nil {
		out.ClusterNetwork.NetworkID = in.CloudProvider.Hetzner.NetworkID
	}

	return nil
}

func Convert_v1alpha1_MachineControllerConfig_To_kubeone_MachineControllerConfig(in *MachineControllerConfig, out *kubeoneapi.MachineControllerConfig, s conversion.Scope) error {
	if err := autoConvert_v1alpha1_MachineControllerConfig_To_kubeone_MachineControllerConfig(in, out, s); err != nil {
		return err
	}

	// The Provider field has been dropped from v1beta1 API.

	return nil
}

func Convert_kubeone_Features_To_v1alpha1_Features(in *kubeoneapi.Features, out *Features, s conversion.Scope) error {
	return autoConvert_kubeone_Features_To_v1alpha1_Features(in, out, s)
}
