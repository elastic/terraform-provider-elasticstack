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

package dashboard

import (
	"context"
	"encoding/json"
	"math"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newXYChartPanelConfigConverter() xyChartPanelConfigConverter {
	return xyChartPanelConfigConverter{
		lensVisualizationBase: lensVisualizationBase{
			visualizationType: string(kbapi.XyChartNoESQLTypeXy),
			hasTFPanelConfig:  func(pm panelModel) bool { return pm.XYChartConfig != nil },
		},
	}
}

type xyChartPanelConfigConverter struct {
	lensVisualizationBase
}

func (c xyChartPanelConfigConverter) populateFromAttributes(ctx context.Context, pm *panelModel, attrs kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	pm.XYChartConfig = &xyChartConfigModel{}
	if xyChart, err := attrs.AsXyChartNoESQL(); err == nil {
		return pm.XYChartConfig.fromAPINoESQL(ctx, xyChart)
	}
	xyChart, err := attrs.AsXyChartESQL()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return pm.XYChartConfig.fromAPIESQL(ctx, xyChart)
}

func (c xyChartPanelConfigConverter) buildAttributes(pm panelModel) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs kbapi.KbnDashboardPanelTypeVisConfig0
	configModel := *pm.XYChartConfig

	if configModel.xyUsesESQL() {
		chart, xyDiags := configModel.toAPIESQL()
		diags.Append(xyDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromXyChartESQL(chart); err != nil {
			return attrs, diagutil.FrameworkDiagFromError(err)
		}
		return attrs, diags
	}

	chart, xyDiags := configModel.toAPINoESQL()
	diags.Append(xyDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromXyChartNoESQL(chart); err != nil {
		return attrs, diagutil.FrameworkDiagFromError(err)
	}
	return attrs, diags
}

type xyChartConfigModel struct {
	Title       types.String           `tfsdk:"title"`
	Description types.String           `tfsdk:"description"`
	Axis        *xyAxisModel           `tfsdk:"axis"`
	Decorations *xyDecorationsModel    `tfsdk:"decorations"`
	Fitting     *xyFittingModel        `tfsdk:"fitting"`
	Layers      []xyLayerModel         `tfsdk:"layers"`
	Legend      *xyLegendModel         `tfsdk:"legend"`
	Query       *filterSimpleModel     `tfsdk:"query"`
	Filters     []chartFilterJSONModel `tfsdk:"filters"`
}

type xyAxisModel struct {
	X          *xyAxisConfigModel `tfsdk:"x"`
	Y          *yAxisConfigModel  `tfsdk:"y"`
	SecondaryY *yAxisConfigModel  `tfsdk:"secondary_y"`
}

func (m *xyAxisModel) fromAPI(apiAxis kbapi.VisApiXyAxisConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiAxis.X != nil {
		xBytes, err := json.Marshal(apiAxis.X)
		if err != nil {
			diags.AddError("Failed to marshal XY chart X axis", err.Error())
			return diags
		}
		var xView xyAxisConfigAPIModel
		if err := json.Unmarshal(xBytes, &xView); err != nil {
			diags.AddError("Failed to decode XY chart X axis", err.Error())
			return diags
		}
		m.X = &xyAxisConfigModel{}
		xDiags := m.X.fromAPI(&xView)
		diags.Append(xDiags...)
		if m.X.isEmpty() {
			m.X = nil
		}
	}

	if apiAxis.Y != nil {
		m.Y = &yAxisConfigModel{}
		yDiags := m.Y.fromAPIY(apiAxis.Y)
		diags.Append(yDiags...)
		if m.Y.isEmpty() {
			m.Y = nil
		}
	}

	if apiAxis.SecondaryY != nil {
		m.SecondaryY = &yAxisConfigModel{}
		secondaryYDiags := m.SecondaryY.fromAPISecondaryY(apiAxis.SecondaryY)
		diags.Append(secondaryYDiags...)
		if m.SecondaryY.isEmpty() {
			m.SecondaryY = nil
		}
	}

	return diags
}

func (m *xyAxisModel) toAPI() (kbapi.VisApiXyAxisConfig, diag.Diagnostics) {
	if m == nil {
		return kbapi.VisApiXyAxisConfig{}, nil
	}

	var diags diag.Diagnostics
	var axis kbapi.VisApiXyAxisConfig

	if m.X != nil {
		xAxis, xDiags := m.X.toAPI()
		diags.Append(xDiags...)
		if !xDiags.HasError() && xAxis != nil {
			xb, err := json.Marshal(xAxis)
			if err != nil {
				diags.AddError("Failed to marshal XY X axis model", err.Error())
				return axis, diags
			}
			partial, err := json.Marshal(axis)
			if err != nil {
				diags.AddError("Failed to marshal XY axis envelope", err.Error())
				return axis, diags
			}
			var env map[string]json.RawMessage
			if err := json.Unmarshal(partial, &env); err != nil {
				diags.AddError("Failed to prepare XY axis merge", err.Error())
				return axis, diags
			}
			env["x"] = json.RawMessage(xb)
			merged, err := json.Marshal(env)
			if err != nil {
				diags.AddError("Failed to marshal merged XY axis", err.Error())
				return axis, diags
			}
			if err := json.Unmarshal(merged, &axis); err != nil {
				diags.AddError("Failed to merge XY X axis into API model", err.Error())
				return axis, diags
			}
		}
	}

	if m.Y != nil {
		yAxis, yDiags := m.Y.toAPIY()
		diags.Append(yDiags...)
		axis.Y = yAxis
	}

	if m.SecondaryY != nil {
		secondaryYAxis, secondaryYDiags := m.SecondaryY.toAPISecondaryY()
		diags.Append(secondaryYDiags...)
		axis.SecondaryY = secondaryYAxis
	}

	return axis, diags
}

type xyAxisConfigModel struct {
	Title            *axisTitleModel      `tfsdk:"title"`
	Ticks            types.Bool           `tfsdk:"ticks"`
	Grid             types.Bool           `tfsdk:"grid"`
	LabelOrientation types.String         `tfsdk:"label_orientation"`
	Scale            types.String         `tfsdk:"scale"`
	DomainJSON       jsontypes.Normalized `tfsdk:"domain_json"`
}

func (m *xyAxisConfigModel) isEmpty() bool {
	if m == nil {
		return true
	}
	if typeutils.IsKnown(m.Ticks) || typeutils.IsKnown(m.Grid) || typeutils.IsKnown(m.LabelOrientation) || typeutils.IsKnown(m.Scale) || typeutils.IsKnown(m.DomainJSON) {
		return false
	}
	return axisTitleIsDefault(m.Title)
}

type xyAxisConfigAPIModel = struct {
	Domain *kbapi.VisApiXyAxisConfig_X_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigXScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}

func (m *xyAxisConfigModel) fromAPI(apiAxis *xyAxisConfigAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiAxis == nil {
		return diags
	}

	if apiAxis.Grid != nil {
		m.Grid = types.BoolValue(apiAxis.Grid.Visible)
	} else {
		m.Grid = types.BoolNull()
	}
	if apiAxis.Ticks != nil {
		m.Ticks = types.BoolValue(apiAxis.Ticks.Visible)
	} else {
		m.Ticks = types.BoolNull()
	}
	if apiAxis.Labels != nil {
		m.LabelOrientation = types.StringValue(string(apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	if apiAxis.Domain != nil {
		domainJSON, err := json.Marshal(apiAxis.Domain)
		if err == nil {
			m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
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

	if typeutils.IsKnown(m.Grid) {
		xAxis.Grid = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Grid.ValueBool()}
	}
	if typeutils.IsKnown(m.Ticks) {
		xAxis.Ticks = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Ticks.ValueBool()}
	}
	if typeutils.IsKnown(m.LabelOrientation) {
		xAxis.Labels = &struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		}{Orientation: kbapi.VisApiOrientation(m.LabelOrientation.ValueString())}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.VisApiXyAxisConfigXScale(m.Scale.ValueString())
		xAxis.Scale = &scale
	}
	if m.Title != nil {
		xAxis.Title = m.Title.toAPI()
	}
	if typeutils.IsKnown(m.DomainJSON) {
		var domain kbapi.VisApiXyAxisConfig_X_Domain
		domainDiags := m.DomainJSON.Unmarshal(&domain)
		diags.Append(domainDiags...)
		if !domainDiags.HasError() {
			xAxis.Domain = &domain
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
	DomainJSON       jsontypes.Normalized `tfsdk:"domain_json"`
}

func (m *yAxisConfigModel) isEmpty() bool {
	if m == nil {
		return true
	}
	if typeutils.IsKnown(m.Ticks) || typeutils.IsKnown(m.Grid) || typeutils.IsKnown(m.LabelOrientation) || typeutils.IsKnown(m.Scale) || typeutils.IsKnown(m.DomainJSON) {
		return false
	}
	return axisTitleIsDefault(m.Title)
}

func (m *yAxisConfigModel) fromAPIY(apiAxis *struct {
	Anchor *kbapi.VisApiXyAxisConfigYAnchor  `json:"anchor,omitempty"`
	Domain kbapi.VisApiXyAxisConfig_Y_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigYScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiAxis == nil {
		return diags
	}

	if apiAxis.Grid != nil {
		m.Grid = types.BoolValue(apiAxis.Grid.Visible)
	} else {
		m.Grid = types.BoolNull()
	}
	if apiAxis.Ticks != nil {
		m.Ticks = types.BoolValue(apiAxis.Ticks.Visible)
	} else {
		m.Ticks = types.BoolNull()
	}
	if apiAxis.Labels != nil {
		m.LabelOrientation = types.StringValue(string(apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func (m *yAxisConfigModel) toAPIY() (*struct {
	Anchor *kbapi.VisApiXyAxisConfigYAnchor  `json:"anchor,omitempty"`
	Domain kbapi.VisApiXyAxisConfig_Y_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigYScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	yAxis := &struct {
		Anchor *kbapi.VisApiXyAxisConfigYAnchor  `json:"anchor,omitempty"`
		Domain kbapi.VisApiXyAxisConfig_Y_Domain `json:"domain"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		} `json:"labels,omitempty"`
		Scale *kbapi.VisApiXyAxisConfigYScale `json:"scale,omitempty"`
		Ticks *struct {
			Visible bool `json:"visible"`
		} `json:"ticks,omitempty"`
		Title *struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		} `json:"title,omitempty"`
	}{}

	if typeutils.IsKnown(m.Grid) {
		yAxis.Grid = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Grid.ValueBool()}
	}
	if typeutils.IsKnown(m.Ticks) {
		yAxis.Ticks = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Ticks.ValueBool()}
	}
	if typeutils.IsKnown(m.LabelOrientation) {
		yAxis.Labels = &struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		}{Orientation: kbapi.VisApiOrientation(m.LabelOrientation.ValueString())}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.VisApiXyAxisConfigYScale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = m.Title.toAPI()
	}
	if typeutils.IsKnown(m.DomainJSON) {
		domainDiags := m.DomainJSON.Unmarshal(&yAxis.Domain)
		diags.Append(domainDiags...)
	}

	return yAxis, diags
}

func (m *yAxisConfigModel) fromAPISecondaryY(apiAxis *struct {
	Anchor *kbapi.VisApiXyAxisConfigSecondaryYAnchor  `json:"anchor,omitempty"`
	Domain kbapi.VisApiXyAxisConfig_SecondaryY_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigSecondaryYScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if apiAxis == nil {
		return diags
	}

	if apiAxis.Grid != nil {
		m.Grid = types.BoolValue(apiAxis.Grid.Visible)
	} else {
		m.Grid = types.BoolNull()
	}
	if apiAxis.Ticks != nil {
		m.Ticks = types.BoolValue(apiAxis.Ticks.Visible)
	} else {
		m.Ticks = types.BoolNull()
	}
	if apiAxis.Labels != nil {
		m.LabelOrientation = types.StringValue(string(apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &axisTitleModel{}
		m.Title.fromAPI(apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func (m *yAxisConfigModel) toAPISecondaryY() (*struct {
	Anchor *kbapi.VisApiXyAxisConfigSecondaryYAnchor  `json:"anchor,omitempty"`
	Domain kbapi.VisApiXyAxisConfig_SecondaryY_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigSecondaryYScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	yAxis := &struct {
		Anchor *kbapi.VisApiXyAxisConfigSecondaryYAnchor  `json:"anchor,omitempty"`
		Domain kbapi.VisApiXyAxisConfig_SecondaryY_Domain `json:"domain"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		} `json:"labels,omitempty"`
		Scale *kbapi.VisApiXyAxisConfigSecondaryYScale `json:"scale,omitempty"`
		Ticks *struct {
			Visible bool `json:"visible"`
		} `json:"ticks,omitempty"`
		Title *struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		} `json:"title,omitempty"`
	}{}

	if typeutils.IsKnown(m.Grid) {
		yAxis.Grid = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Grid.ValueBool()}
	}
	if typeutils.IsKnown(m.Ticks) {
		yAxis.Ticks = &struct {
			Visible bool `json:"visible"`
		}{Visible: m.Ticks.ValueBool()}
	}
	if typeutils.IsKnown(m.LabelOrientation) {
		yAxis.Labels = &struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		}{Orientation: kbapi.VisApiOrientation(m.LabelOrientation.ValueString())}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.VisApiXyAxisConfigSecondaryYScale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = m.Title.toAPI()
	}
	if typeutils.IsKnown(m.DomainJSON) {
		domainDiags := m.DomainJSON.Unmarshal(&yAxis.Domain)
		diags.Append(domainDiags...)
	}

	return yAxis, diags
}

type axisTitleModel struct {
	Value   types.String `tfsdk:"value"`
	Visible types.Bool   `tfsdk:"visible"`
}

func axisTitleIsDefault(title *axisTitleModel) bool {
	if title == nil {
		return true
	}
	if typeutils.IsKnown(title.Value) {
		return false
	}
	if typeutils.IsKnown(title.Visible) {
		return title.Visible.ValueBool()
	}
	return true
}

func (m *axisTitleModel) fromAPI(apiTitle *struct {
	Text    *string `json:"text,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
}) {
	if apiTitle == nil {
		return
	}
	m.Value = types.StringPointerValue(apiTitle.Text)
	m.Visible = types.BoolPointerValue(apiTitle.Visible)
}

func (m *axisTitleModel) toAPI() *struct {
	Text    *string `json:"text,omitempty"`
	Visible *bool   `json:"visible,omitempty"`
} {
	if m == nil {
		return nil
	}

	title := &struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	}{}

	if typeutils.IsKnown(m.Value) {
		title.Text = new(m.Value.ValueString())
	}
	if typeutils.IsKnown(m.Visible) {
		title.Visible = new(m.Visible.ValueBool())
	}

	return title
}

type xyDecorationsModel struct {
	ShowEndZones          types.Bool    `tfsdk:"show_end_zones"`
	ShowCurrentTimeMarker types.Bool    `tfsdk:"show_current_time_marker"`
	PointVisibility       types.String  `tfsdk:"point_visibility"`
	LineInterpolation     types.String  `tfsdk:"line_interpolation"`
	MinimumBarHeight      types.Int64   `tfsdk:"minimum_bar_height"`
	ShowValueLabels       types.Bool    `tfsdk:"show_value_labels"`
	FillOpacity           types.Float64 `tfsdk:"fill_opacity"`
}

func (m *xyDecorationsModel) readFromStyling(s kbapi.XyStyling) {
	if s.Overlays.PartialBuckets != nil && s.Overlays.PartialBuckets.Visible != nil {
		m.ShowEndZones = types.BoolValue(*s.Overlays.PartialBuckets.Visible)
	} else {
		m.ShowEndZones = types.BoolNull()
	}
	if s.Overlays.CurrentTimeMarker != nil && s.Overlays.CurrentTimeMarker.Visible != nil {
		m.ShowCurrentTimeMarker = types.BoolValue(*s.Overlays.CurrentTimeMarker.Visible)
	} else {
		m.ShowCurrentTimeMarker = types.BoolNull()
	}
	if s.Points.Visibility != nil {
		switch *s.Points.Visibility {
		case kbapi.XyStylingPointsVisibilityHidden:
			m.PointVisibility = types.StringValue("never")
		case kbapi.XyStylingPointsVisibilityVisible:
			m.PointVisibility = types.StringValue("always")
		default:
			m.PointVisibility = types.StringValue("auto")
		}
	} else {
		m.PointVisibility = types.StringNull()
	}
	if s.Interpolation != nil {
		m.LineInterpolation = types.StringValue(string(*s.Interpolation))
	} else {
		m.LineInterpolation = types.StringNull()
	}
	if s.Bars.MinimumHeight != nil {
		m.MinimumBarHeight = types.Int64Value(int64(*s.Bars.MinimumHeight))
	} else {
		m.MinimumBarHeight = types.Int64Null()
	}
	if s.Bars.DataLabels != nil && s.Bars.DataLabels.Visible != nil {
		m.ShowValueLabels = types.BoolValue(*s.Bars.DataLabels.Visible)
	} else {
		m.ShowValueLabels = types.BoolNull()
	}
	if s.Areas.FillOpacity != nil {
		val := float64(*s.Areas.FillOpacity)
		m.FillOpacity = types.Float64Value(math.Round(val*100) / 100)
	} else {
		m.FillOpacity = types.Float64Null()
	}
}

func (m *xyDecorationsModel) writeToStyling(s *kbapi.XyStyling) {
	if m == nil {
		return
	}
	if typeutils.IsKnown(m.ShowEndZones) {
		v := m.ShowEndZones.ValueBool()
		s.Overlays.PartialBuckets = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.ShowCurrentTimeMarker) {
		v := m.ShowCurrentTimeMarker.ValueBool()
		s.Overlays.CurrentTimeMarker = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.PointVisibility) {
		switch m.PointVisibility.ValueString() {
		case "never":
			v := kbapi.XyStylingPointsVisibilityHidden
			s.Points.Visibility = &v
		case "always":
			v := kbapi.XyStylingPointsVisibilityVisible
			s.Points.Visibility = &v
		default:
			v := kbapi.XyStylingPointsVisibilityAuto
			s.Points.Visibility = &v
		}
	}
	if typeutils.IsKnown(m.LineInterpolation) {
		interp := kbapi.XyStylingInterpolation(m.LineInterpolation.ValueString())
		s.Interpolation = &interp
	}
	if typeutils.IsKnown(m.MinimumBarHeight) {
		s.Bars.MinimumHeight = new(float32(m.MinimumBarHeight.ValueInt64()))
	}
	if typeutils.IsKnown(m.ShowValueLabels) {
		v := m.ShowValueLabels.ValueBool()
		s.Bars.DataLabels = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.FillOpacity) {
		s.Areas.FillOpacity = new(float32(m.FillOpacity.ValueFloat64()))
	}
}

type xyFittingModel struct {
	Type     types.String `tfsdk:"type"`
	Dotted   types.Bool   `tfsdk:"dotted"`
	EndValue types.String `tfsdk:"end_value"`
}

func (m *xyFittingModel) fromAPI(apiFitting kbapi.XyFitting) {
	m.Type = typeutils.StringishValue(apiFitting.Type)
	m.Dotted = types.BoolPointerValue(apiFitting.Emphasize)
	if apiFitting.Extend != nil {
		m.EndValue = types.StringValue(string(*apiFitting.Extend))
	} else {
		m.EndValue = types.StringNull()
	}
}

func (m *xyFittingModel) toAPI() kbapi.XyFitting {
	out := kbapi.XyFitting{Type: kbapi.XyFittingTypeNone}
	if m == nil {
		return out
	}
	if typeutils.IsKnown(m.Type) {
		out.Type = kbapi.XyFittingType(m.Type.ValueString())
	}
	if typeutils.IsKnown(m.Dotted) {
		out.Emphasize = new(m.Dotted.ValueBool())
	}
	if typeutils.IsKnown(m.EndValue) {
		ext := kbapi.XyFittingExtend(m.EndValue.ValueString())
		out.Extend = &ext
	}
	return out
}

type xyLegendModel struct {
	Visibility         types.String `tfsdk:"visibility"`
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
	m.Position = types.StringNull()
	m.Size = types.StringNull()
	m.Columns = types.Int64Null()
	m.TruncateAfterLines = types.Int64Null()
	m.Alignment = types.StringNull()
	m.Statistics = types.ListNull(types.StringType)

	// Try inside legend first
	legendInside, err := apiLegend.AsXyLegendInside()
	if err == nil && legendInside.Placement == kbapi.XyLegendInsidePlacementInside {
		m.Inside = types.BoolValue(true)
		m.Visibility = typeutils.StringishPointerValue(legendInside.Visibility)
		m.Alignment = typeutils.StringishPointerValue(legendInside.Position)

		if legendInside.Layout != nil && legendInside.Layout.Truncate != nil && legendInside.Layout.Truncate.MaxLines != nil {
			m.TruncateAfterLines = types.Int64Value(int64(*legendInside.Layout.Truncate.MaxLines))
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

	// Try outside vertical legend first since it carries required size information.
	legendOutsideVertical, err := apiLegend.AsXyLegendOutsideVertical()
	if err == nil &&
		legendOutsideVertical.Placement != nil &&
		*legendOutsideVertical.Placement == kbapi.XyLegendOutsideVerticalPlacementOutside &&
		(legendOutsideVertical.Position == nil ||
			*legendOutsideVertical.Position == kbapi.Left ||
			*legendOutsideVertical.Position == kbapi.Right) &&
		legendOutsideVertical.Size != "" {
		m.Inside = types.BoolValue(false)
		m.Visibility = typeutils.StringishPointerValue(legendOutsideVertical.Visibility)
		m.Position = typeutils.StringishPointerValue(legendOutsideVertical.Position)
		m.Size = types.StringValue(string(legendOutsideVertical.Size))

		if legendOutsideVertical.Layout != nil && legendOutsideVertical.Layout.Truncate != nil && legendOutsideVertical.Layout.Truncate.MaxLines != nil {
			m.TruncateAfterLines = types.Int64Value(int64(*legendOutsideVertical.Layout.Truncate.MaxLines))
		}

		if legendOutsideVertical.Statistics != nil {
			stats := make([]types.String, 0, len(*legendOutsideVertical.Statistics))
			for _, s := range *legendOutsideVertical.Statistics {
				stats = append(stats, types.StringValue(string(s)))
			}
			var statsDiags diag.Diagnostics
			m.Statistics, statsDiags = types.ListValueFrom(ctx, types.StringType, stats)
			diags.Append(statsDiags...)
		}
		return diags
	}

	// Try outside horizontal legend
	legendOutsideHorizontal, err := apiLegend.AsXyLegendOutsideHorizontal()
	if err == nil {
		m.Inside = types.BoolValue(false)
		m.Visibility = typeutils.StringishPointerValue(legendOutsideHorizontal.Visibility)
		m.Position = typeutils.StringishPointerValue(legendOutsideHorizontal.Position)

		if legendOutsideHorizontal.Layout != nil {
			if layout, layoutErr := legendOutsideHorizontal.Layout.AsXyLegendOutsideHorizontalLayout0(); layoutErr == nil &&
				layout.Truncate != nil && layout.Truncate.MaxLines != nil {
				m.TruncateAfterLines = types.Int64Value(int64(*layout.Truncate.MaxLines))
			}
		}

		if legendOutsideHorizontal.Statistics != nil {
			stats := make([]types.String, 0, len(*legendOutsideHorizontal.Statistics))
			for _, s := range *legendOutsideHorizontal.Statistics {
				stats = append(stats, types.StringValue(string(s)))
			}
			var statsDiags diag.Diagnostics
			m.Statistics, statsDiags = types.ListValueFrom(ctx, types.StringType, stats)
			diags.Append(statsDiags...)
		}
		return diags
	}

	return diags
}

func (m *xyLegendModel) toAPI() (kbapi.XyLegend, diag.Diagnostics) {
	if m == nil {
		return kbapi.XyLegend{}, nil
	}

	var diags diag.Diagnostics
	isInside := typeutils.IsKnown(m.Inside) && m.Inside.ValueBool()
	insideVisibility := kbapi.XyLegendInsideVisibilityAuto
	outsideHorizontalVisibility := kbapi.XyLegendOutsideHorizontalVisibilityAuto
	outsideVerticalVisibility := kbapi.XyLegendOutsideVerticalVisibilityAuto
	if typeutils.IsKnown(m.Visibility) {
		insideVisibility = kbapi.XyLegendInsideVisibility(m.Visibility.ValueString())
		outsideHorizontalVisibility = kbapi.XyLegendOutsideHorizontalVisibility(m.Visibility.ValueString())
		outsideVerticalVisibility = kbapi.XyLegendOutsideVerticalVisibility(m.Visibility.ValueString())
	}
	statsElemsToStrings := func() ([]string, bool) {
		if !typeutils.IsKnown(m.Statistics) {
			return nil, false
		}

		elems := m.Statistics.Elements()
		if len(elems) == 0 {
			return nil, false
		}

		stats := make([]string, 0, len(elems))
		for _, elem := range elems {
			strVal, ok := elem.(types.String)
			if !ok {
				diags.AddError("Invalid legend statistic value", "Expected statistics element to be a string")
				return nil, false
			}
			if !typeutils.IsKnown(strVal) {
				diags.AddError("Invalid legend statistic value", "Statistics element must be known")
				return nil, false
			}
			stats = append(stats, strVal.ValueString())
		}

		return stats, true
	}

	if isInside {
		var legend kbapi.XyLegendInside
		legend.Placement = kbapi.XyLegendInsidePlacementInside
		legend.Visibility = &insideVisibility

		if typeutils.IsKnown(m.TruncateAfterLines) {
			legend.Layout = &struct {
				Truncate *struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				} `json:"truncate,omitempty"`
				Type kbapi.XyLegendInsideLayoutType `json:"type"`
			}{
				Truncate: &struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				}{
					MaxLines: new(float32(m.TruncateAfterLines.ValueInt64())),
				},
				Type: kbapi.XyLegendInsideLayoutTypeGrid,
			}
		}
		if typeutils.IsKnown(m.Columns) {
			legend.Columns = new(float32(m.Columns.ValueInt64()))
		}
		if typeutils.IsKnown(m.Alignment) {
			pos := kbapi.XyLegendInsidePosition(m.Alignment.ValueString())
			legend.Position = &pos
		}
		if stats, ok := statsElemsToStrings(); ok {
			statsAPI := make([]kbapi.XyLegendInsideStatistics, 0, len(stats))
			for _, s := range stats {
				statsAPI = append(statsAPI, kbapi.XyLegendInsideStatistics(s))
			}
			legend.Statistics = &statsAPI
		}

		var result kbapi.XyLegend
		if err := result.FromXyLegendInside(legend); err != nil {
			diags.AddError("Failed to create inside legend", err.Error())
		}
		return result, diags
	}

	outsidePosition := ""
	if typeutils.IsKnown(m.Position) {
		outsidePosition = m.Position.ValueString()
	}
	isHorizontal := outsidePosition == "top" || outsidePosition == "bottom"

	var result kbapi.XyLegend
	if isHorizontal {
		var legend kbapi.XyLegendOutsideHorizontal
		placement := kbapi.XyLegendOutsideHorizontalPlacementOutside
		legend.Placement = &placement
		legend.Visibility = &outsideHorizontalVisibility
		if outsidePosition != "" {
			pos := kbapi.XyLegendOutsideHorizontalPosition(outsidePosition)
			legend.Position = &pos
		}
		if typeutils.IsKnown(m.TruncateAfterLines) {
			layout := kbapi.XyLegendOutsideHorizontal_Layout{}
			if err := layout.FromXyLegendOutsideHorizontalLayout0(kbapi.XyLegendOutsideHorizontalLayout0{
				Truncate: &struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				}{
					MaxLines: new(float32(m.TruncateAfterLines.ValueInt64())),
				},
				Type: kbapi.XyLegendOutsideHorizontalLayout0TypeGrid,
			}); err != nil {
				diags.AddError("Failed to create horizontal legend layout", err.Error())
				return result, diags
			}
			legend.Layout = &layout
		}
		if stats, ok := statsElemsToStrings(); ok {
			statsAPI := make([]kbapi.XyLegendOutsideHorizontalStatistics, 0, len(stats))
			for _, s := range stats {
				statsAPI = append(statsAPI, kbapi.XyLegendOutsideHorizontalStatistics(s))
			}
			legend.Statistics = &statsAPI
		}
		if err := result.FromXyLegendOutsideHorizontal(legend); err != nil {
			diags.AddError("Failed to create outside horizontal legend", err.Error())
		}
		return result, diags
	}

	var legend kbapi.XyLegendOutsideVertical
	placement := kbapi.XyLegendOutsideVerticalPlacementOutside
	legend.Placement = &placement
	legend.Visibility = &outsideVerticalVisibility
	if outsidePosition != "" {
		pos := kbapi.XyLegendOutsideVerticalPosition(outsidePosition)
		legend.Position = &pos
	}
	if typeutils.IsKnown(m.Size) {
		legend.Size = kbapi.LegendSize(m.Size.ValueString())
	} else {
		legend.Size = kbapi.LegendSizeM
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		legend.Layout = &struct {
			Truncate *struct {
				Enabled  *bool    `json:"enabled,omitempty"`
				MaxLines *float32 `json:"max_lines,omitempty"`
			} `json:"truncate,omitempty"`
			Type kbapi.XyLegendOutsideVerticalLayoutType `json:"type"`
		}{
			Truncate: &struct {
				Enabled  *bool    `json:"enabled,omitempty"`
				MaxLines *float32 `json:"max_lines,omitempty"`
			}{
				MaxLines: new(float32(m.TruncateAfterLines.ValueInt64())),
			},
			Type: kbapi.Grid,
		}
	}
	if stats, ok := statsElemsToStrings(); ok {
		statsAPI := make([]kbapi.XyLegendOutsideVerticalStatistics, 0, len(stats))
		for _, s := range stats {
			statsAPI = append(statsAPI, kbapi.XyLegendOutsideVerticalStatistics(s))
		}
		legend.Statistics = &statsAPI
	}
	if err := result.FromXyLegendOutsideVertical(legend); err != nil {
		diags.AddError("Failed to create outside vertical legend", err.Error())
	}
	return result, diags
}

func (m *xyChartConfigModel) xyUsesESQL() bool {
	if m == nil {
		return false
	}
	for _, layer := range m.Layers {
		if layer.DataLayer != nil && dataSourceJSONIsESQL(layer.DataLayer.DataSourceJSON) {
			return true
		}
		if layer.ReferenceLineLayer != nil && dataSourceJSONIsESQL(layer.ReferenceLineLayer.DataSourceJSON) {
			return true
		}
	}
	return false
}

func dataSourceJSONIsESQL(j jsontypes.Normalized) bool {
	if !typeutils.IsKnown(j) || j.IsNull() {
		return false
	}
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(j.ValueString()), &probe); err != nil {
		return false
	}
	return probe.Type == "esql" || probe.Type == "table"
}

func (m *xyChartConfigModel) stylingToAPI() kbapi.XyStyling {
	fit := kbapi.XyFitting{Type: kbapi.XyFittingTypeNone}
	if m.Fitting != nil {
		fit = m.Fitting.toAPI()
	}
	s := kbapi.XyStyling{
		Areas:    kbapi.XyStylingAreas{},
		Bars:     kbapi.XyStylingBars{},
		Fitting:  fit,
		Overlays: kbapi.XyStylingOverlays{},
		Points:   kbapi.XyStylingPoints{},
	}
	if m.Decorations != nil {
		m.Decorations.writeToStyling(&s)
	}
	return s
}

// toAPINoESQL converts the XY chart config model to a non-ES|QL API payload.
func (m *xyChartConfigModel) toAPINoESQL() (kbapi.XyChartNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.XyChartNoESQL{Type: kbapi.XyChartNoESQLTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = new(m.Description.ValueString())
	}

	if m.Axis != nil {
		axis, axisDiags := m.Axis.toAPI()
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = m.stylingToAPI()
	chart.TimeRange = lensPanelTimeRange()

	if len(m.Layers) > 0 {
		layers := make([]kbapi.XyLayersNoESQL, 0, len(m.Layers))
		for _, layer := range m.Layers {
			apiLayer, layerDiags := layer.toAPILayersNoESQL()
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				layers = append(layers, apiLayer)
			}
		}
		if len(layers) > 0 {
			chart.Layers = layers
		}
	}

	if m.Legend != nil {
		legend, legendDiags := m.Legend.toAPI()
		diags.Append(legendDiags...)
		if !legendDiags.HasError() {
			chart.Legend = legend
		}
	}

	if m.Query != nil {
		chart.Query = m.Query.toAPI()
	}

	chart.Filters = buildFiltersForAPI(m.Filters, &diags)
	return chart, diags
}

// toAPIESQL converts the XY chart config model to an ES|QL API payload.
func (m *xyChartConfigModel) toAPIESQL() (kbapi.XyChartESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.XyChartESQL{Type: kbapi.XyChartESQLTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = new(m.Description.ValueString())
	}

	if m.Axis != nil {
		axis, axisDiags := m.Axis.toAPI()
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = m.stylingToAPI()
	chart.TimeRange = lensPanelTimeRange()

	if len(m.Layers) > 0 {
		layers := make([]kbapi.XyLayerESQL, 0, len(m.Layers))
		for _, layer := range m.Layers {
			apiLayer, layerDiags := layer.toAPILayerESQL()
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				layers = append(layers, apiLayer)
			}
		}
		if len(layers) > 0 {
			chart.Layers = layers
		}
	}

	if m.Legend != nil {
		legend, legendDiags := m.Legend.toAPI()
		diags.Append(legendDiags...)
		if !legendDiags.HasError() {
			chart.Legend = legend
		}
	}

	chart.Filters = buildFiltersForAPI(m.Filters, &diags)
	return chart, diags
}

func (m *xyChartConfigModel) fromAPINoESQL(ctx context.Context, apiChart kbapi.XyChartNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	if len(apiChart.Layers) > 0 {
		m.Layers = make([]xyLayerModel, 0, len(apiChart.Layers))
		for _, apiLayer := range apiChart.Layers {
			layer := xyLayerModel{}
			layerDiags := layer.fromAPILayersNoESQL(apiLayer)
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				m.Layers = append(m.Layers, layer)
			}
		}
	}

	m.Axis = &xyAxisModel{}
	axisDiags := m.Axis.fromAPI(apiChart.Axis)
	diags.Append(axisDiags...)

	m.Decorations = &xyDecorationsModel{}
	m.Decorations.readFromStyling(apiChart.Styling)

	m.Fitting = &xyFittingModel{}
	m.Fitting.fromAPI(apiChart.Styling.Fitting)

	m.Legend = &xyLegendModel{}
	legendDiags := m.Legend.fromAPI(ctx, apiChart.Legend)
	diags.Append(legendDiags...)

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(apiChart.Query)

	m.Filters = populateFiltersFromAPI(apiChart.Filters, &diags)
	return diags
}

func (m *xyChartConfigModel) fromAPIESQL(ctx context.Context, apiChart kbapi.XyChartESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	if len(apiChart.Layers) > 0 {
		m.Layers = make([]xyLayerModel, 0, len(apiChart.Layers))
		for _, apiLayer := range apiChart.Layers {
			layer := xyLayerModel{}
			layerDiags := layer.fromAPILayerESQL(apiLayer)
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				m.Layers = append(m.Layers, layer)
			}
		}
	}

	m.Axis = &xyAxisModel{}
	axisDiags := m.Axis.fromAPI(apiChart.Axis)
	diags.Append(axisDiags...)

	m.Decorations = &xyDecorationsModel{}
	m.Decorations.readFromStyling(apiChart.Styling)

	m.Fitting = &xyFittingModel{}
	m.Fitting.fromAPI(apiChart.Styling.Fitting)

	m.Legend = &xyLegendModel{}
	legendDiags := m.Legend.fromAPI(ctx, apiChart.Legend)
	diags.Append(legendDiags...)

	m.Query = &filterSimpleModel{}
	m.Query.fromAPI(kbapi.FilterSimple{})

	m.Filters = populateFiltersFromAPI(apiChart.Filters, &diags)
	return diags
}

// alignXYChartStateFromPlanPanels preserves practitioner intent for XY charts when Kibana
// injects implicit defaults on read or omits configured fields from the response.
func alignXYChartStateFromPlanPanels(planPanels, statePanels []panelModel) {
	n := min(len(statePanels), len(planPanels))
	for i := range n {
		pp, sp := planPanels[i].XYChartConfig, statePanels[i].XYChartConfig
		if pp == nil || sp == nil {
			continue
		}
		alignXYChartStateFromPlan(pp, sp)
	}
}

func alignXYChartStateFromPlan(plan, state *xyChartConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveKnownStringIfStateBlank(plan.Title, &state.Title)
	preserveKnownStringIfStateBlank(plan.Description, &state.Description)

	alignXYAxisStateFromPlan(plan.Axis, state.Axis)
	alignXYDecorationsStateFromPlan(plan.Decorations, state.Decorations)
	alignXYLegendStateFromPlan(plan.Legend, state.Legend)
	alignXYLayerStateFromPlan(plan.Layers, state.Layers)
}

func alignXYAxisStateFromPlan(plan, state *xyAxisModel) {
	if plan == nil || state == nil {
		return
	}

	alignXYXAxisStateFromPlan(plan.X, state.X)
	alignXYYAxisStateFromPlan(plan.Y, state.Y)

	if plan.SecondaryY != nil && state.SecondaryY == nil {
		state.SecondaryY = cloneYAxisConfigModel(plan.SecondaryY)
		return
	}
	alignXYSecondaryYAxisStateFromPlan(plan.SecondaryY, state.SecondaryY)
}

func alignXYXAxisStateFromPlan(plan, state *xyAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	preserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	preserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	preserveNullStringIfStateEquals(plan.Scale, &state.Scale, string(kbapi.VisApiXyAxisConfigXScaleOrdinal))
	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	preserveNullJSONIfStateMatches(plan.DomainJSON, &state.DomainJSON, `{"type":"fit","rounding":false}`)
	preservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYYAxisStateFromPlan(plan, state *yAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullBoolIfStateEquals(plan.Grid, &state.Grid, true)
	preserveNullBoolIfStateEquals(plan.Ticks, &state.Ticks, true)
	preserveNullStringIfStateEquals(plan.LabelOrientation, &state.LabelOrientation, "horizontal")
	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	preservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYSecondaryYAxisStateFromPlan(plan, state *yAxisConfigModel) {
	if plan == nil || state == nil {
		return
	}

	preserveKnownBoolIfStateNull(plan.Grid, &state.Grid)
	preserveKnownBoolIfStateNull(plan.Ticks, &state.Ticks)
	preserveKnownStringIfStateNull(plan.LabelOrientation, &state.LabelOrientation)
	preserveKnownStringIfStateNull(plan.Scale, &state.Scale)
	preserveKnownAxisTitleIfStateBlank(plan.Title, &state.Title)
	preservePlanJSONIfStateAddsOptionalKeys(plan.DomainJSON, &state.DomainJSON, "rounding")
}

func alignXYDecorationsStateFromPlan(plan, state *xyDecorationsModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullBoolIfStateEquals(plan.ShowEndZones, &state.ShowEndZones, false)
	preserveNullBoolIfStateEquals(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker, false)
	preserveNullStringIfStateEquals(plan.PointVisibility, &state.PointVisibility, "auto")
	preserveNullStringIfStateEquals(plan.LineInterpolation, &state.LineInterpolation, "linear")
	preserveKnownBoolIfStateNull(plan.ShowEndZones, &state.ShowEndZones)
	preserveKnownBoolIfStateNull(plan.ShowCurrentTimeMarker, &state.ShowCurrentTimeMarker)
	preserveKnownStringIfStateNull(plan.PointVisibility, &state.PointVisibility)
	preserveKnownStringIfStateNull(plan.LineInterpolation, &state.LineInterpolation)
	preserveKnownInt64IfStateNull(plan.MinimumBarHeight, &state.MinimumBarHeight)
	preserveKnownBoolIfStateNull(plan.ShowValueLabels, &state.ShowValueLabels)
	preserveKnownFloat64IfStateNull(plan.FillOpacity, &state.FillOpacity)
}

func alignXYLegendStateFromPlan(plan, state *xyLegendModel) {
	if plan == nil || state == nil {
		return
	}

	preserveNullInt64IfStateEquals(plan.TruncateAfterLines, &state.TruncateAfterLines, 1)
	preserveKnownStringIfStateNull(plan.Visibility, &state.Visibility)
	preserveKnownBoolIfStateNull(plan.Inside, &state.Inside)
	preserveKnownStringIfStateNull(plan.Position, &state.Position)
	preserveKnownStringIfStateNull(plan.Size, &state.Size)
	preserveKnownInt64IfStateNull(plan.Columns, &state.Columns)
	preserveKnownStringIfStateNull(plan.Alignment, &state.Alignment)
}

func alignXYLayerStateFromPlan(planLayers, stateLayers []xyLayerModel) {
	n := min(len(stateLayers), len(planLayers))
	for i := range n {
		planLayer, stateLayer := planLayers[i], &stateLayers[i]
		if planLayer.DataLayer != nil && stateLayer.DataLayer != nil {
			preservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.DataSourceJSON, &stateLayer.DataLayer.DataSourceJSON, "time_field")
			preservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.XJSON, &stateLayer.DataLayer.XJSON)
			preservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.BreakdownByJSON, &stateLayer.DataLayer.BreakdownByJSON)

			m := min(len(stateLayer.DataLayer.Y), len(planLayer.DataLayer.Y))
			for j := range m {
				preservePlanJSONIfStateAddsOptionalKeys(planLayer.DataLayer.Y[j].ConfigJSON, &stateLayer.DataLayer.Y[j].ConfigJSON, "axis_id")
			}
		}

		if planLayer.ReferenceLineLayer == nil || stateLayer.ReferenceLineLayer == nil {
			continue
		}

		preservePlanJSONIfStateAddsOptionalKeys(planLayer.ReferenceLineLayer.DataSourceJSON, &stateLayer.ReferenceLineLayer.DataSourceJSON, "time_field")
		m := min(len(stateLayer.ReferenceLineLayer.Thresholds), len(planLayer.ReferenceLineLayer.Thresholds))
		for j := range m {
			preservePlanJSONIfStateAddsOptionalKeys(planLayer.ReferenceLineLayer.Thresholds[j].ValueJSON, &stateLayer.ReferenceLineLayer.Thresholds[j].ValueJSON, "axis_id")
		}
	}
}

func preserveKnownStringIfStateBlank(plan types.String, state *types.String) {
	if !typeutils.IsKnown(plan) {
		return
	}
	if state.IsNull() || state.IsUnknown() || state.ValueString() == "" {
		*state = plan
	}
}

func preserveKnownAxisTitleIfStateBlank(plan *axisTitleModel, state **axisTitleModel) {
	if plan == nil {
		return
	}
	if *state == nil {
		*state = cloneAxisTitleModel(plan)
		return
	}

	preserveKnownStringIfStateBlank(plan.Value, &(*state).Value)
	preserveKnownBoolIfStateNull(plan.Visible, &(*state).Visible)
}

func preserveKnownStringIfStateNull(plan types.String, state *types.String) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownBoolIfStateNull(plan types.Bool, state *types.Bool) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownInt64IfStateNull(plan types.Int64, state *types.Int64) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveKnownFloat64IfStateNull(plan types.Float64, state *types.Float64) {
	if typeutils.IsKnown(plan) && (state.IsNull() || state.IsUnknown()) {
		*state = plan
	}
}

func preserveNullStringIfStateEquals(plan types.String, state *types.String, expected string) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueString() == expected {
		*state = plan
	}
}

func preserveNullBoolIfStateEquals(plan types.Bool, state *types.Bool, expected bool) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueBool() == expected {
		*state = plan
	}
}

func preserveNullInt64IfStateEquals(plan types.Int64, state *types.Int64, expected int64) {
	if !plan.IsNull() || plan.IsUnknown() {
		return
	}
	if typeutils.IsKnown(*state) && state.ValueInt64() == expected {
		*state = plan
	}
}

func preserveNullJSONIfStateMatches(plan jsontypes.Normalized, state *jsontypes.Normalized, expected string) {
	if !plan.IsNull() || plan.IsUnknown() || !typeutils.IsKnown(*state) {
		return
	}
	expectedNormalized := jsontypes.NewNormalizedValue(expected)
	if state.ValueString() == expectedNormalized.ValueString() {
		*state = plan
	}
}

func preservePlanJSONIfStateAddsOptionalKeys(plan jsontypes.Normalized, state *jsontypes.Normalized, optionalKeys ...string) {
	if !typeutils.IsKnown(plan) || !typeutils.IsKnown(*state) {
		return
	}

	var planObj map[string]any
	if err := json.Unmarshal([]byte(plan.ValueString()), &planObj); err != nil {
		return
	}
	var stateObj map[string]any
	if err := json.Unmarshal([]byte(state.ValueString()), &stateObj); err != nil {
		return
	}

	for _, key := range optionalKeys {
		if _, hasPlan := planObj[key]; hasPlan {
			continue
		}
		delete(stateObj, key)
	}

	stateNormalized := normalizeXYPlanComparisonJSON(stateObj)
	planNormalized := normalizeXYPlanComparisonJSON(planObj)
	if reflect.DeepEqual(stateNormalized, planNormalized) {
		*state = plan
	}
}

func normalizeXYPlanComparisonJSON(value any) any {
	switch t := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for key, value := range t {
			out[key] = normalizeXYPlanComparisonJSON(value)
		}
		if formatValue, ok := out["format"]; ok {
			if formatMap, ok := formatValue.(map[string]any); ok {
				if formatBytes, err := json.Marshal(formatMap); err == nil {
					normalizedFormat := normalizeKibanaLensNumberFormatJSONString(string(formatBytes))
					var formatAny any
					if json.Unmarshal([]byte(normalizedFormat), &formatAny) == nil {
						out["format"] = normalizeXYPlanComparisonJSON(formatAny)
					}
				}
			}
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, elem := range t {
			out[i] = normalizeXYPlanComparisonJSON(elem)
		}
		return out
	default:
		return value
	}
}

func cloneAxisTitleModel(model *axisTitleModel) *axisTitleModel {
	if model == nil {
		return nil
	}
	cloned := *model
	return &cloned
}

func cloneYAxisConfigModel(model *yAxisConfigModel) *yAxisConfigModel {
	if model == nil {
		return nil
	}
	cloned := *model
	cloned.Title = cloneAxisTitleModel(model.Title)
	return &cloned
}
