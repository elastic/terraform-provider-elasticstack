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
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster/settings"
	"github.com/google/go-cmp/cmp"
)

// helpers_test.go provides white-box access via the exported test helpers.
// The non-exported functions are exercised through the exported wrappers
// (ExportedExpandSettings, etc.) defined in export_test.go.

func TestExpandSettings_StringValue(t *testing.T) {
	ctx := context.Background()
	settingsList := settings.MakeSettingsListWithValue("mykey", "myval")

	result, diags := settings.ExportedExpandSettings(ctx, settingsList)
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
	ctx := context.Background()
	settingsList := settings.MakeSettingsListWithValueList("listkey", []string{"a", "b"})

	result, diags := settings.ExportedExpandSettings(ctx, settingsList)
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

func TestExpandSettings_EmptyList(t *testing.T) {
	ctx := context.Background()
	settingsList := settings.EmptySettingsList()

	result, diags := settings.ExportedExpandSettings(ctx, settingsList)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result != nil {
		t.Errorf("expected nil for empty list, got %v", result)
	}
}

func TestExpandSettings_DuplicateNameError(t *testing.T) {
	ctx := context.Background()
	settingsList := settings.MakeSettingsListWithDuplicateName("dupkey", "v1", "v2")

	_, diags := settings.ExportedExpandSettings(ctx, settingsList)
	if !diags.HasError() {
		t.Error("expected error for duplicate setting name")
	}
}

func TestExpandSettings_BothValueAndValueList_Error(t *testing.T) {
	ctx := context.Background()
	settingsList := settings.MakeSettingsListBothValues("key", "v", []string{"a"})

	_, diags := settings.ExportedExpandSettings(ctx, settingsList)
	if !diags.HasError() {
		t.Error("expected error when both value and value_list are set")
	}
}

func TestExpandSettings_NeitherValueNorValueList_Error(t *testing.T) {
	ctx := context.Background()
	settingsList := settings.MakeSettingsListNeitherValue("key")

	_, diags := settings.ExportedExpandSettings(ctx, settingsList)
	if !diags.HasError() {
		t.Error("expected error when neither value nor value_list is set")
	}
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
	ctx := context.Background()
	configured := map[string]any{"persistent": map[string]any{"mykey": "anything"}}
	api := map[string]any{"persistent": map[string]any{"mykey": "40mb"}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := settings.ExtractSettingsFromList(ctx, t, result)
	if len(items) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(items))
	}
	if items[0].Name != "mykey" {
		t.Errorf("expected name=mykey, got %s", items[0].Name)
	}
	if items[0].Value != "40mb" {
		t.Errorf("expected value=40mb, got %s", items[0].Value)
	}
	if items[0].ValueList != nil {
		t.Errorf("expected null value_list, got %v", items[0].ValueList)
	}
}

func TestFlattenSettings_ListValue(t *testing.T) {
	ctx := context.Background()
	configured := map[string]any{"transient": map[string]any{"listkey": "anything"}}
	api := map[string]any{"transient": map[string]any{"listkey": []any{"x", "y"}}}

	result, diags := settings.ExportedFlattenSettings(ctx, "transient", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	items := settings.ExtractSettingsFromList(ctx, t, result)
	if len(items) != 1 {
		t.Fatalf("expected 1 setting, got %d", len(items))
	}
	if items[0].Name != "listkey" {
		t.Errorf("expected name=listkey, got %s", items[0].Name)
	}
	if items[0].HasValue {
		t.Errorf("expected null value, got %s", items[0].Value)
	}
	if diff := cmp.Diff([]string{"x", "y"}, items[0].ValueList); diff != "" {
		t.Errorf("value_list mismatch (-want +got):\n%s", diff)
	}
}

func TestFlattenSettings_AbsentFromAPIIsOmitted(t *testing.T) {
	ctx := context.Background()
	configured := map[string]any{"persistent": map[string]any{"gone": "v"}}
	api := map[string]any{"persistent": map[string]any{}} // key not in API response

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	// Result should be an empty list block
	if len(result.Elements()) != 0 {
		t.Errorf("expected empty list when key absent from API, got %d elements", len(result.Elements()))
	}
}

func TestFlattenSettings_EmptyConfigured(t *testing.T) {
	ctx := context.Background()
	configured := map[string]any{}
	api := map[string]any{"persistent": map[string]any{"somekey": "v"}}

	result, diags := settings.ExportedFlattenSettings(ctx, "persistent", configured, api)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result.Elements()) != 0 {
		t.Errorf("expected empty list for unconfigured category, got %d elements", len(result.Elements()))
	}
}
