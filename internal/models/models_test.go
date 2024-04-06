package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSliceOrCSV_UnmarshalJSON9(t *testing.T) {
	tests := []struct {
		name           string
		jsonString     string
		expectedResult StringSliceOrCSV
		expectedErr    error
	}{
		{
			name:           "should handle json arrays",
			jsonString:     `["a", "b", "c"]`,
			expectedResult: StringSliceOrCSV{"a", "b", "c"},
		},
		{
			name:           "should handle csv strings",
			jsonString:     `"a,b,c"`,
			expectedResult: StringSliceOrCSV{"a", "b", "c"},
		},
		{
			name:       "should handle explicit nulls",
			jsonString: `null`,
		},
		{
			name:       "should handle empty strings",
			jsonString: `""`,
		},
		{
			name:        "should fail on invalid data",
			jsonString:  "true",
			expectedErr: ErrInvalidStringSliceOrCSV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualModel StringSliceOrCSV
			err := json.Unmarshal([]byte(tt.jsonString), &actualModel)

			if tt.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.expectedErr)
			}

			require.Equal(t, tt.expectedResult, actualModel)
		})
	}
}
