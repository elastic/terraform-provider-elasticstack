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

func (m tfModel) histogramCustomIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.HistogramCustomIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.HistogramCustomIndicator[0]
	if len(ind.Good) != 1 || len(ind.Total) != 1 {
		diags.AddError("Invalid configuration", "histogram_custom_indicator.good and .total must each have exactly 1 item")
		return true, slo.SloWithSummaryResponseIndicator{}, diags
	}

	good := slo.IndicatorPropertiesHistogramParamsGood{
		Field:       ind.Good[0].Field.ValueString(),
		Aggregation: ind.Good[0].Aggregation.ValueString(),
		Filter:      stringPtr(ind.Good[0].Filter),
		From:        float64Ptr(ind.Good[0].From),
		To:          float64Ptr(ind.Good[0].To),
	}
	total := slo.IndicatorPropertiesHistogramParamsTotal{
		Field:       ind.Total[0].Field.ValueString(),
		Aggregation: ind.Total[0].Aggregation.ValueString(),
		Filter:      stringPtr(ind.Total[0].Filter),
		From:        float64Ptr(ind.Total[0].From),
		To:          float64Ptr(ind.Total[0].To),
	}

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesHistogram: &slo.IndicatorPropertiesHistogram{
			Type: indicatorAddressToType["histogram_custom_indicator"],
			Params: slo.IndicatorPropertiesHistogramParams{
				Index:          ind.Index.ValueString(),
				DataViewId:     stringPtr(ind.DataViewID),
				Filter:         stringPtr(ind.Filter),
				TimestampField: ind.TimestampField.ValueString(),
				Good:           good,
				Total:          total,
			},
		},
	}, diags
}

func (m *tfModel) populateFromHistogramCustomIndicator(apiIndicator *slo.IndicatorPropertiesHistogram) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiIndicator == nil {
		return diags
	}

	p := apiIndicator.Params
	ind := tfHistogramCustomIndicator{
		Index:          types.StringValue(p.Index),
		TimestampField: types.StringValue(p.TimestampField),
		Filter:         stringOrNull(p.Filter),
		DataViewID:     types.StringNull(),
		Good: []tfHistogramRange{{
			Field:       types.StringValue(p.Good.Field),
			Aggregation: types.StringValue(p.Good.Aggregation),
			Filter:      stringOrNull(p.Good.Filter),
			From:        float64OrNull(p.Good.From),
			To:          float64OrNull(p.Good.To),
		}},
		Total: []tfHistogramRange{{
			Field:       types.StringValue(p.Total.Field),
			Aggregation: types.StringValue(p.Total.Aggregation),
			Filter:      stringOrNull(p.Total.Filter),
			From:        float64OrNull(p.Total.From),
			To:          float64OrNull(p.Total.To),
		}},
	}
	if p.DataViewId != nil {
		ind.DataViewID = types.StringValue(*p.DataViewId)
	}

	m.HistogramCustomIndicator = []tfHistogramCustomIndicator{ind}
	return diags
}
