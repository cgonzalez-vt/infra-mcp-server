package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// RDSService provides RDS operations
type RDSService struct {
	clientManager *ClientManager
}

// NewRDSService creates a new RDS service
func NewRDSService(clientManager *ClientManager) *RDSService {
	return &RDSService{
		clientManager: clientManager,
	}
}

// DBInstance represents an RDS database instance
type DBInstance struct {
	Identifier         string
	ARN                string
	Engine             string
	EngineVersion      string
	Status             string
	Endpoint           string
	Port               int32
	InstanceClass      string
	AllocatedStorage   int32
	StorageType        string
	AvailabilityZone   string
	MultiAZ            bool
	PubliclyAccessible bool
	MasterUsername     string
	DBName             string
	VPCSecurityGroups  []string
	DBSubnetGroup      string
}

// DBSnapshot represents an RDS database snapshot
type DBSnapshot struct {
	Identifier       string
	ARN              string
	DBInstanceID     string
	SnapshotType     string
	Status           string
	Engine           string
	AllocatedStorage int32
	SnapshotTime     string
	Port             int32
	AvailabilityZone string
}

// ListDBInstances lists all RDS database instances
func (r *RDSService) ListDBInstances(ctx context.Context, profileID string) ([]DBInstance, error) {
	client, err := r.clientManager.GetRDSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list DB instances: %w", err)
	}

	instances := make([]DBInstance, 0, len(result.DBInstances))
	for _, db := range result.DBInstances {
		instance := DBInstance{
			Identifier:         aws.ToString(db.DBInstanceIdentifier),
			ARN:                aws.ToString(db.DBInstanceArn),
			Engine:             aws.ToString(db.Engine),
			EngineVersion:      aws.ToString(db.EngineVersion),
			Status:             aws.ToString(db.DBInstanceStatus),
			InstanceClass:      aws.ToString(db.DBInstanceClass),
			AllocatedStorage:   aws.ToInt32(db.AllocatedStorage),
			StorageType:        aws.ToString(db.StorageType),
			AvailabilityZone:   aws.ToString(db.AvailabilityZone),
			MultiAZ:            aws.ToBool(db.MultiAZ),
			PubliclyAccessible: aws.ToBool(db.PubliclyAccessible),
			MasterUsername:     aws.ToString(db.MasterUsername),
			DBName:             aws.ToString(db.DBName),
		}

		// Add endpoint information
		if db.Endpoint != nil {
			instance.Endpoint = aws.ToString(db.Endpoint.Address)
			instance.Port = aws.ToInt32(db.Endpoint.Port)
		}

		// Add VPC security groups
		securityGroups := make([]string, 0, len(db.VpcSecurityGroups))
		for _, sg := range db.VpcSecurityGroups {
			securityGroups = append(securityGroups, aws.ToString(sg.VpcSecurityGroupId))
		}
		instance.VPCSecurityGroups = securityGroups

		// Add DB subnet group
		if db.DBSubnetGroup != nil {
			instance.DBSubnetGroup = aws.ToString(db.DBSubnetGroup.DBSubnetGroupName)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// DescribeDBInstance gets detailed information about a specific DB instance
func (r *RDSService) DescribeDBInstance(ctx context.Context, profileID string, identifier string) (*DBInstance, error) {
	client, err := r.clientManager.GetRDSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe DB instance: %w", err)
	}

	if len(result.DBInstances) == 0 {
		return nil, fmt.Errorf("DB instance %s not found", identifier)
	}

	db := result.DBInstances[0]
	instance := &DBInstance{
		Identifier:         aws.ToString(db.DBInstanceIdentifier),
		ARN:                aws.ToString(db.DBInstanceArn),
		Engine:             aws.ToString(db.Engine),
		EngineVersion:      aws.ToString(db.EngineVersion),
		Status:             aws.ToString(db.DBInstanceStatus),
		InstanceClass:      aws.ToString(db.DBInstanceClass),
		AllocatedStorage:   aws.ToInt32(db.AllocatedStorage),
		StorageType:        aws.ToString(db.StorageType),
		AvailabilityZone:   aws.ToString(db.AvailabilityZone),
		MultiAZ:            aws.ToBool(db.MultiAZ),
		PubliclyAccessible: aws.ToBool(db.PubliclyAccessible),
		MasterUsername:     aws.ToString(db.MasterUsername),
		DBName:             aws.ToString(db.DBName),
	}

	if db.Endpoint != nil {
		instance.Endpoint = aws.ToString(db.Endpoint.Address)
		instance.Port = aws.ToInt32(db.Endpoint.Port)
	}

	securityGroups := make([]string, 0, len(db.VpcSecurityGroups))
	for _, sg := range db.VpcSecurityGroups {
		securityGroups = append(securityGroups, aws.ToString(sg.VpcSecurityGroupId))
	}
	instance.VPCSecurityGroups = securityGroups

	if db.DBSubnetGroup != nil {
		instance.DBSubnetGroup = aws.ToString(db.DBSubnetGroup.DBSubnetGroupName)
	}

	return instance, nil
}

// GetDBConnectionInfo returns connection information for a DB instance
func (r *RDSService) GetDBConnectionInfo(ctx context.Context, profileID string, identifier string) (map[string]interface{}, error) {
	instance, err := r.DescribeDBInstance(ctx, profileID, identifier)
	if err != nil {
		return nil, err
	}

	connectionInfo := map[string]interface{}{
		"identifier": instance.Identifier,
		"endpoint":   instance.Endpoint,
		"port":       instance.Port,
		"engine":     instance.Engine,
		"dbName":     instance.DBName,
		"username":   instance.MasterUsername,
		"status":     instance.Status,
	}

	return connectionInfo, nil
}

// ListDBSnapshots lists snapshots for a DB instance
func (r *RDSService) ListDBSnapshots(ctx context.Context, profileID string, identifier string) ([]DBSnapshot, error) {
	client, err := r.clientManager.GetRDSClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &rds.DescribeDBSnapshotsInput{}
	if identifier != "" {
		input.DBInstanceIdentifier = aws.String(identifier)
	}

	result, err := client.DescribeDBSnapshots(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list DB snapshots: %w", err)
	}

	snapshots := make([]DBSnapshot, 0, len(result.DBSnapshots))
	for _, snap := range result.DBSnapshots {
		snapshot := DBSnapshot{
			Identifier:       aws.ToString(snap.DBSnapshotIdentifier),
			ARN:              aws.ToString(snap.DBSnapshotArn),
			DBInstanceID:     aws.ToString(snap.DBInstanceIdentifier),
			SnapshotType:     aws.ToString(snap.SnapshotType),
			Status:           aws.ToString(snap.Status),
			Engine:           aws.ToString(snap.Engine),
			AllocatedStorage: aws.ToInt32(snap.AllocatedStorage),
			Port:             aws.ToInt32(snap.Port),
			AvailabilityZone: aws.ToString(snap.AvailabilityZone),
		}
		if snap.SnapshotCreateTime != nil {
			snapshot.SnapshotTime = snap.SnapshotCreateTime.String()
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// ListDBClusters lists all RDS clusters (Aurora)
func (r *RDSService) ListDBClusters(ctx context.Context, profileID string) ([]map[string]interface{}, error) {
	client, err := r.clientManager.GetRDSClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list DB clusters: %w", err)
	}

	clusters := make([]map[string]interface{}, 0, len(result.DBClusters))
	for _, cluster := range result.DBClusters {
		clusterInfo := map[string]interface{}{
			"identifier":    aws.ToString(cluster.DBClusterIdentifier),
			"arn":           aws.ToString(cluster.DBClusterArn),
			"engine":        aws.ToString(cluster.Engine),
			"engineVersion": aws.ToString(cluster.EngineVersion),
			"status":        aws.ToString(cluster.Status),
			"multiAZ":       aws.ToBool(cluster.MultiAZ),
		}

		if cluster.Endpoint != nil {
			clusterInfo["endpoint"] = *cluster.Endpoint
		}
		if cluster.ReaderEndpoint != nil {
			clusterInfo["readerEndpoint"] = *cluster.ReaderEndpoint
		}
		if cluster.Port != nil {
			clusterInfo["port"] = *cluster.Port
		}

		// Add cluster members
		members := make([]string, 0, len(cluster.DBClusterMembers))
		for _, member := range cluster.DBClusterMembers {
			members = append(members, aws.ToString(member.DBInstanceIdentifier))
		}
		clusterInfo["members"] = members

		clusters = append(clusters, clusterInfo)
	}

	return clusters, nil
}

