package validators

import (
	"testing"
)

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
			if matched != tt.matched {
				t.Errorf("StringMatchesISO8601Regex() failed match = %v, want %v", matched, tt.matched)
			}
		})
	}
}
