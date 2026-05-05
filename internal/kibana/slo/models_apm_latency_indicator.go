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
	"math"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfApmLatencyIndicator struct {
	Index           types.String `tfsdk:"index"`
	Filter          types.String `tfsdk:"filter"`
	Service         types.String `tfsdk:"service"`
	Environment     types.String `tfsdk:"environment"`
	TransactionType types.String `tfsdk:"transaction_type"`
	TransactionName types.String `tfsdk:"transaction_name"`
	Threshold       types.Int64  `tfsdk:"threshold"`
}

func (m tfModel) apmLatencyIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.ApmLatencyIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.ApmLatencyIndicator[0]

	apmLatency := kbapi.SLOsIndicatorPropertiesApmLatency{
		Type: indicatorAddressToType["apm_latency_indicator"],
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			Threshold       float64 `json:"threshold"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Service:         ind.Service.ValueString(),
			Environment:     ind.Environment.ValueString(),
			TransactionType: ind.TransactionType.ValueString(),
			TransactionName: ind.TransactionName.ValueString(),
			Filter:          typeutils.ValueStringPointer(ind.Filter),
			Index:           ind.Index.ValueString(),
			Threshold:       float64(ind.Threshold.ValueInt64()),
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesApmLatency(apmLatency); err != nil {
		diags.AddError("Failed to build APM Latency indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromApmLatencyIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesApmLatency) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiIndicator.Params.Threshold != apiIndicator.Params.Threshold { // NaN check
		diags.AddError("Invalid API response", "indicator.params.threshold must be a valid number")
		return diags
	}
	threshold := float64(apiIndicator.Params.Threshold)
	if math.IsInf(threshold, 0) || threshold < 0 {
		diags.AddError("Invalid API response", "indicator.params.threshold must be a non-negative finite number")
		return diags
	}

	p := apiIndicator.Params
	m.ApmLatencyIndicator = []tfApmLatencyIndicator{{
		Environment:     types.StringValue(p.Environment),
		Service:         types.StringValue(p.Service),
		TransactionType: types.StringValue(p.TransactionType),
		TransactionName: types.StringValue(p.TransactionName),
		Index:           types.StringValue(p.Index),
		Filter:          stringOrNull(p.Filter),
		Threshold:       types.Int64Value(int64(p.Threshold)),
	}}

	return diags
}
