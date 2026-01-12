package common

import (
	"fmt"
	"strings"
	"time"
)

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// StartMillis returns the start time in milliseconds since epoch
func (tr *TimeRange) StartMillis() int64 {
	return tr.Start.UnixMilli()
}

// EndMillis returns the end time in milliseconds since epoch
func (tr *TimeRange) EndMillis() int64 {
	return tr.End.UnixMilli()
}

// AvailableTimeRanges returns a list of available predefined time range names
func AvailableTimeRanges() []string {
	return []string{
		"last_1_hour",
		"last_3_hours",
		"last_6_hours",
		"last_12_hours",
		"last_24_hours",
		"last_2_days",
		"last_3_days",
		"last_7_days",
		"last_14_days",
		"last_30_days",
		"last_60_days",
		"last_90_days",
		"today",
		"yesterday",
		"this_week",
		"last_week",
		"this_month",
		"last_month",
	}
}

// ParseTimeRange parses a time range string and returns the corresponding TimeRange
// It supports predefined ranges (e.g., "last_7_days") and custom epoch milliseconds
func ParseTimeRange(name string) (*TimeRange, error) {
	if name == "" {
		return nil, nil
	}

	now := time.Now()

	switch strings.ToLower(strings.TrimSpace(name)) {
	// Hours-based ranges
	case "last1hour", "last_1_hour", "lasthour", "last_hour":
		start := now.Add(-1 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last3hours", "last_3_hours":
		start := now.Add(-3 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last6hours", "last_6_hours":
		start := now.Add(-6 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last12hours", "last_12_hours":
		start := now.Add(-12 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last24hours", "last_24_hours":
		start := now.Add(-24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	// Days-based ranges
	case "last2days", "last_2_days":
		start := now.Add(-2 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last3days", "last_3_days":
		start := now.Add(-3 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last7days", "last_7_days":
		start := now.Add(-7 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last14days", "last_14_days":
		start := now.Add(-14 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last30days", "last_30_days":
		start := now.Add(-30 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last60days", "last_60_days":
		start := now.Add(-60 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	case "last90days", "last_90_days":
		start := now.Add(-90 * 24 * time.Hour)
		return &TimeRange{Start: start, End: now}, nil

	// Named ranges
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return &TimeRange{Start: start, End: now}, nil

	case "yesterday":
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		start := end.Add(-24 * time.Hour)
		return &TimeRange{Start: start, End: end}, nil

	case "thisweek", "this_week":
		weekday := int(now.Weekday())
		start := now.Add(-time.Duration(weekday) * 24 * time.Hour)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, now.Location())
		return &TimeRange{Start: start, End: now}, nil

	case "lastweek", "last_week":
		weekday := int(now.Weekday())
		thisWeekStart := now.Add(-time.Duration(weekday) * 24 * time.Hour)
		thisWeekStart = time.Date(thisWeekStart.Year(), thisWeekStart.Month(), thisWeekStart.Day(), 0, 0, 0, 0, now.Location())
		lastWeekStart := thisWeekStart.Add(-7 * 24 * time.Hour)
		return &TimeRange{Start: lastWeekStart, End: thisWeekStart}, nil

	case "thismonth", "this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return &TimeRange{Start: start, End: now}, nil

	case "lastmonth", "last_month":
		thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		var lastMonthStart time.Time
		if now.Month() == 1 {
			lastMonthStart = time.Date(now.Year()-1, 12, 1, 0, 0, 0, 0, now.Location())
		} else {
			lastMonthStart = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		}
		return &TimeRange{Start: lastMonthStart, End: thisMonthStart}, nil

	default:
		return nil, fmt.Errorf("unknown time range: %s. Available ranges: %s", name, strings.Join(AvailableTimeRanges(), ", "))
	}
}

// TimeRangeHelpText returns a help text describing available time range options
func TimeRangeHelpText() string {
	return `Human-readable time range. Options: last_1_hour, last_3_hours, last_6_hours, last_12_hours, last_24_hours, last_2_days, last_3_days, last_7_days, last_14_days, last_30_days, last_60_days, last_90_days, today, yesterday, this_week, last_week, this_month, last_month. Takes precedence over date/time parameters if provided.`
}

// ParseDateTime parses a date/time string in various formats and returns the time
// Supported formats:
//   - ISO 8601: "2025-01-09T15:30:00Z", "2025-01-09T15:30:00-05:00"
//   - Date only: "2025-01-09" (assumes midnight UTC)
//   - Date with time: "2025-01-09 15:30:00"
//
// Returns nil if the input is empty
func ParseDateTime(input string) (*time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	// List of formats to try, in order of preference
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,       // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05",  // ISO without timezone (assume UTC)
		"2006-01-02 15:04:05",  // Space-separated datetime
		"2006-01-02",           // Date only (midnight UTC)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date/time '%s'. Supported formats: ISO 8601 (2025-01-09T15:30:00Z), date only (2025-01-09), or datetime (2025-01-09 15:30:00)", input)
}

// ParseDateTimeMillis parses a date/time string and returns epoch milliseconds
// Returns 0 if the input is empty
func ParseDateTimeMillis(input string) (int64, error) {
	t, err := ParseDateTime(input)
	if err != nil {
		return 0, err
	}
	if t == nil {
		return 0, nil
	}
	return t.UnixMilli(), nil
}

