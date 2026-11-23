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

package rolloverindexseries

import (
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func loadedSchema(t *testing.T) schema.Schema {
	t.Helper()
	series := NewRolloverIndexSeriesResource()
	var resp resource.SchemaResponse
	series.Schema(t.Context(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return resp.Schema
}

func TestSchema_registersSeriesAttributes(t *testing.T) {
	series := NewRolloverIndexSeriesResource()
	var meta resource.MetadataResponse
	series.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
	require.Equal(t, "elasticstack_elasticsearch_rollover_index_series", meta.TypeName)

	s := loadedSchema(t)
	require.Equal(t, []string{
		"deletion_protection",
		"id",
		"indices",
		"prefix",
		"rollover_alias",
		"timeouts",
		"write_index",
	}, sortedKeys(s.Attributes))
	require.Contains(t, s.Blocks, "elasticsearch_connection")

	timeoutsAttr, ok := s.Attributes["timeouts"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	for _, operation := range []string{"create", "read", "update", "delete"} {
		require.Contains(t, timeoutsAttr.Attributes, operation)
	}

	writeIndex := s.Attributes["write_index"].(schema.StringAttribute)
	require.True(t, writeIndex.Computed)
	require.False(t, writeIndex.Required)
	require.False(t, writeIndex.Optional)
	require.Nil(t, writeIndex.Default)

	protection := s.Attributes["deletion_protection"].(schema.BoolAttribute)
	require.True(t, protection.Optional)
	require.NotNil(t, protection.Default)
	var defaultResp defaults.BoolResponse
	protection.Default.DefaultBool(t.Context(), defaults.BoolRequest{}, &defaultResp)
	require.False(t, defaultResp.Diagnostics.HasError())
	require.Equal(t, types.BoolValue(true), defaultResp.PlanValue)
}

func TestSchema_namingInputsForceReplacement(t *testing.T) {
	require.False(t, stringRequiresReplace(t, "prefix", "orders", "orders"))
	require.True(t, stringRequiresReplace(t, "prefix", "orders", "billing"))
	require.False(t, stringRequiresReplace(t, "rollover_alias", "orders-write", "orders-write"))
	require.True(t, stringRequiresReplace(t, "rollover_alias", "orders-write", "billing-write"))
}

func TestSchema_unsupportedIndexConfigurationRejected(t *testing.T) {
	s := loadedSchema(t)
	for _, name := range []string{
		"settings",
		"settings_raw",
		"number_of_shards",
		"number_of_replicas",
		"mappings",
		"aliases",
		"alias",
		"filter",
		"routing",
		"index_routing",
		"search_routing",
		"is_write_index",
		"lifecycle",
		"rollover_conditions",
		"initial_index",
		"generation",
	} {
		_, exists := s.Attributes[name]
		require.Falsef(t, exists, "schema exposes unsupported attribute %s", name)
	}
}

func TestSchema_configuredNamesUseLiteralValidators(t *testing.T) {
	wildcard := validateSchemaString(t, "prefix", "orders-*")
	require.True(t, wildcard.HasError())
	require.Contains(t, diagnosticText(wildcard), "orders-*")

	tooLongPrefix := strings.Repeat("a", maxElasticsearchNameBytes-len(bootstrapGenerationSuffix)+1)
	length := validateSchemaString(t, "prefix", tooLongPrefix)
	require.True(t, length.HasError())
	require.Contains(t, diagnosticText(length), bootstrapIndexName(tooLongPrefix))

	comma := validateSchemaString(t, "rollover_alias", "orders-write,other-alias")
	require.True(t, comma.HasError())
	require.Contains(t, diagnosticText(comma), "orders-write,other-alias")
}

func stringRequiresReplace(t *testing.T, name, prior, next string) bool {
	t.Helper()
	attr := loadedSchema(t).Attributes[name].(schema.StringAttribute)
	known := tftypes.NewValue(tftypes.String, "present")
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue(next),
		StateValue:  types.StringValue(prior),
		PlanValue:   types.StringValue(next),
		State:       tfsdk.State{Raw: known},
		Plan:        tfsdk.Plan{Raw: known},
	}
	replace := false
	for _, modifier := range attr.PlanModifiers {
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		modifier.PlanModifyString(t.Context(), req, resp)
		replace = replace || resp.RequiresReplace
	}
	return replace
}

func validateSchemaString(t *testing.T, name, value string) diag.Diagnostics {
	t.Helper()
	attr := loadedSchema(t).Attributes[name].(schema.StringAttribute)
	var diags diag.Diagnostics
	req := validator.StringRequest{
		Path:        path.Root(name),
		ConfigValue: types.StringValue(value),
	}
	for _, v := range attr.Validators {
		resp := &validator.StringResponse{}
		v.ValidateString(t.Context(), req, resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func sortedKeys(attrs map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
