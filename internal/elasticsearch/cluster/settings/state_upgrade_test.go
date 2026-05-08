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

package settings

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestMigrateClusterSettingsStateV0ToV1(t *testing.T) {
	// Mirrors the on-disk shape that the SDKv2 implementation produced: every
	// setting object carries both value and value_list, with the unused one
	// stored as an empty string / empty list rather than null.
	priorState := map[string]any{
		"id": "cluster/cluster-settings",
		"persistent": []any{
			map[string]any{
				"setting": []any{
					map[string]any{
						"name":       "indices.lifecycle.poll_interval",
						"value":      "10m",
						"value_list": []any{},
					},
					map[string]any{
						"name":       "xpack.security.audit.logfile.events.include",
						"value":      "",
						"value_list": []any{"ACCESS_DENIED", "ACCESS_GRANTED"},
					},
				},
			},
		},
		"transient": []any{
			map[string]any{
				"setting": []any{
					map[string]any{
						"name":       "indices.breaker.total.limit",
						"value":      "60%",
						"value_list": []any{},
					},
				},
			},
		},
	}

	priorJSON, err := json.Marshal(priorState)
	if err != nil {
		t.Fatalf("marshal prior state: %v", err)
	}

	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: priorJSON},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.DynamicValue == nil {
		t.Fatal("expected DynamicValue to be set")
	}

	var got map[string]any
	if err := json.Unmarshal(resp.DynamicValue.JSON, &got); err != nil {
		t.Fatalf("unmarshal upgraded state: %v", err)
	}

	want := map[string]any{
		"id": "cluster/cluster-settings",
		"persistent": map[string]any{
			"setting": []any{
				map[string]any{
					"name":  "indices.lifecycle.poll_interval",
					"value": "10m",
				},
				map[string]any{
					"name":       "xpack.security.audit.logfile.events.include",
					"value_list": []any{"ACCESS_DENIED", "ACCESS_GRANTED"},
				},
			},
		},
		"transient": map[string]any{
			"setting": []any{
				map[string]any{
					"name":  "indices.breaker.total.limit",
					"value": "60%",
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("upgraded state mismatch (-want +got):\n%s", diff)
	}
}

func TestMigrateClusterSettingsStateV0ToV1_NilRawState(t *testing.T) {
	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: nil,
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected diagnostics for nil raw state")
	}
}

func TestMigrateClusterSettingsStateV0ToV1_OnlyPersistent(t *testing.T) {
	priorState := map[string]any{
		"id": "cluster/cluster-settings",
		"persistent": []any{
			map[string]any{
				"setting": []any{
					map[string]any{"name": "k", "value": "v", "value_list": []any{}},
				},
			},
		},
	}
	priorJSON, _ := json.Marshal(priorState)

	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: priorJSON},
	}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got map[string]any
	_ = json.Unmarshal(resp.DynamicValue.JSON, &got)

	settings := got["persistent"].(map[string]any)["setting"].([]any)
	first := settings[0].(map[string]any)
	if _, ok := first["value_list"]; ok {
		t.Errorf("expected value_list key to be removed, got %v", first)
	}
	if _, ok := got["transient"]; ok {
		t.Errorf("expected transient to be absent, got %v", got["transient"])
	}
}

func TestMigrateClusterSettingsStateV0ToV1_EmptyCategoryListDropped(t *testing.T) {
	priorState := map[string]any{
		"id":         "cluster/cluster-settings",
		"persistent": []any{},
		"transient":  []any{map[string]any{"setting": []any{map[string]any{"name": "k", "value": "v", "value_list": []any{}}}}},
	}
	priorJSON, _ := json.Marshal(priorState)

	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: priorJSON},
	}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got map[string]any
	_ = json.Unmarshal(resp.DynamicValue.JSON, &got)

	if _, ok := got["persistent"]; ok {
		t.Errorf("expected persistent key to be removed for empty list, got %v", got["persistent"])
	}
	if _, ok := got["transient"].(map[string]any); !ok {
		t.Errorf("expected transient to be unwrapped to a map, got %T", got["transient"])
	}
}

func TestMigrateClusterSettingsStateV0ToV1_RejectsMultipleCategoryBlocks(t *testing.T) {
	priorState := map[string]any{
		"id": "cluster/cluster-settings",
		"persistent": []any{
			map[string]any{"setting": []any{}},
			map[string]any{"setting": []any{}},
		},
	}
	priorJSON, _ := json.Marshal(priorState)

	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: priorJSON},
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for multiple persistent blocks")
	}
}

func TestMigrateClusterSettingsStateV0ToV1_RejectsUnexpectedCategoryShape(t *testing.T) {
	priorState := map[string]any{
		"id":         "cluster/cluster-settings",
		"persistent": map[string]any{"setting": []any{}},
	}
	priorJSON, _ := json.Marshal(priorState)

	resp := &resource.UpgradeStateResponse{}
	migrateClusterSettingsStateV0ToV1(context.Background(), resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: priorJSON},
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for non-list persistent block")
	}
}
