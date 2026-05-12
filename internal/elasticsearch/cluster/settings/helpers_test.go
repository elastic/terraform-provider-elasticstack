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

package settings_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/settings"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateConfigModel_BothEmpty_Error(t *testing.T) {
	diags := settings.ExportedValidateConfigModel(
		mustEmptySettingsBlock(t),
		mustEmptySettingsBlock(t),
	)
	if !diags.HasError() {
		t.Error("expected error when both persistent and transient are empty")
	}
}

func TestValidateConfigModel_BothNull_Error(t *testing.T) {
	diags := settings.ExportedValidateConfigModel(
		settings.NullSettingsBlock(),
		settings.NullSettingsBlock(),
	)
	if !diags.HasError() {
		t.Error("expected error when both persistent and transient are null")
	}
}

func TestValidateConfigModel_PersistentSet_OK(t *testing.T) {
	diags := settings.ExportedValidateConfigModel(
		settings.MakeSettingsBlockWithValue("k", "v"),
		settings.NullSettingsBlock(),
	)
	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags)
	}
}

func TestValidateConfigModel_TransientSet_OK(t *testing.T) {
	diags := settings.ExportedValidateConfigModel(
		settings.NullSettingsBlock(),
		settings.MakeSettingsBlockWithValue("k", "v"),
	)
	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags)
	}
}

func TestExpandSettings_StringValue(t *testing.T) {
	ctx := t.Context()
	block := settings.MakeSettingsBlockWithValue("mykey", "myval")

	result, diags := settings.ExportedExpandSettings(ctx, block)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if got := result["mykey"]; got != "myval" {
		t.Errorf("expected mykey=myval, got %v", got)
	}
}

func TestExpandSettings_ListValue(t *testing.T) {
	ctx := t.Context()
	block := settings.MakeSettingsBlockWithValueList("listkey", []string{"a", "b"})

	result, diags := settings.ExportedExpandSettings(ctx, block)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	got, ok := result["listkey"].([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result["listkey"])
	}
	want := []any{"a", "b"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("list value mismatch (-want +got):\n%s", diff)
	}
}

func TestExpandSettings_NullBlock(t *testing.T) {
	ctx := t.Context()
	result, diags := settings.ExportedExpandSettings(ctx, settings.NullSettingsBlock())
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result != nil {
		t.Errorf("expected nil for null block, got %v", result)
	}
}

func TestExpandSettings_EmptyBlock(t *testing.T) {
	ctx := t.Context()
	result, diags := settings.ExportedExpandSettings(ctx, mustEmptySettingsBlock(t))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result != nil {
		t.Errorf("expected nil for empty block, got %v", result)
	}
}

func mustEmptySettingsBlock(t *testing.T) types.Object {
	t.Helper()

	block, diags := settings.EmptySettingsBlock()
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating empty settings block: %v", diags)
	}
	return block
}

func TestUpdateRemovedSettings_RemovesDeletedKeys(t *testing.T) {
	oldSettings := map[string]any{"keep": "v1", "remove": "v2"}
	newSettings := map[string]any{"keep": "v1"}
	target := map[string]any{
		"persistent": map[string]any{"keep": "v1"},
	}

	settings.ExportedUpdateRemovedSettings("persistent", oldSettings, newSettings, target)

	cat := target["persistent"].(map[string]any)
	if _, ok := cat["remove"]; !ok {
		t.Error("expected 'remove' key to be nulled in target")
	}
	if cat["remove"] != nil {
		t.Errorf("expected nil value for removed key, got %v", cat["remove"])
	}
}

func TestUpdateRemovedSettings_EqualSetsNoChange(t *testing.T) {
	oldSettings := map[string]any{"a": "1"}
	newSettings := map[string]any{"a": "1"}
	target := map[string]any{}

	settings.ExportedUpdateRemovedSettings("persistent", oldSettings, newSettings, target)

	if len(target) != 0 {
		t.Errorf("expected no changes for equal settings, target=%v", target)
	}
}

func TestFlattenSettings_ScalarValue(t *testing.T) {
	ctx := t.Context()
	configured := map[string]any{"persistent": map[string]any{"mykey": "anything"}}
	api := map[string]any{"persistent": map[string]any{"mykey": "40mb"}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := settings.ExtractSettingsFromBlock(ctx, t, result)
	if len(items) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(items))
	}
	if items[0].Name != "mykey" {
		t.Errorf("expected name=mykey, got %s", items[0].Name)
	}
	if items[0].Value != "40mb" {
		t.Errorf("expected value=40mb, got %s", items[0].Value)
	}
	if len(items[0].ValueList) != 0 {
		t.Errorf("expected empty value_list, got %v", items[0].ValueList)
	}
}

func TestFlattenSettings_ListValue(t *testing.T) {
	ctx := t.Context()
	configured := map[string]any{"transient": map[string]any{"listkey": "anything"}}
	api := map[string]any{"transient": map[string]any{"listkey": []any{"x", "y"}}}

	result, diags := settings.ExportedFlattenSettings(ctx, "transient", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := settings.ExtractSettingsFromBlock(ctx, t, result)
	if len(items) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(items))
	}
	if items[0].Name != "listkey" {
		t.Errorf("expected name=listkey, got %s", items[0].Name)
	}
	if items[0].Value != "" {
		t.Errorf("expected empty value, got %s", items[0].Value)
	}
	if diff := cmp.Diff([]string{"x", "y"}, items[0].ValueList); diff != "" {
		t.Errorf("value_list mismatch (-want +got):\n%s", diff)
	}
}

// TestFlattenSettings_NonStringScalar covers the default branch of the type
// switch in flattenSettings: any scalar value the API returns that is neither
// a string nor a []any (for example a JSON number or boolean) must be
// surfaced through "value" via fmt.Sprintf.
func TestFlattenSettings_NonStringScalar(t *testing.T) {
	ctx := t.Context()
	configured := map[string]any{"persistent": map[string]any{
		"numeric": "anything",
		"boolish": "anything",
	}}
	api := map[string]any{"persistent": map[string]any{
		"numeric": float64(42),
		"boolish": true,
	}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := settings.ExtractSettingsFromBlock(ctx, t, result)
	if len(items) != 2 {
		t.Fatalf("expected 2 settings, got %d", len(items))
	}

	got := map[string]string{}
	for _, it := range items {
		got[it.Name] = it.Value
		if len(it.ValueList) != 0 {
			t.Errorf("expected empty value_list for %s, got %v", it.Name, it.ValueList)
		}
	}
	if got["numeric"] != "42" {
		t.Errorf("expected numeric=42, got %q", got["numeric"])
	}
	if got["boolish"] != "true" {
		t.Errorf("expected boolish=true, got %q", got["boolish"])
	}
}

func TestFlattenSettings_AbsentFromAPIIsOmitted(t *testing.T) {
	ctx := t.Context()
	configured := map[string]any{"persistent": map[string]any{"gone": "v"}}
	api := map[string]any{"persistent": map[string]any{}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Result should be a non-null block with an empty setting set.
	items := settings.ExtractSettingsFromBlock(ctx, t, result)
	if len(items) != 0 {
		t.Errorf("expected 0 settings when key absent from API, got %d", len(items))
	}
}

func TestFlattenSettings_EmptyConfiguredReturnsNull(t *testing.T) {
	ctx := t.Context()
	configured := map[string]any{}
	api := map[string]any{"persistent": map[string]any{"somekey": "v"}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !result.IsNull() {
		t.Errorf("expected null block for unconfigured category, got %v", result)
	}
}
