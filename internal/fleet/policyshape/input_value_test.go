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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// defaultsLessInputAttributeTypes reproduces the shape of
// internal/fleet/managedintegration/schema.go's managedIntegrationInputAttributeTypes()
// (enabled/condition/vars/streams, deliberately omitting `defaults`) without
// importing that package (policyshape must not depend on any of its own
// consumers). This is the attribute-types map decodeInputModel's
// defaults-less branch (input_value.go) exists to tolerate: every other test
// in this file builds an InputValue via InputAttributeTypes(), which always
// includes `defaults`, so none of them exercise that branch.
func defaultsLessInputAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		AttrEnabled:   types.BoolType,
		AttrCondition: types.StringType,
		AttrVars:      jsontypes.NormalizedType{},
		AttrStreams: types.MapType{
			ElemType: StreamType(),
		},
	}
}

func mustNewDefaultsLessInputValue(ctx context.Context, t *testing.T, m inputModelSansDefaults) InputValue {
	t.Helper()
	value, diags := basetypes.NewObjectValueFrom(ctx, defaultsLessInputAttributeTypes(), m)
	require.False(t, diags.HasError(), "Failed to create defaults-less InputValue: %v", diags)
	return InputValue{ObjectValue: value}
}

// TestInputValue_decodeInputModel_defaultsLessSchema covers the bug fixed by
// decodeInputModel (input_value.go): before that fix, every one of
// ObjectSemanticEquals/EnabledByDefault/MaybeEnabled called v.As(ctx,
// &InputModel{}, ...) directly, which hard-fails with "Struct defines fields
// not found in object: defaults" against an InputValue built from an
// attribute-types map with no `defaults` key -- exactly what
// managedintegration's managedIntegrationInputAttributeTypes() produces. This test
// exercises all three methods against such a defaults-less InputValue and
// asserts they behave correctly (no error, and semantically sane results)
// rather than hard-failing.
func TestInputValue_decodeInputModel_defaultsLessSchema(t *testing.T) {
	ctx := context.Background()

	enabledStream := mustNewStreamsMap(ctx, t, map[string]InputStreamModel{
		"stream1": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedNull(),
		},
	})

	value1 := mustNewDefaultsLessInputValue(ctx, t, inputModelSansDefaults{
		Enabled:   types.BoolValue(true),
		Condition: types.StringNull(),
		Vars:      jsontypes.NewNormalizedValue(`{"key":"value"}`),
		Streams:   enabledStream,
	})
	value2 := mustNewDefaultsLessInputValue(ctx, t, inputModelSansDefaults{
		Enabled:   types.BoolValue(true),
		Condition: types.StringNull(),
		Vars:      jsontypes.NewNormalizedValue(`{"key":"value"}`),
		Streams:   enabledStream,
	})

	equal, diags := value1.ObjectSemanticEquals(ctx, value2)
	require.False(t, diags.HasError(), "ObjectSemanticEquals should not error on a defaults-less InputValue: %v", diags)
	assert.True(t, equal, "identical defaults-less inputs should be semantically equal")

	enabledByDefault, diags := value1.EnabledByDefault(ctx)
	require.False(t, diags.HasError(), "EnabledByDefault should not error on a defaults-less InputValue: %v", diags)
	assert.False(t, enabledByDefault, "there is no defaults object to consult, so EnabledByDefault should be false")

	// MaybeEnabled treats an input as disabled unless at least one of its
	// streams is enabled (see input_value.go); value1 has one enabled stream,
	// so this should be true.
	maybeEnabled, diags := value1.MaybeEnabled(ctx)
	require.False(t, diags.HasError(), "MaybeEnabled should not error on a defaults-less InputValue: %v", diags)
	assert.True(t, maybeEnabled, "at least one stream is enabled, so MaybeEnabled should be true")
}

// TestInputValue_decodeInputModel_defaultsLessSchema_differentVarsNotEqual
// confirms ObjectSemanticEquals still correctly distinguishes different
// values (not just that it avoids erroring) on a defaults-less InputValue.
func TestInputValue_decodeInputModel_defaultsLessSchema_differentVarsNotEqual(t *testing.T) {
	ctx := context.Background()

	value1 := mustNewDefaultsLessInputValue(ctx, t, inputModelSansDefaults{
		Enabled: types.BoolValue(true),
		Vars:    jsontypes.NewNormalizedValue(`{"key":"value1"}`),
		Streams: types.MapNull(StreamType()),
	})
	value2 := mustNewDefaultsLessInputValue(ctx, t, inputModelSansDefaults{
		Enabled: types.BoolValue(true),
		Vars:    jsontypes.NewNormalizedValue(`{"key":"value2"}`),
		Streams: types.MapNull(StreamType()),
	})

	equal, diags := value1.ObjectSemanticEquals(ctx, value2)
	require.False(t, diags.HasError(), "%v", diags)
	assert.False(t, equal, "different vars should not be semantically equal")
}

func mustNewInputDefaults(ctx context.Context, t *testing.T, defaults InputDefaultsModel) types.Object {
	t.Helper()
	value, diags := types.ObjectValueFrom(ctx, InputDefaultsAttributeTypes(), defaults)
	require.False(t, diags.HasError(), "Failed to create defaults object: %v", diags)
	return value
}
