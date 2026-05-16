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

type tfApmAvailabilityIndicator struct {
	Index           types.String `tfsdk:"index"`
	Filter          types.String `tfsdk:"filter"`
	Service         types.String `tfsdk:"service"`
	Environment     types.String `tfsdk:"environment"`
	TransactionType types.String `tfsdk:"transaction_type"`
	TransactionName types.String `tfsdk:"transaction_name"`
}

func (m tfModel) apmAvailabilityIndicatorToAPI() (bool, kbapi.SLOsSloWithSummaryResponse_Indicator, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if len(m.ApmAvailabilityIndicator) != 1 {
		return false, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}

	ind := m.ApmAvailabilityIndicator[0]

	apmAvailability := kbapi.SLOsIndicatorPropertiesApmAvailability{
		Type: indicatorAddressToType["apm_availability_indicator"],
		Params: struct {
			Environment     string  `json:"environment"`
			Filter          *string `json:"filter,omitempty"`
			Index           string  `json:"index"`
			Service         string  `json:"service"`
			TransactionName string  `json:"transactionName"`
			TransactionType string  `json:"transactionType"`
		}{
			Service:         ind.Service.ValueString(),
			Environment:     ind.Environment.ValueString(),
			TransactionType: ind.TransactionType.ValueString(),
			TransactionName: ind.TransactionName.ValueString(),
			Filter:          typeutils.ValueStringPointer(ind.Filter),
			Index:           ind.Index.ValueString(),
		},
	}

	var result kbapi.SLOsSloWithSummaryResponse_Indicator
	if err := result.FromSLOsIndicatorPropertiesApmAvailability(apmAvailability); err != nil {
		diags.AddError("Failed to build APM Availability indicator", err.Error())
		return true, kbapi.SLOsSloWithSummaryResponse_Indicator{}, diags
	}
	return true, result, diags
}

func (m *tfModel) populateFromApmAvailabilityIndicator(apiIndicator kbapi.SLOsIndicatorPropertiesApmAvailability) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p := apiIndicator.Params
	m.ApmAvailabilityIndicator = []tfApmAvailabilityIndicator{{
		Environment:     types.StringValue(p.Environment),
		Service:         types.StringValue(p.Service),
		TransactionType: types.StringValue(p.TransactionType),
		TransactionName: types.StringValue(p.TransactionName),
		Index:           types.StringValue(p.Index),
		Filter:          types.StringPointerValue(p.Filter),
	}}

	return diags
}
