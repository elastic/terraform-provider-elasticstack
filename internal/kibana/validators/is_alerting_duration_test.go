package validators

import (
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
			if matched != tt.matched {
				t.Errorf("StringMatchesAlertingDurationRegex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
