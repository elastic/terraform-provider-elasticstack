package slo

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	var diags diag.Diagnostics
	if len(m.KqlCustomIndicator) != 1 {
		return false, slo.SloWithSummaryResponseIndicator{}, diags
	}

	ind := m.KqlCustomIndicator[0]

	var filterObj *slo.KqlWithFilters
	if utils.IsKnown(ind.Filter) {
		v := ind.Filter.ValueString()
		filterObj = &slo.KqlWithFilters{String: &v}
	}

	good := slo.KqlWithFiltersGood{}
	if utils.IsKnown(ind.Good) {
		v := ind.Good.ValueString()
		good.String = &v
	}

	total := slo.KqlWithFiltersTotal{}
	if utils.IsKnown(ind.Total) {
		v := ind.Total.ValueString()
		total.String = &v
	}

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
	var diags diag.Diagnostics
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
