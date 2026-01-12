package common

import (
	"testing"
	"time"
)

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		validateFn  func(*TimeRange, time.Time) bool
		description string
	}{
		{
			name:        "empty string returns nil",
			input:       "",
			wantErr:     false,
			validateFn:  func(tr *TimeRange, now time.Time) bool { return tr == nil },
			description: "empty input should return nil",
		},
		{
			name:    "last_1_hour",
			input:   "last_1_hour",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-1 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should return last 1 hour range",
		},
		{
			name:    "last_24_hours",
			input:   "last_24_hours",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should return last 24 hours range",
		},
		{
			name:    "last_7_days",
			input:   "last_7_days",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-7 * 24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should return last 7 days range",
		},
		{
			name:    "last_30_days",
			input:   "last_30_days",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-30 * 24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should return last 30 days range",
		},
		{
			name:    "last_90_days",
			input:   "last_90_days",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-90 * 24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should return last 90 days range",
		},
		{
			name:    "today",
			input:   "today",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				return tr.Start.Equal(expectedStart) && tr.End.Sub(now) < time.Second
			},
			description: "should return today range",
		},
		{
			name:    "this_week",
			input:   "this_week",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				weekday := int(now.Weekday())
				expectedStart := now.Add(-time.Duration(weekday) * 24 * time.Hour)
				expectedStart = time.Date(expectedStart.Year(), expectedStart.Month(), expectedStart.Day(), 0, 0, 0, 0, now.Location())
				return tr.Start.Equal(expectedStart) && tr.End.Sub(now) < time.Second
			},
			description: "should return this week range",
		},
		{
			name:    "this_month",
			input:   "this_month",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				return tr.Start.Equal(expectedStart) && tr.End.Sub(now) < time.Second
			},
			description: "should return this month range",
		},
		{
			name:        "invalid range",
			input:       "invalid_range",
			wantErr:     true,
			validateFn:  nil,
			description: "should error on invalid range",
		},
		{
			name:    "case insensitive",
			input:   "LAST_7_DAYS",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-7 * 24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should be case insensitive",
		},
		{
			name:    "with whitespace",
			input:   "  last_7_days  ",
			wantErr: false,
			validateFn: func(tr *TimeRange, now time.Time) bool {
				expectedStart := now.Add(-7 * 24 * time.Hour)
				return tr.Start.Sub(expectedStart) < time.Second && tr.End.Sub(now) < time.Second
			},
			description: "should trim whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			tr, err := ParseTimeRange(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTimeRange(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeRange(%q) unexpected error: %v", tt.input, err)
				return
			}

			if tt.validateFn != nil && !tt.validateFn(tr, now) {
				if tr != nil {
					t.Errorf("ParseTimeRange(%q) = %v, validation failed: %s", tt.input, tr, tt.description)
				} else {
					t.Errorf("ParseTimeRange(%q) = nil, validation failed: %s", tt.input, tt.description)
				}
			}
		})
	}
}

func TestTimeRangeMillis(t *testing.T) {
	now := time.Now()
	tr := &TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	startMillis := tr.StartMillis()
	endMillis := tr.EndMillis()

	expectedStartMillis := tr.Start.UnixMilli()
	expectedEndMillis := tr.End.UnixMilli()

	if startMillis != expectedStartMillis {
		t.Errorf("StartMillis() = %d, want %d", startMillis, expectedStartMillis)
	}

	if endMillis != expectedEndMillis {
		t.Errorf("EndMillis() = %d, want %d", endMillis, expectedEndMillis)
	}

	// Verify the difference is approximately 1 hour (in milliseconds)
	diff := endMillis - startMillis
	expectedDiff := int64(time.Hour / time.Millisecond)
	if diff != expectedDiff {
		t.Errorf("Time difference = %d ms, want %d ms (1 hour)", diff, expectedDiff)
	}
}

func TestAvailableTimeRanges(t *testing.T) {
	ranges := AvailableTimeRanges()

	if len(ranges) == 0 {
		t.Error("AvailableTimeRanges() returned empty slice")
	}

	// Verify all available ranges can be parsed
	for _, rangeName := range ranges {
		tr, err := ParseTimeRange(rangeName)
		if err != nil {
			t.Errorf("AvailableTimeRanges() contains invalid range %q: %v", rangeName, err)
		}
		if tr == nil {
			t.Errorf("AvailableTimeRanges() contains range %q that parses to nil", rangeName)
		}
	}
}

func TestTimeRangeHelpText(t *testing.T) {
	help := TimeRangeHelpText()

	if help == "" {
		t.Error("TimeRangeHelpText() returned empty string")
	}

	// Verify it mentions some key ranges
	expectedTerms := []string{"last_7_days", "last_30_days", "today", "this_month"}
	for _, term := range expectedTerms {
		if !containsString(help, term) {
			t.Errorf("TimeRangeHelpText() should mention %q", term)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantNil  bool
		expected time.Time
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			wantErr: false,
			wantNil: true,
		},
		{
			name:     "RFC3339 with Z",
			input:    "2025-01-09T15:30:00Z",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 15, 30, 0, 0, time.UTC),
		},
		{
			name:     "RFC3339 with timezone offset",
			input:    "2025-01-09T15:30:00-05:00",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 15, 30, 0, 0, time.FixedZone("", -5*3600)),
		},
		{
			name:     "ISO without timezone",
			input:    "2025-01-09T15:30:00",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 15, 30, 0, 0, time.UTC),
		},
		{
			name:     "date only",
			input:    "2025-01-09",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "datetime with space",
			input:    "2025-01-09 15:30:00",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 15, 30, 0, 0, time.UTC),
		},
		{
			name:     "with whitespace",
			input:    "  2025-01-09  ",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "invalid date",
			input:   "2025-13-45",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTime(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDateTime(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDateTime(%q) unexpected error: %v", tt.input, err)
				return
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("ParseDateTime(%q) = %v, want nil", tt.input, result)
				}
				return
			}

			if result == nil {
				t.Errorf("ParseDateTime(%q) = nil, want %v", tt.input, tt.expected)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ParseDateTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDateTimeMillis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected int64
	}{
		{
			name:     "empty string returns 0",
			input:    "",
			wantErr:  false,
			expected: 0,
		},
		{
			name:     "valid date",
			input:    "2025-01-09T00:00:00Z",
			wantErr:  false,
			expected: time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC).UnixMilli(),
		},
		{
			name:    "invalid date",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTimeMillis(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDateTimeMillis(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDateTimeMillis(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseDateTimeMillis(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

