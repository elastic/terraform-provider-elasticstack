package utils

import (
	"errors"
	"reflect"
	"testing"
)

func TestStringIsDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            interface{}
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid duration string",
			i:    "30s",
			k:    "timeout",
		},
		{
			name:       "invalid duration string",
			i:          "30ss",
			k:          "timeout",
			wantErrors: []error{errors.New(`"timeout" contains an invalid duration: time: unknown unit "ss" in duration "30ss"`)},
		},
		{
			name:       "invalid type",
			i:          30,
			k:          "timeout",
			wantErrors: []error{errors.New("expected type of timeout to be string")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsDuration(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsDuration() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("StringIsDuration() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func TestStringIsElasticDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            interface{}
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid Elastic duration string",
			i:    "30d",
			k:    "delay",
		},
		{
			name:       "invalid Elastic duration unit",
			i:          "12w",
			k:          "delay",
			wantErrors: []error{errors.New(`"delay" contains an invalid duration: not conforming to Elastic time-units format`)},
		},
		{
			name:       "invalid Elastic duration value",
			i:          ".12s",
			k:          "delay",
			wantErrors: []error{errors.New(`"delay" contains an invalid duration: not conforming to Elastic time-units format`)},
		},
		{
			name:       "invalid data type",
			i:          30,
			k:          "delay",
			wantErrors: []error{errors.New("expected type of delay to be string")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsElasticDuration(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsElasticDuration() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("StringIsElasticDuration() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func TestStringIsAlertingDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            interface{}
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid Alerting duration string (30d)",
			i:    "30d",
			k:    "duration",
		},
		{
			name:       "invalid Alerting duration unit (0s)",
			i:          "0s",
			k:          "duration",
			wantErrors: []error{errors.New(`invalid value for duration (string is not a valid Alerting duration in seconds (s), minutes (m), hours (h), or days (d))`)},
		},
		{
			name:       "invalid Alerting duration value (.12y)",
			i:          ".12y",
			k:          "duration",
			wantErrors: []error{errors.New(`invalid value for duration (string is not a valid Alerting duration in seconds (s), minutes (m), hours (h), or days (d))`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsAlertingDuration()(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsAlertingDuration() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("StringIsAlertingDuration() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func TestStringIsMaintenanceWindowOnWeekDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            interface{}
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid on_week_day string (+1MO)",
			i:    "+1MO",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (+2TU)",
			i:    "+2TU",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (+3WE)",
			i:    "+3WE",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (+4TH)",
			i:    "+4TH",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (+5FR)",
			i:    "+5FR",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (-5SA)",
			i:    "-5SA",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (-4SU)",
			i:    "-4SU",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (-3MO)",
			i:    "-3MO",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (-2TU)",
			i:    "-2TU",
			k:    "on_week_day",
		},
		{
			name: "valid on_week_day string (-1WE)",
			i:    "-1WE",
			k:    "on_week_day",
		},
		{
			name:       "invalid on_week_day unit (FOOBAR)",
			i:          "FOOBAR",
			k:          "on_week_day",
			wantErrors: []error{errors.New("invalid value for on_week_day (string is not a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`).)")},
		},
		{
			name:       "invalid on_week_day string (+9MO)",
			i:          "+9MO",
			k:          "on_week_day",
			wantErrors: []error{errors.New("invalid value for on_week_day (string is not a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`).)")},
		},
		{
			name:       "invalid on_week_day string (-7FR)",
			i:          "-7FR",
			k:          "on_week_day",
			wantErrors: []error{errors.New("invalid value for on_week_day (string is not a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`).)")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsMaintenanceWindowOnWeekDay()(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsMaintenanceWindowOnWeekDay() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("StringIsMaintenanceWindowOnWeekDay() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func TestStringIsMaintenanceWindowIntervalFrequency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            interface{}
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid interval/frequency string (2d)",
			i:    "2d",
			k:    "every",
		},
		{
			name: "valid interval/frequency string (5w)",
			i:    "5w",
			k:    "every",
		},
		{
			name: "valid interval/frequency string (3M)",
			i:    "3M",
			k:    "every",
		},
		{
			name: "valid interval/frequency string (1y)",
			i:    "1y",
			k:    "every",
		},
		{
			name:       "invalid interval/frequency string (5m)",
			i:          "5m",
			k:          "every",
			wantErrors: []error{errors.New("invalid value for every (string is not a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.)")},
		},
		{
			name:       "invalid interval/frequency string (-1w)",
			i:          "-1w",
			k:          "every",
			wantErrors: []error{errors.New("invalid value for every (string is not a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.)")},
		},
		{
			name:       "invalid interval/frequency string (invalid)",
			i:          "invalid",
			k:          "every",
			wantErrors: []error{errors.New("invalid value for every (string is not a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.)")},
		},
		{
			name:       "invalid interval/frequency empty string",
			i:          "  ",
			k:          "every",
			wantErrors: []error{errors.New("invalid value for every (string is not a valid interval/frequency. Allowed values are in the `<integer><unit>` format. `<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. For example: `15d`, `2w`, `3m`, `1y`.)")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsMaintenanceWindowIntervalFrequency()(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsMaintenanceWindowIntervalFrequency() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("StringIsMaintenanceWindowIntervalFrequency() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}
