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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

func (m tfModel) kqlCustomIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.KqlCustomIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.KqlCustomIndicator[0]

	var filterObj *kbapi.SLOsKqlWithFilters
	if typeutils.IsKnown(ind.Filter) {
		v := ind.Filter.ValueString()
		var f kbapi.SLOsKqlWithFilters
		if err := f.FromSLOsKqlWithFilters0(v); err != nil {
			diags.AddError("Invalid configuration", "kql_custom_indicator.filter: "+err.Error())
			return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
		}
		filterObj = &f
	}

	// Default good and total to empty string if not provided, as they are required by the API.
	goodStr := ""
	if typeutils.IsKnown(ind.Good) {
		goodStr = ind.Good.ValueString()
	}
	var good kbapi.SLOsKqlWithFiltersGood
	if err := good.FromSLOsKqlWithFiltersGood0(goodStr); err != nil {
		diags.AddError("Invalid configuration", "kql_custom_indicator.good: "+err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	totalStr := ""
	if typeutils.IsKnown(ind.Total) {
		totalStr = ind.Total.ValueString()
	}
	var total kbapi.SLOsKqlWithFiltersTotal
	if err := total.FromSLOsKqlWithFiltersTotal0(totalStr); err != nil {
		diags.AddError("Invalid configuration", "kql_custom_indicator.total: "+err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	kqlIndicator := kbapi.SLOsIndicatorPropertiesCustomKql{
		Type: indicatorAddressToType["kql_custom_indicator"],
		Params: struct {
			DataViewId     *string                       `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
			Filter         *kbapi.SLOsKqlWithFilters     `json:"filter,omitempty"`
			Good           kbapi.SLOsKqlWithFiltersGood  `json:"good"`
			Index          string                        `json:"index"`
			TimestampField string                        `json:"timestampField"`
			Total          kbapi.SLOsKqlWithFiltersTotal `json:"total"`
		}{
			Index:          ind.Index.ValueString(),
			DataViewId:     stringPtr(ind.DataViewID),
			Filter:         filterObj,
			Good:           good,
			Total:          total,
			TimestampField: ind.TimestampField.ValueString(),
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesCustomKql(kqlIndicator); err != nil {
		diags.AddError("Failed to build KQL indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromKqlCustomIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesCustomKql) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p := apiIndicator.Params
	ind := tfKqlCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         types.StringNull(),
		Good:           types.StringNull(),
		Total:          types.StringNull(),
		DataViewID:     types.StringNull(),
	}

	if p.Filter != nil {
		// Try string variant first; fall back to object variant's KqlQuery field.
		// Note: AsSLOsKqlWithFilters0 always succeeds on kbapi unions (json.RawMessage).
		// We dispatch on whether the object variant's KqlQuery is set to distinguish variants.
		if f1, err := p.Filter.AsSLOsKqlWithFilters1(); err == nil && f1.KqlQuery != nil {
			ind.Filter = types.StringValue(*f1.KqlQuery)
		} else if s, err := p.Filter.AsSLOsKqlWithFilters0(); err == nil {
			ind.Filter = types.StringValue(s)
		}
	}

	// Dispatch on object variant's KqlQuery field; fall back to string variant.
	if g1, err := p.Good.AsSLOsKqlWithFiltersGood1(); err == nil && g1.KqlQuery != nil {
		ind.Good = types.StringValue(*g1.KqlQuery)
	} else if s, err := p.Good.AsSLOsKqlWithFiltersGood0(); err == nil {
		ind.Good = types.StringValue(s)
	}
	if t1, err := p.Total.AsSLOsKqlWithFiltersTotal1(); err == nil && t1.KqlQuery != nil {
		ind.Total = types.StringValue(*t1.KqlQuery)
	} else if s, err := p.Total.AsSLOsKqlWithFiltersTotal0(); err == nil {
		ind.Total = types.StringValue(s)
	}

	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	m.KqlCustomIndicator = []tfKqlCustomIndicator{ind}
	return diags
}
