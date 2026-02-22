package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newGaugePanelConfigConverter() gaugePanelConfigConverter {
	return gaugePanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.GaugeNoESQLTypeGauge),
		},
	}
}

type gaugePanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c gaugePanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.GaugeConfig != nil
}

func (c gaugePanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig2()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return nil
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var gaugeChart kbapi.GaugeChartSchema
	if err := json.Unmarshal(attrsJSON, &gaugeChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	gaugeNoESQL, err := gaugeChart.AsGaugeNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.GaugeConfig = &gaugeConfigModel{}
	return pm.GaugeConfig.fromAPI(ctx, gaugeNoESQL)
}

func (c gaugePanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.GaugeConfig

	gaugeNoESQL, gaugeDiags := configModel.toAPI()
	diags.Append(gaugeDiags...)
	if diags.HasError() {
		return diags
	}

	var gaugeChart kbapi.GaugeChartSchema
	if err := gaugeChart.FromGaugeNoESQL(gaugeNoESQL); err != nil {
		diags.AddError("Failed to convert gauge to schema", err.Error())
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromGaugeChartSchema(gaugeChart); err != nil {
		diags.AddError("Failed to create gauge attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_1_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig10Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig10{
		Attributes: configAttrs,
	}

	var config1 kbapi.DashboardPanelItemConfig1
	if err := config1.FromDashboardPanelItemConfig10(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig1(config1); err != nil {
		diags.AddError("Failed to marshal gauge config", err.Error())
	}

	return diags
}

type gaugeConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized                              `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	ShapeJSON           jsontypes.Normalized                              `tfsdk:"shape_json"`
}

func (m *gaugeConfigModel) fromAPI(ctx context.Context, api kbapi.GaugeNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics
	_ = ctx

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)

	datasetBytes, err := api.Dataset.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.DatasetJSON = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.IgnoreGlobalFilters = types.BoolPointerValue(api.IgnoreGlobalFilters)
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(api.Query)

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := api.Metric.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateGaugeMetricDefaults,
	)

	if api.Shape != nil {
		shapeBytes, err := api.Shape.MarshalJSON()
		if err != nil {
			diags.AddError("Failed to marshal shape", err.Error())
			return diags
		}
		m.ShapeJSON = jsontypes.NewNormalizedValue(string(shapeBytes))
	} else {
		m.ShapeJSON = jsontypes.NewNormalizedNull()
	}

	return diags
}

func (m *gaugeConfigModel) toAPI() (kbapi.GaugeNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var api kbapi.GaugeNoESQL

	api.Type = kbapi.GaugeNoESQLTypeGauge

	if !m.Title.IsNull() {
		api.Title = m.Title.ValueStringPointer()
	}

	if !m.Description.IsNull() {
		api.Description = m.Description.ValueStringPointer()
	}

	if typeutils.IsKnown(m.DatasetJSON) {
		if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
			return api, diags
		}
	}

	if !m.IgnoreGlobalFilters.IsNull() {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}

	if !m.Sampling.IsNull() {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
	}

	if m.Query != nil {
		api.Query = m.Query.toAPI()
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		api.Filters = &filters
	}

	if typeutils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return api, diags
		}
	}

	if typeutils.IsKnown(m.ShapeJSON) {
		var shape kbapi.GaugeNoESQL_Shape
		shapeDiags := m.ShapeJSON.Unmarshal(&shape)
		diags.Append(shapeDiags...)
		if !shapeDiags.HasError() {
			api.Shape = &shape
		}
	}

	return api, diags
}
