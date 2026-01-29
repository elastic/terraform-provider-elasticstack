package slo

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
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

func (m tfModel) apmLatencyIndicatorToAPI() (bool, slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(m.ApmLatencyIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.ApmLatencyIndicator[0]

	return true, slo.SloWithSummaryResponseIndicator{
		IndicatorPropertiesApmLatency: &slo.IndicatorPropertiesApmLatency{
			Type: indicatorAddressToType["apm_latency_indicator"],
			Params: slo.IndicatorPropertiesApmLatencyParams{
				Service:         ind.Service.ValueString(),
				Environment:     ind.Environment.ValueString(),
				TransactionType: ind.TransactionType.ValueString(),
				TransactionName: ind.TransactionName.ValueString(),
				Filter:          stringPtr(ind.Filter),
				Index:           ind.Index.ValueString(),
				Threshold:       float64(ind.Threshold.ValueInt64()),
			},
		},
	}, diags
}

func (m *tfModel) populateFromApmLatencyIndicator(apiIndicator *slo.IndicatorPropertiesApmLatency) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiIndicator == nil {
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
