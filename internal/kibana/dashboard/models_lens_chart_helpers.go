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
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// populateFiltersFromAPI converts a slice of kbapi.LensPanelFilters_Item into models.ChartFilterJSONModel
// values, appending any errors to diags.
func populateFiltersFromAPI(filters []kbapi.LensPanelFilters_Item, diags *diag.Diagnostics) []models.ChartFilterJSONModel {
	if len(filters) == 0 {
		return nil
	}
	result := make([]models.ChartFilterJSONModel, 0, len(filters))
	for _, f := range filters {
		fm := models.ChartFilterJSONModel{}
		fd := chartFilterJSONPopulateFromAPIItem(&fm, f)
		diags.Append(fd...)
		if !fd.HasError() {
			result = append(result, fm)
		}
	}
	return result
}

// buildFiltersForAPI converts the model filter slice into the kbapi type, appending errors to diags.
// The returned slice is always non-nil (empty API payload is []kbapi.LensPanelFilters_Item{}).
func buildFiltersForAPI(filters []models.ChartFilterJSONModel, diags *diag.Diagnostics) []kbapi.LensPanelFilters_Item {
	if len(filters) == 0 {
		return []kbapi.LensPanelFilters_Item{}
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
	return items
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

func preservePriorJSONWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current customtypes.JSONWithDefaultsValue[T], diags *diag.Diagnostics) customtypes.JSONWithDefaultsValue[T] {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return current
	}

	eq, d := prior.StringSemanticEquals(ctx, current)
	diags.Append(d...)
	if d.HasError() {
		return current
	}
	if eq {
		return prior
	}
	return current
}

// lensDataSourceIsESQLOrTable reports whether a Lens chart's `data_source` union
// JSON is the ES|QL ("esql") or table ("table") dataset shape. Both map to the
// ES|QL API variant of a chart panel. Returns false on marshal/unmarshal error.
func lensDataSourceIsESQLOrTable(body []byte, err error) bool {
	if err != nil {
		return false
	}
	var ds struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &ds); err != nil {
		return false
	}
	return ds.Type == legacyMetricDatasetTypeESQL || ds.Type == legacyMetricDatasetTypeTable
}

// lensESQLNumberFormatJSONFromAPI marshals a Lens ES|QL dimension `format` union
// value to a normalized Terraform string. Empty or null JSON is replaced with the
// default number-format payload so Terraform state matches what Kibana echoes.
func lensESQLNumberFormatJSONFromAPI(format any, errLabel string, diags *diag.Diagnostics) (jsontypes.Normalized, bool) {
	bytes, err := json.Marshal(format)
	if err != nil {
		diags.AddError("Failed to marshal "+errLabel, err.Error())
		return jsontypes.Normalized{}, false
	}
	if len(bytes) == 0 || string(bytes) == jsonNullString {
		bytes = []byte(defaultNumberFormatJSON)
	}
	return jsontypes.NewNormalizedValue(normalizeKibanaLensNumberFormatJSONString(string(bytes))), true
}

// lensQueryESQLMode returns whether a Lens chart's optional `query` object selects
// ES|QL mode (i.e. `query` is omitted, or both `expression` and `language` are
// null). ok is false when the configuration is still unknown and validation
// should defer.
func lensQueryESQLMode(ctx context.Context, config tfsdk.Config, attrPath path.Path, diags *diag.Diagnostics) (esqlMode bool, ok bool) {
	var queryObj types.Object
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query"), &queryObj)...)
	if diags.HasError() {
		return false, false
	}
	if queryObj.IsUnknown() {
		return false, false
	}
	if queryObj.IsNull() {
		return true, true
	}

	var lang, expr types.String
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("language"), &lang)...)
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("expression"), &expr)...)
	if diags.HasError() {
		return false, false
	}
	return lang.IsNull() && expr.IsNull(), true
}

func preservePriorNormalizedWithDefaultsIfEquivalent[T any](ctx context.Context, prior, current jsontypes.Normalized, defaults func(T) T, diags *diag.Diagnostics) jsontypes.Normalized {
	if prior.IsNull() || prior.IsUnknown() || current.IsNull() || current.IsUnknown() {
		return current
	}

	priorWithDefaults := customtypes.NewJSONWithDefaultsValue(prior.ValueString(), defaults)
	currentWithDefaults := customtypes.NewJSONWithDefaultsValue(current.ValueString(), defaults)
	eq, d := priorWithDefaults.StringSemanticEquals(ctx, currentWithDefaults)
	diags.Append(d...)
	if d.HasError() {
		return current
	}
	if eq {
		return prior
	}
	return current
}
