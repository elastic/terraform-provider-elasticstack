package integration_policy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputsValue_MapSemanticEquals(t *testing.T) {
	ctx := context.Background()
	elemType := getInputsElementType()

	tests := []struct {
		name        string
		value1      InputsValue
		value2      InputsValue
		expected    bool
		expectError bool
	}{
		{
			name:     "both null values are equal",
			value1:   NewInputsNull(elemType),
			value2:   NewInputsNull(elemType),
			expected: true,
		},
		{
			name:     "both unknown values are equal",
			value1:   NewInputsUnknown(elemType),
			value2:   NewInputsUnknown(elemType),
			expected: true,
		},
		{
			name:     "null vs unknown are not equal",
			value1:   NewInputsNull(elemType),
			value2:   NewInputsUnknown(elemType),
			expected: false,
		},
		{
			name: "same enabled inputs are equal",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: true,
		},
		{
			name: "disabled inputs are ignored - input disabled in first value",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(false),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "old_value"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
				"input2": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value2"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input2": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value2"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: true,
		},
		{
			name: "disabled inputs are ignored - input disabled in second value",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
				"input2": {
					Enabled: types.BoolValue(false),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "ignored"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: true,
		},
		{
			name: "disabled inputs are ignored - both have different disabled inputs",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
				"input2": {
					Enabled: types.BoolValue(false),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "disabled_old"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
				"input3": {
					Enabled: types.BoolValue(false),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "disabled_new"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: true,
		},
		{
			name: "different enabled inputs are not equal",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value2"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: false,
		},
		{
			name: "different number of enabled inputs are not equal",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
				"input2": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value2"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: false,
		},
		{
			name: "disabled streams are ignored",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "stream_value"}`),
						},
						"stream2": {
							Enabled: types.BoolValue(false),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "disabled_old"}`),
						},
					}),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value1"}`),
					Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "stream_value"}`),
						},
						"stream3": {
							Enabled: types.BoolValue(false),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "disabled_new"}`),
						},
					}),
				},
			}),
			expected: true,
		},
		{
			name: "vars use semantic equality - whitespace differences ignored",
			value1: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key":"value"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			value2: mustNewInputsValue(ctx, t, map[string]integrationPolicyInputsModel{
				"input1": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
					Streams: types.MapNull(getInputStreamType()),
				},
			}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.value1.MapSemanticEquals(ctx, tt.value2)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func mustNewInputsValue(ctx context.Context, t *testing.T, inputs map[string]integrationPolicyInputsModel) InputsValue {
	t.Helper()
	value, diags := NewInputsValueFrom(ctx, getInputsElementType(), inputs)
	require.False(t, diags.HasError(), "Failed to create InputsValue: %v", diags)
	return value
}

func mustNewStreamsMap(ctx context.Context, t *testing.T, streams map[string]integrationPolicyInputStreamModel) types.Map {
	t.Helper()
	value, diags := types.MapValueFrom(ctx, getInputStreamType(), streams)
	require.False(t, diags.HasError(), "Failed to create streams map: %v", diags)
	return value
}
