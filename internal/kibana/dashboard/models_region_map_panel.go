package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newRegionMapPanelConfigConverter() regionMapPanelConfigConverter {
	return regionMapPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.RegionMapNoESQLTypeRegionMap),
		},
	}
}

type regionMapPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c regionMapPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.RegionMapConfig != nil
}

func (c regionMapPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	cfgMap, err := config.AsDashboardPanelItemConfig2()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]interface{})
	if !ok {
		return nil
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var regionMap kbapi.RegionMapChartSchema
	if err := json.Unmarshal(attrsJSON, &regionMap); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.RegionMapConfig = &regionMapConfigModel{}

	regionMapNoESQL, err := regionMap.AsRegionMapNoESQL()
	if err == nil {
		return pm.RegionMapConfig.fromAPINoESQL(ctx, regionMapNoESQL)
	}

	regionMapESQL, err := regionMap.AsRegionMapESQL()
	if err == nil {
		return pm.RegionMapConfig.fromAPIESQL(ctx, regionMapESQL)
	}

	return diagutil.FrameworkDiagFromError(err)
}

func (c regionMapPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.RegionMapConfig

	regionMap, regionDiags := configModel.toAPI()
	diags.Append(regionDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromRegionMapChartSchema(regionMap); err != nil {
		diags.AddError("Failed to create region map attributes", err.Error())
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
		diags.AddError("Failed to marshal region map config", err.Error())
	}

	return diags
}

type regionMapConfigModel struct {
	Title               types.String                                      `tfsdk:"title"`
	Description         types.String                                      `tfsdk:"description"`
	DatasetJSON         jsontypes.Normalized                              `tfsdk:"dataset_json"`
	IgnoreGlobalFilters types.Bool                                        `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64                                     `tfsdk:"sampling"`
	Query               *filterSimpleModel                                `tfsdk:"query"`
	Filters             []searchFilterModel                               `tfsdk:"filters"`
	MetricJSON          customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"metric_json"`
	RegionJSON          jsontypes.Normalized                              `tfsdk:"region_json"`
}

func (m *regionMapConfigModel) fromAPINoESQL(ctx context.Context, api kbapi.RegionMapNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

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
		populateRegionMapMetricDefaults,
	)

	regionBytes, err := api.Region.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal region", err.Error())
		return diags
	}
	m.RegionJSON = jsontypes.NewNormalizedValue(string(regionBytes))

	return diags
}

func (m *regionMapConfigModel) fromAPIESQL(ctx context.Context, api kbapi.RegionMapESQL) diag.Diagnostics {
	var diags diag.Diagnostics

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

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	metricBytes, err := json.Marshal(api.Metric)
	if err != nil {
		diags.AddError("Failed to marshal metric", err.Error())
		return diags
	}
	m.MetricJSON = customtypes.NewJSONWithDefaultsValue[map[string]any](
		string(metricBytes),
		populateRegionMapMetricDefaults,
	)

	regionBytes, err := json.Marshal(api.Region)
	if err != nil {
		diags.AddError("Failed to marshal region", err.Error())
		return diags
	}
	m.RegionJSON = jsontypes.NewNormalizedValue(string(regionBytes))

	return diags
}

func (m *regionMapConfigModel) toAPI() (kbapi.RegionMapChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m == nil {
		return kbapi.RegionMapChartSchema{}, diags
	}

	if m.Query != nil && utils.IsKnown(m.Query.Query) {
		api := kbapi.RegionMapNoESQL{
			Type: kbapi.RegionMapNoESQLTypeRegionMap,
		}

		if utils.IsKnown(m.Title) {
			api.Title = m.Title.ValueStringPointer()
		}
		if utils.IsKnown(m.Description) {
			api.Description = m.Description.ValueStringPointer()
		}
		if utils.IsKnown(m.DatasetJSON) {
			if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
				return kbapi.RegionMapChartSchema{}, diags
			}
		}
		if utils.IsKnown(m.IgnoreGlobalFilters) {
			api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
		}
		if utils.IsKnown(m.Sampling) {
			sampling := float32(m.Sampling.ValueFloat64())
			api.Sampling = &sampling
		}
		api.Query = m.Query.toAPI()

		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
			for i, filterModel := range m.Filters {
				filter, filterDiags := filterModel.toAPI()
				diags.Append(filterDiags...)
				filters[i] = filter
			}
			api.Filters = &filters
		}

		if utils.IsKnown(m.MetricJSON) {
			if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
				diags.AddError("Failed to unmarshal metric", err.Error())
				return kbapi.RegionMapChartSchema{}, diags
			}
		}
		if utils.IsKnown(m.RegionJSON) {
			if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
				diags.AddError("Failed to unmarshal region", err.Error())
				return kbapi.RegionMapChartSchema{}, diags
			}
		}

		var schema kbapi.RegionMapChartSchema
		if err := schema.FromRegionMapNoESQL(api); err != nil {
			diags.AddError("Failed to create region map schema", err.Error())
		}
		return schema, diags
	}

	api := kbapi.RegionMapESQL{
		Type: kbapi.RegionMapESQLTypeRegionMap,
	}

	if utils.IsKnown(m.Title) {
		api.Title = m.Title.ValueStringPointer()
	}
	if utils.IsKnown(m.Description) {
		api.Description = m.Description.ValueStringPointer()
	}
	if utils.IsKnown(m.DatasetJSON) {
		if err := json.Unmarshal([]byte(m.DatasetJSON.ValueString()), &api.Dataset); err != nil {
			diags.AddError("Failed to unmarshal dataset", err.Error())
			return kbapi.RegionMapChartSchema{}, diags
		}
	}
	if utils.IsKnown(m.IgnoreGlobalFilters) {
		api.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if utils.IsKnown(m.Sampling) {
		sampling := float32(m.Sampling.ValueFloat64())
		api.Sampling = &sampling
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

	if utils.IsKnown(m.MetricJSON) {
		if err := json.Unmarshal([]byte(m.MetricJSON.ValueString()), &api.Metric); err != nil {
			diags.AddError("Failed to unmarshal metric", err.Error())
			return kbapi.RegionMapChartSchema{}, diags
		}
	}
	if utils.IsKnown(m.RegionJSON) {
		if err := json.Unmarshal([]byte(m.RegionJSON.ValueString()), &api.Region); err != nil {
			diags.AddError("Failed to unmarshal region", err.Error())
			return kbapi.RegionMapChartSchema{}, diags
		}
	}

	var schema kbapi.RegionMapChartSchema
	if err := schema.FromRegionMapESQL(api); err != nil {
		diags.AddError("Failed to create region map schema", err.Error())
	}
	return schema, diags
}
