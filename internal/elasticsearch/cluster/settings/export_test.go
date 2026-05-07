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

// Package settings_test helpers: export_test.go lives in package settings
// so it can access unexported identifiers, but is only compiled during tests.

package settings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ExportedExpandSettings exposes expandSettings for white-box testing.
func ExportedExpandSettings(ctx context.Context, list types.List) (map[string]any, diag.Diagnostics) {
	return expandSettings(ctx, list)
}

// ExportedUpdateRemovedSettings exposes updateRemovedSettings for white-box testing.
func ExportedUpdateRemovedSettings(name string, oldSettings, newSettings, target map[string]any) {
	updateRemovedSettings(name, oldSettings, newSettings, target)
}

// ExportedFlattenSettings exposes flattenSettings for white-box testing.
func ExportedFlattenSettings(ctx context.Context, category string, configured, api map[string]any) (types.List, diag.Diagnostics) {
	return flattenSettings(ctx, category, configured, api)
}

// SettingItem is a simplified view of a settingModel for assertions in tests.
type SettingItem struct {
	Name      string
	Value     string
	HasValue  bool
	ValueList []string
}

// ExtractSettingsFromList decodes the block list returned by flattenSettings
// into a simple []SettingItem for easy assertions in tests.
func ExtractSettingsFromList(ctx context.Context, t testing.TB, list types.List) []SettingItem {
	t.Helper()

	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	var blocks []settingsBlockModel
	diags := list.ElementsAs(ctx, &blocks, false)
	if diags.HasError() {
		t.Fatalf("ExtractSettingsFromList: failed to decode block list: %v", diags)
	}
	if len(blocks) == 0 {
		return nil
	}

	var models []settingModel
	diags = blocks[0].Setting.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		t.Fatalf("ExtractSettingsFromList: failed to decode setting set: %v", diags)
	}

	items := make([]SettingItem, 0, len(models))
	for _, m := range models {
		item := SettingItem{Name: m.Name.ValueString()}
		if !m.Value.IsNull() && !m.Value.IsUnknown() {
			item.Value = m.Value.ValueString()
			item.HasValue = true
		}
		if !m.ValueList.IsNull() && !m.ValueList.IsUnknown() {
			var vals []string
			diags = m.ValueList.ElementsAs(ctx, &vals, false)
			if diags.HasError() {
				t.Fatalf("ExtractSettingsFromList: failed to decode value_list: %v", diags)
			}
			item.ValueList = vals
		}
		items = append(items, item)
	}
	return items
}

// MakeSettingsListWithValue builds a types.List containing one settingsBlockModel
// with a single setting using a scalar value.
func MakeSettingsListWithValue(name, value string) types.List {
	sm := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringValue(value),
		ValueList: types.ListNull(types.StringType),
	}
	return makeSettingsListFromModels([]settingModel{sm})
}

// MakeSettingsListWithValueList builds a types.List with a single setting using a list value.
func MakeSettingsListWithValueList(name string, vals []string) types.List {
	attrVals := make([]attr.Value, len(vals))
	for i, v := range vals {
		attrVals[i] = types.StringValue(v)
	}
	sm := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringNull(),
		ValueList: types.ListValueMust(types.StringType, attrVals),
	}
	return makeSettingsListFromModels([]settingModel{sm})
}

// MakeSettingsListWithDuplicateName builds a list with two settings sharing the same name.
func MakeSettingsListWithDuplicateName(name, v1, v2 string) types.List {
	s1 := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringValue(v1),
		ValueList: types.ListNull(types.StringType),
	}
	s2 := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringValue(v2),
		ValueList: types.ListNull(types.StringType),
	}
	return makeSettingsListFromModels([]settingModel{s1, s2})
}

// MakeSettingsListBothValues builds a list where both value and value_list are set.
func MakeSettingsListBothValues(name, value string, listVals []string) types.List {
	attrVals := make([]attr.Value, len(listVals))
	for i, v := range listVals {
		attrVals[i] = types.StringValue(v)
	}
	sm := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringValue(value),
		ValueList: types.ListValueMust(types.StringType, attrVals),
	}
	return makeSettingsListFromModels([]settingModel{sm})
}

// MakeSettingsListNeitherValue builds a list where neither value nor value_list is set.
func MakeSettingsListNeitherValue(name string) types.List {
	sm := settingModel{
		Name:      types.StringValue(name),
		Value:     types.StringNull(),
		ValueList: types.ListNull(types.StringType),
	}
	return makeSettingsListFromModels([]settingModel{sm})
}

// EmptySettingsList returns a types.List with zero elements.
func EmptySettingsList() types.List {
	return types.ListValueMust(settingsListElemType(), []attr.Value{})
}

func makeSettingsListFromModels(models []settingModel) types.List {
	ctx := context.Background()
	settingAttr := settingModelAttrTypes()
	settingVals := make([]attr.Value, len(models))
	for i, m := range models {
		obj, diags := types.ObjectValueFrom(ctx, settingAttr, m)
		if diags.HasError() {
			panic("makeSettingsListFromModels: failed to create object")
		}
		settingVals[i] = obj
	}

	settingSet, diags := types.SetValue(types.ObjectType{AttrTypes: settingAttr}, settingVals)
	if diags.HasError() {
		panic("makeSettingsListFromModels: failed to create set")
	}

	block := settingsBlockModel{Setting: settingSet}
	blockAttr := settingsBlockAttrTypes()
	blockObj, diags := types.ObjectValueFrom(ctx, blockAttr, block)
	if diags.HasError() {
		panic("makeSettingsListFromModels: failed to create block object")
	}

	list, diags := types.ListValue(settingsListElemType(), []attr.Value{blockObj})
	if diags.HasError() {
		panic("makeSettingsListFromModels: failed to create list")
	}
	return list
}
