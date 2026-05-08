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

// Package settings helpers: export_test.go lives in package settings
// so it can access unexported identifiers, but is only compiled during tests.

package settings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ExportedExpandSettings exposes expandSettings for white-box testing.
func ExportedExpandSettings(ctx context.Context, block types.Object) (map[string]any, diag.Diagnostics) {
	return expandSettings(ctx, block)
}

// ExportedUpdateRemovedSettings exposes updateRemovedSettings for white-box testing.
func ExportedUpdateRemovedSettings(name string, oldSettings, newSettings, target map[string]any) {
	updateRemovedSettings(name, oldSettings, newSettings, target)
}

// ExportedFlattenSettings exposes flattenSettings for white-box testing.
func ExportedFlattenSettings(ctx context.Context, category string, configured, api map[string]any) (types.Object, diag.Diagnostics) {
	return flattenSettings(ctx, category, configured, api)
}

// ExportedValidateConfigModel exposes validateConfigModel for white-box testing
// of the ValidateConfig rule without needing to construct a tfsdk.Config.
func ExportedValidateConfigModel(persistent, transient types.Object) diag.Diagnostics {
	return validateConfigModel(tfModel{Persistent: persistent, Transient: transient})
}

// SettingItem is a simplified view of a settingModel for assertions in tests.
type SettingItem struct {
	Name      string
	Value     string
	ValueList []string
}

// ExtractSettingsFromBlock decodes a SingleNestedBlock object returned by
// flattenSettings into a simple []SettingItem for easy assertions in tests.
func ExtractSettingsFromBlock(ctx context.Context, t testing.TB, block types.Object) []SettingItem {
	t.Helper()

	if block.IsNull() || block.IsUnknown() {
		return nil
	}

	var blockModel settingsBlockModel
	diags := block.As(ctx, &blockModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("ExtractSettingsFromBlock: failed to decode block object: %v", diags)
	}

	if blockModel.Setting.IsNull() || blockModel.Setting.IsUnknown() {
		return nil
	}

	var models []settingModel
	diags = blockModel.Setting.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		t.Fatalf("ExtractSettingsFromBlock: failed to decode setting set: %v", diags)
	}

	items := make([]SettingItem, 0, len(models))
	for _, m := range models {
		item := SettingItem{Name: m.Name.ValueString(), Value: m.Value.ValueString()}
		if !m.ValueList.IsNull() && !m.ValueList.IsUnknown() {
			var vals []string
			diags = m.ValueList.ElementsAs(ctx, &vals, false)
			if diags.HasError() {
				t.Fatalf("ExtractSettingsFromBlock: failed to decode value_list: %v", diags)
			}
			item.ValueList = vals
		}
		items = append(items, item)
	}
	return items
}

// MakeSettingsBlockWithValue builds a non-null block object with a single
// setting using a scalar value.
func MakeSettingsBlockWithValue(name, value string) types.Object {
	return makeBlock([]settingModel{{
		Name:      types.StringValue(name),
		Value:     types.StringValue(value),
		ValueList: types.ListNull(types.StringType),
	}})
}

// MakeSettingsBlockWithValueList builds a non-null block object with a single
// setting using a list value.
func MakeSettingsBlockWithValueList(name string, vals []string) types.Object {
	attrVals := make([]attr.Value, len(vals))
	for i, v := range vals {
		attrVals[i] = types.StringValue(v)
	}
	return makeBlock([]settingModel{{
		Name:      types.StringValue(name),
		Value:     types.StringNull(),
		ValueList: types.ListValueMust(types.StringType, attrVals),
	}})
}

// EmptySettingsBlock returns a non-null block object with zero settings.
func EmptySettingsBlock() (types.Object, diag.Diagnostics) {
	return emptySettingsBlock()
}

// NullSettingsBlock returns a typed null block object.
func NullSettingsBlock() types.Object {
	return nullSettingsBlock()
}

func makeBlock(models []settingModel) types.Object {
	ctx := context.Background()
	settingAttr := settingModelAttrTypes()
	settingVals := make([]attr.Value, len(models))
	for i, m := range models {
		obj, diags := types.ObjectValueFrom(ctx, settingAttr, m)
		if diags.HasError() {
			panic("makeBlock: failed to create object")
		}
		settingVals[i] = obj
	}
	settingSet, diags := types.SetValue(types.ObjectType{AttrTypes: settingAttr}, settingVals)
	if diags.HasError() {
		panic("makeBlock: failed to create set")
	}
	block := settingsBlockModel{Setting: settingSet}
	blockObj, diags := types.ObjectValueFrom(ctx, settingsBlockAttrTypes(), block)
	if diags.HasError() {
		panic("makeBlock: failed to create block object")
	}
	return blockObj
}
