package dashboard

import (
	"context"
	"encoding/json"
	"math"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newXYChartPanelConfigConverter() xyChartPanelConfigConverter {
	return xyChartPanelConfigConverter{
		lensPanelConfigConverter: lensPanelConfigConverter{
			visualizationType: string(kbapi.Xy),
		},
	}
}

type xyChartPanelConfigConverter struct {
	lensPanelConfigConverter
}

func (c xyChartPanelConfigConverter) handlesTFPanelConfig(pm panelModel) bool {
	return pm.XYChartConfig != nil
}

func (c xyChartPanelConfigConverter) populateFromAPIPanel(ctx context.Context, pm *panelModel, config kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	// Try to extract the XY chart config from the panel config
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

	// Marshal and unmarshal to get the XyChartSchema
	attrsJSON, err := json.Marshal(attrsMap)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var xyChart kbapi.XyChartSchema
	if err := json.Unmarshal(attrsJSON, &xyChart); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	// Populate the model
	pm.XYChartConfig = &xyChartConfigModel{}
	return pm.XYChartConfig.fromAPI(ctx, xyChart)
}

func (c xyChartPanelConfigConverter) mapPanelToAPI(pm panelModel, apiConfig *kbapi.DashboardPanelItem_Config) diag.Diagnostics {
	var diags diag.Diagnostics
	configModel := *pm.XYChartConfig

	// Convert the structured model to API schema
	xyChart, xyDiags := configModel.toAPI()
	diags.Append(xyDiags...)
	if diags.HasError() {
		return diags
	}

	// Create the nested Config1 structure
	var attrs0 kbapi.DashboardPanelItemConfig10Attributes0
	if err := attrs0.FromXyChartSchema(xyChart); err != nil {
		diags.AddError("Failed to create XY chart attributes", err.Error())
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
		diags.AddError("Failed to marshal XY chart config", err.Error())
	}

	return diags
}

type xyChartConfigModel struct {
	Title       types.String        `tfsdk:"title"`
	Description types.String        `tfsdk:"description"`
	Axis        *xyAxisModel        `tfsdk:"axis"`
	Decorations *xyDecorationsModel `tfsdk:"decorations"`
	Fitting     *xyFittingModel     `tfsdk:"fitting"`
	Layers      []xyLayerModel      `tfsdk:"layers"`
	Legend      *xyLegendModel      `tfsdk:"legend"`
	Query       *filterSimpleModel  `tfsdk:"query"`
	Filters     []searchFilterModel `tfsdk:"filters"`
}

type xyAxisModel struct {
	X     *xyAxisConfigModel `tfsdk:"x"`
	Left  *yAxisConfigModel  `tfsdk:"left"`
	Right *yAxisConfigModel  `tfsdk:"right"`
}

func (m *xyAxisModel) fromAPI(apiAxis kbapi.XyAxis) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiAxis.X != nil {
		m.X = &xyAxisConfigModel{}
		xDiags := m.X.fromAPI(apiAxis.X)
		diags.Append(xDiags...)
	}

	if apiAxis.Left != nil {
		m.Left = &yAxisConfigModel{}
		leftDiags := m.Left.fromAPILeft(apiAxis.Left)
		diags.Append(leftDiags...)
	}

	if apiAxis.Right != nil {
		m.Right = &yAxisConfigModel{}
		rightDiags := m.Right.fromAPIRight(apiAxis.Right)
		diags.Append(rightDiags...)
	}

	return diags
}

func (m *xyAxisModel) toAPI() (kbapi.XyAxis, diag.Diagnostics) {
	if m == nil {
		return kbapi.XyAxis{}, nil
	}

	var diags diag.Diagnostics
	var axis kbapi.XyAxis

	if m.X != nil {
		xAxis, xDiags := m.X.toAPI()
		diags.Append(xDiags...)
		axis.X = xAxis
	}

	if m.Left != nil {
		leftAxis, leftDiags := m.Left.toAPILeft()
		diags.Append(leftDiags...)
		axis.Left = leftAxis
	}

	if m.Right != nil {
		rightAxis, rightDiags := m.Right.toAPIRight()
		diags.Append(rightDiags...)
		axis.Right = rightAxis
	}

	return axis, diags
}

type xyAxisConfigModel struct {
	Title            *axisTitleModel      `tfsdk:"title"`
	Ticks            types.Bool           `tfsdk:"ticks"`
	Grid             types.Bool           `tfsdk:"grid"`
	LabelOrientation types.String         `tfsdk:"label_orientation"`
	Extent           jsontypes.Normalized `tfsdk:"extent"`
}

type xyAxisConfigAPIModel = struct {
	Extent           *kbapi.XyAxis_X_Extent         `json:"extent,omitempty"`
	Grid             *bool                          `json:"grid,omitempty"`
	LabelOrientation *kbapi.XyAxisXLabelOrientation `json:"label_orientation,omitempty"`
	Ticks            *bool                          `json:"ticks,omitempty"`
	Title            *struct {
		Value   *string `json:"value,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}

func (m *xyAxisConfigModel) fromAPI(apiAxis *xyAxisConfigAPIModel) diag.Diagnostics {
	if apiAxis == nil {
		return nil
	}

	var diags diag.Diagnostics

	m.Grid = types.BoolPointerValue(apiAxis.Grid)
	m.Ticks = types.BoolPointerValue(apiAxis.Ticks)
	m.LabelOrientation = typeutils.StringishPointerValue(apiAxis.LabelOrientation)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	if apiAxis.Extent != nil {
		extentJSON, err := json.Marshal(apiAxis.Extent)
		if err == nil {
			m.Extent = jsontypes.NewNormalizedValue(string(extentJSON))
		}
	}

	return diags
}

func (m *xyAxisConfigModel) toAPI() (*xyAxisConfigAPIModel, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	xAxis := &xyAxisConfigAPIModel{}

	if utils.IsKnown(m.Grid) {
		xAxis.Grid = utils.Pointer(m.Grid.ValueBool())
	}
	if utils.IsKnown(m.Ticks) {
		xAxis.Ticks = utils.Pointer(m.Ticks.ValueBool())
	}
	if utils.IsKnown(m.LabelOrientation) {
		labelOrient := kbapi.XyAxisXLabelOrientation(m.LabelOrientation.ValueString())
		xAxis.LabelOrientation = &labelOrient
	}
	if m.Title != nil {
		xAxis.Title = m.Title.toAPI()
	}
	if utils.IsKnown(m.Extent) {
		var extent kbapi.XyAxis_X_Extent
		extentDiags := m.Extent.Unmarshal(&extent)
		diags.Append(extentDiags...)
		if !extentDiags.HasError() {
			xAxis.Extent = &extent
		}
	}

	return xAxis, diags
}

type yAxisConfigModel struct {
	Title            *axisTitleModel      `tfsdk:"title"`
	Ticks            types.Bool           `tfsdk:"ticks"`
	Grid             types.Bool           `tfsdk:"grid"`
	LabelOrientation types.String         `tfsdk:"label_orientation"`
	Scale            types.String         `tfsdk:"scale"`
	Extent           jsontypes.Normalized `tfsdk:"extent"`
}

type leftYAxisConfigAPIModel = struct {
	Extent           *kbapi.XyAxis_Left_Extent         `json:"extent,omitempty"`
	Grid             *bool                             `json:"grid,omitempty"`
	LabelOrientation *kbapi.XyAxisLeftLabelOrientation `json:"label_orientation,omitempty"`
	Scale            *kbapi.XyAxisLeftScale            `json:"scale,omitempty"`
	Ticks            *bool                             `json:"ticks,omitempty"`
	Title            *struct {
		Value   *string `json:"value,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}

func (m *yAxisConfigModel) fromAPILeft(apiAxis *leftYAxisConfigAPIModel) diag.Diagnostics {
	if apiAxis == nil {
		return nil
	}

	var diags diag.Diagnostics

	m.Grid = types.BoolPointerValue(apiAxis.Grid)
	m.Ticks = types.BoolPointerValue(apiAxis.Ticks)
	m.LabelOrientation = typeutils.StringishPointerValue(apiAxis.LabelOrientation)
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	if apiAxis.Extent != nil {
		extentJSON, err := json.Marshal(apiAxis.Extent)
		if err == nil {
			m.Extent = jsontypes.NewNormalizedValue(string(extentJSON))
		}
	}

	return diags
}

func (m *yAxisConfigModel) toAPILeft() (*leftYAxisConfigAPIModel, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	yAxis := &leftYAxisConfigAPIModel{}

	if utils.IsKnown(m.Grid) {
		yAxis.Grid = utils.Pointer(m.Grid.ValueBool())
	}
	if utils.IsKnown(m.Ticks) {
		yAxis.Ticks = utils.Pointer(m.Ticks.ValueBool())
	}
	if utils.IsKnown(m.LabelOrientation) {
		labelOrient := kbapi.XyAxisLeftLabelOrientation(m.LabelOrientation.ValueString())
		yAxis.LabelOrientation = &labelOrient
	}
	if utils.IsKnown(m.Scale) {
		scale := kbapi.XyAxisLeftScale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = m.Title.toAPI()
	}
	if utils.IsKnown(m.Extent) {
		var extent kbapi.XyAxis_Left_Extent
		extentDiags := m.Extent.Unmarshal(&extent)
		diags.Append(extentDiags...)
		if !extentDiags.HasError() {
			yAxis.Extent = &extent
		}
	}

	return yAxis, diags
}

type rightYAxisConfigAPIModel = struct {
	// Extent Y-axis extent configuration defining how the axis bounds are calculated
	Extent *kbapi.XyAxis_Right_Extent `json:"extent,omitempty"`

	// Grid Whether to show grid lines for this axis
	Grid *bool `json:"grid,omitempty"`

	// LabelOrientation Orientation of the axis labels
	LabelOrientation *kbapi.XyAxisRightLabelOrientation `json:"label_orientation,omitempty"`

	// Scale Y-axis scale type for data transformation
	Scale *kbapi.XyAxisRightScale `json:"scale,omitempty"`

	// Ticks Whether to show tick marks on the axis
	Ticks *bool `json:"ticks,omitempty"`

	// Title Axis title configuration
	Title *struct {
		// Value Axis title text
		Value *string `json:"value,omitempty"`

		// Visible Whether to show the title
		Visible *bool `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}

func (m *yAxisConfigModel) fromAPIRight(apiAxis *rightYAxisConfigAPIModel) diag.Diagnostics {
	if apiAxis == nil {
		return nil
	}

	var diags diag.Diagnostics

	m.Grid = types.BoolPointerValue(apiAxis.Grid)
	m.Ticks = types.BoolPointerValue(apiAxis.Ticks)
	m.LabelOrientation = typeutils.StringishPointerValue(apiAxis.LabelOrientation)
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	if apiAxis.Extent != nil {
		extentJSON, err := json.Marshal(apiAxis.Extent)
		if err == nil {
			m.Extent = jsontypes.NewNormalizedValue(string(extentJSON))
		}
	}

	return diags
}

func (m *yAxisConfigModel) toAPIRight() (*rightYAxisConfigAPIModel, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	yAxis := &rightYAxisConfigAPIModel{}

	if utils.IsKnown(m.Grid) {
		yAxis.Grid = utils.Pointer(m.Grid.ValueBool())
	}
	if utils.IsKnown(m.Ticks) {
		yAxis.Ticks = utils.Pointer(m.Ticks.ValueBool())
	}
	if utils.IsKnown(m.LabelOrientation) {
		labelOrient := kbapi.XyAxisRightLabelOrientation(m.LabelOrientation.ValueString())
		yAxis.LabelOrientation = &labelOrient
	}
	if utils.IsKnown(m.Scale) {
		scale := kbapi.XyAxisRightScale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = m.Title.toAPI()
	}
	if utils.IsKnown(m.Extent) {
		var extent kbapi.XyAxis_Right_Extent
		extentDiags := m.Extent.Unmarshal(&extent)
		diags.Append(extentDiags...)
		if !extentDiags.HasError() {
			yAxis.Extent = &extent
		}
	}

	return yAxis, diags
}

type axisTitleModel struct {
	Value   types.String `tfsdk:"value"`
	Visible types.Bool   `tfsdk:"visible"`
}

func (m *axisTitleModel) fromAPI(apiTitle *struct {
	Value   *string `json:"value,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
}) {
	if apiTitle == nil {
		return
	}
	m.Value = types.StringPointerValue(apiTitle.Value)
	m.Visible = types.BoolPointerValue(apiTitle.Visible)
}

func (m *axisTitleModel) toAPI() *struct {
	Value   *string `json:"value,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}

	title := &struct {
		Value   *string `json:"value,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	}{}

	if utils.IsKnown(m.Value) {
		title.Value = utils.Pointer(m.Value.ValueString())
	}
	if utils.IsKnown(m.Visible) {
		title.Visible = utils.Pointer(m.Visible.ValueBool())
	}

	return title
}

type xyDecorationsModel struct {
	EndZones          types.Bool    `tfsdk:"end_zones"`
	CurrentTimeMarker types.Bool    `tfsdk:"current_time_marker"`
	PointVisibility   types.Bool    `tfsdk:"point_visibility"`
	LineInterpolation types.String  `tfsdk:"line_interpolation"`
	MinimumBarHeight  types.Int64   `tfsdk:"minimum_bar_height"`
	ShowValueLabels   types.Bool    `tfsdk:"show_value_labels"`
	FillOpacity       types.Float64 `tfsdk:"fill_opacity"`
	ValueLabels       types.Bool    `tfsdk:"value_labels"`
}

func (m *xyDecorationsModel) fromAPI(apiDecorations kbapi.XyDecorations) {
	m.EndZones = types.BoolPointerValue(apiDecorations.EndZones)
	m.CurrentTimeMarker = types.BoolPointerValue(apiDecorations.CurrentTimeMarker)
	m.PointVisibility = types.BoolPointerValue(apiDecorations.PointVisibility)
	m.LineInterpolation = typeutils.StringishPointerValue(apiDecorations.LineInterpolation)
	m.ShowValueLabels = types.BoolPointerValue(apiDecorations.ShowValueLabels)
	m.ValueLabels = types.BoolPointerValue(apiDecorations.ValueLabels)

	if apiDecorations.MinimumBarHeight != nil {
		m.MinimumBarHeight = types.Int64Value(int64(*apiDecorations.MinimumBarHeight))
	} else {
		m.MinimumBarHeight = types.Int64Null()
	}

	if apiDecorations.FillOpacity != nil {
		// Round to 2 decimal places to avoid float32 precision issues
		val := float64(*apiDecorations.FillOpacity)
		m.FillOpacity = types.Float64Value(math.Round(val*100) / 100)
	} else {
		m.FillOpacity = types.Float64Null()
	}
}

func (m *xyDecorationsModel) toAPI() kbapi.XyDecorations {
	if m == nil {
		return kbapi.XyDecorations{}
	}

	var decorations kbapi.XyDecorations

	if utils.IsKnown(m.EndZones) {
		decorations.EndZones = utils.Pointer(m.EndZones.ValueBool())
	}
	if utils.IsKnown(m.CurrentTimeMarker) {
		decorations.CurrentTimeMarker = utils.Pointer(m.CurrentTimeMarker.ValueBool())
	}
	if utils.IsKnown(m.PointVisibility) {
		decorations.PointVisibility = utils.Pointer(m.PointVisibility.ValueBool())
	}
	if utils.IsKnown(m.LineInterpolation) {
		interp := kbapi.XyDecorationsLineInterpolation(m.LineInterpolation.ValueString())
		decorations.LineInterpolation = &interp
	}
	if utils.IsKnown(m.MinimumBarHeight) {
		decorations.MinimumBarHeight = utils.Pointer(float32(m.MinimumBarHeight.ValueInt64()))
	}
	if utils.IsKnown(m.ShowValueLabels) {
		decorations.ShowValueLabels = utils.Pointer(m.ShowValueLabels.ValueBool())
	}
	if utils.IsKnown(m.FillOpacity) {
		decorations.FillOpacity = utils.Pointer(float32(m.FillOpacity.ValueFloat64()))
	}
	if utils.IsKnown(m.ValueLabels) {
		decorations.ValueLabels = utils.Pointer(m.ValueLabels.ValueBool())
	}

	return decorations
}

type xyFittingModel struct {
	Type     types.String `tfsdk:"type"`
	Dotted   types.Bool   `tfsdk:"dotted"`
	EndValue types.String `tfsdk:"end_value"`
}

func (m *xyFittingModel) fromAPI(apiFitting kbapi.XyFitting) {
	m.Type = typeutils.StringishValue(apiFitting.Type)
	m.Dotted = types.BoolPointerValue(apiFitting.Dotted)
	m.EndValue = typeutils.StringishPointerValue(apiFitting.EndValue)
}

func (m *xyFittingModel) toAPI() kbapi.XyFitting {
	if m == nil {
		return kbapi.XyFitting{}
	}

	var fitting kbapi.XyFitting

	if utils.IsKnown(m.Type) {
		fitting.Type = kbapi.XyFittingType(m.Type.ValueString())
	}
	if utils.IsKnown(m.Dotted) {
		fitting.Dotted = utils.Pointer(m.Dotted.ValueBool())
	}
	if utils.IsKnown(m.EndValue) {
		endVal := kbapi.XyFittingEndValue(m.EndValue.ValueString())
		fitting.EndValue = &endVal
	}

	return fitting
}

type xyLegendModel struct {
	Visible            types.Bool   `tfsdk:"visible"`
	Statistics         types.List   `tfsdk:"statistics"`
	TruncateAfterLines types.Int64  `tfsdk:"truncate_after_lines"`
	Inside             types.Bool   `tfsdk:"inside"`
	Position           types.String `tfsdk:"position"`
	Size               types.String `tfsdk:"size"`
	Columns            types.Int64  `tfsdk:"columns"`
	Alignment          types.String `tfsdk:"alignment"`
}

func (m *xyLegendModel) fromAPI(ctx context.Context, apiLegend kbapi.XyLegend) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try inside legend first
	legendInside, err := apiLegend.AsXyLegendInside()
	if err == nil && legendInside.Inside {
		m.Inside = types.BoolValue(true)
		m.Visible = types.BoolPointerValue(legendInside.Visible)
		m.Alignment = typeutils.StringishPointerValue(legendInside.Alignment)

		if legendInside.TruncateAfterLines != nil {
			m.TruncateAfterLines = types.Int64Value(int64(*legendInside.TruncateAfterLines))
		} else {
			m.TruncateAfterLines = types.Int64Null()
		}

		if legendInside.Columns != nil {
			m.Columns = types.Int64Value(int64(*legendInside.Columns))
		} else {
			m.Columns = types.Int64Null()
		}

		if legendInside.Statistics != nil {
			stats := make([]types.String, 0, len(*legendInside.Statistics))
			for _, s := range *legendInside.Statistics {
				stats = append(stats, types.StringValue(string(s)))
			}
			var statsDiags diag.Diagnostics
			m.Statistics, statsDiags = types.ListValueFrom(ctx, types.StringType, stats)
			diags.Append(statsDiags...)
		} else {
			m.Statistics = types.ListNull(types.StringType)
		}
		return diags
	}

	// Try outside legend
	legendOutside, err := apiLegend.AsXyLegendOutside()
	if err == nil {
		m.Inside = types.BoolValue(false)
		m.Visible = types.BoolPointerValue(legendOutside.Visible)
		m.Position = typeutils.StringishPointerValue(legendOutside.Position)
		m.Size = typeutils.StringishPointerValue(legendOutside.Size)

		if legendOutside.TruncateAfterLines != nil {
			m.TruncateAfterLines = types.Int64Value(int64(*legendOutside.TruncateAfterLines))
		} else {
			m.TruncateAfterLines = types.Int64Null()
		}

		if legendOutside.Statistics != nil {
			stats := make([]types.String, 0, len(*legendOutside.Statistics))
			for _, s := range *legendOutside.Statistics {
				stats = append(stats, types.StringValue(string(s)))
			}
			var statsDiags diag.Diagnostics
			m.Statistics, statsDiags = types.ListValueFrom(ctx, types.StringType, stats)
			diags.Append(statsDiags...)
		} else {
			m.Statistics = types.ListNull(types.StringType)
		}
	}

	return diags
}

func (m *xyLegendModel) toAPI() (kbapi.XyLegend, diag.Diagnostics) {
	if m == nil {
		return kbapi.XyLegend{}, nil
	}

	var diags diag.Diagnostics
	isInside := utils.IsKnown(m.Inside) && m.Inside.ValueBool()

	if isInside {
		var legend kbapi.XyLegendInside
		legend.Inside = true

		if utils.IsKnown(m.Visible) {
			legend.Visible = utils.Pointer(m.Visible.ValueBool())
		}
		if utils.IsKnown(m.TruncateAfterLines) {
			legend.TruncateAfterLines = utils.Pointer(float32(m.TruncateAfterLines.ValueInt64()))
		}
		if utils.IsKnown(m.Columns) {
			legend.Columns = utils.Pointer(float32(m.Columns.ValueInt64()))
		}
		if utils.IsKnown(m.Alignment) {
			align := kbapi.XyLegendInsideAlignment(m.Alignment.ValueString())
			legend.Alignment = &align
		}

		var result kbapi.XyLegend
		if err := result.FromXyLegendInside(legend); err != nil {
			diags.AddError("Failed to create inside legend", err.Error())
		}
		return result, diags
	}

	// Outside legend
	var legend kbapi.XyLegendOutside

	if utils.IsKnown(m.Visible) {
		legend.Visible = utils.Pointer(m.Visible.ValueBool())
	}
	if utils.IsKnown(m.TruncateAfterLines) {
		legend.TruncateAfterLines = utils.Pointer(float32(m.TruncateAfterLines.ValueInt64()))
	}
	if utils.IsKnown(m.Position) {
		pos := kbapi.XyLegendOutsidePosition(m.Position.ValueString())
		legend.Position = &pos
	}
	if utils.IsKnown(m.Size) {
		size := kbapi.XyLegendOutsideSize(m.Size.ValueString())
		legend.Size = &size
	}

	var result kbapi.XyLegend
	if err := result.FromXyLegendOutside(legend); err != nil {
		diags.AddError("Failed to create outside legend", err.Error())
	}
	return result, diags
}

type filterSimpleModel struct {
	Language types.String `tfsdk:"language"`
	Query    types.String `tfsdk:"query"`
}

func (m *filterSimpleModel) fromAPI(apiQuery kbapi.FilterSimpleSchema) {
	m.Query = types.StringValue(apiQuery.Query)
	m.Language = typeutils.StringishPointerValue(apiQuery.Language)
}

func (m *filterSimpleModel) toAPI() kbapi.FilterSimpleSchema {
	if m == nil {
		return kbapi.FilterSimpleSchema{}
	}

	query := kbapi.FilterSimpleSchema{
		Query: m.Query.ValueString(),
	}
	if utils.IsKnown(m.Language) {
		lang := kbapi.FilterSimpleSchemaLanguage(m.Language.ValueString())
		query.Language = &lang
	}
	return query
}

type searchFilterModel struct {
	Query    types.String         `tfsdk:"query"`
	Meta     jsontypes.Normalized `tfsdk:"meta"`
	Language types.String         `tfsdk:"language"`
}

func (m *searchFilterModel) fromAPI(apiFilter kbapi.SearchFilterSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to extract from SearchFilterSchema0
	filterSchema, err := apiFilter.AsSearchFilterSchema0()
	if err != nil {
		diags.AddError("Failed to extract search filter", err.Error())
		return diags
	}

	// Extract string from union type
	queryStr, queryErr := filterSchema.Query.AsSearchFilterSchema0Query0()
	if queryErr != nil {
		diags.AddError("Failed to extract search filter query", queryErr.Error())
		return diags
	}

	m.Query = types.StringValue(queryStr)
	m.Language = typeutils.StringishPointerValue(filterSchema.Language)

	if filterSchema.Meta != nil {
		metaJSON, err := json.Marshal(filterSchema.Meta)
		if err == nil {
			m.Meta = jsontypes.NewNormalizedValue(string(metaJSON))
		}
	}

	return diags
}

func (m *searchFilterModel) toAPI() (kbapi.SearchFilterSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := kbapi.SearchFilterSchema0{}
	if utils.IsKnown(m.Query) {
		query := m.Query.ValueString()
		var queryUnion kbapi.SearchFilterSchema_0_Query
		if err := queryUnion.FromSearchFilterSchema0Query0(query); err != nil {
			diags.AddError("Failed to create search filter query", err.Error())
			return kbapi.SearchFilterSchema{}, diags
		}
		filter.Query = queryUnion
	}
	if utils.IsKnown(m.Language) {
		lang := kbapi.SearchFilterSchema0Language(m.Language.ValueString())
		filter.Language = &lang
	}

	var result kbapi.SearchFilterSchema
	if err := result.FromSearchFilterSchema0(filter); err != nil {
		diags.AddError("Failed to create search filter", err.Error())
	}
	return result, diags
}

type sectionModel struct {
	Title     types.String     `tfsdk:"title"`
	ID        types.String     `tfsdk:"id"`
	Collapsed types.Bool       `tfsdk:"collapsed"`
	Grid      sectionGridModel `tfsdk:"grid"`
	Panels    []panelModel     `tfsdk:"panels"`
}

type sectionGridModel struct {
	Y types.Int64 `tfsdk:"y"`
}

// toAPI converts the XY chart config model to API schema
func (m *xyChartConfigModel) toAPI() (kbapi.XyChartSchema, diag.Diagnostics) {
	var diags diag.Diagnostics

	xyChart := kbapi.XyChartSchema{
		Type: kbapi.Xy,
	}

	// Convert title and description
	if utils.IsKnown(m.Title) {
		xyChart.Title = utils.Pointer(m.Title.ValueString())
	}
	if utils.IsKnown(m.Description) {
		xyChart.Description = utils.Pointer(m.Description.ValueString())
	}

	// Convert axis
	if m.Axis != nil {
		axis, axisDiags := m.Axis.toAPI()
		diags.Append(axisDiags...)
		xyChart.Axis = axis
	}

	// Convert decorations
	if m.Decorations != nil {
		xyChart.Decorations = m.Decorations.toAPI()
	}

	// Convert fitting
	if m.Fitting != nil {
		xyChart.Fitting = m.Fitting.toAPI()
	}

	// Convert layers
	if len(m.Layers) > 0 {
		layers := make([]kbapi.XyChartSchema_Layers_Item, 0, len(m.Layers))
		for _, layer := range m.Layers {
			apiLayer, layerDiags := layer.toAPI()
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				layers = append(layers, apiLayer)
			}
		}
		if len(layers) > 0 {
			xyChart.Layers = layers
		}
	}

	// Convert legend
	if m.Legend != nil {
		legend, legendDiags := m.Legend.toAPI()
		diags.Append(legendDiags...)
		if !legendDiags.HasError() {
			xyChart.Legend = legend
		}
	}

	// Convert query
	if m.Query != nil {
		xyChart.Query = m.Query.toAPI()
	}

	// Convert filters
	if len(m.Filters) > 0 {
		filters := make([]kbapi.SearchFilterSchema, 0, len(m.Filters))
		for _, f := range m.Filters {
			filter, filterDiags := f.toAPI()
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				filters = append(filters, filter)
			}
		}
		if len(filters) > 0 {
			xyChart.Filters = &filters
		}
	}

	return xyChart, diags
}

// fromAPI populates the XY chart config model from API response
func (m *xyChartConfigModel) fromAPI(ctx context.Context, apiChart kbapi.XyChartSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	// Convert layers
	if len(apiChart.Layers) > 0 {
		m.Layers = make([]xyLayerModel, 0, len(apiChart.Layers))
		for _, apiLayer := range apiChart.Layers {
			layer := xyLayerModel{}
			layerDiags := layer.fromAPI(apiLayer)
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				m.Layers = append(m.Layers, layer)
			}
		}
	}

	// Convert axis
	m.Axis = &xyAxisModel{}
	axisDiags := m.Axis.fromAPI(apiChart.Axis)
	diags.Append(axisDiags...)

	// Convert decorations
	m.Decorations = &xyDecorationsModel{}
	m.Decorations.fromAPI(apiChart.Decorations)

	// Convert fitting
	m.Fitting = &xyFittingModel{}
	m.Fitting.fromAPI(apiChart.Fitting)

	// Convert legend
	m.Legend = &xyLegendModel{}
	legendDiags := m.Legend.fromAPI(ctx, apiChart.Legend)
	diags.Append(legendDiags...)

	// Convert query
	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(apiChart.Query)

	// Convert filters
	if apiChart.Filters != nil && len(*apiChart.Filters) > 0 {
		m.Filters = make([]searchFilterModel, 0, len(*apiChart.Filters))
		for _, f := range *apiChart.Filters {
			filter := searchFilterModel{}
			filterDiags := filter.fromAPI(f)
			diags.Append(filterDiags...)
			if !filterDiags.HasError() {
				m.Filters = append(m.Filters, filter)
			}
		}
	}

	return diags
}
