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

package index

import (
	"context"
	"encoding/json"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const sortConfigPrivateStateKey = "sort_config"

type sortTuple struct {
	Fields  []string
	Orders  []string
	Missing []string
	Mode    []string
}

// readSortTuple parses the four sort setting slices from the model's raw
// settings. Returns ok=false (with no error) when SettingsRaw is unknown/null.
func readSortTuple(model tfModel) (sortTuple, bool, diag.Diagnostics) {
	if !typeutils.IsKnown(model.SettingsRaw) {
		return sortTuple{}, false, nil
	}
	var settings map[string]any
	if diags := model.SettingsRaw.Unmarshal(&settings); diags.HasError() {
		return sortTuple{}, false, diags
	}
	return sortTuple{
		Fields:  extractSortSetting(settings, indexparent.SettingSortField),
		Orders:  extractSortSetting(settings, indexparent.SettingSortOrder),
		Missing: extractSortSetting(settings, indexparent.SettingSortMissing),
		Mode:    extractSortSetting(settings, indexparent.SettingSortMode),
	}, true, nil
}

// saveSortConfig extracts sort settings from the model's raw settings JSON
// and stores the ordered sort configuration in private state.
func saveSortConfig(ctx context.Context, model tfModel, priv privateData) diag.Diagnostics {
	st, ok, diags := readSortTuple(model)
	if !ok || diags.HasError() {
		return diags
	}

	fields := st.Fields
	orders := st.Orders
	missing := st.Missing
	mode := st.Mode

	// Only save to private state if there are sort fields configured.
	if len(fields) == 0 {
		return nil
	}

	ps := sortPrivateState{
		Fields:  fields,
		Orders:  orders,
		Missing: missing,
		Mode:    mode,
	}

	data, err := json.Marshal(ps)
	if err != nil {
		diags.AddError("failed to marshal sort config", err.Error())
		return diags
	}

	diags.Append(priv.SetKey(ctx, sortConfigPrivateStateKey, data)...)
	return diags
}

// extractSortSetting extracts a string slice from the settings map for the
// given key. The settings map may have keys with "index." prefix or bare keys.
func extractSortSetting(settings map[string]any, key string) []string {
	prefixed := "index." + key
	for _, lookup := range []string{prefixed, key} {
		if v, ok := settings[lookup]; ok && v != nil {
			if result := stringSliceFromAny(v); result != nil {
				return result
			}
		}
	}
	return nil
}

// populateSortFromSettings populates the Sort ListNestedAttribute from the
// model's raw settings JSON. Only called when the sort attribute is non-null
// in the current state.
func populateSortFromSettings(ctx context.Context, model *tfModel) diag.Diagnostics {
	st, ok, diags := readSortTuple(*model)
	if !ok || diags.HasError() {
		return diags
	}

	fields := st.Fields
	orders := st.Orders
	missing := st.Missing
	mode := st.Mode

	if len(fields) == 0 {
		return nil
	}

	entries := make([]sortEntryModel, len(fields))
	for i, f := range fields {
		entries[i].Field = types.StringValue(f)

		if i < len(orders) && orders[i] != "" {
			entries[i].Order = types.StringValue(orders[i])
		} else {
			entries[i].Order = types.StringNull()
		}

		if i < len(missing) && missing[i] != "" {
			entries[i].Missing = types.StringValue(missing[i])
		} else {
			entries[i].Missing = types.StringNull()
		}

		if i < len(mode) && mode[i] != "" {
			entries[i].Mode = types.StringValue(mode[i])
		} else {
			entries[i].Mode = types.StringNull()
		}
	}

	elemType := sortElementType(ctx)
	sortList, listDiags := types.ListValueFrom(ctx, elemType, entries)
	diags.Append(listDiags...)
	if diags.HasError() {
		return diags
	}

	model.Sort = sortList
	return nil
}

// populateLegacySortFromSettings populates SortField and SortOrder from the
// model's raw settings JSON. Only called when the sort attribute is null/unknown
// in the current state (legacy path).
func populateLegacySortFromSettings(ctx context.Context, model *tfModel) diag.Diagnostics {
	st, ok, diags := readSortTuple(*model)
	if !ok || diags.HasError() {
		return diags
	}

	fields := st.Fields
	orders := st.Orders

	if len(fields) > 0 {
		fieldSet, setDiags := types.SetValueFrom(ctx, types.StringType, fields)
		diags.Append(setDiags...)
		if diags.HasError() {
			return diags
		}
		model.SortField = fieldSet
	}

	if len(orders) > 0 {
		orderList, listDiags := types.ListValueFrom(ctx, types.StringType, orders)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		model.SortOrder = orderList
	}

	return nil
}

// sortElementType returns the object type for a single sort entry ListNestedAttribute element.
func sortElementType(_ context.Context) attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			attrField:   types.StringType,
			attrOrder:   types.StringType,
			attrMissing: types.StringType,
			attrMode:    types.StringType,
		},
	}
}
