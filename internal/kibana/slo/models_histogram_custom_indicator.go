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

type tfHistogramCustomIndicator struct {
	Index          types.String       `tfsdk:"index"`
	DataViewID     types.String       `tfsdk:"data_view_id"`
	Filter         types.String       `tfsdk:"filter"`
	TimestampField types.String       `tfsdk:"timestamp_field"`
	Good           []tfHistogramRange `tfsdk:"good"`
	Total          []tfHistogramRange `tfsdk:"total"`
}

type tfHistogramRange struct {
	Aggregation types.String  `tfsdk:"aggregation"`
	Field       types.String  `tfsdk:"field"`
	Filter      types.String  `tfsdk:"filter"`
	From        types.Float64 `tfsdk:"from"`
	To          types.Float64 `tfsdk:"to"`
}

func (m tfModel) histogramCustomIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.HistogramCustomIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.HistogramCustomIndicator[0]
	if len(ind.Good) != 1 || len(ind.Total) != 1 {
		diags.AddError("Invalid configuration", "histogram_custom_indicator.good and .total must each have exactly 1 item")
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	histIndicator := kbapi.SLOsIndicatorPropertiesHistogram{
		Type: indicatorAddressToType["histogram_custom_indicator"],
		Params: struct {
			DataViewId *string `json:"dataViewId,omitempty"` //nolint:revive // var-naming: API struct field
			Filter     *string `json:"filter,omitempty"`
			Good       struct {
				Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation `json:"aggregation"`
				Field       string                                                      `json:"field"`
				Filter      *string                                                     `json:"filter,omitempty"`
				From        *float64                                                    `json:"from,omitempty"`
				To          *float64                                                    `json:"to,omitempty"`
			} `json:"good"`
			Index          string `json:"index"`
			TimestampField string `json:"timestampField"`
			Total          struct {
				Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation `json:"aggregation"`
				Field       string                                                       `json:"field"`
				Filter      *string                                                      `json:"filter,omitempty"`
				From        *float64                                                     `json:"from,omitempty"`
				To          *float64                                                     `json:"to,omitempty"`
			} `json:"total"`
		}{
			Index:          ind.Index.ValueString(),
			DataViewId:     typeutils.ValueStringPointer(ind.DataViewID),
			Filter:         typeutils.ValueStringPointer(ind.Filter),
			TimestampField: ind.TimestampField.ValueString(),
			Good: struct {
				Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation `json:"aggregation"`
				Field       string                                                      `json:"field"`
				Filter      *string                                                     `json:"filter,omitempty"`
				From        *float64                                                    `json:"from,omitempty"`
				To          *float64                                                    `json:"to,omitempty"`
			}{
				Field:       ind.Good[0].Field.ValueString(),
				Aggregation: kbapi.SLOsIndicatorPropertiesHistogramParamsGoodAggregation(ind.Good[0].Aggregation.ValueString()),
				Filter:      typeutils.ValueStringPointer(ind.Good[0].Filter),
				From:        typeutils.Float64PointerValue(ind.Good[0].From),
				To:          typeutils.Float64PointerValue(ind.Good[0].To),
			},
			Total: struct {
				Aggregation kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation `json:"aggregation"`
				Field       string                                                       `json:"field"`
				Filter      *string                                                      `json:"filter,omitempty"`
				From        *float64                                                     `json:"from,omitempty"`
				To          *float64                                                     `json:"to,omitempty"`
			}{
				Field:       ind.Total[0].Field.ValueString(),
				Aggregation: kbapi.SLOsIndicatorPropertiesHistogramParamsTotalAggregation(ind.Total[0].Aggregation.ValueString()),
				Filter:      typeutils.ValueStringPointer(ind.Total[0].Filter),
				From:        typeutils.Float64PointerValue(ind.Total[0].From),
				To:          typeutils.Float64PointerValue(ind.Total[0].To),
			},
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesHistogram(histIndicator); err != nil {
		diags.AddError("Failed to build Histogram indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromHistogramCustomIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesHistogram) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p := apiIndicator.Params
	ind := tfHistogramCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         types.StringPointerValue(p.Filter),
		DataViewID:     types.StringNull(),
		Good: []tfHistogramRange{{
			Field:       types.StringValue(p.Good.Field),
			Aggregation: types.StringValue(string(p.Good.Aggregation)),
			Filter:      types.StringPointerValue(p.Good.Filter),
			From:        types.Float64PointerValue(p.Good.From),
			To:          types.Float64PointerValue(p.Good.To),
		}},
		Total: []tfHistogramRange{{
			Field:       types.StringValue(p.Total.Field),
			Aggregation: types.StringValue(string(p.Total.Aggregation)),
			Filter:      types.StringPointerValue(p.Total.Filter),
			From:        types.Float64PointerValue(p.Total.From),
			To:          types.Float64PointerValue(p.Total.To),
		}},
	}
	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	m.HistogramCustomIndicator = []tfHistogramCustomIndicator{ind}
	return diags
}
