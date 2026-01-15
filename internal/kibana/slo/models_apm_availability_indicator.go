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
	var diags diag.Diagnostics
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
	var diags diag.Diagnostics
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
