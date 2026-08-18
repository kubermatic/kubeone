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

package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/cloudprovider"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/provisioner"
	"k8c.io/kubeone/pkg/state"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	awstypes "k8c.io/machine-controller/sdk/cloudprovider/aws"
	"k8c.io/machine-controller/sdk/jsonutil"
	"k8c.io/machine-controller/sdk/providerconfig"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const awsAPIServerPort = 6443

var (
	_ cloudprovider.ControlPlaneCloudProvider = &Provider{}
	_ cloudprovider.LoadBalancerProvider      = &Provider{}
)

func init() {
	cloudprovider.Register(&Provider{})
}

type Provider struct{}

func (p *Provider) Name() string { return "AWS" }

func (p *Provider) Enabled(s *state.State) bool {
	return s.Cluster.CloudProvider.AWS != nil && len(s.Cluster.ControlPlane.NodeSets) > 0
}

func (p *Provider) MatchesConfig(cluster *kubeoneapi.KubeOneCluster) bool {
	return cluster.CloudProvider.AWS != nil
}

func (p *Provider) HasLoadBalancer(s *state.State) bool {
	return p.Enabled(s)
}

func (p *Provider) GenerateMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	return generateAWSControlPlaneMachines(clusterName, nodeSet, kubeletVersion)
}

func (p *Provider) EnsureVM(s *state.State, capimachine clusterv1alpha1.Machine) error {
	provMachines, err := provisioner.FindOrCreateMachines(s.Context, []clusterv1alpha1.Machine{capimachine}, s.Logger)
	if err != nil {
		return err
	}

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	client, err := awsELBv2Client(s)
	if err != nil {
		return err
	}

	tgArn, err := findAWSTargetGroupArn(s, client)
	if err != nil {
		return err
	}

	for _, m := range provMachines {
		if m.PrivateAddress == "" {
			continue
		}

		_, err := client.RegisterTargets(s.Context, &elasticloadbalancingv2.RegisterTargetsInput{
			TargetGroupArn: aws.String(tgArn),
			Targets: []elbv2types.TargetDescription{
				{
					Id:   aws.String(m.PrivateAddress),
					Port: aws.Int32(awsAPIServerPort),
				},
			},
		})
		if err != nil {
			return fail.Cloud(err, "aws", "registering target group member")
		}
	}

	return nil
}

func (p *Provider) LookupVMs(s *state.State) error {
	capimachines, err := p.GenerateMachines(
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

	s.Cluster.ControlPlane.Hosts = append(s.Cluster.ControlPlane.Hosts, cloudprovider.HostConfigsFromMachines(provMachines, s.Cluster.ControlPlane.NodeSets)...)

	return nil
}

func (p *Provider) EnsureLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, err := awsELBv2Client(s)
	if err != nil {
		return err
	}

	lbName := s.Cluster.CloudProvider.AWS.ControlPlane.LoadBalancer.Name
	ctx := s.Context

	describeOut, describeErr := client.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{lbName},
	})

	var notFound *elbv2types.LoadBalancerNotFoundException

	var realLB *elbv2types.LoadBalancer

	switch {
	case describeErr == nil && len(describeOut.LoadBalancers) > 0:
		s.Logger.Debugf("loadbalancer %q already exists", lbName)
		realLB = &describeOut.LoadBalancers[0]
	case describeErr == nil || errors.As(describeErr, &notFound):
		s.Logger.Debugf("no existing loadbalancer found, creating a new one")

		realLB, err = createAWSLoadBalancer(ctx, client, s.Cluster)
		if err != nil {
			return fail.Cloud(err, "aws", "creating loadbalancer")
		}
	default:
		return fail.Cloud(describeErr, "aws", "describing loadbalancers")
	}

	s.Cluster.APIEndpoint.Host = *realLB.DNSName
	s.Cluster.APIEndpoint.Port = awsAPIServerPort

	return nil
}

func (p *Provider) LookupLoadBalancer(s *state.State) error {
	if s.Cluster.APIEndpoint.Host != "" {
		return nil
	}

	client, err := awsELBv2Client(s)
	if err != nil {
		return err
	}

	lbName := s.Cluster.CloudProvider.AWS.ControlPlane.LoadBalancer.Name

	describeOut, err := client.DescribeLoadBalancers(s.Context, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{lbName},
	})
	if err != nil {
		return fail.Cloud(err, "aws", "describing loadbalancers")
	}

	if len(describeOut.LoadBalancers) == 0 {
		return fail.Cloud(fmt.Errorf("no load balancer found with name: %s", lbName), "aws", "looking up loadbalancer")
	}

	realLB := describeOut.LoadBalancers[0]
	s.Logger.Debugf("found loadbalancer %q with arn: %s", lbName, *realLB.LoadBalancerArn)
	s.Cluster.APIEndpoint.Host = *realLB.DNSName
	s.Cluster.APIEndpoint.Port = awsAPIServerPort

	return nil
}

func awsELBv2Client(s *state.State) (*elasticloadbalancingv2.Client, error) {
	providerCreds, err := credentials.ProviderCredentials(s.Cluster.CloudProvider, s.CredentialsFilePath, credentials.TypeUniversal)
	if err != nil {
		return nil, err
	}

	region := s.Cluster.CloudProvider.AWS.Region

	cfg, err := awsconfig.LoadDefaultConfig(
		s.Context,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(awscreds.NewStaticCredentialsProvider(
			providerCreds[credentials.AWSAccessKeyID],
			providerCreds[credentials.AWSSecretAccessKey],
			"",
		)),
	)
	if err != nil {
		return nil, fail.Cloud(err, "aws", "loading AWS SDK config")
	}

	return elasticloadbalancingv2.NewFromConfig(cfg), nil
}

func findAWSTargetGroupArn(s *state.State, client *elasticloadbalancingv2.Client) (string, error) {
	tgName := awsTargetGroupName(s.Cluster.CloudProvider.AWS.ControlPlane.LoadBalancer.Name)

	out, err := client.DescribeTargetGroups(s.Context, &elasticloadbalancingv2.DescribeTargetGroupsInput{
		Names: []string{tgName},
	})
	if err != nil {
		return "", fail.Cloud(err, "aws", "describing target groups")
	}

	if len(out.TargetGroups) == 0 {
		return "", fail.Cloud(fmt.Errorf("no target group found with name: %s", tgName), "aws", "looking up target group")
	}

	return *out.TargetGroups[0].TargetGroupArn, nil
}

func createAWSLoadBalancer(ctx context.Context, client *elasticloadbalancingv2.Client, cluster *kubeoneapi.KubeOneCluster) (*elbv2types.LoadBalancer, error) {
	lbSpec := cluster.CloudProvider.AWS.ControlPlane.LoadBalancer

	vpcID, subnetIDs, err := awsNetworkFromNodeSets(cluster.ControlPlane.NodeSets)
	if err != nil {
		return nil, err
	}

	sdkTags := awsSDKTags(awsTags(cluster.Name, lbSpec.Tags))

	tgOut, err := client.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:                aws.String(awsTargetGroupName(lbSpec.Name)),
		Protocol:            elbv2types.ProtocolEnumTcp,
		Port:                aws.Int32(awsAPIServerPort),
		VpcId:               aws.String(vpcID),
		TargetType:          elbv2types.TargetTypeEnumIp,
		HealthCheckProtocol: elbv2types.ProtocolEnumTcp,
		HealthCheckPort:     aws.String(strconv.Itoa(awsAPIServerPort)),
		Tags:                sdkTags,
	})
	if err != nil {
		return nil, fmt.Errorf("creating target group: %w", err)
	}

	scheme := elbv2types.LoadBalancerSchemeEnumInternetFacing
	if lbSpec.Internal != nil && *lbSpec.Internal {
		scheme = elbv2types.LoadBalancerSchemeEnumInternal
	}

	lbOut, err := client.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:           aws.String(lbSpec.Name),
		Type:           elbv2types.LoadBalancerTypeEnumNetwork,
		Scheme:         scheme,
		Subnets:        subnetIDs,
		SecurityGroups: lbSpec.SecurityGroupIDs,
		Tags:           sdkTags,
	})
	if err != nil {
		return nil, fmt.Errorf("creating load balancer: %w", err)
	}

	newLB := lbOut.LoadBalancers[0]

	_, err = client.CreateListener(ctx, &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: newLB.LoadBalancerArn,
		Protocol:        elbv2types.ProtocolEnumTcp,
		Port:            aws.Int32(awsAPIServerPort),
		DefaultActions: []elbv2types.Action{
			{
				Type:           elbv2types.ActionTypeEnumForward,
				TargetGroupArn: tgOut.TargetGroups[0].TargetGroupArn,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating listener: %w", err)
	}

	return &newLB, nil
}

func awsLabels(clusterName string) map[string]string {
	return map[string]string{
		"kubeone_cluster_name": clusterName,
		"kubeone_role":         "control-plane",
	}
}

func awsTags(clusterName string, extra map[string]string) map[string]string {
	tags := map[string]string{}
	maps.Copy(tags, extra)
	maps.Copy(tags, awsLabels(clusterName))

	return tags
}

func awsSDKTags(tags map[string]string) []elbv2types.Tag {
	sdkTags := make([]elbv2types.Tag, 0, len(tags))
	for k, v := range tags {
		sdkTags = append(sdkTags, elbv2types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	return sdkTags
}

// awsTargetGroupName derives a target group name from the LB name, truncated to AWS's 32-char limit.
func awsTargetGroupName(lbName string) string {
	name := lbName + "-tg"
	if len(name) > 32 {
		name = name[:32]
	}

	return name
}

func awsRawConfigFromNodeSet(node kubeoneapi.NodeSet) (*awstypes.RawConfig, error) {
	var cfg awstypes.RawConfig
	if err := jsonutil.StrictUnmarshal(node.CloudProviderSpec, &cfg); err != nil {
		return nil, fail.Config(err, "decode aws config")
	}

	return &cfg, nil
}

// awsNetworkFromNodeSets collects the VPC (must be identical across nodeSets) and the unique subnet IDs
// used by the control-plane nodeSets, needed to create the NLB in the right network.
func awsNetworkFromNodeSets(nodeSets []kubeoneapi.NodeSet) (string, []string, error) {
	var vpcID string

	subnetSeen := map[string]bool{}

	var subnetIDs []string

	for _, node := range nodeSets {
		cfg, err := awsRawConfigFromNodeSet(node)
		if err != nil {
			return "", nil, err
		}

		if cfg.VpcID.Value == "" || cfg.SubnetID.Value == "" {
			return "", nil, fail.Config(
				fmt.Errorf("vpcId/subnetId missing from control-plane nodeSet %q cloudProviderSpec", node.Name),
				"reading aws config",
			)
		}

		switch {
		case vpcID == "":
			vpcID = cfg.VpcID.Value
		case vpcID != cfg.VpcID.Value:
			return "", nil, fail.Config(errors.New("control-plane nodeSets must use the same vpcId"), "reading aws config")
		}

		if !subnetSeen[cfg.SubnetID.Value] {
			subnetSeen[cfg.SubnetID.Value] = true

			subnetIDs = append(subnetIDs, cfg.SubnetID.Value)
		}
	}

	return vpcID, subnetIDs, nil
}

func generateAWSControlPlaneMachines(clusterName string, nodeSet []kubeoneapi.NodeSet, kubeletVersion string) ([]clusterv1alpha1.Machine, error) {
	var machines []clusterv1alpha1.Machine

	for _, node := range nodeSet {
		timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		nodeLabels := map[string]string{
			"kubeone_own_since_timestamp": timestamp,
		}
		maps.Copy(nodeLabels, awsLabels(clusterName))

		if node.NodeSettings.Labels == nil {
			node.NodeSettings.Labels = map[string]string{}
		}
		maps.Copy(node.NodeSettings.Labels, nodeLabels)

		for idx := range node.Replicas {
			osSpecRaw, err := json.Marshal(node.OperatingSystemSpec)
			if err != nil {
				return nil, err
			}

			awsConfig, err := awsRawConfigFromNodeSet(node)
			if err != nil {
				return nil, err
			}

			if awsConfig.Tags == nil {
				awsConfig.Tags = map[string]string{}
			}

			maps.Copy(awsConfig.Tags, nodeLabels)

			awsSpecRaw, err := json.Marshal(awsConfig)
			if err != nil {
				return nil, fail.Config(err, "marshaling cloud provider spec")
			}

			providerConfig := providerconfig.Config{
				SSHPublicKeys: node.SSH.PublicKeys,
				CloudProvider: providerconfig.CloudProviderAWS,
				CloudProviderSpec: runtime.RawExtension{
					Raw: awsSpecRaw,
				},
				OperatingSystem: providerconfig.OperatingSystem(node.OperatingSystem),
				OperatingSystemSpec: runtime.RawExtension{
					Raw: osSpecRaw,
				},
			}

			providerSpecRaw, err := json.Marshal(providerConfig)
			if err != nil {
				return nil, fail.Cloud(err, "aws", "json marshaling provider config")
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
