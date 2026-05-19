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

package securitydetectionrule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func flattenActionAlertsFilter(ctx context.Context, apiFilter *kbapi.SecurityDetectionsAPIRuleActionAlertsFilter, diags *diag.Diagnostics) types.Object {
	if apiFilter == nil {
		return types.ObjectNull(getAlertsFilterAttrTypes())
	}

	filter := *apiFilter
	filterModel := ActionAlertsFilterModel{
		Query:     types.ObjectNull(getAlertsFilterQueryAttrTypes()),
		Timeframe: types.ObjectNull(getAlertsFilterTimeframeAttrTypes()),
	}

	if queryRaw, ok := filter["query"]; ok && queryRaw != nil {
		queryMap, ok := queryRaw.(map[string]any)
		if !ok {
			diags.AddError("Error reading alerts_filter", fmt.Sprintf("query must be an object, got %T", queryRaw))
			return types.ObjectNull(getAlertsFilterAttrTypes())
		}

		queryModel := ActionAlertsFilterQueryModel{
			Kql:         types.StringNull(),
			FiltersJSON: jsontypes.NewNormalizedNull(),
		}

		if kql, ok := queryMap["kql"].(string); ok {
			queryModel.Kql = types.StringValue(kql)
		}

		if filters, ok := queryMap["filters"]; ok && filters != nil {
			jsonBytes, err := json.Marshal(filters)
			if err != nil {
				diags.AddError("Error marshaling alerts_filter query filters", err.Error())
				return types.ObjectNull(getAlertsFilterAttrTypes())
			}
			queryModel.FiltersJSON = jsontypes.NewNormalizedValue(string(jsonBytes))
		} else {
			queryModel.FiltersJSON = jsontypes.NewNormalizedValue("[]")
		}

		queryObj, d := types.ObjectValueFrom(ctx, getAlertsFilterQueryAttrTypes(), queryModel)
		diags.Append(d...)
		filterModel.Query = queryObj
	}

	if timeframeRaw, ok := filter["timeframe"]; ok && timeframeRaw != nil {
		timeframeMap, ok := timeframeRaw.(map[string]any)
		if !ok {
			diags.AddError("Error reading alerts_filter", fmt.Sprintf("timeframe must be an object, got %T", timeframeRaw))
			return types.ObjectNull(getAlertsFilterAttrTypes())
		}

		timeframeModel := ActionAlertsFilterTimeframeModel{
			Days:       types.ListNull(types.Int64Type),
			Timezone:   types.StringNull(),
			HoursStart: types.StringNull(),
			HoursEnd:   types.StringNull(),
		}

		if daysRaw, ok := timeframeMap["days"]; ok && daysRaw != nil {
			daysList, d := alertsFilterDaysFromAPI(ctx, daysRaw)
			diags.Append(d...)
			timeframeModel.Days = daysList
		}

		if tz, ok := timeframeMap["timezone"].(string); ok {
			timeframeModel.Timezone = types.StringValue(tz)
		}

		if hoursRaw, ok := timeframeMap["hours"]; ok && hoursRaw != nil {
			hoursMap, ok := hoursRaw.(map[string]any)
			if !ok {
				diags.AddError("Error reading alerts_filter", fmt.Sprintf("timeframe.hours must be an object, got %T", hoursRaw))
				return types.ObjectNull(getAlertsFilterAttrTypes())
			}
			if start, ok := hoursMap["start"].(string); ok {
				timeframeModel.HoursStart = types.StringValue(start)
			}
			if end, ok := hoursMap["end"].(string); ok {
				timeframeModel.HoursEnd = types.StringValue(end)
			}
		}

		tfObj, d := types.ObjectValueFrom(ctx, getAlertsFilterTimeframeAttrTypes(), timeframeModel)
		diags.Append(d...)
		filterModel.Timeframe = tfObj
	}

	filterObj, d := types.ObjectValueFrom(ctx, getAlertsFilterAttrTypes(), filterModel)
	diags.Append(d...)
	return filterObj
}

func alertsFilterDaysFromAPI(ctx context.Context, daysRaw any) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	daysSlice, ok := daysRaw.([]any)
	if !ok {
		diags.AddError("Error reading alerts_filter timeframe days", fmt.Sprintf("days must be an array, got %T", daysRaw))
		return types.ListNull(types.Int64Type), diags
	}

	days := make([]int64, 0, len(daysSlice))
	for _, d := range daysSlice {
		switch v := d.(type) {
		case float64:
			days = append(days, int64(v))
		case int:
			days = append(days, int64(v))
		case int64:
			days = append(days, v)
		case json.Number:
			day, err := v.Int64()
			if err != nil {
				diags.AddError("Error reading alerts_filter timeframe days", err.Error())
				return types.ListNull(types.Int64Type), diags
			}
			days = append(days, day)
		default:
			diags.AddError("Error reading alerts_filter timeframe days", fmt.Sprintf("unexpected day value type %T", d))
			return types.ListNull(types.Int64Type), diags
		}
	}

	daysList, d := types.ListValueFrom(ctx, types.Int64Type, days)
	diags.Append(d...)
	return daysList, diags
}

func expandActionAlertsFilter(ctx context.Context, alertsFilter types.Object, diags *diag.Diagnostics) *kbapi.SecurityDetectionsAPIRuleActionAlertsFilter {
	if !typeutils.IsKnown(alertsFilter) || alertsFilter.IsNull() {
		return nil
	}

	var filterModel ActionAlertsFilterModel
	diags.Append(alertsFilter.As(ctx, &filterModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	result := make(map[string]any)

	if typeutils.IsKnown(filterModel.Query) && !filterModel.Query.IsNull() {
		var queryModel ActionAlertsFilterQueryModel
		diags.Append(filterModel.Query.As(ctx, &queryModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}

		queryMap := make(map[string]any)

		if typeutils.IsKnown(queryModel.Kql) {
			queryMap["kql"] = queryModel.Kql.ValueString()
		}

		filtersSlice := []any{}
		if typeutils.IsKnown(queryModel.FiltersJSON) && queryModel.FiltersJSON.ValueString() != "" {
			if err := json.Unmarshal([]byte(queryModel.FiltersJSON.ValueString()), &filtersSlice); err != nil {
				diags.AddError("Error unmarshaling alerts_filter filters_json", err.Error())
				return nil
			}
		}
		queryMap["filters"] = filtersSlice

		if len(queryMap) > 0 {
			result["query"] = queryMap
		}
	}

	if typeutils.IsKnown(filterModel.Timeframe) && !filterModel.Timeframe.IsNull() {
		var tfModel ActionAlertsFilterTimeframeModel
		diags.Append(filterModel.Timeframe.As(ctx, &tfModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil
		}

		timeframeMap := make(map[string]any)

		if typeutils.IsKnown(tfModel.Days) {
			var days []int64
			diags.Append(tfModel.Days.ElementsAs(ctx, &days, false)...)
			if diags.HasError() {
				return nil
			}
			daysInt := make([]int, len(days))
			for i, d := range days {
				daysInt[i] = int(d)
			}
			timeframeMap["days"] = daysInt
		}

		if typeutils.IsKnown(tfModel.Timezone) {
			timeframeMap["timezone"] = tfModel.Timezone.ValueString()
		}

		hoursMap := make(map[string]any)
		if typeutils.IsKnown(tfModel.HoursStart) {
			hoursMap["start"] = tfModel.HoursStart.ValueString()
		}
		if typeutils.IsKnown(tfModel.HoursEnd) {
			hoursMap["end"] = tfModel.HoursEnd.ValueString()
		}
		if len(hoursMap) > 0 {
			timeframeMap["hours"] = hoursMap
		}

		if len(timeframeMap) > 0 {
			result["timeframe"] = timeframeMap
		}
	}

	if len(result) == 0 {
		return nil
	}

	apiFilter := kbapi.SecurityDetectionsAPIRuleActionAlertsFilter(result)
	return &apiFilter
}
