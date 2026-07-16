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

package ilm

import (
	"context"
	"encoding/json"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func testResourceSchema(t *testing.T) schema.Schema {
	t.Helper()
	ctx := context.Background()
	var resp resource.SchemaResponse
	newResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp.Schema
}

func runUpgrade(t *testing.T, raw map[string]any) *resource.UpgradeStateResponse {
	t.Helper()
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)

	r := newResource()
	upgraders := r.UpgradeState(context.Background())
	up, ok := upgraders[0]
	require.True(t, ok)

	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: rawJSON},
	}
	resp := &resource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), req, resp)
	return resp
}

func requireUpgradedStateDecodes(t *testing.T, resp *resource.UpgradeStateResponse) {
	t.Helper()
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)
	require.NotNil(t, resp.DynamicValue.JSON)

	ctx := context.Background()
	sch := testResourceSchema(t)
	tfTyp := sch.Type().TerraformType(ctx)
	raw, err := resp.DynamicValue.Unmarshal(tfTyp)
	require.NoError(t, err)

	state := tfsdk.State{Schema: sch, Raw: raw}
	var model tfModel
	diags := state.Get(ctx, &model)
	require.False(t, diags.HasError(), "%s", diags)
}

func baseILMState() map[string]any {
	return map[string]any{
		"id":   "cluster-uuid/policy-x",
		"name": "policy-x",
		"elasticsearch_connection": []any{
			map[string]any{"username": "u"},
		},
	}
}

func TestILMResourceUpgradeState(t *testing.T) {
	t.Parallel()

	raw := baseILMState()
	raw["hot"] = []any{
		map[string]any{
			"min_age": "1h",
			"set_priority": []any{
				map[string]any{"priority": float64(10)},
			},
			"rollover": []any{
				map[string]any{"max_age": "1d"},
			},
			"readonly": []any{
				map[string]any{"enabled": true},
			},
		},
	}
	raw["delete"] = []any{
		map[string]any{
			"min_age": "0ms",
			"delete": []any{
				map[string]any{"delete_searchable_snapshot": true},
			},
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)

	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	hot, ok := got["hot"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "1h", hot["min_age"])
	sp, ok := hot["set_priority"].(map[string]any)
	require.True(t, ok)
	priority, ok := sp["priority"].(float64)
	require.True(t, ok)
	require.InEpsilon(t, 10.0, priority, 0.0001)
	_, listSP := hot["set_priority"].([]any)
	require.False(t, listSP)

	del, ok := got["delete"].(map[string]any)
	require.True(t, ok)
	innerDel, ok := del["delete"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, innerDel["delete_searchable_snapshot"])

	conn, ok := got["elasticsearch_connection"].([]any)
	require.True(t, ok)
	require.Len(t, conn, 1)

	requireUpgradedStateDecodes(t, resp)
}

func TestMigrateILMStateV0ToV1_nullifyEmptyStrings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		patch  map[string]any
		assert func(t *testing.T, got map[string]any)
	}{
		{
			name: "metadata_empty_string",
			patch: map[string]any{
				"metadata": "",
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Nil(t, got["metadata"])
			},
		},
		{
			name: "allocate_include_empty_string",
			patch: map[string]any{
				"warm": []any{
					map[string]any{
						"min_age": "7d",
						"allocate": []any{
							map[string]any{
								"include":            "",
								"number_of_replicas": float64(1),
							},
						},
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				warm, ok := got["warm"].(map[string]any)
				require.True(t, ok)
				allocate, ok := warm["allocate"].(map[string]any)
				require.True(t, ok)
				require.Nil(t, allocate["include"])
				replicas, ok := allocate["number_of_replicas"].(float64)
				require.True(t, ok)
				require.InEpsilon(t, 1.0, replicas, 0.0001)
				_, hasExclude := allocate["exclude"]
				require.False(t, hasExclude)
				_, hasRequire := allocate["require"]
				require.False(t, hasRequire)
			},
		},
		{
			name: "allocate_all_json_attrs_empty_string",
			patch: map[string]any{
				"warm": []any{
					map[string]any{
						"min_age": "7d",
						"allocate": []any{
							map[string]any{
								"include":            "",
								"exclude":            "",
								"require":            "",
								"number_of_replicas": float64(2),
							},
						},
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				warm, ok := got["warm"].(map[string]any)
				require.True(t, ok)
				allocate, ok := warm["allocate"].(map[string]any)
				require.True(t, ok)
				require.Nil(t, allocate["include"])
				require.Nil(t, allocate["exclude"])
				require.Nil(t, allocate["require"])
			},
		},
		{
			name: "metadata_and_allocate_empty_strings",
			patch: map[string]any{
				"metadata": "",
				"cold": []any{
					map[string]any{
						"min_age": "30d",
						"allocate": []any{
							map[string]any{
								"include": "",
								"exclude": "",
								"require": "",
							},
						},
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Nil(t, got["metadata"])
				cold, ok := got["cold"].(map[string]any)
				require.True(t, ok)
				allocate, ok := cold["allocate"].(map[string]any)
				require.True(t, ok)
				require.Nil(t, allocate["include"])
				require.Nil(t, allocate["exclude"])
				require.Nil(t, allocate["require"])
			},
		},
		{
			name: "non_empty_json_strings_preserved",
			patch: map[string]any{
				"metadata": `{"k":"v"}`,
				"warm": []any{
					map[string]any{
						"min_age": "7d",
						"allocate": []any{
							map[string]any{
								"include": `{"box_type":"warm"}`,
								"exclude": `{"box_type":"cold"}`,
								"require": `{"data":"hot"}`,
							},
						},
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				metadata, ok := got["metadata"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"k":"v"}`, metadata)
				warm, ok := got["warm"].(map[string]any)
				require.True(t, ok)
				allocate, ok := warm["allocate"].(map[string]any)
				require.True(t, ok)
				include, ok := allocate["include"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"box_type":"warm"}`, include)
				exclude, ok := allocate["exclude"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"box_type":"cold"}`, exclude)
				requireVal, ok := allocate["require"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"data":"hot"}`, requireVal)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseILMState()
			maps.Copy(raw, tc.patch)

			resp := runUpgrade(t, raw)
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			var got map[string]any
			require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
			tc.assert(t, got)
			requireUpgradedStateDecodes(t, resp)
		})
	}
}
