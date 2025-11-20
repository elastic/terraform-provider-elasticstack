package integration_policy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputValue_ObjectSemanticEquals(t *testing.T) {
	ctx := context.Background()
	attrTypes := getInputsAttributeTypes()

	tests := []struct {
		name        string
		value1      InputValue
		value2      InputValue
		expected    bool
		expectError bool
	}{
		{
			name:     "both null values are equal",
			value1:   NewInputNull(attrTypes),
			value2:   NewInputNull(attrTypes),
			expected: true,
		},
		{
			name:     "both unknown values are equal",
			value1:   NewInputUnknown(attrTypes),
			value2:   NewInputUnknown(attrTypes),
			expected: true,
		},
		{
			name:     "null vs unknown are equal",
			value1:   NewInputNull(attrTypes),
			value2:   NewInputUnknown(attrTypes),
			expected: true,
		},
		{
			name: "same inputs without defaults are equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
		{
			name: "different vars are not equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value1"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value2"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: false,
		},
		{
			name: "unset vars use defaults - equal when defaults match",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
		{
			name: "unset vars use defaults - not equal when different from defaults",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "other_value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: false,
		},
		{
			name: "both unset vars use same defaults - equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
		{
			name: "unset streams use defaults - equal when defaults match",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars: jsontypes.NewNormalizedNull(),
					Streams: map[string]inputDefaultsStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
						},
					},
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
					},
				}),
			}),
			expected: true,
		},
		{
			name: "stream vars use defaults - equal when defaults match",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars: jsontypes.NewNormalizedNull(),
					Streams: map[string]inputDefaultsStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
						},
					},
				}),
				Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						Vars:    jsontypes.NewNormalizedNull(),
					},
				}),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
					},
				}),
			}),
			expected: true,
		},
		{
			name: "disabled streams are ignored",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
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
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
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
			}),
			expected: true,
		},
		{
			name: "vars use semantic equality - whitespace differences ignored",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key":"value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.value1.ObjectSemanticEquals(ctx, tt.value2)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func mustNewInputValue(ctx context.Context, t *testing.T, input integrationPolicyInputsModel) InputValue {
	t.Helper()
	value, diags := NewInputValueFrom(ctx, getInputsAttributeTypes(), input)
	require.False(t, diags.HasError(), "Failed to create InputValue: %v", diags)
	return value
}

func mustNewInputDefaults(ctx context.Context, t *testing.T, defaults inputDefaultsModel) types.Object {
	t.Helper()
	value, diags := types.ObjectValueFrom(ctx, getInputDefaultsAttrTypes(), defaults)
	require.False(t, diags.HasError(), "Failed to create defaults object: %v", diags)
	return value
}
