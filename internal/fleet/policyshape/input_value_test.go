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

package policyshape

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
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			expected: true,
		},
		{
			name: "different vars are not equal",
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value1"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value2"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			expected: false,
		},
		{
			name: "unset vars use defaults - equal when defaults match",
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			expected: true,
		},
		{
			name: "unset vars use defaults - not equal when different from defaults",
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "other_value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			expected: false,
		},
		{
			name: "both unset vars use same defaults - equal",
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedNull(),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars:    jsontypes.NewNormalizedValue(`{"key": "default_value"}`),
					Streams: nil,
				}),
				Streams: types.MapNull(StreamType()),
			}),
			expected: true,
		},
		{
			name: "unset streams use defaults - equal when defaults match",
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars: jsontypes.NewNormalizedNull(),
					Streams: map[string]InputDefaultsStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
						},
					},
				}),
				Streams: types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
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
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled: types.BoolValue(true),
				Vars:    jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: mustNewInputDefaults(ctx, t, InputDefaultsModel{
					Vars: jsontypes.NewNormalizedNull(),
					Streams: map[string]InputDefaultsStreamModel{
						"stream1": {
							Enabled: types.BoolValue(true),
							Vars:    jsontypes.NewNormalizedValue(`{"stream_key": "default_stream"}`),
						},
					},
				}),
				Streams: mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
					"stream1": {
						Enabled: types.BoolValue(true),
						Vars:    jsontypes.NewNormalizedNull(),
					},
				}),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
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
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
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
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams: mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
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
			value1: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key":"value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
			}),
			value2: mustNewInputValue(ctx, t, InputModel{
				Enabled:  types.BoolValue(true),
				Vars:     jsontypes.NewNormalizedValue(`{"key": "value"}`),
				Defaults: types.ObjectNull(InputDefaultsAttributeTypes()),
				Streams:  types.MapNull(StreamType()),
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

func mustNewInputValue(ctx context.Context, t *testing.T, input InputModel) InputValue {
	t.Helper()
	value, diags := basetypes.NewObjectValueFrom(ctx, InputAttributeTypes(), input)
	require.False(t, diags.HasError(), "Failed to create InputValue: %v", diags)
	return InputValue{ObjectValue: value}
}

func mustNewInputDefaults(ctx context.Context, t *testing.T, defaults InputDefaultsModel) types.Object {
	t.Helper()
	value, diags := types.ObjectValueFrom(ctx, InputDefaultsAttributeTypes(), defaults)
	require.False(t, diags.HasError(), "Failed to create defaults object: %v", diags)
	return value
}
