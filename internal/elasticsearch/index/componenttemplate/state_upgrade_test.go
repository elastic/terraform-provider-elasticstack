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

package componenttemplate

import (
	"context"
	"encoding/json"
	"maps"
	"strings"
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
	var model Data
	diags := state.Get(ctx, &model)
	require.False(t, diags.HasError(), "%s", diags)
}

func baseComponentTemplateState() map[string]any {
	return map[string]any{
		"id":   "cluster-uuid/my-template",
		"name": "my-template",
		"elasticsearch_connection": []any{
			map[string]any{"username": "u"},
		},
	}
}

func TestComponentTemplateUpgradeState_template_path(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		patch       map[string]any
		assert      func(t *testing.T, got map[string]any)
		expectErr   bool
		errContains string
	}{
		{
			name:  "absent_unchanged",
			patch: map[string]any{},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				_, ok := got["template"]
				require.False(t, ok)
			},
		},
		{
			name: "explicit_null",
			patch: map[string]any{
				"template": nil,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				v, ok := got["template"]
				require.True(t, ok)
				require.Nil(t, v)
			},
		},
		{
			name: "empty_list_to_null",
			patch: map[string]any{
				"template": []any{},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				v, ok := got["template"]
				require.True(t, ok)
				require.Nil(t, v)
			},
		},
		{
			name: "singleton_list_to_object",
			patch: map[string]any{
				"template": []any{
					map[string]any{
						"mappings": `{"properties":{"a":{"type":"keyword"}}}`,
						"settings": `{"index":{"number_of_shards":"1"}}`,
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				tmpl, ok := got["template"].(map[string]any)
				require.True(t, ok)
				mappings, ok := tmpl["mappings"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"properties":{"a":{"type":"keyword"}}}`, mappings)
				settings, ok := tmpl["settings"].(string)
				require.True(t, ok)
				require.JSONEq(t, `{"index":{"number_of_shards":"1"}}`, settings)
			},
		},
		{
			name: "multi_element_error",
			patch: map[string]any{
				"template": []any{
					map[string]any{"mappings": `{}`},
					map[string]any{"mappings": `{}`},
				},
			},
			expectErr:   true,
			errContains: `unexpected multi-element array at path "template"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseComponentTemplateState()
			maps.Copy(raw, tc.patch)
			if tc.name == "absent_unchanged" {
				delete(raw, "template")
			}

			resp := runUpgrade(t, raw)
			if tc.expectErr {
				require.True(t, resp.Diagnostics.HasError())
				var b strings.Builder
				for _, d := range resp.Diagnostics.Errors() {
					b.WriteString(d.Detail())
				}
				require.Contains(t, b.String(), tc.errContains)
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
			var got map[string]any
			require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
			tc.assert(t, got)
			requireUpgradedStateDecodes(t, resp)
		})
	}
}

func TestComponentTemplateUpgradeState_drops_version_zero(t *testing.T) {
	t.Parallel()
	raw := baseComponentTemplateState()
	raw["version"] = float64(0)
	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	_, ok := got["version"]
	require.False(t, ok)
	requireUpgradedStateDecodes(t, resp)
}

func TestComponentTemplateUpgradeState_normalizes_alias_routing_echo(t *testing.T) {
	t.Parallel()
	raw := baseComponentTemplateState()
	raw["template"] = []any{
		map[string]any{
			"alias": []any{
				map[string]any{
					"name":           "routing_only_alias",
					"routing":        "",
					"index_routing":  "shard-a",
					"search_routing": "shard-a",
					"is_hidden":      false,
					"is_write_index": false,
					"filter":         nil,
				},
			},
		},
	}
	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	tmpl := got["template"].(map[string]any)
	aliases := tmpl["alias"].([]any)
	am := aliases[0].(map[string]any)
	require.Equal(t, "shard-a", am["routing"])
	require.Empty(t, am["index_routing"])
	require.Empty(t, am["search_routing"])
	requireUpgradedStateDecodes(t, resp)
}

func TestComponentTemplateUpgradeState_combined_v0_realistic(t *testing.T) {
	t.Parallel()

	raw := baseComponentTemplateState()
	raw["metadata"] = `{"k":"v"}`
	raw["version"] = float64(3)
	raw["template"] = []any{
		map[string]any{
			"mappings": `{"properties":{"x":{"type":"text"}}}`,
			"settings": `{"index.number_of_replicas":"0"}`,
			"alias": []any{
				map[string]any{
					"name":           "primary",
					"is_write_index": true,
				},
			},
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))

	metadata, ok := got["metadata"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"k":"v"}`, metadata)
	require.InEpsilon(t, 3.0, got["version"], 0.0001)

	tmpl, ok := got["template"].(map[string]any)
	require.True(t, ok)
	mappings, ok := tmpl["mappings"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"properties":{"x":{"type":"text"}}}`, mappings)
	settings, ok := tmpl["settings"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"index.number_of_replicas":"0"}`, settings)

	aliasList, ok := tmpl["alias"].([]any)
	require.True(t, ok)
	require.Len(t, aliasList, 1)

	requireUpgradedStateDecodes(t, resp)
}

func TestComponentTemplateUpgradeState_invalid_json(t *testing.T) {
	t.Parallel()

	r := newResource()
	up := r.UpgradeState(context.Background())[0]
	resp := &resource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{`)},
	}, resp)
	require.True(t, resp.Diagnostics.HasError())
}
