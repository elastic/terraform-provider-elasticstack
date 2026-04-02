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

package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// populateFiltersFromAPI converts a slice of kbapi.LensPanelFilters_Item into chartFilterJSONModel
// values, appending any errors to diags.
func populateFiltersFromAPI(filters []kbapi.LensPanelFilters_Item, diags *diag.Diagnostics) []chartFilterJSONModel {
	if len(filters) == 0 {
		return nil
	}
	result := make([]chartFilterJSONModel, 0, len(filters))
	for _, f := range filters {
		fm := chartFilterJSONModel{}
		fd := fm.populateFromAPIItem(f)
		diags.Append(fd...)
		if !fd.HasError() {
			result = append(result, fm)
		}
	}
	return result
}

// buildFiltersForAPI converts the model filter slice into the kbapi type, appending errors to diags.
// The returned slice is always non-nil (empty API payload is []kbapi.LensPanelFilters_Item{}).
func buildFiltersForAPI(filters []chartFilterJSONModel, diags *diag.Diagnostics) []kbapi.LensPanelFilters_Item {
	result := []kbapi.LensPanelFilters_Item{}
	if len(filters) == 0 {
		return result
	}
	items := make([]kbapi.LensPanelFilters_Item, 0, len(filters))
	for _, f := range filters {
		var item kbapi.LensPanelFilters_Item
		fd := decodeChartFilterJSON(f.FilterJSON, &item)
		diags.Append(fd...)
		if !fd.HasError() {
			items = append(items, item)
		}
	}
	if len(items) > 0 {
		return items
	}
	return result
}

// marshalToNormalized stores the already-marshaled bytes as a jsontypes.Normalized value,
// or adds an error to diags and returns (zero, false) on failure.
func marshalToNormalized(bytes []byte, err error, fieldName string, diags *diag.Diagnostics) (jsontypes.Normalized, bool) {
	if err != nil {
		diags.AddError("Failed to marshal "+fieldName, err.Error())
		return jsontypes.Normalized{}, false
	}
	return jsontypes.NewNormalizedValue(string(bytes)), true
}

// marshalToJSONWithDefaults stores the already-marshaled bytes as a JSONWithDefaultsValue,
// or adds an error to diags and returns (zero, false) on failure.
func marshalToJSONWithDefaults[T any](bytes []byte, err error, fieldName string, defaults func(T) T, diags *diag.Diagnostics) (customtypes.JSONWithDefaultsValue[T], bool) {
	if err != nil {
		diags.AddError("Failed to marshal "+fieldName, err.Error())
		return customtypes.JSONWithDefaultsValue[T]{}, false
	}
	return customtypes.NewJSONWithDefaultsValue(string(bytes), defaults), true
}
