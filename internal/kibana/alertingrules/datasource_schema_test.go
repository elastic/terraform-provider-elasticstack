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

package alertingrules

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestNewDataSource_schemaDeclaresInputsAndComputedRules(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()
	_, ok := ds.(datasource.DataSourceWithConfigure)
	require.True(t, ok)

	resp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError())

	for _, name := range []string{"space_id", "rule_id", "filter", "id", "rules"} {
		_, ok := resp.Schema.Attributes[name]
		require.Truef(t, ok, "schema missing attribute %q", name)
	}
	_, hasConnection := resp.Schema.Blocks["kibana_connection"]
	require.True(t, hasConnection, "envelope should inject kibana_connection")
}

func TestSchema_rulesElementAttributes(t *testing.T) {
	t.Parallel()

	sch := getDataSourceSchema(context.Background())
	nested, ok := sch.Attributes["rules"].(schema.ListNestedAttribute)
	require.True(t, ok)

	for _, name := range []string{
		"id", "name", "rule_type_id", "consumer", "enabled", "tags",
		"scheduled_task_id", "last_execution_status", "last_execution_date",
	} {
		_, exists := nested.NestedObject.Attributes[name]
		require.Truef(t, exists, "rules element missing %q", name)
	}
}

func TestSchema_ruleIDAndFilterTogetherIsError(t *testing.T) {
	t.Parallel()

	diags := validateSelectors(t, tftypes.NewValue(tftypes.String, "abc"), tftypes.NewValue(tftypes.String, "alert.attributes.enabled: true"))
	require.True(t, diags.HasError())
}

func TestSchema_emptyRuleIDIsError(t *testing.T) {
	t.Parallel()

	diags := validateSelectors(t, tftypes.NewValue(tftypes.String, ""), tftypes.NewValue(tftypes.String, nil))
	require.True(t, diags.HasError())
}

func TestSchema_emptyFilterIsError(t *testing.T) {
	t.Parallel()

	diags := validateSelectors(t, tftypes.NewValue(tftypes.String, nil), tftypes.NewValue(tftypes.String, ""))
	require.True(t, diags.HasError())
}

func TestSchema_neitherSelectorIsValid(t *testing.T) {
	t.Parallel()

	diags := validateSelectors(t, tftypes.NewValue(tftypes.String, nil), tftypes.NewValue(tftypes.String, nil))
	require.False(t, diags.HasError())
}

func validateSelectors(t *testing.T, ruleID, filter tftypes.Value) diag.Diagnostics {
	t.Helper()

	ctx := context.Background()
	sch := getDataSourceSchema(ctx)
	cfg := tfsdk.Config{
		Schema: sch,
		Raw:    tftypes.NewValue(sch.Type().TerraformType(ctx), selectorConfigValues(sch, ruleID, filter)),
	}

	var diags diag.Diagnostics
	for _, name := range []string{"rule_id", "filter"} {
		attr, ok := sch.Attributes[name].(schema.StringAttribute)
		require.Truef(t, ok, "%s is not a string attribute", name)

		var value types.String
		diags.Append(cfg.GetAttribute(ctx, path.Root(name), &value)...)
		for _, v := range attr.Validators {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{
				Path:           path.Root(name),
				PathExpression: path.MatchRoot(name),
				Config:         cfg,
				ConfigValue:    value,
			}, resp)
			diags.Append(resp.Diagnostics...)
		}
	}
	return diags
}

func selectorConfigValues(sch schema.Schema, ruleID, filter tftypes.Value) map[string]tftypes.Value {
	ctx := context.Background()
	tfType := sch.Type().TerraformType(ctx).(tftypes.Object)
	return map[string]tftypes.Value{
		"space_id": tftypes.NewValue(tftypes.String, nil),
		"rule_id":  ruleID,
		"filter":   filter,
		"id":       tftypes.NewValue(tftypes.String, nil),
		"rules":    tftypes.NewValue(tfType.AttributeTypes["rules"], nil),
	}
}
