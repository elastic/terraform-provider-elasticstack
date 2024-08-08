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
