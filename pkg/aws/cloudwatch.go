package aws

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// CloudWatchService provides CloudWatch Logs operations
type CloudWatchService struct {
	clientManager *ClientManager
}

// NewCloudWatchService creates a new CloudWatch service
func NewCloudWatchService(clientManager *ClientManager) *CloudWatchService {
	return &CloudWatchService{
		clientManager: clientManager,
	}
}

// LogGroup represents a CloudWatch log group
type LogGroup struct {
	Name          string
	ARN           string
	CreationTime  int64
	RetentionDays int32
	StoredBytes   int64
	LogGroupClass string
}

// LogStream represents a CloudWatch log stream
type LogStream struct {
	Name              string
	ARN               string
	CreationTime      int64
	FirstEventTime    int64
	LastEventTime     int64
	LastIngestionTime int64
	StoredBytes       int64
}

// LogEvent represents a CloudWatch log event
type LogEvent struct {
	Timestamp     int64
	Message       string
	IngestionTime int64
}

// ListLogGroups lists all CloudWatch log groups
func (cw *CloudWatchService) ListLogGroups(ctx context.Context, profileID string, prefix string, limit int32) ([]LogGroup, error) {
	client, err := cw.clientManager.GetCloudWatchLogsClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	if prefix != "" {
		input.LogGroupNamePrefix = aws.String(prefix)
	}
	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	result, err := client.DescribeLogGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list log groups: %w", err)
	}

	logGroups := make([]LogGroup, 0, len(result.LogGroups))
	for _, lg := range result.LogGroups {
		logGroup := LogGroup{
			Name:         aws.ToString(lg.LogGroupName),
			ARN:          aws.ToString(lg.Arn),
			CreationTime: aws.ToInt64(lg.CreationTime),
			StoredBytes:  aws.ToInt64(lg.StoredBytes),
		}
		if lg.RetentionInDays != nil {
			logGroup.RetentionDays = *lg.RetentionInDays
		}
		if lg.LogGroupClass != "" {
			logGroup.LogGroupClass = string(lg.LogGroupClass)
		}
		logGroups = append(logGroups, logGroup)
	}

	return logGroups, nil
}

// GetLogStreams gets log streams for a log group
func (cw *CloudWatchService) GetLogStreams(ctx context.Context, profileID string, logGroupName string, limit int32) ([]LogStream, error) {
	client, err := cw.clientManager.GetCloudWatchLogsClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
		Descending:   aws.Bool(true),
		OrderBy:      types.OrderByLastEventTime,
	}
	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	result, err := client.DescribeLogStreams(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get log streams: %w", err)
	}

	logStreams := make([]LogStream, 0, len(result.LogStreams))
	for _, ls := range result.LogStreams {
		logStream := LogStream{
			Name:              aws.ToString(ls.LogStreamName),
			ARN:               aws.ToString(ls.Arn),
			CreationTime:      aws.ToInt64(ls.CreationTime),
			FirstEventTime:    aws.ToInt64(ls.FirstEventTimestamp),
			LastEventTime:     aws.ToInt64(ls.LastEventTimestamp),
			LastIngestionTime: aws.ToInt64(ls.LastIngestionTime),
			StoredBytes:       aws.ToInt64(ls.StoredBytes),
		}
		logStreams = append(logStreams, logStream)
	}

	return logStreams, nil
}

// QueryLogsResult contains log events and pagination information
type QueryLogsResult struct {
	Events        []LogEvent `json:"events"`
	TotalReturned int        `json:"total_returned"`
	HasMore       bool       `json:"has_more"`
	StartTime     int64      `json:"start_time_ms"`
	EndTime       int64      `json:"end_time_ms"`
	TimeRangeInfo string     `json:"time_range_info"`
}

// QueryLogs queries log events with optional filter pattern
func (cw *CloudWatchService) QueryLogs(ctx context.Context, profileID string, logGroupName string, filterPattern string, startTime int64, endTime int64, limit int32) ([]LogEvent, error) {
	result, err := cw.QueryLogsWithPagination(ctx, profileID, logGroupName, filterPattern, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	return result.Events, nil
}

// QueryLogsWithPagination queries log events with full pagination support
// It will fetch multiple pages up to the specified limit
func (cw *CloudWatchService) QueryLogsWithPagination(ctx context.Context, profileID string, logGroupName string, filterPattern string, startTime int64, endTime int64, limit int32) (*QueryLogsResult, error) {
	client, err := cw.clientManager.GetCloudWatchLogsClient(profileID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}

	// AWS FilterLogEvents has a max limit of 10000 per call
	const maxPerCall int32 = 10000
	callLimit := limit
	if callLimit > maxPerCall {
		callLimit = maxPerCall
	}

	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroupName),
		Limit:        aws.Int32(callLimit),
	}

	if filterPattern != "" {
		input.FilterPattern = aws.String(filterPattern)
	}
	if startTime > 0 {
		input.StartTime = aws.Int64(startTime)
	}
	if endTime > 0 {
		input.EndTime = aws.Int64(endTime)
	}

	allEvents := make([]LogEvent, 0)
	hasMore := false

	// Paginate through results
	for {
		result, err := client.FilterLogEvents(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to query logs: %w", err)
		}

		for _, event := range result.Events {
			logEvent := LogEvent{
				Timestamp:     aws.ToInt64(event.Timestamp),
				Message:       aws.ToString(event.Message),
				IngestionTime: aws.ToInt64(event.IngestionTime),
			}
			allEvents = append(allEvents, logEvent)
		}

		// Check if we've reached our limit
		if int32(len(allEvents)) >= limit {
			allEvents = allEvents[:limit]
			hasMore = result.NextToken != nil
			break
		}

		// Check if there are more pages
		if result.NextToken == nil {
			break
		}

		input.NextToken = result.NextToken
	}

	// Build time range info for context
	timeRangeInfo := fmt.Sprintf("Queried from %s to %s",
		time.UnixMilli(startTime).Format(time.RFC3339),
		time.UnixMilli(endTime).Format(time.RFC3339))

	return &QueryLogsResult{
		Events:        allEvents,
		TotalReturned: len(allEvents),
		HasMore:       hasMore,
		StartTime:     startTime,
		EndTime:       endTime,
		TimeRangeInfo: timeRangeInfo,
	}, nil
}

// TailLogs gets the most recent log events from a log group
func (cw *CloudWatchService) TailLogs(ctx context.Context, profileID string, logGroupName string, lines int32) ([]LogEvent, error) {
	if lines <= 0 {
		lines = 100
	}

	// Get recent time range (last hour)
	endTime := time.Now().Unix() * 1000
	startTime := time.Now().Add(-1*time.Hour).Unix() * 1000

	events, err := cw.QueryLogs(ctx, profileID, logGroupName, "", startTime, endTime, lines)
	if err != nil {
		return nil, err
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	return events, nil
}

// GetLogEventsByStream gets log events from a specific log stream
func (cw *CloudWatchService) GetLogEventsByStream(ctx context.Context, profileID string, logGroupName string, logStreamName string, limit int32, startFromHead bool) ([]LogEvent, error) {
	client, err := cw.clientManager.GetCloudWatchLogsClient(profileID)
	if err != nil {
		return nil, err
	}

	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		StartFromHead: aws.Bool(startFromHead),
	}
	if limit > 0 {
		input.Limit = aws.Int32(limit)
	}

	result, err := client.GetLogEvents(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get log events: %w", err)
	}

	logEvents := make([]LogEvent, 0, len(result.Events))
	for _, event := range result.Events {
		logEvent := LogEvent{
			Timestamp:     aws.ToInt64(event.Timestamp),
			Message:       aws.ToString(event.Message),
			IngestionTime: aws.ToInt64(event.IngestionTime),
		}
		logEvents = append(logEvents, logEvent)
	}

	return logEvents, nil
}

// InsightsQueryResult contains CloudWatch Logs Insights query results
type InsightsQueryResult struct {
	QueryID       string              `json:"query_id"`
	Status        string              `json:"status"`
	Results       []map[string]string `json:"results"`
	TotalRecords  int                 `json:"total_records"`
	BytesScanned  float64             `json:"bytes_scanned"`
	StartTime     int64               `json:"start_time_ms"`
	EndTime       int64               `json:"end_time_ms"`
	TimeRangeInfo string              `json:"time_range_info"`
}// RunInsightsQuery executes a CloudWatch Logs Insights query and waits for results
// This is more powerful than FilterLogEvents for complex queries over large time ranges
func (cw *CloudWatchService) RunInsightsQuery(ctx context.Context, profileID string, logGroupNames []string, queryString string, startTime int64, endTime int64, limit int32) (*InsightsQueryResult, error) {
	client, err := cw.clientManager.GetCloudWatchLogsClient(profileID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000 // CloudWatch Logs Insights max limit
	}

	// Start the query
	startQueryInput := &cloudwatchlogs.StartQueryInput{
		LogGroupNames: logGroupNames,
		StartTime:     aws.Int64(startTime / 1000), // Insights uses seconds, not milliseconds
		EndTime:       aws.Int64(endTime / 1000),
		QueryString:   aws.String(queryString),
		Limit:         aws.Int32(limit),
	}

	startResult, err := client.StartQuery(ctx, startQueryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to start insights query: %w", err)
	}

	queryID := aws.ToString(startResult.QueryId)

	// Poll for results (with timeout)
	const maxWait = 60 * time.Second
	const pollInterval = 500 * time.Millisecond
	deadline := time.Now().Add(maxWait)

	var queryResults *cloudwatchlogs.GetQueryResultsOutput
	for time.Now().Before(deadline) {
		getResultsInput := &cloudwatchlogs.GetQueryResultsInput{
			QueryId: aws.String(queryID),
		}

		queryResults, err = client.GetQueryResults(ctx, getResultsInput)
		if err != nil {
			return nil, fmt.Errorf("failed to get query results: %w", err)
		}

		status := queryResults.Status
		if status == types.QueryStatusComplete || status == types.QueryStatusFailed || status == types.QueryStatusCancelled {
			break
		}

		time.Sleep(pollInterval)
	}

	if queryResults == nil {
		return nil, fmt.Errorf("query timed out after %v", maxWait)
	}

	// Parse results
	results := make([]map[string]string, 0, len(queryResults.Results))
	for _, row := range queryResults.Results {
		rowMap := make(map[string]string)
		for _, field := range row {
			fieldName := aws.ToString(field.Field)
			fieldValue := aws.ToString(field.Value)
			rowMap[fieldName] = fieldValue
		}
		results = append(results, rowMap)
	}

	// Build time range info
	timeRangeInfo := fmt.Sprintf("Insights query from %s to %s",
		time.UnixMilli(startTime).Format(time.RFC3339),
		time.UnixMilli(endTime).Format(time.RFC3339))

	bytesScanned := float64(0)
	if queryResults.Statistics != nil {
		bytesScanned = queryResults.Statistics.BytesScanned
	}

	return &InsightsQueryResult{
		QueryID:       queryID,
		Status:        string(queryResults.Status),
		Results:       results,
		TotalRecords:  len(results),
		BytesScanned:  bytesScanned,
		StartTime:     startTime,
		EndTime:       endTime,
		TimeRangeInfo: timeRangeInfo,
	}, nil
}
