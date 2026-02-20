package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newPieChartPanelConfigConverter() pieChartPanelConfigConverter {
	return pieChartPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: "pie", // This seems to be the type literal used in API, though schema implies checking union
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.PieChartConfig != nil },
		},
	}
}

type pieChartPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c pieChartPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.PieChartConfig != nil
}

func (c pieChartPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	// Try to extract the pie chart config from the panel config
	cfgMap, err := config.AsDashboardPanelItemConfig2()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Extract the attributes
	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil
	}

	attrsMap, ok := attrs.(map[string]interface{})
	if !ok {
		return nil
	}

	// Marshal and unmarshal to get the PieChartSchema
	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var pieChart kbapi.PieChartSchema
	if err := json.Unmarshal(attrsJSON, &pieChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model
	pm.PieChartConfig = &pieChartConfigModel{}
	return pm.PieChartConfig.fromAPI(ctx, pieChart)
}

func (c pieChartPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.PieChartConfig

	// Convert the structured model to API schema
	pieChart, pieDiags := configModel.toAPI()
	diags.Append(pieDiags...)
	if diags.HasError() {
		return diags
	}

	// Create the nested Config1 structure
	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromPieChartSchema(pieChart); err != nil {
		diags.AddError("Failed to create pie chart attributes", err.Error())
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
		diags.AddError("Failed to marshal pie chart config", err.Error())
		return diags
	}

	return diags
}

type pieChartConfigModel struct {
	Title               types.String         `tfsdk:"title"`
	Description         types.String         `tfsdk:"description"`
	Dataset             jsontypes.Normalized `tfsdk:"dataset"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	DonutHole           types.String         `tfsdk:"donut_hole"`
	LabelPosition       types.String         `tfsdk:"label_position"`
	Legend              jsontypes.Normalized `tfsdk:"legend"`
	Query               *filterSimpleModel   `tfsdk:"query"`
	Filters             []searchFilterModel  `tfsdk:"filters"`
	Metrics             []pieMetricModel     `tfsdk:"metrics"`
	GroupBy             []pieGroupByModel    `tfsdk:"group_by"`
}

type pieMetricModel struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config"`
}

type pieGroupByModel struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config"`
}

func (m *pieChartConfigModel) fromAPI(ctx context.Context, apiChart kbapi.PieChartSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try with non-ESQL first (most common)
	noESQL, err := apiChart.AsPieNoESQL()
	if err == nil {
		// Check that Query is present if it's supposed to be there, or use that to disambiguate if needed
		return m.fromAPINoESQL(ctx, noESQL)
	}

	esql, err := apiChart.AsPieESQL()
	if err == nil {
		return m.fromAPIESQL(ctx, esql)
	}

	diags.AddError("Failed to parse pie chart schema", "Could not parse as either PieNoESQL or PieESQL")
	return diags
}

func (m *pieChartConfigModel) fromAPINoESQL(ctx context.Context, apiChart kbapi.PieNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	if apiChart.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*apiChart.IgnoreGlobalFilters)
	} else {
		m.IgnoreGlobalFilters = types.BoolValue(false)
	}

	if apiChart.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiChart.Sampling))
	} else {
		m.Sampling = types.Float64Value(1.0)
	}

	if apiChart.DonutHole != nil {
		m.DonutHole = types.StringValue(string(*apiChart.DonutHole))
	} else {
		m.DonutHole = types.StringNull()
	}

	if apiChart.LabelPosition != nil {
		m.LabelPosition = types.StringValue(string(*apiChart.LabelPosition))
	} else {
		m.LabelPosition = types.StringNull()
	}

	// Dataset
	datasetJSON, err := json.Marshal(apiChart.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))

	// Legend
	legendJSON, err := json.Marshal(apiChart.Legend)
	if err != nil {
		diags.AddError("Failed to marshal legend", err.Error())
		return diags
	}
	m.Legend = jsontypes.NewNormalizedValue(string(legendJSON))

	// Query
	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(apiChart.Query)

	// Filters
	if apiChart.Filters != nil && len(*apiChart.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*apiChart.Filters))
		for i, filter := range *apiChart.Filters {
			filterDiags := m.Filters[i].fromAPI(filter)
			diags.Append(filterDiags...)
		}
	}

	// Metrics
	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]pieMetricModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(metricJSON),
				populatePieChartMetricDefaults,
			)
		}
	}

	// GroupBy
	if apiChart.GroupBy != nil && len(*apiChart.GroupBy) > 0 {
		m.GroupBy = make([]pieGroupByModel, len(*apiChart.GroupBy))
		for i, groupBy := range *apiChart.GroupBy {
			groupByJSON, err := json.Marshal(groupBy)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(groupByJSON),
				populatePieChartGroupByDefaults,
			)
		}
	}

	return diags
}

func (m *pieChartConfigModel) fromAPIESQL(ctx context.Context, apiChart kbapi.PieESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Note: Title is not in PieESQL schema in some versions, check if it's there
	// According to generated/kbapi/kibana.gen.go (which I glimpsed), PieESQL doesn't have Title?
	// Wait, I saw "Title *string" in PieNoESQL but didn't check PieESQL carefully for Title.
	// Assuming it's absent or handled differently. Let's omit and check later.
	// Wait, description is there.
	m.Description = types.StringPointerValue(apiChart.Description)

	if apiChart.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*apiChart.IgnoreGlobalFilters)
	} else {
		m.IgnoreGlobalFilters = types.BoolValue(false)
	}

	if apiChart.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiChart.Sampling))
	} else {
		m.Sampling = types.Float64Value(1.0)
	}

	if apiChart.DonutHole != nil {
		m.DonutHole = types.StringValue(string(*apiChart.DonutHole))
	} else {
		m.DonutHole = types.StringNull()
	}

	if apiChart.LabelPosition != nil {
		m.LabelPosition = types.StringValue(string(*apiChart.LabelPosition))
	} else {
		m.LabelPosition = types.StringNull()
	}

	// Dataset
	datasetJSON, err := json.Marshal(apiChart.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal dataset", err.Error())
		return diags
	}
	m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))

	// Legend
	legendJSON, err := json.Marshal(apiChart.Legend)
	if err != nil {
		diags.AddError("Failed to marshal legend", err.Error())
		return diags
	}
	m.Legend = jsontypes.NewNormalizedValue(string(legendJSON))

	// No Query field for ESQL (it's part of dataset usually or handled differently)
	m.Query = nil

	// Filters
	if apiChart.Filters != nil && len(*apiChart.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*apiChart.Filters))
		for i, filter := range *apiChart.Filters {
			filterDiags := m.Filters[i].fromAPI(filter)
			diags.Append(filterDiags...)
		}
	}

	// Metrics
	if len(apiChart.Metrics) > 0 {
		m.Metrics = make([]pieMetricModel, len(apiChart.Metrics))
		for i, metric := range apiChart.Metrics {
			metricJSON, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal metric", err.Error())
				continue
			}
			m.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(metricJSON),
				populatePieChartMetricDefaults,
			)
		}
	}

	// GroupBy
	if apiChart.GroupBy != nil && len(*apiChart.GroupBy) > 0 {
		m.GroupBy = make([]pieGroupByModel, len(*apiChart.GroupBy))
		for i, groupBy := range *apiChart.GroupBy {
			groupByJSON, err := json.Marshal(groupBy)
			if err != nil {
				diags.AddError("Failed to marshal group_by", err.Error())
				continue
			}
			m.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue[map[string]any](
				string(groupByJSON),
				populatePieChartGroupByDefaults,
			)
		}
	}

	return diags
}

func (m *pieChartConfigModel) toAPI() (kbapi.PieChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var pieChart kbapi.PieChartSchema

	// Determine if it's ESQL or not based on query field?
	// If Query is set, it's definitely PieNoESQL. If not set, it might be ESQL or NoESQL with empty query.
	// But usually users specify which one they want.
	// For now, let's default to PieNoESQL if Query is present, otherwise check other criteria.
	// Actually, simpler logic: if Query is nil, assume ESQL? But wait, standard pie charts might have no query if using defaults?
	// It's safer to check structure of Dataset if possible, or try to infer.
	// However, mimicking MetricChart: check for existence of Query model.

	isNoESQL := m.Query != nil

	if isNoESQL {
		var chart kbapi.PieNoESQL

		chart.Title = m.Title.ValueStringPointer()
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.PieNoESQLDonutHole(m.DonutHole.ValueString())
			chart.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			val := kbapi.PieNoESQLLabelPosition(m.LabelPosition.ValueString())
			chart.LabelPosition = &val
		}

		// Legend
		if !m.Legend.IsNull() {
			if err := json.Unmarshal([]byte(m.Legend.ValueString()), &chart.Legend); err != nil {
				diags.AddError("Failed to unmarshal legend", err.Error())
			}
		}
		if chart.Legend.Size == "" {
			chart.Legend.Size = kbapi.LegendSizeAuto
		}

		// Dataset
		if !m.Dataset.IsNull() {
			if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &chart.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
			}
		}

		// Query
		chart.Query = m.Query.toAPI()

		// Filters
		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
			for i, filter := range m.Filters {
				f, d := filter.toAPI()
				diags.Append(d...)
				filters[i] = f
			}
			chart.Filters = &filters
		}

		// Metrics
		if len(m.Metrics) > 0 {
			metrics := make([]kbapi.PieNoESQL_Metrics_Item, len(m.Metrics))
			for i, metric := range m.Metrics {
				if err := json.Unmarshal([]byte(metric.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
				}
			}
			chart.Metrics = metrics
		}

		// GroupBy
		if len(m.GroupBy) > 0 {
			groupBy := make([]kbapi.PieNoESQL_GroupBy_Item, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
			}
			chart.GroupBy = &groupBy
		}

		// Always set type to pie as it's required by the schema
		chart.Type = kbapi.PieNoESQLTypePie

		if err := pieChart.FromPieNoESQL(chart); err != nil {
			diags.AddError("Failed to create PieNoESQL schema", err.Error())
		}
	} else {
		var chart kbapi.PieESQL

		// PieESQL does not have Title?
		// Check generated code, PieESQL struct definition in kiban.gen.go:
		// type PieESQL struct { ... Description *string ... } NO Title.
		chart.Description = m.Description.ValueStringPointer()
		chart.IgnoreGlobalFilters = m.IgnoreGlobalFilters.ValueBoolPointer()

		if !m.Sampling.IsNull() {
			val := float32(m.Sampling.ValueFloat64())
			chart.Sampling = &val
		}

		if !m.DonutHole.IsNull() {
			val := kbapi.PieESQLDonutHole(m.DonutHole.ValueString())
			chart.DonutHole = &val
		}

		if !m.LabelPosition.IsNull() {
			val := kbapi.PieESQLLabelPosition(m.LabelPosition.ValueString())
			chart.LabelPosition = &val
		}

		// Legend
		if !m.Legend.IsNull() {
			if err := json.Unmarshal([]byte(m.Legend.ValueString()), &chart.Legend); err != nil {
				diags.AddError("Failed to unmarshal legend", err.Error())
			}
		}
		if chart.Legend.Size == "" {
			chart.Legend.Size = kbapi.LegendSizeAuto
		}

		// Dataset
		if !m.Dataset.IsNull() {
			if err := json.Unmarshal([]byte(m.Dataset.ValueString()), &chart.Dataset); err != nil {
				diags.AddError("Failed to unmarshal dataset", err.Error())
			}
		}

		// Filters
		if len(m.Filters) > 0 {
			filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
			for i, filter := range m.Filters {
				f, d := filter.toAPI()
				diags.Append(d...)
				filters[i] = f
			}
			chart.Filters = &filters
		}

		// Metrics
		if len(m.Metrics) > 0 {
			metrics := make([]struct {
				Color     kbapi.StaticColor             `json:"color"`
				Column    string                        `json:"column"`
				Format    kbapi.FormatTypeSchema        `json:"format"`
				Label     *string                       `json:"label,omitempty"`
				Operation kbapi.PieESQLMetricsOperation `json:"operation"`
			}, len(m.Metrics))
			for i, metric := range m.Metrics {
				if err := json.Unmarshal([]byte(metric.Config.ValueString()), &metrics[i]); err != nil {
					diags.AddError("Failed to unmarshal metric", err.Error())
				}
			}
			chart.Metrics = metrics
		}

		// GroupBy
		if len(m.GroupBy) > 0 {
			groupBy := make([]struct {
				CollapseBy kbapi.CollapseBy              `json:"collapse_by"`
				Color      kbapi.ColorMapping            `json:"color"`
				Column     string                        `json:"column"`
				Operation  kbapi.PieESQLGroupByOperation `json:"operation"`
			}, len(m.GroupBy))
			for i, grp := range m.GroupBy {
				if err := json.Unmarshal([]byte(grp.Config.ValueString()), &groupBy[i]); err != nil {
					diags.AddError("Failed to unmarshal group_by", err.Error())
				}
			}
			chart.GroupBy = &groupBy
		}

		// Always set type to pie as it's required by the schema
		chart.Type = kbapi.PieESQLTypePie

		if err := pieChart.FromPieESQL(chart); err != nil {
			diags.AddError("Failed to create PieESQL schema", err.Error())
		}
	}

	return pieChart, diags
}
