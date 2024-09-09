package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestDurationType_ValueType(t *testing.T) {
	require.Equal(t, Duration{}, DurationType{}.ValueType(context.Background()))
}

func TestDurationType_ValueFromString(t *testing.T) {
	stringValue := basetypes.NewStringValue("duration")
	expectedResult := Duration{StringValue: stringValue}
	durationValue, diags := DurationType{}.ValueFromString(context.Background(), stringValue)

	require.Nil(t, diags)
	require.Equal(t, expectedResult, durationValue)
}

func TestDurationType_ValueFromTerraform(t *testing.T) {
	tests := []struct {
		name          string
		tfValue       tftypes.Value
		expectedValue attr.Value
		expectedError string
	}{
		{
			name:          "should return an error if the tf value is not a string",
			tfValue:       tftypes.NewValue(tftypes.Bool, true),
			expectedValue: nil,
			expectedError: "expected string",
		},
		{
			name:          "should return a new duration value if the tf value is a string",
			tfValue:       tftypes.NewValue(tftypes.String, "3h"),
			expectedValue: NewDurationValue("3h"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := DurationType{}.ValueFromTerraform(context.Background(), tt.tfValue)

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, tt.expectedValue, val)
		})
	}
}
