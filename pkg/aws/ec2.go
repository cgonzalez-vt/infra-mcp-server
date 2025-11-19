package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Service provides EC2 operations
type EC2Service struct {
	clientManager *ClientManager
}

// NewEC2Service creates a new EC2 service
func NewEC2Service(clientManager *ClientManager) *EC2Service {
	return &EC2Service{
		clientManager: clientManager,
	}
}

// Instance represents an EC2 instance
type Instance struct {
	InstanceID       string
	InstanceType     string
	State            string
	PrivateIP        string
	PublicIP         string
	LaunchTime       string
	AvailabilityZone string
	VpcID            string
	SubnetID         string
	SecurityGroups   []string
	Tags             map[string]string
}

// ListInstances lists all EC2 instances
func (e *EC2Service) ListInstances(ctx context.Context, profileID string) ([]Instance, error) {
	client, err := e.clientManager.GetEC2Client(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	instances := make([]Instance, 0)
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {
			instance := Instance{
				InstanceID:       aws.ToString(inst.InstanceId),
				InstanceType:     string(inst.InstanceType),
				State:            string(inst.State.Name),
				PrivateIP:        aws.ToString(inst.PrivateIpAddress),
				PublicIP:         aws.ToString(inst.PublicIpAddress),
				AvailabilityZone: aws.ToString(inst.Placement.AvailabilityZone),
				VpcID:            aws.ToString(inst.VpcId),
				SubnetID:         aws.ToString(inst.SubnetId),
			}

			if inst.LaunchTime != nil {
				instance.LaunchTime = inst.LaunchTime.String()
			}

			// Add security groups
			securityGroups := make([]string, 0, len(inst.SecurityGroups))
			for _, sg := range inst.SecurityGroups {
				securityGroups = append(securityGroups, aws.ToString(sg.GroupId))
			}
			instance.SecurityGroups = securityGroups

			// Add tags
			tags := make(map[string]string)
			for _, tag := range inst.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
			instance.Tags = tags

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// DescribeInstance gets detailed information about a specific instance
func (e *EC2Service) DescribeInstance(ctx context.Context, profileID string, instanceID string) (*Instance, error) {
	client, err := e.clientManager.GetEC2Client(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance %s not found", instanceID)
	}

	inst := result.Reservations[0].Instances[0]
	instance := &Instance{
		InstanceID:       aws.ToString(inst.InstanceId),
		InstanceType:     string(inst.InstanceType),
		State:            string(inst.State.Name),
		PrivateIP:        aws.ToString(inst.PrivateIpAddress),
		PublicIP:         aws.ToString(inst.PublicIpAddress),
		AvailabilityZone: aws.ToString(inst.Placement.AvailabilityZone),
		VpcID:            aws.ToString(inst.VpcId),
		SubnetID:         aws.ToString(inst.SubnetId),
	}

	if inst.LaunchTime != nil {
		instance.LaunchTime = inst.LaunchTime.String()
	}

	securityGroups := make([]string, 0, len(inst.SecurityGroups))
	for _, sg := range inst.SecurityGroups {
		securityGroups = append(securityGroups, aws.ToString(sg.GroupId))
	}
	instance.SecurityGroups = securityGroups

	tags := make(map[string]string)
	for _, tag := range inst.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	instance.Tags = tags

	return instance, nil
}

// ListVPCs lists all VPCs
func (e *EC2Service) ListVPCs(ctx context.Context, profileID string) ([]map[string]interface{}, error) {
	client, err := e.clientManager.GetEC2Client(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list VPCs: %w", err)
	}

	vpcs := make([]map[string]interface{}, 0, len(result.Vpcs))
	for _, vpc := range result.Vpcs {
		vpcInfo := map[string]interface{}{
			"vpcId":     aws.ToString(vpc.VpcId),
			"cidrBlock": aws.ToString(vpc.CidrBlock),
			"state":     string(vpc.State),
			"isDefault": aws.ToBool(vpc.IsDefault),
		}

		tags := make(map[string]string)
		for _, tag := range vpc.Tags {
			tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
		vpcInfo["tags"] = tags

		vpcs = append(vpcs, vpcInfo)
	}

	return vpcs, nil
}

// ListSecurityGroups lists security groups
func (e *EC2Service) ListSecurityGroups(ctx context.Context, profileID string, vpcID string) ([]map[string]interface{}, error) {
	client, err := e.clientManager.GetEC2Client(profileID)
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeSecurityGroupsInput{}
	if vpcID != "" {
		input.Filters = []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		}
	}

	result, err := client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list security groups: %w", err)
	}

	securityGroups := make([]map[string]interface{}, 0, len(result.SecurityGroups))
	for _, sg := range result.SecurityGroups {
		sgInfo := map[string]interface{}{
			"groupId":     aws.ToString(sg.GroupId),
			"groupName":   aws.ToString(sg.GroupName),
			"description": aws.ToString(sg.Description),
			"vpcId":       aws.ToString(sg.VpcId),
		}

		tags := make(map[string]string)
		for _, tag := range sg.Tags {
			tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
		sgInfo["tags"] = tags

		securityGroups = append(securityGroups, sgInfo)
	}

	return securityGroups, nil
}

