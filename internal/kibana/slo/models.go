package slo

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var tfSettingsAttrTypes = map[string]attr.Type{
	"sync_delay":               types.StringType,
	"frequency":                types.StringType,
	"prevent_initial_backfill": types.BoolType,
}

type tfModel struct {
	ID types.String `tfsdk:"id"`

	SloID        types.String `tfsdk:"slo_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	SpaceID      types.String `tfsdk:"space_id"`
	BudgetMethod types.String `tfsdk:"budgeting_method"`

	TimeWindow []tfTimeWindow `tfsdk:"time_window"`
	Objective  []tfObjective  `tfsdk:"objective"`
	Settings   types.Object   `tfsdk:"settings"`

	GroupBy types.List     `tfsdk:"group_by"`
	Tags    []types.String `tfsdk:"tags"`

	MetricCustomIndicator    []tfMetricCustomIndicator    `tfsdk:"metric_custom_indicator"`
	HistogramCustomIndicator []tfHistogramCustomIndicator `tfsdk:"histogram_custom_indicator"`
	ApmLatencyIndicator      []tfApmLatencyIndicator      `tfsdk:"apm_latency_indicator"`
	ApmAvailabilityIndicator []tfApmAvailabilityIndicator `tfsdk:"apm_availability_indicator"`
	KqlCustomIndicator       []tfKqlCustomIndicator       `tfsdk:"kql_custom_indicator"`
	TimesliceMetricIndicator []tfTimesliceMetricIndicator `tfsdk:"timeslice_metric_indicator"`
}

type tfTimeWindow struct {
	Duration types.String `tfsdk:"duration"`
	Type     types.String `tfsdk:"type"`
}

type tfObjective struct {
	Target          types.Float64 `tfsdk:"target"`
	TimesliceTarget types.Float64 `tfsdk:"timeslice_target"`
	TimesliceWindow types.String  `tfsdk:"timeslice_window"`
}

type tfSettings struct {
	SyncDelay              types.String `tfsdk:"sync_delay"`
	Frequency              types.String `tfsdk:"frequency"`
	PreventInitialBackfill types.Bool   `tfsdk:"prevent_initial_backfill"`
}

func (m tfModel) toAPIModel() (models.Slo, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(m.TimeWindow) != 1 {
		diags.AddError("Invalid configuration", "time_window must have exactly 1 item")
		return models.Slo{}, diags
	}
	if len(m.Objective) != 1 {
		diags.AddError("Invalid configuration", "objective must have exactly 1 item")
		return models.Slo{}, diags
	}

	indicator, diags := m.indicatorToAPI()
	if diags.HasError() {
		return models.Slo{}, diags
	}

	tw := slo.TimeWindow{
		Type:     m.TimeWindow[0].Type.ValueString(),
		Duration: m.TimeWindow[0].Duration.ValueString(),
	}

	obj := slo.Objective{
		Target: m.Objective[0].Target.ValueFloat64(),
	}
	if utils.IsKnown(m.Objective[0].TimesliceTarget) {
		v := m.Objective[0].TimesliceTarget.ValueFloat64()
		obj.TimesliceTarget = &v
	}
	if utils.IsKnown(m.Objective[0].TimesliceWindow) {
		v := m.Objective[0].TimesliceWindow.ValueString()
		obj.TimesliceWindow = &v
	}

	var settings *slo.Settings
	if utils.IsKnown(m.Settings) {
		settingsModel, settingsDiags := tfSettingsFromObject(m.Settings)
		diags.Append(settingsDiags...)
		if diags.HasError() {
			return models.Slo{}, diags
		}
		settings = settingsModel.toAPIModel()
	}

	apiModel := models.Slo{
		Name:            m.Name.ValueString(),
		Description:     m.Description.ValueString(),
		Indicator:       indicator,
		TimeWindow:      tw,
		BudgetingMethod: slo.BudgetingMethod(m.BudgetMethod.ValueString()),
		Objective:       obj,
		Settings:        settings,
		SpaceID:         m.SpaceID.ValueString(),
	}

	if utils.IsKnown(m.SloID) {
		sloID := m.SloID.ValueString()
		if sloID != "" {
			apiModel.SloID = sloID
		}
	}

	if utils.IsKnown(m.GroupBy) {
		for i, v := range m.GroupBy.Elements() {
			g, ok := v.(types.String)
			if !ok {
				diags.AddError("Invalid configuration", fmt.Sprintf("group_by[%d] is not a string", i))
				continue
			}
			if utils.IsKnown(g) {
				apiModel.GroupBy = append(apiModel.GroupBy, g.ValueString())
			}
		}
	}
	for _, t := range m.Tags {
		if utils.IsKnown(t) {
			apiModel.Tags = append(apiModel.Tags, t.ValueString())
		}
	}

	return apiModel, diags
}

func (m tfModel) indicatorToAPI() (slo.SloWithSummaryResponseIndicator, diag.Diagnostics) {
	var diags diag.Diagnostics

	if ok, ind, indDiags := m.kqlCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.apmAvailabilityIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.apmLatencyIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.histogramCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.metricCustomIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	if ok, ind, indDiags := m.timesliceMetricIndicatorToAPI(); ok {
		diags.Append(indDiags...)
		return ind, diags
	}

	diags.AddError(
		"Invalid configuration",
		"exactly one indicator block must be set",
	)
	return slo.SloWithSummaryResponseIndicator{}, diags
}

func (m *tfModel) populateFromAPI(apiModel *models.Slo) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiModel == nil {
		return diags
	}

	m.SloID = types.StringValue(apiModel.SloID)
	m.SpaceID = types.StringValue(apiModel.SpaceID)
	m.Name = types.StringValue(apiModel.Name)
	m.Description = types.StringValue(apiModel.Description)
	m.BudgetMethod = types.StringValue(string(apiModel.BudgetingMethod))

	m.TimeWindow = []tfTimeWindow{{
		Duration: types.StringValue(apiModel.TimeWindow.Duration),
		Type:     types.StringValue(apiModel.TimeWindow.Type),
	}}

	obj := tfObjective{
		Target: types.Float64Value(apiModel.Objective.Target),
	}
	if apiModel.Objective.TimesliceTarget != nil {
		obj.TimesliceTarget = types.Float64Value(*apiModel.Objective.TimesliceTarget)
	} else {
		obj.TimesliceTarget = types.Float64Null()
	}
	if apiModel.Objective.TimesliceWindow != nil {
		obj.TimesliceWindow = types.StringValue(*apiModel.Objective.TimesliceWindow)
	} else {
		obj.TimesliceWindow = types.StringNull()
	}
	m.Objective = []tfObjective{obj}

	if utils.IsKnown(m.Settings) && apiModel.Settings != nil {
		attrValues := map[string]attr.Value{
			"sync_delay":               types.StringPointerValue(apiModel.Settings.SyncDelay),
			"frequency":                types.StringPointerValue(apiModel.Settings.Frequency),
			"prevent_initial_backfill": types.BoolPointerValue(apiModel.Settings.PreventInitialBackfill),
		}
		settingsObj, objDiags := types.ObjectValue(tfSettingsAttrTypes, attrValues)
		diags.Append(objDiags...)
		m.Settings = settingsObj
	} else {
		m.Settings = types.ObjectNull(tfSettingsAttrTypes)
	}

	if len(apiModel.GroupBy) > 0 {
		groupByValues := make([]attr.Value, len(apiModel.GroupBy))
		for i, g := range apiModel.GroupBy {
			groupByValues[i] = types.StringValue(g)
		}
		groupByList, listDiags := types.ListValue(types.StringType, groupByValues)
		diags.Append(listDiags...)
		m.GroupBy = groupByList
	} else {
		m.GroupBy = types.ListNull(types.StringType)
	}
	m.Tags = nil
	for _, t := range apiModel.Tags {
		m.Tags = append(m.Tags, types.StringValue(t))
	}

	m.MetricCustomIndicator = nil
	m.HistogramCustomIndicator = nil
	m.ApmLatencyIndicator = nil
	m.ApmAvailabilityIndicator = nil
	m.KqlCustomIndicator = nil
	m.TimesliceMetricIndicator = nil

	switch {
	case apiModel.Indicator.IndicatorPropertiesApmAvailability != nil:
		diags.Append(m.populateFromApmAvailabilityIndicator(apiModel.Indicator.IndicatorPropertiesApmAvailability)...)

	case apiModel.Indicator.IndicatorPropertiesApmLatency != nil:
		diags.Append(m.populateFromApmLatencyIndicator(apiModel.Indicator.IndicatorPropertiesApmLatency)...)

	case apiModel.Indicator.IndicatorPropertiesCustomKql != nil:
		diags.Append(m.populateFromKqlCustomIndicator(apiModel.Indicator.IndicatorPropertiesCustomKql)...)

	case apiModel.Indicator.IndicatorPropertiesHistogram != nil:
		diags.Append(m.populateFromHistogramCustomIndicator(apiModel.Indicator.IndicatorPropertiesHistogram)...)

	case apiModel.Indicator.IndicatorPropertiesCustomMetric != nil:
		diags.Append(m.populateFromMetricCustomIndicator(apiModel.Indicator.IndicatorPropertiesCustomMetric)...)

	case apiModel.Indicator.IndicatorPropertiesTimesliceMetric != nil:
		diags.Append(m.populateFromTimesliceMetricIndicator(apiModel.Indicator.IndicatorPropertiesTimesliceMetric)...)

	default:
		diags.AddError("Unexpected API response", "indicator not set")
		return diags
	}

	return diags
}

func tfSettingsFromObject(obj types.Object) (tfSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := obj.Attributes()

	syncDelayVal, ok := attrs["sync_delay"].(types.String)
	if !ok {
		diags.AddError("Invalid configuration", "settings.sync_delay is not a string")
		return tfSettings{}, diags
	}

	frequencyVal, ok := attrs["frequency"].(types.String)
	if !ok {
		diags.AddError("Invalid configuration", "settings.frequency is not a string")
		return tfSettings{}, diags
	}

	preventInitialBackfillVal, ok := attrs["prevent_initial_backfill"].(types.Bool)
	if !ok {
		diags.AddError("Invalid configuration", "settings.prevent_initial_backfill is not a bool")
		return tfSettings{}, diags
	}

	return tfSettings{
		SyncDelay:              syncDelayVal,
		Frequency:              frequencyVal,
		PreventInitialBackfill: preventInitialBackfillVal,
	}, diags
}

func (s tfSettings) toAPIModel() *slo.Settings {
	settings := slo.Settings{}
	hasAny := false

	if utils.IsKnown(s.SyncDelay) {
		v := s.SyncDelay.ValueString()
		settings.SyncDelay = &v
		hasAny = true
	}
	if utils.IsKnown(s.Frequency) {
		v := s.Frequency.ValueString()
		settings.Frequency = &v
		hasAny = true
	}
	if utils.IsKnown(s.PreventInitialBackfill) {
		v := s.PreventInitialBackfill.ValueBool()
		settings.PreventInitialBackfill = &v
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return &settings
}

func stringPtr(v types.String) *string {
	if !utils.IsKnown(v) {
		return nil
	}
	s := v.ValueString()
	return &s
}

func float64Ptr(v types.Float64) *float64 {
	if !utils.IsKnown(v) {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

func stringOrNull(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func float64OrNull(v *float64) types.Float64 {
	if v == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*v)
}

func (m tfModel) hasDataViewID() bool {
	return (len(m.MetricCustomIndicator) == 1 && utils.IsKnown(m.MetricCustomIndicator[0].DataViewID) && m.MetricCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.HistogramCustomIndicator) == 1 && utils.IsKnown(m.HistogramCustomIndicator[0].DataViewID) && m.HistogramCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.KqlCustomIndicator) == 1 && utils.IsKnown(m.KqlCustomIndicator[0].DataViewID) && m.KqlCustomIndicator[0].DataViewID.ValueString() != "") ||
		(len(m.TimesliceMetricIndicator) == 1 && utils.IsKnown(m.TimesliceMetricIndicator[0].DataViewID) && m.TimesliceMetricIndicator[0].DataViewID.ValueString() != "")
}
