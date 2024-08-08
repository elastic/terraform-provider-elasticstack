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

func TestDuration_StringSemanticEquals(t *testing.T) {
	tests := []struct {
		name               string
		duration           Duration
		otherVal           basetypes.StringValuable
		expectedEqual      bool
		expectedErrorDiags bool
	}{
		{
			name:               "should error if the other value is not a duration",
			duration:           NewDurationValue("3h"),
			otherVal:           basetypes.NewStringValue("3d"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
		{
			name:          "two null values are semantically equal",
			duration:      NewDurationNull(),
			otherVal:      NewDurationNull(),
			expectedEqual: true,
		},
		{
			name:          "null is not equal to unknown",
			duration:      NewDurationNull(),
			otherVal:      NewDurationUnknown(),
			expectedEqual: false,
		},
		{
			name:          "null is not equal to a string value",
			duration:      NewDurationNull(),
			otherVal:      NewDurationValue("3h"),
			expectedEqual: false,
		},
		{
			name:          "two unknown values are semantically equal",
			duration:      NewDurationUnknown(),
			otherVal:      NewDurationUnknown(),
			expectedEqual: true,
		},
		{
			name:          "unknown is not equal to a string value",
			duration:      NewDurationUnknown(),
			otherVal:      NewDurationValue("3h"),
			expectedEqual: false,
		},
		{
			name:          "two equal values are semantically equal",
			duration:      NewDurationValue("3h"),
			otherVal:      NewDurationValue("3h"),
			expectedEqual: true,
		},
		{
			name:          "two semantically equal values are semantically equal",
			duration:      NewDurationValue("3h"),
			otherVal:      NewDurationValue("180m"),
			expectedEqual: true,
		},
		{
			name:               "errors if this value is invalid",
			duration:           NewDurationValue("not a duration"),
			otherVal:           NewDurationValue("180m"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
		{
			name:               "errors if the other value is invalid",
			duration:           NewDurationValue("3h"),
			otherVal:           NewDurationValue("not a duration"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEqual, diags := tt.duration.StringSemanticEquals(context.Background(), tt.otherVal)

			require.Equal(t, tt.expectedEqual, isEqual)
			require.Equal(t, tt.expectedErrorDiags, diags.HasError())
		})
	}
}
