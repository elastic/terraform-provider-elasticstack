package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringIsJSONObject(t *testing.T) {
	tests := []struct {
		name                  string
		fieldVal              interface{}
		expectedErrsToContain []string
	}{
		{
			name:     "should not return an error for a valid json object",
			fieldVal: "{}",
		},
		{
			name:     "should return an error if the field is not a string",
			fieldVal: true,
			expectedErrsToContain: []string{
				"expected type of field-name to be string",
			},
		},
		{
			name:     "should return an error if the field is valid json, but not an object",
			fieldVal: "[]",
			expectedErrsToContain: []string{
				"expected field-name to be a JSON object. Check the documentation for the expected format.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, errors := stringIsJSONObject(tt.fieldVal, "field-name")
			require.Empty(t, warnings)

			require.Equal(t, len(tt.expectedErrsToContain), len(errors))
			for i, err := range errors {
				require.ErrorContains(t, err, tt.expectedErrsToContain[i])
			}
		})
	}
}
