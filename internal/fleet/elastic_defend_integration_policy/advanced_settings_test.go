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

package elasticdefendintegrationpolicy_test

import (
	"context"
	"testing"

	edip "github.com/elastic/terraform-provider-elasticstack/internal/fleet/elastic_defend_integration_policy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateAdvancedSettingKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	valid, diags := edip.AdvancedSettingsMapFromTerraform(ctx, types.MapValueMust(types.StringType, map[string]attr.Value{
		"linux.advanced.artifacts.global.base_url": types.StringValue("http://mirror.example.com"),
	}))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if valid["linux.advanced.artifacts.global.base_url"] != "http://mirror.example.com" {
		t.Fatalf("unexpected value: %v", valid)
	}

	_, diags = edip.AdvancedSettingsMapFromTerraform(ctx, types.MapValueMust(types.StringType, map[string]attr.Value{
		"artifacts.global.base_url": types.StringValue("http://mirror.example.com"),
	}))
	if !diags.HasError() {
		t.Fatal("expected invalid key diagnostic")
	}
}

func TestAdvancedSettingsRoundTrip(t *testing.T) {
	t.Parallel()

	policyData := map[string]any{
		"linux": map[string]any{
			"advanced": map[string]any{
				"artifacts": map[string]any{
					"global": map[string]any{
						"base_url": "http://mirror.example.com",
						"interval": "7200",
					},
				},
			},
		},
		"windows": map[string]any{
			"advanced": map[string]any{
				"artifacts": map[string]any{
					"global": map[string]any{
						"base_url": "http://win-mirror.example.com",
					},
				},
			},
		},
	}

	flat := edip.AdvancedSettingsFromPolicyData(policyData)
	if flat["linux.advanced.artifacts.global.base_url"] != "http://mirror.example.com" {
		t.Fatalf("unexpected linux base_url: %v", flat)
	}
	if flat["linux.advanced.artifacts.global.interval"] != "7200" {
		t.Fatalf("unexpected linux interval: %v", flat)
	}
	if flat["windows.advanced.artifacts.global.base_url"] != "http://win-mirror.example.com" {
		t.Fatalf("unexpected windows base_url: %v", flat)
	}

	nested := map[string]any{}
	edip.MergeAdvancedSettingsIntoPolicy(nested, flat, nil)

	linux := nested["linux"].(map[string]any)
	advanced := linux["advanced"].(map[string]any)
	artifacts := advanced["artifacts"].(map[string]any)
	global := artifacts["global"].(map[string]any)
	if global["base_url"] != "http://mirror.example.com" {
		t.Fatalf("unexpected nested base_url: %v", global["base_url"])
	}
	if global["interval"] != "7200" {
		t.Fatalf("unexpected nested interval: %v", global["interval"])
	}
}

func TestMergeAdvancedSettingsClearsPriorOS(t *testing.T) {
	t.Parallel()

	policy := map[string]any{
		"linux": map[string]any{
			"malware": map[string]any{"mode": "prevent"},
		},
	}

	prior := map[string]string{
		"linux.advanced.artifacts.global.base_url": "http://old.example.com",
	}

	edip.MergeAdvancedSettingsIntoPolicy(policy, map[string]string{}, prior)

	linux := policy["linux"].(map[string]any)
	advanced := linux["advanced"].(map[string]any)
	if len(advanced) != 0 {
		t.Fatalf("expected empty advanced map, got %v", advanced)
	}
	if linux["malware"].(map[string]any)["mode"] != "prevent" {
		t.Fatalf("expected malware mode preserved")
	}
}

func TestBuildPolicyPayloadIncludesAdvancedSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	policyObj := types.ObjectValueMust(map[string]attr.Type{
		"windows": types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"mac":     types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"linux": types.ObjectType{AttrTypes: map[string]attr.Type{
			"events":              types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"malware":             types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"memory_protection":   types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"behavior_protection": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"popup":               types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"logging":             types.ObjectType{AttrTypes: map[string]attr.Type{}},
		}},
	}, map[string]attr.Value{
		"windows": types.ObjectNull(map[string]attr.Type{}),
		"mac":     types.ObjectNull(map[string]attr.Type{}),
		"linux": types.ObjectValueMust(map[string]attr.Type{
			"events":              types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"malware":             types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"memory_protection":   types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"behavior_protection": types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"popup":               types.ObjectType{AttrTypes: map[string]attr.Type{}},
			"logging":             types.ObjectType{AttrTypes: map[string]attr.Type{}},
		}, map[string]attr.Value{
			"events":              types.ObjectNull(map[string]attr.Type{}),
			"malware":             types.ObjectNull(map[string]attr.Type{}),
			"memory_protection":   types.ObjectNull(map[string]attr.Type{}),
			"behavior_protection": types.ObjectNull(map[string]attr.Type{}),
			"popup":               types.ObjectNull(map[string]attr.Type{}),
			"logging":             types.ObjectNull(map[string]attr.Type{}),
		}),
	})

	model := &edip.ElasticDefendIntegrationPolicyModel{
		Policy: policyObj,
		AdvancedSettings: types.MapValueMust(types.StringType, map[string]attr.Value{
			"linux.advanced.artifacts.global.base_url": types.StringValue("http://10.0.0.33"),
		}),
	}

	payload, diags := edip.BuildPolicyPayload(ctx, model, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	linux := payload["linux"].(map[string]any)
	advanced := linux["advanced"].(map[string]any)
	global := advanced["artifacts"].(map[string]any)["global"].(map[string]any)
	if global["base_url"] != "http://10.0.0.33" {
		t.Fatalf("unexpected base_url: %v", global["base_url"])
	}
}
