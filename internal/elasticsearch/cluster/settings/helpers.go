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
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// settingsBlockAttrTypes returns the attr.Type map for settingsBlockModel.
func settingsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"setting": types.SetType{ElemType: types.ObjectType{AttrTypes: settingModelAttrTypes()}},
	}
}

// settingsListElemType returns the attr.Type for elements of the persistent/transient list.
func settingsListElemType() attr.Type {
	return types.ObjectType{AttrTypes: settingsBlockAttrTypes()}
}

// emptySettingsBlockList returns an empty types.List with the correct element type
// for persistent/transient attributes.
func emptySettingsBlockList() types.List {
	return types.ListValueMust(settingsListElemType(), []attr.Value{})
}

// expandSettings converts a types.List of settingsBlockModel into the flat settings
// map used by the Elasticsearch API helpers (map[name]any). Returns nil when
// the list is null, unknown, or empty.
func expandSettings(ctx context.Context, settingsList types.List) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if settingsList.IsNull() || settingsList.IsUnknown() || len(settingsList.Elements()) == 0 {
		return nil, diags
	}

	var blocks []settingsBlockModel
	diags.Append(settingsList.ElementsAs(ctx, &blocks, false)...)
	if diags.HasError() {
		return nil, diags
	}

	if len(blocks) == 0 {
		return nil, diags
	}

	block := blocks[0]
	var settingModels []settingModel
	diags.Append(block.Setting.ElementsAs(ctx, &settingModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make(map[string]any, len(settingModels))
	for _, s := range settingModels {
		name := s.Name.ValueString()
		if _, exists := result[name]; exists {
			diags.AddError(
				fmt.Sprintf(`Unable to set "%s"`, name),
				fmt.Sprintf(`Found setting "%s" have been already configured.`, name),
			)
			return nil, diags
		}

		hasValue := !s.Value.IsNull() && !s.Value.IsUnknown() && s.Value.ValueString() != ""
		hasValueList := !s.ValueList.IsNull() && !s.ValueList.IsUnknown() && len(s.ValueList.Elements()) > 0

		if hasValue && hasValueList {
			diags.AddError(
				`Only one of "value" or "value_list" can be set.`,
				`Only one of "value" or "value_list" can be set.`,
			)
			return nil, diags
		}
		if !hasValue && !hasValueList {
			diags.AddError(
				`At least one of "value" or "value_list" must be set to not empty value.`,
				`At least one of "value" or "value_list" must be set to not empty value.`,
			)
			return nil, diags
		}

		if hasValue {
			result[name] = s.Value.ValueString()
		} else {
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
	if reflect.DeepEqual(oldSettings, newSettings) {
		return
	}
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
// (persistent/transient) into a types.List of settingsBlockModel, containing
// only settings that are tracked in configuredSettings.
func flattenSettings(ctx context.Context, category string, configuredSettings, apiResponse map[string]any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	configured, _ := configuredSettings[category].(map[string]any)
	if len(configured) == 0 {
		return emptySettingsBlockList(), diags
	}

	apiCategory, _ := apiResponse[category].(map[string]any)

	settingModelAttr := settingModelAttrTypes()
	var settingValues []attr.Value

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
			sm.ValueList = types.ListValueMust(types.StringType, []attr.Value{})
		case []any:
			sm.Value = types.StringValue("")
			vals := make([]attr.Value, len(t))
			for i, item := range t {
				vals[i] = types.StringValue(fmt.Sprintf("%v", item))
			}
			listVal, ds := types.ListValue(types.StringType, vals)
			diags.Append(ds...)
			if diags.HasError() {
				return types.ListNull(settingsListElemType()), diags
			}
			sm.ValueList = listVal
		default:
			sm.Value = types.StringValue(fmt.Sprintf("%v", v))
			sm.ValueList = types.ListValueMust(types.StringType, []attr.Value{})
		}

		obj, ds := types.ObjectValueFrom(ctx, settingModelAttr, sm)
		diags.Append(ds...)
		if diags.HasError() {
			return types.ListNull(settingsListElemType()), diags
		}
		settingValues = append(settingValues, obj)
	}

	if len(settingValues) == 0 {
		return emptySettingsBlockList(), diags
	}

	settingSet, ds := types.SetValue(types.ObjectType{AttrTypes: settingModelAttr}, settingValues)
	diags.Append(ds...)
	if diags.HasError() {
		return types.ListNull(settingsListElemType()), diags
	}

	block := settingsBlockModel{Setting: settingSet}
	blockAttr := settingsBlockAttrTypes()
	blockObj, ds := types.ObjectValueFrom(ctx, blockAttr, block)
	diags.Append(ds...)
	if diags.HasError() {
		return types.ListNull(settingsListElemType()), diags
	}

	result, ds := types.ListValue(settingsListElemType(), []attr.Value{blockObj})
	diags.Append(ds...)
	return result, diags
}
