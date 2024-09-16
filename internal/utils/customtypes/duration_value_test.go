package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestDuration_Type(t *testing.T) {
	require.Equal(t, DurationType{}, Duration{}.Type(context.Background()))
}

func TestDuration_Equal(t *testing.T) {
	tests := []struct {
		name          string
		expectedEqual bool
		val           Duration
		otherVal      attr.Value
	}{
		{
			name:          "not equal if the other value is not a duration",
			expectedEqual: false,
			val:           NewDurationValue("3h"),
			otherVal:      basetypes.NewBoolValue(true),
		},
		{
			name:          "not equal if the durations are not equal",
			expectedEqual: false,
			val:           NewDurationValue("3h"),
			otherVal:      NewDurationValue("1m"),
		},
		{
			name:          "not equal if the durations are semantically equal but string values are not equal",
			expectedEqual: false,
			val:           NewDurationValue("60s"),
			otherVal:      NewDurationValue("1m"),
		},
		{
			name:          "equal if the duration string values are equal",
			expectedEqual: true,
			val:           NewDurationValue("3h"),
			otherVal:      NewDurationValue("3h"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedEqual, tt.val.Equal(tt.otherVal))
		})
	}
}

func TestDuration_ValidateAttribute(t *testing.T) {
	tests := []struct {
		name          string
		duration      Duration
		expectedDiags diag.Diagnostics
	}{
		{
			name:     "unknown is valid",
			duration: NewDurationNull(),
		},
		{
			name:     "null is valid",
			duration: NewDurationUnknown(),
		},
		{
			name:     "valid durations are valid",
			duration: NewDurationValue("3h"),
		},
		{
			name:     "non-duration strings are invalid",
			duration: NewDurationValue("not a duration"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("duration"),
					"Invalid Duration string value",
					`A string value was provided that is not a valid Go duration\n\nGiven value "not a duration"\n`,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			tt.duration.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{
					Path: path.Root("duration"),
				},
				&resp,
			)

			if tt.expectedDiags == nil {
				require.Nil(t, resp.Diagnostics)
			} else {
				require.Equal(t, tt.expectedDiags, resp.Diagnostics)
			}
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
