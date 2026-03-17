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

type tfApmAvailabilityIndicator struct {
	Index           types.String `tfsdk:"index"`
	Filter          types.String `tfsdk:"filter"`
	Service         types.String `tfsdk:"service"`
	Environment     types.String `tfsdk:"environment"`
	TransactionType types.String `tfsdk:"transaction_type"`
	TransactionName types.String `tfsdk:"transaction_name"`
}

func (m tfModel) apmAvailabilityIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.ApmAvailabilityIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.ApmAvailabilityIndicator[0]

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesApmAvailability: &slo.IndicatorPropertiesApmAvailability{
			Type: indicatorAddressToType["apm_availability_indicator"],
			Params: slo.IndicatorPropertiesApmAvailabilityParams{
				Service:         ind.Service.ValueString(),
				Environment:     ind.Environment.ValueString(),
				TransactionType: ind.TransactionType.ValueString(),
				TransactionName: ind.TransactionName.ValueString(),
				Filter:          stringPtr(ind.Filter),
				Index:           ind.Index.ValueString(),
			},
		},
	}, diags
}

func (m *tfModel) populateFromApmAvailabilityIndicator(apiIndicator *slo.IndicatorPropertiesApmAvailability) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiIndicator == nil {
		return diags
	}

	p := apiIndicator.Params
	m.ApmAvailabilityIndicator = []tfApmAvailabilityIndicator{{
		Environment:     types.StringValue(p.Environment),
		Service:         types.StringValue(p.Service),
		TransactionType: types.StringValue(p.TransactionType),
		TransactionName: types.StringValue(p.TransactionName),
		Index:           types.StringValue(p.Index),
		Filter:          stringOrNull(p.Filter),
	}}

	return diags
}
