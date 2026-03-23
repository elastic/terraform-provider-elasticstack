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

package slo

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfKqlCustomIndicator struct {
	Index          types.String `tfsdk:"index"`
	DataViewID     types.String `tfsdk:"data_view_id"`
	Filter         types.String `tfsdk:"filter"`
	Good           types.String `tfsdk:"good"`
	Total          types.String `tfsdk:"total"`
	TimestampField types.String `tfsdk:"timestamp_field"`
}

func (m tfModel) kqlCustomIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.KqlCustomIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.KqlCustomIndicator[0]

	var filterObj *slo.KqlWithFilters
	if typeutils.IsKnown(ind.Filter) {
		v := ind.Filter.ValueString()
		filterObj = &slo.KqlWithFilters{String: &v}
	}

	// Default good and total to empty string if not provided, as they are required by the API
	// and must be marshallable to valid JSON
	goodStr := ""
	if typeutils.IsKnown(ind.Good) {
		goodStr = ind.Good.ValueString()
	}
	good := slo.KqlWithFiltersGood{String: &goodStr}

	totalStr := ""
	if typeutils.IsKnown(ind.Total) {
		totalStr = ind.Total.ValueString()
	}
	total := slo.KqlWithFiltersTotal{String: &totalStr}

	params := slo.IndicatorPropertiesCustomKqlParams{
		Index:          ind.Index.ValueString(),
		DataViewId:     stringPtr(ind.DataViewID),
		Filter:         filterObj,
		Good:           good,
		Total:          total,
		TimestampField: ind.TimestampField.ValueString(),
	}

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesCustomKql: &slo.IndicatorPropertiesCustomKql{
			Type:   indicatorAddressToType["kql_custom_indicator"],
			Params: params,
		},
	}, diags
}

func (m *tfModel) populateFromKqlCustomIndicator(apiIndicator *slo.IndicatorPropertiesCustomKql) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiIndicator == nil {
		return diags
	}

	p := apiIndicator.Params
	ind := tfKqlCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         types.StringNull(),
		Good:           types.StringNull(),
		Total:          types.StringNull(),
		DataViewID:     types.StringNull(),
	}
	if p.Filter != nil && p.Filter.String != nil {
		ind.Filter = types.StringValue(*p.Filter.String)
	}
	// Handle good and total fields - these are always present in the API response
	// If they are empty strings, preserve that in the state
	if p.Good.String != nil {
		ind.Good = types.StringValue(*p.Good.String)
	}
	if p.Total.String != nil {
		ind.Total = types.StringValue(*p.Total.String)
	}
	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	m.KqlCustomIndicator = []tfKqlCustomIndicator{ind}
	return diags
}
