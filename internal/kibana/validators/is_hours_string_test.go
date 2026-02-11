package validators

import (
	"testing"
)

func TestStringMatchesHoursRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		hours   string
		matched bool
	}{
		{
			name:    "valid hours (00:00)",
			hours:   "00:00",
			matched: true,
		},
		{
			name:    "valid hours (09:30)",
			hours:   "09:30",
			matched: true,
		},
		{
			name:    "valid hours (14:45)",
			hours:   "14:45",
			matched: true,
		},
		{
			name:    "valid hours (23:59)",
			hours:   "23:59",
			matched: true,
		},
		{
			name:    "valid hours single digit hour (9:30)",
			hours:   "9:30",
			matched: true,
		},
		{
			name:    "invalid hours (24:00)",
			hours:   "24:00",
			matched: false,
		},
		{
			name:    "invalid hours (12:60)",
			hours:   "12:60",
			matched: false,
		},
		{
			name:    "invalid hours (25:00)",
			hours:   "25:00",
			matched: false,
		},
		{
			name:    "invalid hours format (1200)",
			hours:   "1200",
			matched: false,
		},
		{
			name:    "invalid hours format (12)",
			hours:   "12",
			matched: false,
		},
		{
			name:    "invalid hours empty string",
			hours:   "",
			matched: false,
		},
		{
			name:    "invalid hours format (12:00:00)",
			hours:   "12:00:00",
			matched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesHoursRegex(tt.hours)
			if matched != tt.matched {
				t.Errorf("StringMatchesHoursRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
