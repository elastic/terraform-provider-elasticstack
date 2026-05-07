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

package template

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
	var model Model
	diags := state.Get(ctx, &model)
	require.False(t, diags.HasError(), "%s", diags)
}

func baseIndexTemplateState() map[string]any {
	return map[string]any{
		"id":             "cluster-uuid/my-template",
		"name":           "my-template",
		"index_patterns": []any{"logs-*"},
		"elasticsearch_connection": []any{
			map[string]any{"username": "u"},
		},
	}
}

func TestIndexTemplateUpgradeState_data_stream_path(t *testing.T) {
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
				_, ok := got["data_stream"]
				require.False(t, ok)
			},
		},
		{
			name: "explicit_null",
			patch: map[string]any{
				"data_stream": nil,
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				v, ok := got["data_stream"]
				require.True(t, ok)
				require.Nil(t, v)
			},
		},
		{
			name: "empty_list_to_null",
			patch: map[string]any{
				"data_stream": []any{},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				v, ok := got["data_stream"]
				require.True(t, ok)
				require.Nil(t, v)
			},
		},
		{
			name: "singleton_list_to_object",
			patch: map[string]any{
				"data_stream": []any{
					map[string]any{
						"hidden":               true,
						"allow_custom_routing": false,
					},
				},
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				ds, ok := got["data_stream"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, true, ds["hidden"])
				require.Equal(t, false, ds["allow_custom_routing"])
			},
		},
		{
			name: "multi_element_error",
			patch: map[string]any{
				"data_stream": []any{
					map[string]any{"hidden": true},
					map[string]any{"hidden": false},
				},
			},
			expectErr:   true,
			errContains: `unexpected multi-element array at path "data_stream"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseIndexTemplateState()
			maps.Copy(raw, tc.patch)
			if tc.name == "absent_unchanged" {
				delete(raw, "data_stream")
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

func TestIndexTemplateUpgradeState_template_lifecycle_path(t *testing.T) {
	t.Parallel()

	raw := baseIndexTemplateState()
	raw["template"] = []any{
		map[string]any{
			"lifecycle": []any{
				map[string]any{"data_retention": "30d"},
			},
			"mappings": `{"properties":{"a":{"type":"keyword"}}}`,
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	tmpl := got["template"].(map[string]any)
	mappings, ok := tmpl["mappings"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"properties":{"a":{"type":"keyword"}}}`, mappings)
	lc := tmpl["lifecycle"].(map[string]any)
	require.Equal(t, "30d", lc["data_retention"])
	requireUpgradedStateDecodes(t, resp)
}

func TestIndexTemplateUpgradeState_template_data_stream_options_path(t *testing.T) {
	t.Parallel()

	raw := baseIndexTemplateState()
	raw["template"] = []any{
		map[string]any{
			"data_stream_options": []any{
				map[string]any{
					"failure_store": []any{
						map[string]any{
							"enabled": true,
							"lifecycle": []any{
								map[string]any{"data_retention": "7d"},
							},
						},
					},
				},
			},
		},
	}

	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	tmpl := got["template"].(map[string]any)
	dso := tmpl["data_stream_options"].(map[string]any)
	fs := dso["failure_store"].(map[string]any)
	require.Equal(t, true, fs["enabled"])
	lc := fs["lifecycle"].(map[string]any)
	require.Equal(t, "7d", lc["data_retention"])
	requireUpgradedStateDecodes(t, resp)
}

func TestIndexTemplateUpgradeState_per_path_table(t *testing.T) {
	t.Parallel()

	type tc struct {
		name        string
		build       func(base map[string]any)
		errContains string
		assert      func(t *testing.T, got map[string]any)
	}

	cases := []tc{
		{
			name: "template_absent",
			build: func(base map[string]any) {
				delete(base, "template")
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				_, ok := got["template"]
				require.False(t, ok)
			},
		},
		{
			name: "template_null",
			build: func(base map[string]any) {
				base["template"] = nil
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Nil(t, got["template"])
			},
		},
		{
			name: "template_empty_list",
			build: func(base map[string]any) {
				base["template"] = []any{}
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				require.Nil(t, got["template"])
			},
		},
		{
			name: "template_multi",
			build: func(base map[string]any) {
				base["template"] = []any{map[string]any{}, map[string]any{}}
			},
			errContains: `unexpected multi-element array at path "template"`,
		},
		{
			name: "template_lifecycle_null_inside_template_object",
			build: func(base map[string]any) {
				base["template"] = map[string]any{
					"lifecycle": nil,
				}
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				tmpl := got["template"].(map[string]any)
				require.Nil(t, tmpl["lifecycle"])
			},
		},
		{
			name: "template_lifecycle_empty_list",
			build: func(base map[string]any) {
				base["template"] = []any{
					map[string]any{"lifecycle": []any{}},
				}
			},
			assert: func(t *testing.T, got map[string]any) {
				t.Helper()
				tmpl := got["template"].(map[string]any)
				require.Nil(t, tmpl["lifecycle"])
			},
		},
		{
			name: "template_lifecycle_multi",
			build: func(base map[string]any) {
				base["template"] = []any{
					map[string]any{
						"lifecycle": []any{
							map[string]any{"data_retention": "1d"},
							map[string]any{"data_retention": "2d"},
						},
					},
				}
			},
			errContains: `unexpected multi-element array at path "template.lifecycle"`,
		},
		{
			name: "template_data_stream_options_multi",
			build: func(base map[string]any) {
				base["template"] = []any{
					map[string]any{
						"data_stream_options": []any{
							map[string]any{},
							map[string]any{},
						},
					},
				}
			},
			errContains: `unexpected multi-element array at path "template.data_stream_options"`,
		},
		{
			name: "failure_store_multi",
			build: func(base map[string]any) {
				base["template"] = []any{
					map[string]any{
						"data_stream_options": []any{
							map[string]any{
								"failure_store": []any{
									map[string]any{"enabled": true},
									map[string]any{"enabled": false},
								},
							},
						},
					},
				}
			},
			errContains: `unexpected multi-element array at path "template.data_stream_options.failure_store"`,
		},
		{
			name: "failure_store_lifecycle_multi",
			build: func(base map[string]any) {
				base["template"] = []any{
					map[string]any{
						"data_stream_options": []any{
							map[string]any{
								"failure_store": []any{
									map[string]any{
										"enabled": true,
										"lifecycle": []any{
											map[string]any{"data_retention": "1d"},
											map[string]any{"data_retention": "2d"},
										},
									},
								},
							},
						},
					},
				}
			},
			errContains: `unexpected multi-element array at path "template.data_stream_options.failure_store.lifecycle"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := baseIndexTemplateState()
			tc.build(raw)
			resp := runUpgrade(t, raw)
			if tc.errContains != "" {
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

func TestIndexTemplateUpgradeState_drops_version_zero(t *testing.T) {
	t.Parallel()
	raw := baseIndexTemplateState()
	raw["version"] = float64(0)
	resp := runUpgrade(t, raw)
	require.False(t, resp.Diagnostics.HasError(), "%s", resp.Diagnostics)
	var got map[string]any
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &got))
	_, ok := got["version"]
	require.False(t, ok)
	requireUpgradedStateDecodes(t, resp)
}

func TestIndexTemplateUpgradeState_normalizes_alias_routing_echo(t *testing.T) {
	t.Parallel()
	raw := baseIndexTemplateState()
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

func TestIndexTemplateUpgradeState_preserves_non_collapsed_bytes(t *testing.T) {
	t.Parallel()

	meta := `{"team":"search"}`
	raw := baseIndexTemplateState()
	raw["metadata"] = meta
	raw["composed_of"] = []any{"ct1", "ct2"}
	raw["template"] = []any{
		map[string]any{
			"settings": `{"index":{"number_of_shards":"1"}}`,
			"alias": []any{
				map[string]any{
					"name":           "a1",
					"routing":        "r1",
					"index_routing":  "",
					"search_routing": "",
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
	require.Equal(t, meta, got["metadata"])
	require.Equal(t, []any{"ct1", "ct2"}, got["composed_of"])
	tmpl := got["template"].(map[string]any)
	settings, ok := tmpl["settings"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"index":{"number_of_shards":"1"}}`, settings)
	aliases, ok := tmpl["alias"].([]any)
	require.True(t, ok)
	require.Len(t, aliases, 1)
	requireUpgradedStateDecodes(t, resp)
}

func TestIndexTemplateUpgradeState_combined_v0_realistic(t *testing.T) {
	t.Parallel()

	raw := baseIndexTemplateState()
	raw["metadata"] = `{"k":"v"}`
	raw["data_stream"] = []any{
		map[string]any{"hidden": true, "allow_custom_routing": false},
	}
	raw["template"] = []any{
		map[string]any{
			"mappings": `{"properties":{"x":{"type":"text"}}}`,
			"settings": `{"index.number_of_replicas":"0"}`,
			"lifecycle": []any{
				map[string]any{"data_retention": "10d"},
			},
			"data_stream_options": []any{
				map[string]any{
					"failure_store": []any{
						map[string]any{
							"enabled": true,
							"lifecycle": []any{
								map[string]any{"data_retention": "30d"},
							},
						},
					},
				},
			},
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

	ds, ok := got["data_stream"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, ds["hidden"])

	tmpl, ok := got["template"].(map[string]any)
	require.True(t, ok)
	combinedMappings, ok := tmpl["mappings"].(string)
	require.True(t, ok)
	require.JSONEq(t, `{"properties":{"x":{"type":"text"}}}`, combinedMappings)

	lc, ok := tmpl["lifecycle"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "10d", lc["data_retention"])

	dso, ok := tmpl["data_stream_options"].(map[string]any)
	require.True(t, ok)
	fs, ok := dso["failure_store"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, fs["enabled"])
	fslc, ok := fs["lifecycle"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "30d", fslc["data_retention"])

	aliasList, ok := tmpl["alias"].([]any)
	require.True(t, ok)
	require.Len(t, aliasList, 1)

	requireUpgradedStateDecodes(t, resp)
}

func TestIndexTemplateUpgradeState_invalid_json(t *testing.T) {
	t.Parallel()

	r := newResource()
	up := r.UpgradeState(context.Background())[0]
	resp := &resource.UpgradeStateResponse{}
	up.StateUpgrader(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{`)},
	}, resp)
	require.True(t, resp.Diagnostics.HasError())
}
