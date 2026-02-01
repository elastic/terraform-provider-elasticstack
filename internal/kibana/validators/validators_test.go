package validators

import (
	"reflect"
	"testing"
)

func TestStringMatchesAlertingDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration string
		matched  bool
	}{
		{
			name:     "valid Alerting duration string (30d)",
			duration: "30d",
			matched:  true,
		},
		{
			name:     "invalid Alerting duration unit (0s)",
			duration: "0s",
			matched:  false,
		},
		{
			name:     "invalid Alerting duration value (.12y)",
			duration: ".12y",
			matched:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesAlertingDurationRegex(tt.duration)
			if !reflect.DeepEqual(matched, tt.matched) {
				t.Errorf("StringMatchesAlertingDurationRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}

func TestStringMatchesISO8601(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		date    string
		matched bool
	}{
		{
			name:    "valid complete date 1",
			date:    "1994-11-05T13:15:30Z",
			matched: true,
		},
		{
			name:    "valid complete date 2",
			date:    "1997-07-04T19:20+01:00",
			matched: true,
		},
		{
			name:    "valid complete date 3",
			date:    "1994-11-05T08:15:30-05:00",
			matched: true,
		},
		{
			name:    "valid complete date plus hours, minutes and seconds",
			date:    "1997-07-16T19:20:30+01:00",
			matched: true,
		},
		{
			name:    "valid complete date plus hours, minutes, seconds and a decimal fraction of a second",
			date:    "1997-07-16T19:20:30.45+01:00",
			matched: true,
		}, {
			name:    "invalid year",
			date:    "1997",
			matched: false,
		},
		{
			name:    "invalid year and month",
			date:    "1997-07",
			matched: false,
		},
		{
			name:    "invalid complete date",
			date:    "1997-07-04",
			matched: false,
		},
		{
			name:    "invalid hours and minutes",
			date:    "1997-40-04T30:220+01:00",
			matched: false,
		},
		{
			name:    "invalid  seconds",
			date:    "1997-07-16T19:20:80+01:00",
			matched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesISO8601Regex(tt.date)
			if !reflect.DeepEqual(matched, tt.matched) {
				t.Errorf("StringMatchesISO8601Regex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}

func TestStringMatchesOnWeekDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		onWeekDay string
		matched   bool
	}{
		{
			name:      "valid on_week_day string (+1MO)",
			onWeekDay: "+1MO",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+2TU)",
			onWeekDay: "+2TU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+3WE)",
			onWeekDay: "+3WE",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+4TH)",
			onWeekDay: "+4TH",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (+5FR)",
			onWeekDay: "+5FR",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-5SA)",
			onWeekDay: "-5SA",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-4SU)",
			onWeekDay: "-4SU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-3MO)",
			onWeekDay: "-3MO",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-2TU)",
			onWeekDay: "-2TU",
			matched:   true,
		},
		{
			name:      "valid on_week_day string (-1WE)",
			onWeekDay: "-1WE",
			matched:   true,
		},
		{
			name:      "invalid on_week_day unit (FOOBAR)",
			onWeekDay: "FOOBAR",
			matched:   false,
		},
		{
			name:      "invalid on_week_day string (+9MO)",
			onWeekDay: "+9MO",
			matched:   false,
		},
		{
			name:      "invalid on_week_day string (-7FR)",
			onWeekDay: "-7FR",
			matched:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesOnWeekDayRegex(tt.onWeekDay)
			if !reflect.DeepEqual(matched, tt.matched) {
				t.Errorf("StringMatchesOnWeekDayRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}

func TestStringMatchesIntervalFrequencyRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		intervalFrequency string
		matched           bool
	}{

		{
			name:              "valid interval/frequency string (2d)",
			intervalFrequency: "2d",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (5w)",
			intervalFrequency: "5w",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (3M)",
			intervalFrequency: "3M",
			matched:           true,
		},
		{
			name:              "valid interval/frequency string (1y)",
			intervalFrequency: "1y",
			matched:           true,
		},
		{
			name:              "invalid interval/frequency string (5m)",
			intervalFrequency: "5m",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency string (-1w)",
			intervalFrequency: "-1w",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency string (invalid)",
			intervalFrequency: "invalid",
			matched:           false,
		},
		{
			name:              "invalid interval/frequency empty string",
			intervalFrequency: "  ",
			matched:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, _ := StringMatchesIntervalFrequencyRegex(tt.intervalFrequency)
			if !reflect.DeepEqual(matched, tt.matched) {
				t.Errorf("StringMatchesOnWeekDayRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}

func TestStringIsHours(t *testing.T) {
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
			name:    "valid hours (09:00)",
			hours:   "09:00",
			matched: true,
		},
		{
			name:    "valid hours (9:00)",
			hours:   "9:00",
			matched: true,
		},
		{
			name:    "valid hours (12:30)",
			hours:   "12:30",
			matched: true,
		},
		{
			name:    "valid hours (23:59)",
			hours:   "23:59",
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
			name:    "invalid hours format (12-00)",
			hours:   "12-00",
			matched: false,
		},
		{
			name:    "invalid hours format (abc)",
			hours:   "abc",
			matched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := hoursPattern.MatchString(tt.hours)
			if matched != tt.matched {
				t.Errorf("hoursPattern.MatchString(%q) = %v, want %v", tt.hours, matched, tt.matched)
			}
		})
	}
}
