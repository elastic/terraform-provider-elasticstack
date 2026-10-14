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

package casetemplate

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuppressSemanticYAMLDiff(t *testing.T) {
	const state = "name: triage\nseverity: low\nfields:\n  - name: a\n    type: keyword\n"

	tests := []struct {
		name        string
		state       types.String
		plan        types.String
		wantPlanned types.String
	}{
		{
			name:        "whitespace and comment only change keeps state value",
			state:       types.StringValue(state),
			plan:        types.StringValue("# comment\nname:   triage\nseverity: low\n\nfields:\n    - name: a\n      type: keyword\n"),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "key order only change keeps state value",
			state:       types.StringValue(state),
			plan:        types.StringValue("severity: low\nname: triage\nfields:\n  - type: keyword\n    name: a\n"),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "different value keeps planned value",
			state:       types.StringValue(state),
			plan:        types.StringValue("name: triage\nseverity: high\nfields:\n  - name: a\n    type: keyword\n"),
			wantPlanned: types.StringValue("name: triage\nseverity: high\nfields:\n  - name: a\n    type: keyword\n"),
		},
		{
			name:        "list order change keeps planned value",
			state:       types.StringValue("fields: [a, b]\n"),
			plan:        types.StringValue("fields: [b, a]\n"),
			wantPlanned: types.StringValue("fields: [b, a]\n"),
		},
		{
			name:        "null state keeps planned value on create",
			state:       types.StringNull(),
			plan:        types.StringValue(state),
			wantPlanned: types.StringValue(state),
		},
		{
			name:        "unknown plan is left unchanged",
			state:       types.StringValue(state),
			plan:        types.StringUnknown(),
			wantPlanned: types.StringUnknown(),
		},
		{
			name:        "invalid YAML falls back to the planned value",
			state:       types.StringValue(state),
			plan:        types.StringValue("a: [unterminated"),
			wantPlanned: types.StringValue("a: [unterminated"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{StateValue: tc.state, PlanValue: tc.plan, ConfigValue: tc.plan}
			resp := &planmodifier.StringResponse{PlanValue: tc.plan}

			suppressSemanticYAMLDiff().PlanModifyString(context.Background(), req, resp)

			assert.False(t, resp.Diagnostics.HasError())
			assert.True(t, tc.wantPlanned.Equal(resp.PlanValue), "got %s", resp.PlanValue)
		})
	}
}

type attrValues map[string]tftypes.Value

func strVal(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

func tagsVal(tags ...string) tftypes.Value {
	elems := make([]tftypes.Value, len(tags))
	for i, tag := range tags {
		elems[i] = strVal(tag)
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elems)
}

func buildObject(t *testing.T, values attrValues) tftypes.Value {
	t.Helper()
	objType, ok := fullSchema(t).Type().TerraformType(context.Background()).(tftypes.Object)
	require.True(t, ok)
	full := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		if v, ok := values[name]; ok {
			full[name] = v
		} else {
			full[name] = tftypes.NewValue(attrType, nil)
		}
	}
	return tftypes.NewValue(objType, full)
}

// fullSchema returns the resource schema including the attributes and blocks
// injected by the Kibana resource envelope (kibana_connection, timeouts).
func fullSchema(t *testing.T) rschema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	newResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	return resp.Schema
}

func merge(base attrValues, overrides attrValues) attrValues {
	out := attrValues{}
	maps.Copy(out, base)
	maps.Copy(out, overrides)
	return out
}

func TestOnlyFormattingChanged(t *testing.T) {
	const stateYAML = "name: triage\nseverity: low\n"
	const reformattedYAML = "# comment\nseverity: low\nname:   triage\n"

	base := attrValues{
		"owner":           strVal("cases"),
		"name":            strVal("t"),
		"description":     strVal("d"),
		"tags":            tagsVal("a", "b"),
		"space_id":        strVal("default"),
		"is_enabled":      tftypes.NewValue(tftypes.Bool, true),
		"definition_yaml": strVal(stateYAML),
	}
	configBase := merge(base, attrValues{"definition_yaml": strVal(reformattedYAML)})

	tests := []struct {
		name         string
		config       attrValues
		state        attrValues
		nullState    bool
		wantKeepsOld bool
	}{
		{name: "reformat only keeps state", config: configBase, state: base, wantKeepsOld: true},
		{name: "reformat and name change stays unknown", config: merge(configBase, attrValues{"name": strVal("renamed")}), state: base},
		{name: "reformat and tags change stays unknown", config: merge(configBase, attrValues{"tags": tagsVal("b", "a")}), state: base},
		{name: "reformat and description removal stays unknown", config: merge(configBase, attrValues{"description": tftypes.NewValue(tftypes.String, nil)}), state: base},
		{name: "reformat and owner change stays unknown", config: merge(configBase, attrValues{"owner": strVal("observability")}), state: base},
		{name: "reformat and is_enabled change stays unknown", config: merge(configBase, attrValues{"is_enabled": tftypes.NewValue(tftypes.Bool, false)}), state: base},
		{name: "reformat with is_enabled unset keeps state", config: merge(configBase, attrValues{"is_enabled": tftypes.NewValue(tftypes.Bool, nil)}), state: base, wantKeepsOld: true},
		{name: "reformat with space_id unset keeps state", config: merge(configBase, attrValues{"space_id": tftypes.NewValue(tftypes.String, nil)}), state: base, wantKeepsOld: true},
		{name: "structural yaml change stays unknown", config: merge(configBase, attrValues{"definition_yaml": strVal("name: triage\nseverity: high\n")}), state: base},
		{name: "reformat and timeouts change stays unknown", config: merge(configBase, attrValues{"timeouts": timeoutsVal(t, "5m")}), state: base},
		{name: "reformat and kibana_connection change stays unknown", config: merge(configBase, attrValues{"kibana_connection": connectionVal(t, "rotated-key")}), state: base},
		{name: "create with null state stays unknown", config: configBase, nullState: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			s := fullSchema(t)
			cfg := tfsdk.Config{Schema: s, Raw: buildObject(t, tc.config)}
			plan := tfsdk.Plan{Schema: s, Raw: buildObject(t, tc.config)}
			objType := s.Type().TerraformType(ctx)
			state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(objType, nil)}
			if !tc.nullState {
				state.Raw = buildObject(t, tc.state)
			}

			var diags diag.Diagnostics
			got := onlyFormattingChanged(ctx, cfg, plan, state, &diags)

			require.False(t, diags.HasError(), "%s", diags)
			assert.Equal(t, tc.wantKeepsOld, got)
		})
	}
}

// TestOnlyFormattingChangedCoversConfigurableAttributes fails when a
// configurable attribute is added to the schema without being compared by
// onlyFormattingChanged, which would silently pin computed values on a plan
// that does send a PUT.
func TestOnlyFormattingChangedCoversConfigurableAttributes(t *testing.T) {
	const stateYAML = "name: triage\nseverity: low\n"
	const reformattedYAML = "# comment\nseverity: low\nname:   triage\n"

	base := attrValues{
		"owner":           strVal("cases"),
		"name":            strVal("t"),
		"description":     strVal("d"),
		"tags":            tagsVal("a", "b"),
		"space_id":        strVal("default"),
		"is_enabled":      tftypes.NewValue(tftypes.Bool, true),
		"definition_yaml": strVal(stateYAML),
	}
	configBase := merge(base, attrValues{"definition_yaml": strVal(reformattedYAML)})

	changes := map[string]tftypes.Value{
		"owner":             strVal("observability"),
		"name":              strVal("renamed"),
		"description":       strVal("other"),
		"tags":              tagsVal("b", "a"),
		"space_id":          strVal("other"),
		"is_enabled":        tftypes.NewValue(tftypes.Bool, false),
		"timeouts":          timeoutsVal(t, "5m"),
		"kibana_connection": connectionVal(t, "rotated-key"),
	}

	s := fullSchema(t)
	var configurable []string
	for name, a := range s.Attributes {
		if a.IsRequired() || a.IsOptional() {
			configurable = append(configurable, name)
		}
	}
	for name := range s.Blocks {
		configurable = append(configurable, name)
	}

	for _, name := range configurable {
		if name == "definition_yaml" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			change, ok := changes[name]
			require.True(t, ok, "configurable attribute %q has no change case; add it here and to onlyFormattingChanged", name)

			ctx := context.Background()
			config := merge(configBase, attrValues{name: change})
			cfg := tfsdk.Config{Schema: s, Raw: buildObject(t, config)}
			plan := tfsdk.Plan{Schema: s, Raw: buildObject(t, config)}
			state := tfsdk.State{Schema: s, Raw: buildObject(t, base)}

			var diags diag.Diagnostics
			got := onlyFormattingChanged(ctx, cfg, plan, state, &diags)

			require.False(t, diags.HasError(), "%s", diags)
			assert.False(t, got, "a change to %q must not be treated as formatting-only", name)
		})
	}
}

// nullObject builds an object value of the given type with all attributes null,
// overridden by the supplied values.
func nullObject(objType tftypes.Object, values map[string]tftypes.Value) tftypes.Value {
	full := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		if v, ok := values[name]; ok {
			full[name] = v
		} else {
			full[name] = tftypes.NewValue(attrType, nil)
		}
	}
	return tftypes.NewValue(objType, full)
}

func schemaAttrType(t *testing.T, name string) tftypes.Type {
	t.Helper()
	objType, ok := fullSchema(t).Type().TerraformType(context.Background()).(tftypes.Object)
	require.True(t, ok)
	attrType, ok := objType.AttributeTypes[name]
	require.True(t, ok, "attribute %s missing from schema", name)
	return attrType
}

func timeoutsVal(t *testing.T, update string) tftypes.Value {
	t.Helper()
	objType, ok := schemaAttrType(t, "timeouts").(tftypes.Object)
	require.True(t, ok)
	return nullObject(objType, map[string]tftypes.Value{"update": strVal(update)})
}

func connectionVal(t *testing.T, apiKey string) tftypes.Value {
	t.Helper()
	listType, ok := schemaAttrType(t, "kibana_connection").(tftypes.List)
	require.True(t, ok)
	elemType, ok := listType.ElementType.(tftypes.Object)
	require.True(t, ok)
	return tftypes.NewValue(listType, []tftypes.Value{
		nullObject(elemType, map[string]tftypes.Value{"api_key": strVal(apiKey)}),
	})
}
