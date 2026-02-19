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

func newMosaicPanelConfigConverter() mosaicPanelConfigConverter {
	return mosaicPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.MosaicNoESQLTypeMosaic),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.MosaicConfig != nil },
		},
	}
}

type mosaicPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c mosaicPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.MosaicConfig != nil
}

func (c mosaicPanelConfigConverter) handlesAPIPanelConfig(pm *panelModel, panelType string, cfg kbapi.DashboardPanelItem_Config) bool {
	if c.hasTFPanelConfig != nil && pm != nil && !c.hasTFPanelConfig(*pm) {
		return false
	}

	if panelType != "lens" {
		return false
	}

	attrsMap, err := getLensPanelAttributesMap(cfg)
	if err != nil {
		return false
	}

	vizType, ok := attrsMap["type"].(string)
	if !ok {
		return false
	}

	return vizType == string(kbapi.MosaicNoESQLTypeMosaic)
}

func (c mosaicPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	attrsMap, err := getLensPanelAttributesMap(config)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var mosaicChart kbapi.MosaicChartSchema
	if err := json.Unmarshal(attrsJSON, &mosaicChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	pm.MosaicConfig = &mosaicConfigModel{}

	isESQL := false
	if dataset, ok := attrsMap["dataset"].(map[string]interface{}); ok {
		if t, ok := dataset["type"].(string); ok && t == "esql" {
			isESQL = true
		}
	}

	if isESQL {
		mosaicESQL, err := mosaicChart.AsMosaicESQL()
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return pm.MosaicConfig.fromAPIESQL(ctx, mosaicESQL)
	}

	mosaicNoESQL, err := mosaicChart.AsMosaicNoESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.MosaicConfig.fromAPINoESQL(ctx, mosaicNoESQL)
}

func getLensPanelAttributesMap(cfg kbapi.DashboardPanelItem_Config) (map[string]interface{}, error) {
	// Prefer the typed union accessor when possible.
	if cfgMap, err := cfg.AsDashboardPanelItemConfig2(); err == nil {
		if attrs, ok := cfgMap["attributes"]; ok {
			if attrsMap, ok := attrs.(map[string]interface{}); ok {
				return attrsMap, nil
			}
		}
	}

	// Fall back to generic JSON parsing for other config variants.
	configBytes, err := cfg.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var cfgMap map[string]interface{}
	if err := json.Unmarshal(configBytes, &cfgMap); err != nil {
		return nil, err
	}

	attrs, ok := cfgMap["attributes"]
	if !ok {
		return nil, nil
	}

	attrsMap, ok := attrs.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	return attrsMap, nil
}

func (c mosaicPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	if pm.MosaicConfig == nil {
		return diags
	}

	mosaicChart, mosaicDiags := pm.MosaicConfig.toAPI()
	diags.Append(mosaicDiags...)
	if diags.HasError() {
		return diags
	}

	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromMosaicChartSchema(mosaicChart); err != nil {
		diags.AddError("Failed to create Mosaic chart attributes", err.Error())
		return diags
	}

	var configAttrs kbapi.DashboardPanelItem_Config_1_0_Attributes
	if err := configAttrs.FromDashboardPanelItemConfig10Attributes0(attrs0); err != nil {
		diags.AddError("Failed to create config attributes", err.Error())
		return diags
	}

	config10 := kbapi.DashboardPanelItemConfig10{Attributes: configAttrs}

	var config1 kbapi.DashboardPanelItemConfig1
	if err := config1.FromDashboardPanelItemConfig10(config10); err != nil {
		diags.AddError("Failed to create config1", err.Error())
		return diags
	}

	if err := apiConfig.FromDashboardPanelItemConfig1(config1); err != nil {
		diags.AddError("Failed to create config", err.Error())
		return diags
	}

	return diags
}

type mosaicConfigModel struct {
	Title               types.String             `tfsdk:"title"`
	Description         types.String             `tfsdk:"description"`
	IgnoreGlobalFilters types.Bool               `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64            `tfsdk:"sampling"`
	Legend              *mosaicLegendModel       `tfsdk:"legend"`
	Filters             []searchFilterModel      `tfsdk:"filters"`
	ValueDisplay        *mosaicValueDisplayModel `tfsdk:"value_display"`
	Esql                *mosaicEsqlModel         `tfsdk:"esql"`
	Standard            *mosaicStandardModel     `tfsdk:"standard"`
}

type mosaicLegendModel struct {
	Nested             types.Bool    `tfsdk:"nested"`
	Size               types.String  `tfsdk:"size"`
	TruncateAfterLines types.Float64 `tfsdk:"truncate_after_lines"`
	Visible            types.String  `tfsdk:"visible"`
}

type mosaicEsqlModel struct {
	Dataset          jsontypes.Normalized   `tfsdk:"dataset"`
	GroupBreakdownBy []mosaicOperationModel `tfsdk:"group_breakdown_by"`
	GroupBy          []mosaicOperationModel `tfsdk:"group_by"`
	Metrics          []mosaicOperationModel `tfsdk:"metrics"`
}

type mosaicStandardModel struct {
	Dataset          jsontypes.Normalized   `tfsdk:"dataset"`
	Query            *filterSimpleModel     `tfsdk:"query"`
	GroupBreakdownBy []mosaicOperationModel `tfsdk:"group_breakdown_by"`
	GroupBy          []mosaicOperationModel `tfsdk:"group_by"`
	Metrics          []mosaicOperationModel `tfsdk:"metrics"`
}

type mosaicOperationModel struct {
	Config customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"config"`
}

type mosaicValueDisplayModel struct {
	Mode            types.String  `tfsdk:"mode"`
	PercentDecimals types.Float64 `tfsdk:"percent_decimals"`
}

func (m *mosaicLegendModel) fromAPI(api kbapi.MosaicLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLines = types.Float64Value(float64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLines = types.Float64Null()
	}
	if api.Visible != nil {
		m.Visible = types.StringValue(string(*api.Visible))
	} else {
		m.Visible = types.StringNull()
	}
}

func (m *mosaicLegendModel) toAPI() kbapi.MosaicLegend {
	legend := kbapi.MosaicLegend{Size: kbapi.LegendSizeAuto}
	if !m.Nested.IsNull() {
		legend.Nested = m.Nested.ValueBoolPointer()
	}
	if !m.Size.IsNull() {
		legend.Size = kbapi.LegendSize(m.Size.ValueString())
	}
	if !m.TruncateAfterLines.IsNull() {
		v := float32(m.TruncateAfterLines.ValueFloat64())
		legend.TruncateAfterLines = &v
	}
	if !m.Visible.IsNull() {
		visible := kbapi.MosaicLegendVisible(m.Visible.ValueString())
		legend.Visible = &visible
	}
	return legend
}

func (m *mosaicConfigModel) fromAPIESQL(_ context.Context, api kbapi.MosaicESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	if api.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*api.IgnoreGlobalFilters)
	} else {
		m.IgnoreGlobalFilters = types.BoolValue(false)
	}
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Value(1)
	}

	m.Legend = &mosaicLegendModel{}
	m.Legend.fromAPI(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &mosaicValueDisplayModel{Mode: types.StringValue(string(api.ValueDisplay.Mode))}
		if api.ValueDisplay.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.ValueDisplay.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	m.Esql = &mosaicEsqlModel{}
	datasetBytes, err := json.Marshal(api.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal mosaic dataset", err.Error())
		return diags
	}
	m.Esql.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	if api.GroupBreakdownBy != nil && len(*api.GroupBreakdownBy) > 0 {
		m.Esql.GroupBreakdownBy = make([]mosaicOperationModel, len(*api.GroupBreakdownBy))
		for i, op := range *api.GroupBreakdownBy {
			b, err := json.Marshal(op)
			if err != nil {
				diags.AddError("Failed to marshal mosaic group_breakdown_by", err.Error())
				return diags
			}
			m.Esql.GroupBreakdownBy[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.Esql.GroupBy = make([]mosaicOperationModel, len(*api.GroupBy))
		for i, op := range *api.GroupBy {
			b, err := json.Marshal(op)
			if err != nil {
				diags.AddError("Failed to marshal mosaic group_by", err.Error())
				return diags
			}
			m.Esql.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	if len(api.Metrics) > 0 {
		m.Esql.Metrics = make([]mosaicOperationModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			b, err := json.Marshal(metric)
			if err != nil {
				diags.AddError("Failed to marshal mosaic metric", err.Error())
				return diags
			}
			m.Esql.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	return diags
}

func (m *mosaicConfigModel) fromAPINoESQL(_ context.Context, api kbapi.MosaicNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(api.Title)
	m.Description = types.StringPointerValue(api.Description)
	if api.IgnoreGlobalFilters != nil {
		m.IgnoreGlobalFilters = types.BoolValue(*api.IgnoreGlobalFilters)
	} else {
		m.IgnoreGlobalFilters = types.BoolValue(false)
	}
	if api.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*api.Sampling))
	} else {
		m.Sampling = types.Float64Value(1)
	}

	m.Legend = &mosaicLegendModel{}
	m.Legend.fromAPI(api.Legend)

	if api.ValueDisplay != nil {
		m.ValueDisplay = &mosaicValueDisplayModel{Mode: types.StringValue(string(api.ValueDisplay.Mode))}
		if api.ValueDisplay.PercentDecimals != nil {
			m.ValueDisplay.PercentDecimals = types.Float64Value(float64(*api.ValueDisplay.PercentDecimals))
		} else {
			m.ValueDisplay.PercentDecimals = types.Float64Null()
		}
	}

	if api.Filters != nil && len(*api.Filters) > 0 {
		m.Filters = make([]searchFilterModel, len(*api.Filters))
		for i, filterSchema := range *api.Filters {
			filterDiags := m.Filters[i].fromAPI(filterSchema)
			diags.Append(filterDiags...)
		}
	}

	m.Standard = &mosaicStandardModel{}
	datasetBytes, err := json.Marshal(api.Dataset)
	if err != nil {
		diags.AddError("Failed to marshal mosaic dataset", err.Error())
		return diags
	}
	m.Standard.Dataset = jsontypes.NewNormalizedValue(string(datasetBytes))

	m.Standard.Query = &filterSimpleModel{}
	m.Standard.Query.fromAPI(api.Query)

	if api.GroupBreakdownBy != nil && len(*api.GroupBreakdownBy) > 0 {
		m.Standard.GroupBreakdownBy = make([]mosaicOperationModel, len(*api.GroupBreakdownBy))
		for i, op := range *api.GroupBreakdownBy {
			b, err := op.MarshalJSON()
			if err != nil {
				diags.AddError("Failed to marshal mosaic group_breakdown_by", err.Error())
				return diags
			}
			m.Standard.GroupBreakdownBy[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	if api.GroupBy != nil && len(*api.GroupBy) > 0 {
		m.Standard.GroupBy = make([]mosaicOperationModel, len(*api.GroupBy))
		for i, op := range *api.GroupBy {
			b, err := op.MarshalJSON()
			if err != nil {
				diags.AddError("Failed to marshal mosaic group_by", err.Error())
				return diags
			}
			m.Standard.GroupBy[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	if len(api.Metrics) > 0 {
		m.Standard.Metrics = make([]mosaicOperationModel, len(api.Metrics))
		for i, metric := range api.Metrics {
			b, err := metric.MarshalJSON()
			if err != nil {
				diags.AddError("Failed to marshal mosaic metric", err.Error())
				return diags
			}
			m.Standard.Metrics[i].Config = customtypes.NewJSONWithDefaultsValue(string(b), populateMosaicOperationDefaults)
		}
	}

	return diags
}

func (m *mosaicConfigModel) toAPI() (kbapi.MosaicChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics
	var mosaicChart kbapi.MosaicChartSchema

	switch {
	case m.Esql != nil && m.Standard != nil:
		diags.AddError("Invalid mosaic config", "Only one of 'esql' or 'standard' can be configured")
		return mosaicChart, diags
	case m.Esql == nil && m.Standard == nil:
		diags.AddError("Invalid mosaic config", "One of 'esql' or 'standard' must be configured")
		return mosaicChart, diags
	case m.Esql != nil:
		api, apiDiags := m.toAPIESQL()
		diags.Append(apiDiags...)
		if diags.HasError() {
			return mosaicChart, diags
		}
		if err := mosaicChart.FromMosaicESQL(api); err != nil {
			diags.AddError("Failed to convert Mosaic ES|QL config", err.Error())
		}
		return mosaicChart, diags
	default:
		api, apiDiags := m.toAPINoESQL()
		diags.Append(apiDiags...)
		if diags.HasError() {
			return mosaicChart, diags
		}
		if err := mosaicChart.FromMosaicNoESQL(api); err != nil {
			diags.AddError("Failed to convert Mosaic standard config", err.Error())
		}
		return mosaicChart, diags
	}
}

func (m *mosaicConfigModel) toAPIESQL() (kbapi.MosaicESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Esql == nil {
		return kbapi.MosaicESQL{}, diags
	}

	attrs := map[string]interface{}{
		"type":   kbapi.MosaicESQLTypeMosaic,
		"legend": kbapi.MosaicLegend{Size: kbapi.LegendSizeAuto},
	}

	if !m.Title.IsNull() {
		attrs["title"] = m.Title.ValueStringPointer()
	}
	if !m.Description.IsNull() {
		attrs["description"] = m.Description.ValueStringPointer()
	}
	if !m.IgnoreGlobalFilters.IsNull() {
		attrs["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if !m.Sampling.IsNull() {
		sampling := float32(m.Sampling.ValueFloat64())
		attrs["sampling"] = &sampling
	}

	if m.Legend != nil {
		attrs["legend"] = m.Legend.toAPI()
	}

	if m.ValueDisplay != nil && !m.ValueDisplay.Mode.IsNull() {
		valueDisplay := map[string]interface{}{
			"mode": kbapi.MosaicESQLValueDisplayMode(m.ValueDisplay.Mode.ValueString()),
		}
		if !m.ValueDisplay.PercentDecimals.IsNull() {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			valueDisplay["percent_decimals"] = &p
		}
		attrs["value_display"] = valueDisplay
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		attrs["filters"] = &filters
	}

	if m.Esql.Dataset.IsNull() {
		diags.AddError("Invalid mosaic ES|QL config", "'dataset' must be set")
		return kbapi.MosaicESQL{}, diags
	}
	var datasetAny interface{}
	if err := json.Unmarshal([]byte(m.Esql.Dataset.ValueString()), &datasetAny); err != nil {
		diags.AddError("Failed to unmarshal mosaic ES|QL dataset", err.Error())
		return kbapi.MosaicESQL{}, diags
	}
	attrs["dataset"] = datasetAny

	if len(m.Esql.GroupBreakdownBy) > 0 {
		group := make([]interface{}, len(m.Esql.GroupBreakdownBy))
		for i, opModel := range m.Esql.GroupBreakdownBy {
			if opModel.Config.IsNull() {
				diags.AddError("Invalid mosaic ES|QL config", "'group_breakdown_by.config' must be set")
				return kbapi.MosaicESQL{}, diags
			}
			if err := json.Unmarshal([]byte(opModel.Config.ValueString()), &group[i]); err != nil {
				diags.AddError("Failed to unmarshal mosaic ES|QL group_breakdown_by", err.Error())
				return kbapi.MosaicESQL{}, diags
			}
		}
		attrs["group_breakdown_by"] = &group
	}

	if len(m.Esql.GroupBy) > 0 {
		group := make([]interface{}, len(m.Esql.GroupBy))
		for i, opModel := range m.Esql.GroupBy {
			if opModel.Config.IsNull() {
				diags.AddError("Invalid mosaic ES|QL config", "'group_by.config' must be set")
				return kbapi.MosaicESQL{}, diags
			}
			if err := json.Unmarshal([]byte(opModel.Config.ValueString()), &group[i]); err != nil {
				diags.AddError("Failed to unmarshal mosaic ES|QL group_by", err.Error())
				return kbapi.MosaicESQL{}, diags
			}
		}
		attrs["group_by"] = &group
	}

	if len(m.Esql.Metrics) != 1 {
		diags.AddError("Invalid mosaic ES|QL config", "'metrics' must have exactly 1 item")
		return kbapi.MosaicESQL{}, diags
	}
	metric := make([]interface{}, 1)
	if m.Esql.Metrics[0].Config.IsNull() {
		diags.AddError("Invalid mosaic ES|QL config", "'metrics.0.config' must be set")
		return kbapi.MosaicESQL{}, diags
	}
	if err := json.Unmarshal([]byte(m.Esql.Metrics[0].Config.ValueString()), &metric[0]); err != nil {
		diags.AddError("Failed to unmarshal mosaic ES|QL metric", err.Error())
		return kbapi.MosaicESQL{}, diags
	}
	attrs["metrics"] = metric

	attrsBytes, err := json.Marshal(attrs)
	if err != nil {
		diags.AddError("Failed to marshal Mosaic ES|QL config", err.Error())
		return kbapi.MosaicESQL{}, diags
	}

	var api kbapi.MosaicESQL
	if err := json.Unmarshal(attrsBytes, &api); err != nil {
		diags.AddError("Failed to decode Mosaic ES|QL config", err.Error())
		return kbapi.MosaicESQL{}, diags
	}

	return api, diags
}

func (m *mosaicConfigModel) toAPINoESQL() (kbapi.MosaicNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Standard == nil {
		return kbapi.MosaicNoESQL{}, diags
	}

	attrs := map[string]interface{}{
		"type":   kbapi.MosaicNoESQLTypeMosaic,
		"legend": kbapi.MosaicLegend{Size: kbapi.LegendSizeAuto},
	}

	if !m.Title.IsNull() {
		attrs["title"] = m.Title.ValueStringPointer()
	}
	if !m.Description.IsNull() {
		attrs["description"] = m.Description.ValueStringPointer()
	}
	if !m.IgnoreGlobalFilters.IsNull() {
		attrs["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBoolPointer()
	}
	if !m.Sampling.IsNull() {
		sampling := float32(m.Sampling.ValueFloat64())
		attrs["sampling"] = &sampling
	}

	if m.Legend != nil {
		attrs["legend"] = m.Legend.toAPI()
	}

	if m.ValueDisplay != nil && !m.ValueDisplay.Mode.IsNull() {
		valueDisplay := map[string]interface{}{
			"mode": kbapi.MosaicNoESQLValueDisplayMode(m.ValueDisplay.Mode.ValueString()),
		}
		if !m.ValueDisplay.PercentDecimals.IsNull() {
			p := float32(m.ValueDisplay.PercentDecimals.ValueFloat64())
			valueDisplay["percent_decimals"] = &p
		}
		attrs["value_display"] = valueDisplay
	}

	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, len(m.Filters))
		for i, filterModel := range m.Filters {
			filter, filterDiags := filterModel.toAPI()
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		attrs["filters"] = &filters
	}

	if m.Standard.Dataset.IsNull() {
		diags.AddError("Invalid mosaic standard config", "'dataset' must be set")
		return kbapi.MosaicNoESQL{}, diags
	}
	var datasetAny interface{}
	if err := json.Unmarshal([]byte(m.Standard.Dataset.ValueString()), &datasetAny); err != nil {
		diags.AddError("Failed to unmarshal mosaic standard dataset", err.Error())
		return kbapi.MosaicNoESQL{}, diags
	}
	attrs["dataset"] = datasetAny

	if m.Standard.Query == nil {
		diags.AddError("Invalid mosaic standard config", "'query' must be set")
		return kbapi.MosaicNoESQL{}, diags
	}
	attrs["query"] = m.Standard.Query.toAPI()

	if len(m.Standard.GroupBreakdownBy) > 0 {
		group := make([]interface{}, len(m.Standard.GroupBreakdownBy))
		for i, opModel := range m.Standard.GroupBreakdownBy {
			if opModel.Config.IsNull() {
				diags.AddError("Invalid mosaic standard config", "'group_breakdown_by.config' must be set")
				return kbapi.MosaicNoESQL{}, diags
			}
			if err := json.Unmarshal([]byte(opModel.Config.ValueString()), &group[i]); err != nil {
				diags.AddError("Failed to unmarshal mosaic standard group_breakdown_by", err.Error())
				return kbapi.MosaicNoESQL{}, diags
			}
		}
		attrs["group_breakdown_by"] = &group
	}

	if len(m.Standard.GroupBy) > 0 {
		group := make([]interface{}, len(m.Standard.GroupBy))
		for i, opModel := range m.Standard.GroupBy {
			if opModel.Config.IsNull() {
				diags.AddError("Invalid mosaic standard config", "'group_by.config' must be set")
				return kbapi.MosaicNoESQL{}, diags
			}
			if err := json.Unmarshal([]byte(opModel.Config.ValueString()), &group[i]); err != nil {
				diags.AddError("Failed to unmarshal mosaic standard group_by", err.Error())
				return kbapi.MosaicNoESQL{}, diags
			}
		}
		attrs["group_by"] = &group
	}

	if len(m.Standard.Metrics) != 1 {
		diags.AddError("Invalid mosaic standard config", "'metrics' must have exactly 1 item")
		return kbapi.MosaicNoESQL{}, diags
	}
	metric := make([]interface{}, 1)
	if m.Standard.Metrics[0].Config.IsNull() {
		diags.AddError("Invalid mosaic standard config", "'metrics.0.config' must be set")
		return kbapi.MosaicNoESQL{}, diags
	}
	if err := json.Unmarshal([]byte(m.Standard.Metrics[0].Config.ValueString()), &metric[0]); err != nil {
		diags.AddError("Failed to unmarshal mosaic standard metric", err.Error())
		return kbapi.MosaicNoESQL{}, diags
	}
	attrs["metrics"] = metric

	attrsBytes, err := json.Marshal(attrs)
	if err != nil {
		diags.AddError("Failed to marshal Mosaic standard config", err.Error())
		return kbapi.MosaicNoESQL{}, diags
	}

	var api kbapi.MosaicNoESQL
	if err := json.Unmarshal(attrsBytes, &api); err != nil {
		diags.AddError("Failed to decode Mosaic standard config", err.Error())
		return kbapi.MosaicNoESQL{}, diags
	}

	return api, diags
}
