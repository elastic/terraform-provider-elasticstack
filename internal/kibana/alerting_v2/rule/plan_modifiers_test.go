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

package rule

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

type attrValues map[string]tftypes.Value

func strVal(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }

// fullSchema returns the resource schema including the attributes and blocks
// injected by the Kibana resource envelope (kibana_connection, timeouts).
func fullSchema(t *testing.T) rschema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	newResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	return resp.Schema
}

func rootObjectType(t *testing.T) tftypes.Object {
	t.Helper()
	objType, ok := fullSchema(t).Type().TerraformType(context.Background()).(tftypes.Object)
	require.True(t, ok)
	return objType
}

// objectVal builds an object of objType from values, filling every attribute
// not given with null.
func objectVal(objType tftypes.Object, values attrValues) tftypes.Value {
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

func nestedObjectType(t *testing.T, root tftypes.Object, name string) tftypes.Object {
	t.Helper()
	objType, ok := root.AttributeTypes[name].(tftypes.Object)
	require.True(t, ok)
	return objType
}

func artifactsVal(t *testing.T, runbookData string) tftypes.Value {
	t.Helper()
	listType, ok := rootObjectType(t).AttributeTypes["artifacts"].(tftypes.List)
	require.True(t, ok)
	elemType, ok := listType.ElementType.(tftypes.Object)
	require.True(t, ok)
	return tftypes.NewValue(listType, []tftypes.Value{
		objectVal(elemType, attrValues{"id": strVal("rb"), "type": strVal("runbook"), "data": strVal(runbookData)}),
	})
}

func ruleObject(t *testing.T, overrides attrValues) tftypes.Value {
	t.Helper()
	root := rootObjectType(t)
	nested := func(name string) tftypes.Object { return nestedObjectType(t, root, name) }
	values := attrValues{
		"id":         strVal("default/my-rule"),
		"rule_id":    strVal("my-rule"),
		"space_id":   strVal("default"),
		"kind":       strVal("alert"),
		"enabled":    tftypes.NewValue(tftypes.Bool, true),
		"time_field": strVal("@timestamp"),
		"metadata":   objectVal(nested("metadata"), attrValues{"name": strVal("CPU high")}),
		"schedule":   objectVal(nested("schedule"), attrValues{"every": strVal("1m")}),
		"query":      objectVal(nested("query"), attrValues{"base": strVal("FROM logs-*")}),
		"recovery":   objectVal(nested("recovery"), attrValues{"strategy": strVal("no_breach")}),
		"no_data":    objectVal(nested("no_data"), attrValues{"strategy": strVal("ignore")}),
		"version":    tftypes.NewValue(tftypes.Number, 3),
		"created_at": strVal("2026-01-02T03:04:05.000Z"),
		"updated_at": strVal("2026-01-03T03:04:05.000Z"),
	}
	maps.Copy(values, overrides)
	return objectVal(root, values)
}

func modifyPlan(t *testing.T, plan, state tftypes.Value) *resource.ModifyPlanResponse {
	t.Helper()
	schema := fullSchema(t)
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: schema, Raw: plan},
		Plan:   tfsdk.Plan{Schema: schema, Raw: plan},
		State:  tfsdk.State{Schema: schema, Raw: state},
	}
	resp := &resource.ModifyPlanResponse{Plan: req.Plan}
	newResource().ModifyPlan(context.Background(), req, resp)
	return resp
}

func TestResource_modifyPlan_leavesUnknownArtifactsUntouched(t *testing.T) {
	t.Parallel()

	listType := rootObjectType(t).AttributeTypes["artifacts"]
	plan := ruleObject(t, attrValues{
		"artifacts":  tftypes.NewValue(listType, tftypes.UnknownValue),
		"version":    tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"updated_at": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})
	state := ruleObject(t, attrValues{"artifacts": artifactsVal(t, `{"content":"# Steps"}`)})

	resp := modifyPlan(t, plan, state)

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	require.True(t, resp.Plan.Raw.Equal(plan))
}

func TestResource_modifyPlan_reformatOnlyDataKeepsPriorStateAndComputedValues(t *testing.T) {
	t.Parallel()

	state := ruleObject(t, attrValues{"artifacts": artifactsVal(t, `{"content":"# Steps"}`)})
	plan := ruleObject(t, attrValues{
		"artifacts":  artifactsVal(t, "{\n  \"content\":   \"# Steps\"\n}\n"),
		"version":    tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"updated_at": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})

	resp := modifyPlan(t, plan, state)

	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	require.True(t, resp.Plan.Raw.Equal(state), "plan should equal prior state, got %s", resp.Plan.Raw)
}

func TestResource_modifyPlan_realChangesKeepVersionUnknown(t *testing.T) {
	t.Parallel()

	state := ruleObject(t, attrValues{"artifacts": artifactsVal(t, `{"content":"# Steps"}`)})

	for name, tc := range map[string]struct {
		overrides attrValues
		wantData  string
	}{
		"different data keeps the planned data": {
			overrides: attrValues{"artifacts": artifactsVal(t, `{"content":"# Other steps"}`)},
			wantData:  `{"content":"# Other steps"}`,
		},
		"reformatted data with a renamed rule restores the prior data only": {
			overrides: attrValues{
				"artifacts": artifactsVal(t, "{ \"content\" : \"# Steps\" }"),
				"metadata":  objectVal(nestedObjectType(t, rootObjectType(t), "metadata"), attrValues{"name": strVal("renamed")}),
			},
			wantData: `{"content":"# Steps"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tc.overrides["version"] = tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)
			tc.overrides["updated_at"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			plan := ruleObject(t, tc.overrides)

			resp := modifyPlan(t, plan, state)

			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
			var planned Model
			require.False(t, resp.Plan.Get(context.Background(), &planned).HasError())
			require.Equal(t, tc.wantData, planned.Artifacts[0].Data.ValueString())
			require.True(t, planned.Version.IsUnknown(), "version must stay unknown for a real update")
		})
	}
}
