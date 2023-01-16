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
