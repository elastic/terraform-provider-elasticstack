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
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// --- ValidateConfigBlockPresent ---

func TestValidateConfigBlockPresent_missing_addsError(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{}
	diags := panelkit.ValidateConfigBlockPresent(attrs, "my_config", path.Empty(), "Missing config", "Config is required.")
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing config", diags[0].Summary())
	assert.Equal(t, "Config is required.", diags[0].Detail())
}

func TestValidateConfigBlockPresent_null_addsError(t *testing.T) {
	t.Parallel()
	objType := map[string]attr.Type{"slo_id": types.StringType}
	attrs := map[string]attr.Value{"my_config": types.ObjectNull(objType)}
	diags := panelkit.ValidateConfigBlockPresent(attrs, "my_config", path.Empty(), "Missing config", "Config is required.")
	require.True(t, diags.HasError())
}

func TestValidateConfigBlockPresent_unknown_noError(t *testing.T) {
	t.Parallel()
	objType := map[string]attr.Type{"slo_id": types.StringType}
	attrs := map[string]attr.Value{"my_config": types.ObjectUnknown(objType)}
	diags := panelkit.ValidateConfigBlockPresent(attrs, "my_config", path.Empty(), "Missing config", "Config is required.")
	assert.False(t, diags.HasError())
}

func TestValidateConfigBlockPresent_concreteSet_noError(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("my_config", map[string]attr.Value{"slo_id": types.StringValue("id")})
	diags := panelkit.ValidateConfigBlockPresent(attrs, "my_config", path.Empty(), "Missing config", "Config is required.")
	assert.False(t, diags.HasError())
}

func TestValidateConfigBlockPresent_usesGivenErrPath(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{}
	errPath := path.Root("panel").AtName("my_config")
	diags := panelkit.ValidateConfigBlockPresent(attrs, "my_config", errPath, "Missing config", "Config is required.")
	require.True(t, diags.HasError())
	withPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, errPath.String(), withPath.Path().String())
}

// --- ValidateRequiredStringFields ---

func TestValidateRequiredStringFields_missingBlock_addsMissingConfigError(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{}
	diags := panelkit.ValidateRequiredStringFields(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "Invalid config", "data_view_id", "metric_field")
	require.True(t, diags.HasError())
	assert.Equal(t, "Missing config", diags[0].Summary())
}

func TestValidateRequiredStringFields_nestedUnknown_defersNoError(t *testing.T) {
	t.Parallel()
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{"data_view_id": types.StringType, "metric_field": types.StringType}}
	attrs := map[string]attr.Value{"my_config": types.ObjectUnknown(objType.AttrTypes)}
	diags := panelkit.ValidateRequiredStringFields(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "Invalid config", "data_view_id", "metric_field")
	assert.False(t, diags.HasError())
}

func TestValidateRequiredStringFields_nested_missingFields_addsOneErrorPerField(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("my_config", map[string]attr.Value{
		"data_view_id": types.StringNull(),
		"metric_field": types.StringNull(),
	})
	diags := panelkit.ValidateRequiredStringFields(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "Invalid config", "data_view_id", "metric_field")
	require.Len(t, diags.Errors(), 2)
	assert.Equal(t, "Invalid config", diags[0].Summary())
	assert.Equal(t, "`data_view_id` is required.", diags[0].Detail())
	assert.Equal(t, "`metric_field` is required.", diags[1].Detail())
}

func TestValidateRequiredStringFields_nested_allPresent_noError(t *testing.T) {
	t.Parallel()
	attrs := makeNestedStringAttrs("my_config", map[string]attr.Value{
		"data_view_id": types.StringValue("logs-*"),
		"metric_field": types.StringValue("bytes"),
	})
	diags := panelkit.ValidateRequiredStringFields(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "Invalid config", "data_view_id", "metric_field")
	assert.False(t, diags.HasError())
}

func TestValidateRequiredStringFields_flat_missingOneField(t *testing.T) {
	t.Parallel()
	attrs := map[string]attr.Value{
		"data_view_id": types.StringValue("logs-*"),
		"metric_field": types.StringNull(),
	}
	diags := panelkit.ValidateRequiredStringFields(attrs, path.Empty(), "my_config",
		"Missing config", "Config is required.", "Invalid config", "data_view_id", "metric_field")
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "`metric_field` is required.", diags[0].Detail())
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
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{"slo_id": types.StringType}}
	attrs := map[string]attr.Value{"my_config": types.ObjectUnknown(objType.AttrTypes)}
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
