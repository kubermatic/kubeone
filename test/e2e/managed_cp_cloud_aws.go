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

package e2e

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

const (
	// tags applied by KubeOne to the control plane instances, load balancer
	// and target group, see pkg/cloudprovider/aws
	awsClusterNameTag = "kubeone_cluster_name"
	awsRoleTag        = "kubeone_role"
	awsControlPlane   = "control-plane"

	// elbv2 DescribeTags accepts at most 20 resource ARNs
	awsDescribeTagsBatchSize = 20

	awsJanitorWaitTimeout = 10 * time.Minute
)

// awsClusterTag is the tag key applied to all the instances (control plane and
// workers) and load balancers of the cluster, including the ones created by
// the AWS CCM.
func awsClusterTag(clusterName string) string {
	return "kubernetes.io/cluster/" + clusterName
}

type awsCloud struct {
	ec2         *ec2.Client
	elbv2       *elasticloadbalancingv2.Client
	clusterName string
}

func newAWSCloud(ctx context.Context, clusterName string, tfOutput map[string]any) (*awsCloud, error) {
	networking, _ := tfOutput["networking"].(map[string]any)

	region, _ := networking["region"].(string)
	if region == "" {
		return nil, errors.New("region is missing from the networking terraform output")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	return &awsCloud{
		ec2:         ec2.NewFromConfig(cfg),
		elbv2:       elasticloadbalancingv2.NewFromConfig(cfg),
		clusterName: clusterName,
	}, nil
}

func (a *awsCloud) ControlPlaneVMs(ctx context.Context) (int, error) {
	instances, err := a.instances(ctx,
		[]ec2types.InstanceStateName{ec2types.InstanceStateNamePending, ec2types.InstanceStateNameRunning},
		ec2types.Filter{Name: aws.String("tag:" + awsClusterNameTag), Values: []string{a.clusterName}},
		ec2types.Filter{Name: aws.String("tag:" + awsRoleTag), Values: []string{awsControlPlane}},
	)

	return len(instances), err
}

func (a *awsCloud) LoadBalancers(ctx context.Context) ([]string, error) {
	// the CCM doesn't set awsClusterNameTag, only KubeOne does
	lbs, err := a.loadBalancers(ctx, func(tags map[string]string) bool {
		return tags[awsClusterNameTag] == a.clusterName
	})
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, lb := range lbs {
		addresses = append(addresses, aws.ToString(lb.DNSName))
	}

	return addresses, nil
}

func (a *awsCloud) Janitor(ctx context.Context) error {
	var errs []error

	if err := a.terminateInstances(ctx); err != nil {
		errs = append(errs, err)
	}

	// load balancers created by KubeOne and by the CCM
	clusterOwned := func(tags map[string]string) bool {
		_, ok := tags[awsClusterTag(a.clusterName)]

		return ok
	}

	lbs, err := a.loadBalancers(ctx, clusterOwned)
	if err != nil {
		errs = append(errs, err)
	}

	var deletedLBs []string
	for _, lb := range lbs {
		if _, err = a.elbv2.DeleteLoadBalancer(ctx, &elasticloadbalancingv2.DeleteLoadBalancerInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		}); err != nil {
			errs = append(errs, fmt.Errorf("deleting load balancer %q: %w", aws.ToString(lb.LoadBalancerName), err))

			continue
		}
		deletedLBs = append(deletedLBs, aws.ToString(lb.LoadBalancerArn))
	}

	// target groups can be deleted only once they're not used by any listener
	if len(deletedLBs) > 0 {
		waiter := elasticloadbalancingv2.NewLoadBalancersDeletedWaiter(a.elbv2)
		if err = waiter.Wait(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			LoadBalancerArns: deletedLBs,
		}, awsJanitorWaitTimeout); err != nil {
			errs = append(errs, fmt.Errorf("waiting for load balancers to be deleted: %w", err))
		}
	}

	tgs, err := a.targetGroups(ctx, clusterOwned)
	if err != nil {
		errs = append(errs, err)
	}

	for _, tg := range tgs {
		if _, err = a.elbv2.DeleteTargetGroup(ctx, &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tg.TargetGroupArn,
		}); err != nil {
			errs = append(errs, fmt.Errorf("deleting target group %q: %w", aws.ToString(tg.TargetGroupName), err))
		}
	}

	return errors.Join(errs...)
}

// terminateInstances terminates all the instances of the cluster and waits for
// them to be gone, as they'd block terraform destroy of the subnet and the
// security groups otherwise.
func (a *awsCloud) terminateInstances(ctx context.Context) error {
	instances, err := a.instances(ctx,
		[]ec2types.InstanceStateName{
			ec2types.InstanceStateNamePending,
			ec2types.InstanceStateNameRunning,
			ec2types.InstanceStateNameStopping,
			ec2types.InstanceStateNameStopped,
			ec2types.InstanceStateNameShuttingDown,
		},
		ec2types.Filter{Name: aws.String("tag-key"), Values: []string{awsClusterTag(a.clusterName)}},
	)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		return nil
	}

	var ids []string
	for _, instance := range instances {
		ids = append(ids, aws.ToString(instance.InstanceId))
	}

	if _, err = a.ec2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: ids}); err != nil {
		return fmt.Errorf("terminating instances %v: %w", ids, err)
	}

	waiter := ec2.NewInstanceTerminatedWaiter(a.ec2)
	if err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{InstanceIds: ids}, awsJanitorWaitTimeout); err != nil {
		return fmt.Errorf("waiting for instances %v to terminate: %w", ids, err)
	}

	return nil
}

func (a *awsCloud) instances(ctx context.Context, states []ec2types.InstanceStateName, filters ...ec2types.Filter) ([]ec2types.Instance, error) {
	stateValues := make([]string, 0, len(states))
	for _, state := range states {
		stateValues = append(stateValues, string(state))
	}

	filters = append(filters, ec2types.Filter{Name: aws.String("instance-state-name"), Values: stateValues})

	var instances []ec2types.Instance

	paginator := ec2.NewDescribeInstancesPaginator(a.ec2, &ec2.DescribeInstancesInput{Filters: filters})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describing instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}

	return instances, nil
}

func (a *awsCloud) loadBalancers(ctx context.Context, match func(tags map[string]string) bool) ([]elbv2types.LoadBalancer, error) {
	var lbs []elbv2types.LoadBalancer

	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(a.elbv2, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describing load balancers: %w", err)
		}
		lbs = append(lbs, page.LoadBalancers...)
	}

	arns := make([]string, 0, len(lbs))
	for _, lb := range lbs {
		arns = append(arns, aws.ToString(lb.LoadBalancerArn))
	}

	tags, err := a.elbv2Tags(ctx, arns)
	if err != nil {
		return nil, err
	}

	var matched []elbv2types.LoadBalancer
	for _, lb := range lbs {
		if match(tags[aws.ToString(lb.LoadBalancerArn)]) {
			matched = append(matched, lb)
		}
	}

	return matched, nil
}

func (a *awsCloud) targetGroups(ctx context.Context, match func(tags map[string]string) bool) ([]elbv2types.TargetGroup, error) {
	var tgs []elbv2types.TargetGroup

	paginator := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(a.elbv2, &elasticloadbalancingv2.DescribeTargetGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describing target groups: %w", err)
		}
		tgs = append(tgs, page.TargetGroups...)
	}

	arns := make([]string, 0, len(tgs))
	for _, tg := range tgs {
		arns = append(arns, aws.ToString(tg.TargetGroupArn))
	}

	tags, err := a.elbv2Tags(ctx, arns)
	if err != nil {
		return nil, err
	}

	var matched []elbv2types.TargetGroup
	for _, tg := range tgs {
		if match(tags[aws.ToString(tg.TargetGroupArn)]) {
			matched = append(matched, tg)
		}
	}

	return matched, nil
}

// elbv2Tags returns the tags of the elbv2 resources, keyed by the resource ARN
func (a *awsCloud) elbv2Tags(ctx context.Context, arns []string) (map[string]map[string]string, error) {
	tags := map[string]map[string]string{}

	for start := 0; start < len(arns); start += awsDescribeTagsBatchSize {
		batch := arns[start:min(start+awsDescribeTagsBatchSize, len(arns))]

		out, err := a.elbv2.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{ResourceArns: batch})
		if err != nil {
			return nil, fmt.Errorf("describing tags: %w", err)
		}

		for _, desc := range out.TagDescriptions {
			resourceTags := map[string]string{}
			for _, tag := range desc.Tags {
				resourceTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
			tags[aws.ToString(desc.ResourceArn)] = resourceTags
		}
	}

	return tags, nil
}
