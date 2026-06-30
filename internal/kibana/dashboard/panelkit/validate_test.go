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

package panelkit_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeNestedStringAttrs builds an attrs map with a nested object under cfgKey containing string fields.
func makeNestedStringAttrs(cfgKey string, fields map[string]attr.Value) map[string]attr.Value {
	attrTypes := make(map[string]attr.Type, len(fields))
	for k := range fields {
		attrTypes[k] = types.StringType
	}
	obj, _ := types.ObjectValue(attrTypes, fields)
	return map[string]attr.Value{cfgKey: obj}
}

// --- ResolveConfigBlock ---

func TestResolveConfigBlock_unshaped_addsError(t *testing.T) {
	t.Parallel()
	// attrs has neither the flat key nor the config key
	attrs := map[string]attr.Value{"other": types.StringValue("x")}
	flat, _, _, skip, diags := panelkit.ResolveConfigBlock(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "slo_id")
	assert.False(t, flat)
	assert.True(t, skip)
	require.True(t, diags.HasError())
}

func TestResolveConfigBlock_flat_noError(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{"slo_id": types.StringValue("my-slo")}
	flat, _, _, skip, diags := panelkit.ResolveConfigBlock(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "slo_id")
	assert.True(t, flat)
	assert.False(t, skip)
	assert.False(t, diags.HasError())
}

func TestResolveConfigBlock_nestedUnknown_skipsNoError(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{"my_config": types.StringUnknown()}
	_, _, _, skip, diags := panelkit.ResolveConfigBlock(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "slo_id")
	assert.True(t, skip)
	assert.False(t, diags.HasError())
}

func TestResolveConfigBlock_nestedNull_addsError(t *testing.T) {
	t.Parallel()
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{"slo_id": types.StringType}}
	attrs := map[string]attr.Value{"my_config": types.ObjectNull(objType.AttrTypes)}
	_, _, _, skip, diags := panelkit.ResolveConfigBlock(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "slo_id")
	assert.True(t, skip)
	require.True(t, diags.HasError())
}

func TestResolveConfigBlock_nested_valid_noError(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("my_config", map[string]attr.Value{"slo_id": types.StringValue("id")})
	flat, obj, _, skip, diags := panelkit.ResolveConfigBlock(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "slo_id")
	assert.False(t, flat)
	assert.False(t, skip)
	assert.False(t, diags.HasError())
	assert.Equal(t, "id", obj.Attributes()["slo_id"].(types.String).ValueString())
}

// --- ValidateRequiredStringField ---

func TestValidateRequiredStringField_flat_present(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{"slo_id": types.StringValue("my-id")}
	deferred, diags := panelkit.ValidateRequiredStringField(attrs, types.Object{}, true, path.Empty(), "slo_id", "Err", "slo_id required.")
	assert.False(t, deferred)
	assert.False(t, diags.HasError())
}

func TestValidateRequiredStringField_flat_missing(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{"slo_id": types.StringNull()}
	deferred, diags := panelkit.ValidateRequiredStringField(attrs, types.Object{}, true, path.Empty(), "slo_id", "Err", "slo_id required.")
	assert.False(t, deferred)
	require.True(t, diags.HasError())
}

func TestValidateRequiredStringField_flat_unknown_defers(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{"slo_id": types.StringUnknown()}
	deferred, diags := panelkit.ValidateRequiredStringField(attrs, types.Object{}, true, path.Empty(), "slo_id", "Err", "slo_id required.")
	assert.True(t, deferred)
	assert.False(t, diags.HasError())
}

func TestValidateRequiredStringField_nested_present(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("cfg", map[string]attr.Value{"slo_id": types.StringValue("x")})
	raw := attrs["cfg"].(types.Object)
	deferred, diags := panelkit.ValidateRequiredStringField(nil, raw, false, path.Empty(), "slo_id", "Err", "slo_id required.")
	assert.False(t, deferred)
	assert.False(t, diags.HasError())
}

func TestValidateRequiredStringField_nested_missing(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("cfg", map[string]attr.Value{"slo_id": types.StringNull()})
	raw := attrs["cfg"].(types.Object)
	deferred, diags := panelkit.ValidateRequiredStringField(nil, raw, false, path.Empty(), "slo_id", "Err", "slo_id required.")
	assert.False(t, deferred)
	require.True(t, diags.HasError())
}
