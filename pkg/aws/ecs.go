package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ECSService provides ECS operations
type ECSService struct {
	clientManager *ClientManager
}

// NewECSService creates a new ECS service
func NewECSService(clientManager *ClientManager) *ECSService {
	return &ECSService{
		clientManager: clientManager,
	}
}

// Cluster represents an ECS cluster
type Cluster struct {
	ARN                          string
	Name                         string
	Status                       string
	RegisteredContainerInstances int32
	RunningTasksCount            int32
	PendingTasksCount            int32
	ActiveServicesCount          int32
}

// Service represents an ECS service
type Service struct {
	ARN            string
	Name           string
	Status         string
	DesiredCount   int32
	RunningCount   int32
	PendingCount   int32
	LaunchType     string
	TaskDefinition string
	ClusterARN     string
}

// Task represents an ECS task
type Task struct {
	ARN               string
	ClusterARN        string
	TaskDefinitionARN string
	LastStatus        string
	DesiredStatus     string
	LaunchType        string
	CPU               string
	Memory            string
	Containers        []Container
}

// Container represents a container in an ECS task
type Container struct {
	Name       string
	ARN        string
	LastStatus string
	RuntimeID  string
	ExitCode   *int32
}

// ListClusters lists all ECS clusters
func (e *ECSService) ListClusters(ctx context.Context, profileID string) ([]string, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	return result.ClusterArns, nil
}

// DescribeCluster gets detailed information about a cluster
func (e *ECSService) DescribeCluster(ctx context.Context, profileID string, clusterName string) (*Cluster, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: []string{clusterName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	if len(result.Clusters) == 0 {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	c := result.Clusters[0]
	cluster := &Cluster{
		ARN:                          aws.ToString(c.ClusterArn),
		Name:                         aws.ToString(c.ClusterName),
		Status:                       aws.ToString(c.Status),
		RegisteredContainerInstances: c.RegisteredContainerInstancesCount,
		RunningTasksCount:            c.RunningTasksCount,
		PendingTasksCount:            c.PendingTasksCount,
		ActiveServicesCount:          c.ActiveServicesCount,
	}

	return cluster, nil
}

// ListServices lists services in a cluster
func (e *ECSService) ListServices(ctx context.Context, profileID string, clusterName string) ([]string, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.ListServices(ctx, &ecs.ListServicesInput{
		Cluster: aws.String(clusterName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	return result.ServiceArns, nil
}

// DescribeService gets detailed information about a service
func (e *ECSService) DescribeService(ctx context.Context, profileID string, clusterName string, serviceName string) (*Service, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterName),
		Services: []string{serviceName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe service: %w", err)
	}

	if len(result.Services) == 0 {
		return nil, fmt.Errorf("service %s not found in cluster %s", serviceName, clusterName)
	}

	s := result.Services[0]
	service := &Service{
		ARN:            aws.ToString(s.ServiceArn),
		Name:           aws.ToString(s.ServiceName),
		Status:         aws.ToString(s.Status),
		DesiredCount:   s.DesiredCount,
		RunningCount:   s.RunningCount,
		PendingCount:   s.PendingCount,
		LaunchType:     string(s.LaunchType),
		TaskDefinition: aws.ToString(s.TaskDefinition),
		ClusterARN:     aws.ToString(s.ClusterArn),
	}

	return service, nil
}

// ListTasks lists tasks in a cluster, optionally filtered by service
func (e *ECSService) ListTasks(ctx context.Context, profileID string, clusterName string, serviceName string) ([]string, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &ecs.ListTasksInput{
		Cluster: aws.String(clusterName),
	}
	if serviceName != "" {
		input.ServiceName = aws.String(serviceName)
	}

	result, err := client.ListTasks(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return result.TaskArns, nil
}

// DescribeTask gets detailed information about a task
func (e *ECSService) DescribeTask(ctx context.Context, profileID string, clusterName string, taskARN string) (*Task, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   []string{taskARN},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe task: %w", err)
	}

	if len(result.Tasks) == 0 {
		return nil, fmt.Errorf("task %s not found in cluster %s", taskARN, clusterName)
	}

	t := result.Tasks[0]
	task := &Task{
		ARN:               aws.ToString(t.TaskArn),
		ClusterARN:        aws.ToString(t.ClusterArn),
		TaskDefinitionARN: aws.ToString(t.TaskDefinitionArn),
		LastStatus:        aws.ToString(t.LastStatus),
		DesiredStatus:     aws.ToString(t.DesiredStatus),
		LaunchType:        string(t.LaunchType),
		CPU:               aws.ToString(t.Cpu),
		Memory:            aws.ToString(t.Memory),
		Containers:        make([]Container, 0, len(t.Containers)),
	}

	for _, c := range t.Containers {
		container := Container{
			Name:       aws.ToString(c.Name),
			ARN:        aws.ToString(c.ContainerArn),
			LastStatus: aws.ToString(c.LastStatus),
			RuntimeID:  aws.ToString(c.RuntimeId),
			ExitCode:   c.ExitCode,
		}
		task.Containers = append(task.Containers, container)
	}

	return task, nil
}

// DescribeTaskDefinition gets information about a task definition
func (e *ECSService) DescribeTaskDefinition(ctx context.Context, profileID string, taskDefinitionARN string) (map[string]interface{}, error) {
	client, err := e.clientManager.GetECSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe task definition: %w", err)
	}

	td := result.TaskDefinition
	taskDef := map[string]interface{}{
		"family":                  aws.ToString(td.Family),
		"taskDefinitionArn":       aws.ToString(td.TaskDefinitionArn),
		"revision":                td.Revision,
		"status":                  string(td.Status),
		"requiresCompatibilities": td.RequiresCompatibilities,
		"networkMode":             string(td.NetworkMode),
		"cpu":                     aws.ToString(td.Cpu),
		"memory":                  aws.ToString(td.Memory),
	}

	// Add container definitions
	containers := make([]map[string]interface{}, 0, len(td.ContainerDefinitions))
	for _, c := range td.ContainerDefinitions {
		container := map[string]interface{}{
			"name":   aws.ToString(c.Name),
			"image":  aws.ToString(c.Image),
			"cpu":    c.Cpu,
			"memory": c.Memory,
		}
		if c.MemoryReservation != nil {
			container["memoryReservation"] = *c.MemoryReservation
		}
		if c.Essential != nil {
			container["essential"] = *c.Essential
		}
		containers = append(containers, container)
	}
	taskDef["containerDefinitions"] = containers

	return taskDef, nil
}
