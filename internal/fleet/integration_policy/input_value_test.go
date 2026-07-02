// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package integrationpolicy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputValue_ObjectSemanticEquals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		value1      InputValue
		value2      InputValue
		expected    bool
		expectError bool
	}{
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
			name: "stream vars with injected data_stream keys are semantically equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						// Plan-side vars contain only user-configured keys.
						Vars: jsontypes.NewNormalizedValue(`{"stream_key": "stream_value"}`),
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
						// API-side vars contain the server-managed data_stream.* keys.
						Vars: jsontypes.NewNormalizedValue(`{"data_stream.dataset": "gcp_pubsub.generic", "data_stream.type": "logs", "stream_key": "stream_value"}`),
					},
				}),
			}),
			expected: true,
		},
		{
			name: "stream vars without server-managed keys are compared normally",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]integrationPolicyInputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						Vars:    jsontypes.NewNormalizedValue(`{"threshold": 42}`),
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
						Vars:    jsontypes.NewNormalizedValue(`{"threshold": 99}`),
					},
				}),
			}),
			expected: false,
		},
		{
			name: "defaults null in plan, populated in API is semantically equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				// Plan-side defaults is null.
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				// API-side defaults is populated.
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
		{
			name: "defaults populated in plan, null in API is semantically equal",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				// Old-side defaults is populated.
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				// New-side defaults is null.
				Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
				Streams:  types.MapNull(getInputStreamType()),
			}),
			expected: true,
		},
		{
			name: "differing fully-known defaults do not block equality",
			value1: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_a"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
			}),
			value2: mustNewInputValue(ctx, t, integrationPolicyInputsModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, inputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_b"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(getInputStreamType()),
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
	value, diags := basetypes.NewObjectValueFrom(ctx, getInputsAttributeTypes(), input)
	require.False(t, diags.HasError(), "Failed to create InputValue: %v", diags)
	return InputValue{ObjectValue: value}
}

func mustNewInputDefaults(ctx context.Context, t *testing.T, defaults inputDefaultsModel) types.Object {
	t.Helper()
	value, diags := types.ObjectValueFrom(ctx, getInputDefaultsAttrTypes(), defaults)
	require.False(t, diags.HasError(), "Failed to create defaults object: %v", diags)
	return value
}

func TestStripServerManagedVarsKeys(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		input    jsontypes.Normalized
		expected jsontypes.Normalized
	}{
		{
			name:     "null input returned unchanged",
			input:    jsontypes.NewNormalizedNull(),
			expected: jsontypes.NewNormalizedNull(),
		},
		{
			name:     "both server-managed keys stripped",
			input:    jsontypes.NewNormalizedValue(`{"data_stream.dataset":"gcp_pubsub.generic","data_stream.type":"logs","stream_key":"stream_value"}`),
			expected: jsontypes.NewNormalizedValue(`{"stream_key":"stream_value"}`),
		},
		{
			name:     "no server-managed keys returned unchanged",
			input:    jsontypes.NewNormalizedValue(`{"key1":"value1","key2":"value2"}`),
			expected: jsontypes.NewNormalizedValue(`{"key1":"value1","key2":"value2"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := stripServerManagedVarsKeys(tt.input)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			if tt.input.IsNull() {
				require.True(t, result.IsNull(), "expected null result for null input")
				return
			}

			// For non-null cases, compare the normalized JSON values semantically.
			// Use StringSemanticEquals which is key-order-insensitive.
			equal, d := result.StringSemanticEquals(ctx, tt.expected)
			require.False(t, d.HasError(), "semantic equality failed: %v", d)
			require.True(t, equal, "expected %s to equal %s", result.ValueString(), tt.expected.ValueString())
		})
	}
}
