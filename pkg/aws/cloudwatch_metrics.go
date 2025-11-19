package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// CloudWatchMetricsService provides CloudWatch Metrics operations
type CloudWatchMetricsService struct {
	clientManager *ClientManager
}

// NewCloudWatchMetricsService creates a new CloudWatch Metrics service
func NewCloudWatchMetricsService(clientManager *ClientManager) *CloudWatchMetricsService {
	return &CloudWatchMetricsService{
		clientManager: clientManager,
	}
}

// Metric represents a CloudWatch metric
type Metric struct {
	Namespace  string
	MetricName string
	Dimensions map[string]string
}

// MetricDataPoint represents a metric data point
type MetricDataPoint struct {
	Timestamp time.Time
	Value     float64
	Unit      string
}

// ListMetrics lists available CloudWatch metrics
func (cm *CloudWatchMetricsService) ListMetrics(ctx context.Context, profileID string, namespace string, metricName string) ([]Metric, error) {
	client, err := cm.clientManager.GetCloudWatchClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &cloudwatch.ListMetricsInput{}
	if namespace != "" {
		input.Namespace = aws.String(namespace)
	}
	if metricName != "" {
		input.MetricName = aws.String(metricName)
	}

	result, err := client.ListMetrics(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics: %w", err)
	}

	metrics := make([]Metric, 0, len(result.Metrics))
	for _, m := range result.Metrics {
		metric := Metric{
			Namespace:  aws.ToString(m.Namespace),
			MetricName: aws.ToString(m.MetricName),
			Dimensions: make(map[string]string),
		}

		for _, dim := range m.Dimensions {
			metric.Dimensions[aws.ToString(dim.Name)] = aws.ToString(dim.Value)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetMetricStatistics gets statistics for a metric
func (cm *CloudWatchMetricsService) GetMetricStatistics(ctx context.Context, profileID string, namespace string, metricName string, dimensions map[string]string, startTime time.Time, endTime time.Time, period int32, statistics []string) ([]MetricDataPoint, error) {
	client, err := cm.clientManager.GetCloudWatchClient(profileID)
	if err != nil {
		return nil, err
	}

	// Convert dimensions map to CloudWatch dimensions
	cwDimensions := make([]types.Dimension, 0, len(dimensions))
	for name, value := range dimensions {
		cwDimensions = append(cwDimensions, types.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	// Convert statistics strings to types
	cwStatistics := make([]types.Statistic, 0, len(statistics))
	for _, stat := range statistics {
		cwStatistics = append(cwStatistics, types.Statistic(stat))
	}

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: cwDimensions,
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(period),
		Statistics: cwStatistics,
	}

	result, err := client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric statistics: %w", err)
	}

	dataPoints := make([]MetricDataPoint, 0, len(result.Datapoints))
	for _, dp := range result.Datapoints {
		dataPoint := MetricDataPoint{
			Timestamp: aws.ToTime(dp.Timestamp),
			Unit:      string(dp.Unit),
		}

		// Get the first available statistic value
		if dp.Average != nil {
			dataPoint.Value = *dp.Average
		} else if dp.Sum != nil {
			dataPoint.Value = *dp.Sum
		} else if dp.Maximum != nil {
			dataPoint.Value = *dp.Maximum
		} else if dp.Minimum != nil {
			dataPoint.Value = *dp.Minimum
		} else if dp.SampleCount != nil {
			dataPoint.Value = *dp.SampleCount
		}

		dataPoints = append(dataPoints, dataPoint)
	}

	return dataPoints, nil
}

// GetRDSMetrics gets common RDS metrics for a database instance
func (cm *CloudWatchMetricsService) GetRDSMetrics(ctx context.Context, profileID string, dbInstanceIdentifier string, hoursBack int) (map[string][]MetricDataPoint, error) {
	endTime := time.Now()
	startTime := endTime.Add(time.Duration(-hoursBack) * time.Hour)

	dimensions := map[string]string{
		"DBInstanceIdentifier": dbInstanceIdentifier,
	}

	metrics := map[string][]MetricDataPoint{}

	// Common RDS metrics
	metricNames := []string{
		"CPUUtilization",
		"DatabaseConnections",
		"FreeableMemory",
		"FreeStorageSpace",
		"ReadLatency",
		"WriteLatency",
	}

	for _, metricName := range metricNames {
		dataPoints, err := cm.GetMetricStatistics(ctx, profileID, "AWS/RDS", metricName, dimensions, startTime, endTime, 300, []string{"Average"})
		if err != nil {
			// Continue on error, just skip this metric
			continue
		}
		metrics[metricName] = dataPoints
	}

	return metrics, nil
}

// GetECSMetrics gets common ECS metrics for a cluster/service
func (cm *CloudWatchMetricsService) GetECSMetrics(ctx context.Context, profileID string, clusterName string, serviceName string, hoursBack int) (map[string][]MetricDataPoint, error) {
	endTime := time.Now()
	startTime := endTime.Add(time.Duration(-hoursBack) * time.Hour)

	dimensions := map[string]string{
		"ClusterName": clusterName,
		"ServiceName": serviceName,
	}

	metrics := map[string][]MetricDataPoint{}

	// Common ECS metrics
	metricNames := []string{
		"CPUUtilization",
		"MemoryUtilization",
	}

	for _, metricName := range metricNames {
		dataPoints, err := cm.GetMetricStatistics(ctx, profileID, "AWS/ECS", metricName, dimensions, startTime, endTime, 300, []string{"Average"})
		if err != nil {
			// Continue on error, just skip this metric
			continue
		}
		metrics[metricName] = dataPoints
	}

	return metrics, nil
}

