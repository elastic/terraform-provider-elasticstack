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

package mlanomalycharts

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type severityRange struct {
	Min int64  `json:"min"`
	Max *int64 `json:"max,omitempty"`
}

var canonicalSeverityBands = map[string]severityRange{
	severityLow:      {Min: 0, Max: new(int64(3))},
	severityWarning:  {Min: 3, Max: new(int64(25))},
	severityMinor:    {Min: 25, Max: new(int64(50))},
	severityMajor:    {Min: 50, Max: new(int64(75))},
	severityCritical: {Min: 75, Max: nil},
}

func severityRangesEqual(aMin int64, aMax *int64, bMin int64, bMax *int64) bool {
	if aMin != bMin {
		return false
	}
	switch {
	case aMax == nil && bMax == nil:
		return true
	case aMax == nil || bMax == nil:
		return false
	default:
		return *aMax == *bMax
	}
}

func canonicalSeverityForRange(minVal int64, maxVal *int64) (string, bool) {
	for name, band := range canonicalSeverityBands {
		if severityRangesEqual(minVal, maxVal, band.Min, band.Max) {
			return name, true
		}
	}
	return "", false
}

func buildSeverityThresholdItem(item models.MlAnomalyChartsSeverityThresholdModel) (kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item, diag.Diagnostics) {
	var out kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item
	var payload severityRange

	switch {
	case typeutils.IsKnown(item.Severity):
		band, ok := canonicalSeverityBands[item.Severity.ValueString()]
		if !ok {
			var diags diag.Diagnostics
			diags.AddError("Invalid ML anomaly charts configuration", "Unknown severity value.")
			return out, diags
		}
		payload = band
	case typeutils.IsKnown(item.Min):
		payload.Min = item.Min.ValueInt64()
		payload.Max = typeutils.Int64Pointer(item.Max)
	default:
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly charts configuration", "Each `severity_threshold` entry must set either `severity` or `min`.")
		return out, diags
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly charts configuration", err.Error())
		return out, diags
	}
	if err := out.UnmarshalJSON(raw); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly charts configuration", err.Error())
		return out, diags
	}
	return out, nil
}

func parseSeverityThresholdFromAPIItem(item kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item) (severityRange, diag.Diagnostics) {
	raw, err := json.Marshal(item)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly charts panel configuration on read", err.Error())
		return severityRange{}, diags
	}
	var parsed severityRange
	if err := json.Unmarshal(raw, &parsed); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid ML anomaly charts panel configuration on read", err.Error())
		return severityRange{}, diags
	}
	return parsed, nil
}

func severityThresholdRawModel(minVal int64, maxVal *int64) models.MlAnomalyChartsSeverityThresholdModel {
	out := models.MlAnomalyChartsSeverityThresholdModel{
		Severity: types.StringNull(),
		Min:      types.Int64Value(minVal),
	}
	if maxVal == nil {
		out.Max = types.Int64Null()
	} else {
		out.Max = types.Int64Value(*maxVal)
	}
	return out
}

func severityThresholdNamedModel(severity string) models.MlAnomalyChartsSeverityThresholdModel {
	return models.MlAnomalyChartsSeverityThresholdModel{
		Severity: types.StringValue(severity),
		Min:      types.Int64Null(),
		Max:      types.Int64Null(),
	}
}

func severityThresholdFromAPI(
	apiItem kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item,
	priorItem *models.MlAnomalyChartsSeverityThresholdModel,
) (models.MlAnomalyChartsSeverityThresholdModel, diag.Diagnostics) {
	parsed, diags := parseSeverityThresholdFromAPIItem(apiItem)
	if diags.HasError() {
		return models.MlAnomalyChartsSeverityThresholdModel{}, diags
	}

	// Prefer the named form when there is no prior state to preserve, or when the
	// prior item itself was authored as a named severity.
	preferNamed := priorItem == nil || (typeutils.IsKnown(priorItem.Severity) && !typeutils.IsKnown(priorItem.Min))
	if preferNamed {
		if severity, ok := canonicalSeverityForRange(parsed.Min, parsed.Max); ok {
			return severityThresholdNamedModel(severity), nil
		}
	}
	return severityThresholdRawModel(parsed.Min, parsed.Max), nil
}

func buildSeverityThresholdItems(items []models.MlAnomalyChartsSeverityThresholdModel) (*[]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item, diag.Diagnostics) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item, len(items))
	for i, item := range items {
		built, diags := buildSeverityThresholdItem(item)
		if diags.HasError() {
			return nil, diags
		}
		out[i] = built
	}
	return &out, nil
}

func readSeverityThresholdFromAPI(
	apiItems *[]kbapi.KibanaHTTPAPIsMlAnomalyCharts_SeverityThreshold_Item,
	priorItems []models.MlAnomalyChartsSeverityThresholdModel,
) ([]models.MlAnomalyChartsSeverityThresholdModel, diag.Diagnostics) {
	if apiItems == nil || len(*apiItems) == 0 {
		return nil, nil
	}
	result := make([]models.MlAnomalyChartsSeverityThresholdModel, len(*apiItems))
	for i, apiItem := range *apiItems {
		var priorItem *models.MlAnomalyChartsSeverityThresholdModel
		if i < len(priorItems) {
			priorItem = &priorItems[i]
		}
		item, diags := severityThresholdFromAPI(apiItem, priorItem)
		if diags.HasError() {
			return nil, diags
		}
		result[i] = item
	}
	return result, nil
}
