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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// settingModelAttrTypes returns the attr.Type map for settingModel. Used when
// constructing types.Set or types.Object from settingModel values.
func settingModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":       types.StringType,
		"value":      types.StringType,
		"value_list": types.ListType{ElemType: types.StringType},
	}
}

// settingsBlockAttrTypes returns the attr.Type map for the persistent /
// transient single nested block.
func settingsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"setting": types.SetType{ElemType: types.ObjectType{AttrTypes: settingModelAttrTypes()}},
	}
}

// emptySettingsBlock returns a non-null Object with an empty setting set,
// used when the API has no values for a category we are tracking.
func emptySettingsBlock() (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	emptySet, ds := types.SetValue(types.ObjectType{AttrTypes: settingModelAttrTypes()}, []attr.Value{})
	diags.Append(ds...)
	if diags.HasError() {
		return nullSettingsBlock(), diags
	}

	obj, ds := types.ObjectValue(settingsBlockAttrTypes(), map[string]attr.Value{
		"setting": emptySet,
	})
	diags.Append(ds...)
	if diags.HasError() {
		return nullSettingsBlock(), diags
	}
	return obj, diags
}

// nullSettingsBlock returns a typed null Object, used to represent the
// "category not configured" case in state.
func nullSettingsBlock() types.Object {
	return types.ObjectNull(settingsBlockAttrTypes())
}

// expandSettings converts a SingleNestedBlock object representing a category
// (persistent/transient) into the flat settings map used by the Elasticsearch
// API helpers (map[name]any). Returns nil when the object is null, unknown,
// or has no setting elements.
func expandSettings(ctx context.Context, block types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if block.IsNull() || block.IsUnknown() {
		return nil, diags
	}

	var blockModel settingsBlockModel
	diags.Append(block.As(ctx, &blockModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	if blockModel.Setting.IsNull() || blockModel.Setting.IsUnknown() || len(blockModel.Setting.Elements()) == 0 {
		return nil, diags
	}

	var settingModels []settingModel
	diags.Append(blockModel.Setting.ElementsAs(ctx, &settingModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make(map[string]any, len(settingModels))
	for _, s := range settingModels {
		name := s.Name.ValueString()
		hasValue := !s.Value.IsNull() && !s.Value.IsUnknown() && s.Value.ValueString() != ""
		hasValueList := !s.ValueList.IsNull() && !s.ValueList.IsUnknown() && len(s.ValueList.Elements()) > 0

		if hasValue {
			result[name] = s.Value.ValueString()
			continue
		}
		if hasValueList {
			var vals []string
			diags.Append(s.ValueList.ElementsAs(ctx, &vals, false)...)
			if diags.HasError() {
				return nil, diags
			}
			valsAny := make([]any, len(vals))
			for i, v := range vals {
				valsAny[i] = v
			}
			result[name] = valsAny
		}
	}

	return result, diags
}

// getConfiguredSettings returns the combined flat settings map for both persistent
// and transient categories from the model.
func getConfiguredSettings(ctx context.Context, state tfModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := make(map[string]any)

	if persistentMap, ds := expandSettings(ctx, state.Persistent); ds.HasError() {
		diags.Append(ds...)
	} else if persistentMap != nil {
		settings["persistent"] = persistentMap
	}

	if !diags.HasError() {
		if transientMap, ds := expandSettings(ctx, state.Transient); ds.HasError() {
			diags.Append(ds...)
		} else if transientMap != nil {
			settings["transient"] = transientMap
		}
	}

	return settings, diags
}

// updateRemovedSettings adds null entries to targetMap for any settings present in
// oldSettings that are absent from newSettings, so Elasticsearch removes them.
func updateRemovedSettings(name string, oldSettings, newSettings map[string]any, targetMap map[string]any) {
	for k := range oldSettings {
		if _, ok := newSettings[k]; !ok {
			if targetMap[name] == nil {
				targetMap[name] = make(map[string]any)
			}
			targetMap[name].(map[string]any)[k] = nil
		}
	}
}

// flattenSettings converts the flat API response for a single category
// (persistent/transient) into a SingleNestedBlock object containing only
// settings that are tracked in configuredSettings.
//
// Returns a typed null Object when the category is not configured, and an
// empty-set Object when configured-but-no-tracked-keys-present-in-API.
func flattenSettings(ctx context.Context, category string, configuredSettings, apiResponse map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	configured, _ := configuredSettings[category].(map[string]any)
	if len(configured) == 0 {
		return nullSettingsBlock(), diags
	}

	apiCategory, _ := apiResponse[category].(map[string]any)

	settingModelAttr := settingModelAttrTypes()
	settingValues := make([]attr.Value, 0, len(configured))

	for k := range configured {
		v, ok := apiCategory[k]
		if !ok {
			// Setting was removed from ES; omit from state.
			continue
		}

		sm := settingModel{Name: types.StringValue(k)}
		switch t := v.(type) {
		case string:
			sm.Value = types.StringValue(t)
			sm.ValueList = types.ListNull(types.StringType)
		case []any:
			sm.Value = types.StringNull()
			vals := make([]attr.Value, len(t))
			for i, item := range t {
				vals[i] = types.StringValue(fmt.Sprintf("%v", item))
			}
			listVal, ds := types.ListValue(types.StringType, vals)
			diags.Append(ds...)
			if diags.HasError() {
				return nullSettingsBlock(), diags
			}
			sm.ValueList = listVal
		default:
			sm.Value = types.StringValue(fmt.Sprintf("%v", v))
			sm.ValueList = types.ListNull(types.StringType)
		}

		obj, ds := types.ObjectValueFrom(ctx, settingModelAttr, sm)
		diags.Append(ds...)
		if diags.HasError() {
			return nullSettingsBlock(), diags
		}
		settingValues = append(settingValues, obj)
	}

	if len(settingValues) == 0 {
		block, ds := emptySettingsBlock()
		diags.Append(ds...)
		if diags.HasError() {
			return nullSettingsBlock(), diags
		}
		return block, diags
	}

	settingSet, ds := types.SetValue(types.ObjectType{AttrTypes: settingModelAttr}, settingValues)
	diags.Append(ds...)
	if diags.HasError() {
		return nullSettingsBlock(), diags
	}

	block := settingsBlockModel{Setting: settingSet}
	blockObj, ds := types.ObjectValueFrom(ctx, settingsBlockAttrTypes(), block)
	diags.Append(ds...)
	if diags.HasError() {
		return nullSettingsBlock(), diags
	}
	return blockObj, diags
}
