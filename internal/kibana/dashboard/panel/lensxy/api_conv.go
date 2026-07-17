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

package lensxy

import (
	"context"
	"encoding/json"
	"math"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func xyAxisFromAPI(m *models.XYAxisModel, apiAxis *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiAxis == nil {
		return diags
	}

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
		m.X = &models.XYAxisConfigModel{}
		xDiags := xyAxisConfigFromAPI(m.X, &xView)
		diags.Append(xDiags...)
		if xyAxisConfigIsEmpty(m.X) {
			m.X = nil
		}
	}

	if apiAxis.Y != nil {
		m.Y = &models.YAxisConfigModel{}
		yDiags := yAxisConfigFromAPIY(m.Y, apiAxis.Y)
		diags.Append(yDiags...)
		if yAxisConfigIsEmpty(m.Y) {
			m.Y = nil
		}
	}

	if apiAxis.Y2 != nil {
		m.Y2 = &models.YAxisConfigModel{}
		y2Diags := yAxisConfigFromAPIY2(m.Y2, apiAxis.Y2)
		diags.Append(y2Diags...)
		if yAxisConfigIsEmpty(m.Y2) {
			m.Y2 = nil
		}
	}

	return diags
}

func xyAxisToAPI(m *models.XYAxisModel) (*kbapi.KibanaHTTPAPIsVisApiXyAxisConfig, diag.Diagnostics) {
	if m == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	axis := &kbapi.KibanaHTTPAPIsVisApiXyAxisConfig{}

	if m.X != nil {
		xAxis, xDiags := xyAxisConfigToAPI(m.X)
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
		yAxis, yDiags := yAxisConfigToAPIY(m.Y)
		diags.Append(yDiags...)
		axis.Y = yAxis
	}

	if m.Y2 != nil {
		y2Axis, y2Diags := yAxisConfigToAPIY2(m.Y2)
		diags.Append(y2Diags...)
		axis.Y2 = y2Axis
	}

	return axis, diags
}

func xyAxisConfigIsEmpty(m *models.XYAxisConfigModel) bool {
	if m == nil {
		return true
	}
	if typeutils.IsKnown(m.Ticks) || typeutils.IsKnown(m.Grid) || typeutils.IsKnown(m.LabelOrientation) || typeutils.IsKnown(m.Scale) || typeutils.IsKnown(m.DomainJSON) {
		return false
	}
	return axisTitleIsDefault(m.Title)
}

type xyAxisConfigAPIModel = struct {
	Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScale `json:"scale,omitempty"`
	Ticks *struct {
		Visible bool `json:"visible"`
	} `json:"ticks,omitempty"`
	Title *struct {
		Text    *string `json:"text,omitempty"`
		Visible *bool   `json:"visible,omitempty"`
	} `json:"title,omitempty"`
}

func xyAxisConfigFromAPI(m *models.XYAxisConfigModel, apiAxis *xyAxisConfigAPIModel) diag.Diagnostics {
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
	if apiAxis.Labels != nil && apiAxis.Labels.Orientation != nil {
		m.LabelOrientation = types.StringValue(string(*apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, apiAxis.Title)
	}

	if apiAxis.Domain != nil {
		domainJSON, err := json.Marshal(apiAxis.Domain)
		if err == nil {
			m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
		}
	}

	return diags
}

func xyAxisConfigToAPI(m *models.XYAxisConfigModel) (*xyAxisConfigAPIModel, diag.Diagnostics) {
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
		orientation := kbapi.KibanaHTTPAPIsVisApiOrientation(m.LabelOrientation.ValueString())
		xAxis.Labels = &struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		}{Orientation: &orientation}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScale(m.Scale.ValueString())
		xAxis.Scale = &scale
	}
	if m.Title != nil {
		xAxis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	if typeutils.IsKnown(m.DomainJSON) {
		var domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Domain
		domainDiags := m.DomainJSON.Unmarshal(&domain)
		diags.Append(domainDiags...)
		if !domainDiags.HasError() {
			xAxis.Domain = &domain
		}
	}

	return xAxis, diags
}

func yAxisConfigIsEmpty(m *models.YAxisConfigModel) bool {
	if m == nil {
		return true
	}
	if typeutils.IsKnown(m.Ticks) || typeutils.IsKnown(m.Grid) || typeutils.IsKnown(m.LabelOrientation) || typeutils.IsKnown(m.Scale) || typeutils.IsKnown(m.DomainJSON) {
		return false
	}
	return axisTitleIsDefault(m.Title)
}

func yAxisConfigFromAPIY(m *models.YAxisConfigModel, apiAxis *struct {
	Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScale `json:"scale,omitempty"`
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
	if apiAxis.Labels != nil && apiAxis.Labels.Orientation != nil {
		m.LabelOrientation = types.StringValue(string(*apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func yAxisConfigToAPIY(m *models.YAxisConfigModel) (*struct {
	Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScale `json:"scale,omitempty"`
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
		Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		} `json:"labels,omitempty"`
		Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScale `json:"scale,omitempty"`
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
		orientation := kbapi.KibanaHTTPAPIsVisApiOrientation(m.LabelOrientation.ValueString())
		yAxis.Labels = &struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		}{Orientation: &orientation}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	if typeutils.IsKnown(m.DomainJSON) {
		domainDiags := m.DomainJSON.Unmarshal(&yAxis.Domain)
		diags.Append(domainDiags...)
	}

	return yAxis, diags
}

func yAxisConfigFromAPIY2(m *models.YAxisConfigModel, apiAxis *struct {
	Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
	if apiAxis.Labels != nil && apiAxis.Labels.Orientation != nil {
		m.LabelOrientation = types.StringValue(string(*apiAxis.Labels.Orientation))
	} else {
		m.LabelOrientation = types.StringNull()
	}
	m.Scale = typeutils.StringishPointerValue(apiAxis.Scale)

	if apiAxis.Title != nil {
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func yAxisConfigToAPIY2(m *models.YAxisConfigModel) (*struct {
	Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
		Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		} `json:"labels,omitempty"`
		Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
		orientation := kbapi.KibanaHTTPAPIsVisApiOrientation(m.LabelOrientation.ValueString())
		yAxis.Labels = &struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		}{Orientation: &orientation}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2Scale(m.Scale.ValueString())
		yAxis.Scale = &scale
	}
	if m.Title != nil {
		yAxis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	if typeutils.IsKnown(m.DomainJSON) {
		domainDiags := m.DomainJSON.Unmarshal(&yAxis.Domain)
		diags.Append(domainDiags...)
	}

	return yAxis, diags
}

func axisTitleIsDefault(title *models.AxisTitleModel) bool {
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

func xyDecorationsReadFromStyling(m *models.XYDecorationsModel, s *kbapi.KibanaHTTPAPIsXyStyling) {
	if s == nil {
		m.ShowEndZones = types.BoolNull()
		m.ShowCurrentTimeMarker = types.BoolNull()
		m.PointVisibility = types.StringNull()
		m.LineInterpolation = types.StringNull()
		m.MinimumBarHeight = types.Int64Null()
		m.ShowValueLabels = types.BoolNull()
		m.FillOpacity = types.Float64Null()
		return
	}
	if s.Overlays != nil && s.Overlays.PartialBuckets != nil && s.Overlays.PartialBuckets.Visible != nil {
		m.ShowEndZones = types.BoolValue(*s.Overlays.PartialBuckets.Visible)
	} else {
		m.ShowEndZones = types.BoolNull()
	}
	if s.Overlays != nil && s.Overlays.CurrentTimeMarker != nil && s.Overlays.CurrentTimeMarker.Visible != nil {
		m.ShowCurrentTimeMarker = types.BoolValue(*s.Overlays.CurrentTimeMarker.Visible)
	} else {
		m.ShowCurrentTimeMarker = types.BoolNull()
	}
	if s.Points != nil && s.Points.Visibility != nil {
		switch *s.Points.Visibility {
		case kbapi.KibanaHTTPAPIsXyStylingPointsVisibilityHidden:
			m.PointVisibility = types.StringValue("never")
		case kbapi.KibanaHTTPAPIsXyStylingPointsVisibilityVisible:
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
	if s.Bars != nil && s.Bars.MinimumHeight != nil {
		m.MinimumBarHeight = types.Int64Value(int64(*s.Bars.MinimumHeight))
	} else {
		m.MinimumBarHeight = types.Int64Null()
	}
	if s.Bars != nil && s.Bars.DataLabels != nil && s.Bars.DataLabels.Visible != nil {
		m.ShowValueLabels = types.BoolValue(*s.Bars.DataLabels.Visible)
	} else {
		m.ShowValueLabels = types.BoolNull()
	}
	if s.Areas != nil && s.Areas.FillOpacity != nil {
		val := float64(*s.Areas.FillOpacity)
		m.FillOpacity = types.Float64Value(math.Round(val*100) / 100)
	} else {
		m.FillOpacity = types.Float64Null()
	}
}

func xyDecorationsWriteToStyling(m *models.XYDecorationsModel, s *kbapi.KibanaHTTPAPIsXyStyling) {
	if m == nil || s == nil {
		return
	}
	if typeutils.IsKnown(m.ShowEndZones) {
		if s.Overlays == nil {
			s.Overlays = &kbapi.KibanaHTTPAPIsXyStylingOverlays{}
		}
		v := m.ShowEndZones.ValueBool()
		s.Overlays.PartialBuckets = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.ShowCurrentTimeMarker) {
		if s.Overlays == nil {
			s.Overlays = &kbapi.KibanaHTTPAPIsXyStylingOverlays{}
		}
		v := m.ShowCurrentTimeMarker.ValueBool()
		s.Overlays.CurrentTimeMarker = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.PointVisibility) {
		if s.Points == nil {
			s.Points = &kbapi.KibanaHTTPAPIsXyStylingPoints{}
		}
		switch m.PointVisibility.ValueString() {
		case "never":
			v := kbapi.KibanaHTTPAPIsXyStylingPointsVisibilityHidden
			s.Points.Visibility = &v
		case "always":
			v := kbapi.KibanaHTTPAPIsXyStylingPointsVisibilityVisible
			s.Points.Visibility = &v
		default:
			v := kbapi.KibanaHTTPAPIsXyStylingPointsVisibilityAuto
			s.Points.Visibility = &v
		}
	}
	if typeutils.IsKnown(m.LineInterpolation) {
		interp := kbapi.KibanaHTTPAPIsXyStylingInterpolation(m.LineInterpolation.ValueString())
		s.Interpolation = &interp
	}
	if typeutils.IsKnown(m.MinimumBarHeight) {
		if s.Bars == nil {
			s.Bars = &kbapi.KibanaHTTPAPIsXyStylingBars{}
		}
		s.Bars.MinimumHeight = new(float32(m.MinimumBarHeight.ValueInt64()))
	}
	if typeutils.IsKnown(m.ShowValueLabels) {
		if s.Bars == nil {
			s.Bars = &kbapi.KibanaHTTPAPIsXyStylingBars{}
		}
		v := m.ShowValueLabels.ValueBool()
		s.Bars.DataLabels = &struct {
			Visible *bool `json:"visible,omitempty"`
		}{Visible: &v}
	}
	if typeutils.IsKnown(m.FillOpacity) {
		if s.Areas == nil {
			s.Areas = &kbapi.KibanaHTTPAPIsXyStylingAreas{}
		}
		s.Areas.FillOpacity = new(float32(m.FillOpacity.ValueFloat64()))
	}
}

func xyFittingFromAPI(m *models.XYFittingModel, apiFitting *kbapi.KibanaHTTPAPIsXyFitting) {
	if apiFitting == nil {
		m.Type = types.StringNull()
		m.Dotted = types.BoolNull()
		m.EndValue = types.StringNull()
		return
	}
	m.Type = typeutils.NonEmptyStringishValue(apiFitting.Type)
	m.Dotted = types.BoolPointerValue(apiFitting.Emphasize)
	if apiFitting.Extend != nil {
		m.EndValue = typeutils.NonEmptyStringishValue(*apiFitting.Extend)
	} else {
		m.EndValue = types.StringNull()
	}
}

func xyFittingToAPI(m *models.XYFittingModel) kbapi.KibanaHTTPAPIsXyFitting {
	out := kbapi.KibanaHTTPAPIsXyFitting{Type: kbapi.KibanaHTTPAPIsXyFittingTypeNone}
	if m == nil {
		return out
	}
	if typeutils.IsKnown(m.Type) {
		out.Type = kbapi.KibanaHTTPAPIsXyFittingType(m.Type.ValueString())
	}
	if typeutils.IsKnown(m.Dotted) {
		out.Emphasize = new(m.Dotted.ValueBool())
	}
	if typeutils.IsKnown(m.EndValue) {
		ext := kbapi.KibanaHTTPAPIsXyFittingExtend(m.EndValue.ValueString())
		out.Extend = &ext
	}
	return out
}

func xyLegendFromAPI(ctx context.Context, m *models.XYLegendModel, apiLegend *kbapi.KibanaHTTPAPIsXyLegend) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Position = types.StringNull()
	m.Size = types.StringNull()
	m.Columns = types.Int64Null()
	m.TruncateAfterLines = types.Int64Null()
	m.Alignment = types.StringNull()
	m.Statistics = types.ListNull(types.StringType)

	if apiLegend == nil {
		return diags
	}

	// Try inside legend first
	legendInside, err := apiLegend.AsKibanaHTTPAPIsXyLegendInside()
	if err == nil && legendInside.Placement == kbapi.Inside {
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
	legendOutsideVertical, err := apiLegend.AsKibanaHTTPAPIsXyLegendOutsideVertical()
	if err == nil &&
		legendOutsideVertical.Placement != nil &&
		*legendOutsideVertical.Placement == kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalPlacementOutside &&
		(legendOutsideVertical.Position == nil ||
			*legendOutsideVertical.Position == kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalPositionLeft ||
			*legendOutsideVertical.Position == kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalPositionRight) &&
		legendOutsideVertical.Size != nil {
		m.Inside = types.BoolValue(false)
		m.Visibility = typeutils.StringishPointerValue(legendOutsideVertical.Visibility)
		m.Position = typeutils.StringishPointerValue(legendOutsideVertical.Position)
		m.Size = types.StringValue(string(*legendOutsideVertical.Size))

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
	legendOutsideHorizontal, err := apiLegend.AsKibanaHTTPAPIsXyLegendOutsideHorizontal()
	if err == nil {
		m.Inside = types.BoolValue(false)
		m.Visibility = typeutils.StringishPointerValue(legendOutsideHorizontal.Visibility)
		m.Position = typeutils.StringishPointerValue(legendOutsideHorizontal.Position)

		if legendOutsideHorizontal.Layout != nil {
			if layout, layoutErr := legendOutsideHorizontal.Layout.AsKibanaHTTPAPIsXyLegendOutsideHorizontalLayout0(); layoutErr == nil &&
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

	return xyLegendFromAPIFlatFallback(ctx, m, apiLegend)
}

// xyLegendFromAPIFlatFallback handles dashboard API responses that return outside legend
// fields at the top level without union discriminator tags.
func xyLegendFromAPIFlatFallback(ctx context.Context, m *models.XYLegendModel, apiLegend *kbapi.KibanaHTTPAPIsXyLegend) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiLegend == nil {
		return diags
	}

	raw, err := json.Marshal(apiLegend)
	if err != nil {
		return diags
	}

	var flat struct {
		Visibility *string `json:"visibility"`
		Placement  *string `json:"placement"`
		Size       *string `json:"size"`
		Position   *string `json:"position"`
		Layout     *struct {
			Type     *string `json:"type"`
			Truncate *struct {
				MaxLines *float32 `json:"max_lines"`
			} `json:"truncate"`
		} `json:"layout"`
	}
	if err := json.Unmarshal(raw, &flat); err != nil {
		return diags
	}
	if flat.Placement == nil || *flat.Placement != "outside" {
		return diags
	}

	m.Inside = types.BoolValue(false)
	m.Visibility = typeutils.StringishPointerValue(flat.Visibility)
	m.Position = typeutils.StringishPointerValue(flat.Position)
	if flat.Size != nil {
		m.Size = types.StringValue(*flat.Size)
	}
	if flat.Layout != nil && flat.Layout.Truncate != nil && flat.Layout.Truncate.MaxLines != nil {
		m.TruncateAfterLines = types.Int64Value(int64(*flat.Layout.Truncate.MaxLines))
	}
	_ = ctx
	return diags
}

func xyLegendToAPI(m *models.XYLegendModel) (kbapi.KibanaHTTPAPIsXyLegend, diag.Diagnostics) {
	if m == nil {
		return kbapi.KibanaHTTPAPIsXyLegend{}, nil
	}

	var diags diag.Diagnostics
	isInside := typeutils.IsKnown(m.Inside) && m.Inside.ValueBool()
	insideVisibility := kbapi.KibanaHTTPAPIsXyLegendInsideVisibilityAuto
	outsideHorizontalVisibility := kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalVisibilityAuto
	outsideVerticalVisibility := kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalVisibilityAuto
	if typeutils.IsKnown(m.Visibility) {
		insideVisibility = kbapi.KibanaHTTPAPIsXyLegendInsideVisibility(m.Visibility.ValueString())
		outsideHorizontalVisibility = kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalVisibility(m.Visibility.ValueString())
		outsideVerticalVisibility = kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalVisibility(m.Visibility.ValueString())
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
		var legend kbapi.KibanaHTTPAPIsXyLegendInside
		legend.Placement = kbapi.Inside
		legend.Visibility = &insideVisibility

		if typeutils.IsKnown(m.TruncateAfterLines) {
			legend.Layout = &struct {
				Truncate *struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				} `json:"truncate,omitempty"`
				Type kbapi.KibanaHTTPAPIsXyLegendInsideLayoutType `json:"type"`
			}{
				Truncate: &struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				}{
					MaxLines: new(float32(m.TruncateAfterLines.ValueInt64())),
				},
				Type: kbapi.KibanaHTTPAPIsXyLegendInsideLayoutTypeGrid,
			}
		}
		if typeutils.IsKnown(m.Columns) {
			legend.Columns = new(float32(m.Columns.ValueInt64()))
		}
		if typeutils.IsKnown(m.Alignment) {
			pos := kbapi.KibanaHTTPAPIsXyLegendInsidePosition(m.Alignment.ValueString())
			legend.Position = &pos
		}
		if stats, ok := statsElemsToStrings(); ok {
			statsAPI := make([]kbapi.KibanaHTTPAPIsXyLegendInsideStatistics, 0, len(stats))
			for _, s := range stats {
				statsAPI = append(statsAPI, kbapi.KibanaHTTPAPIsXyLegendInsideStatistics(s))
			}
			legend.Statistics = &statsAPI
		}

		var result kbapi.KibanaHTTPAPIsXyLegend
		if err := result.FromKibanaHTTPAPIsXyLegendInside(legend); err != nil {
			diags.AddError("Failed to create inside legend", err.Error())
		}
		return result, diags
	}

	outsidePosition := ""
	if typeutils.IsKnown(m.Position) {
		outsidePosition = m.Position.ValueString()
	}
	isHorizontal := outsidePosition == "top" || outsidePosition == "bottom"

	var result kbapi.KibanaHTTPAPIsXyLegend
	if isHorizontal {
		var legend kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontal
		placement := kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalPlacementOutside
		legend.Placement = &placement
		legend.Visibility = &outsideHorizontalVisibility
		if outsidePosition != "" {
			pos := kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalPosition(outsidePosition)
			legend.Position = &pos
		}
		if typeutils.IsKnown(m.TruncateAfterLines) {
			layout := kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontal_Layout{}
			if err := layout.FromKibanaHTTPAPIsXyLegendOutsideHorizontalLayout0(kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalLayout0{
				Truncate: &struct {
					Enabled  *bool    `json:"enabled,omitempty"`
					MaxLines *float32 `json:"max_lines,omitempty"`
				}{
					MaxLines: new(float32(m.TruncateAfterLines.ValueInt64())),
				},
				Type: kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalLayout0TypeGrid,
			}); err != nil {
				diags.AddError("Failed to create horizontal legend layout", err.Error())
				return result, diags
			}
			legend.Layout = &layout
		}
		if stats, ok := statsElemsToStrings(); ok {
			statsAPI := make([]kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalStatistics, 0, len(stats))
			for _, s := range stats {
				statsAPI = append(statsAPI, kbapi.KibanaHTTPAPIsXyLegendOutsideHorizontalStatistics(s))
			}
			legend.Statistics = &statsAPI
		}
		if err := result.FromKibanaHTTPAPIsXyLegendOutsideHorizontal(legend); err != nil {
			diags.AddError("Failed to create outside horizontal legend", err.Error())
		}
		return result, diags
	}

	var legend kbapi.KibanaHTTPAPIsXyLegendOutsideVertical
	placement := kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalPlacementOutside
	legend.Placement = &placement
	legend.Visibility = &outsideVerticalVisibility
	if outsidePosition != "" {
		pos := kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalPosition(outsidePosition)
		legend.Position = &pos
	}
	if typeutils.IsKnown(m.Size) {
		size := kbapi.KibanaHTTPAPIsLegendSize(m.Size.ValueString())
		legend.Size = &size
	} else {
		size := kbapi.KibanaHTTPAPIsLegendSizeM
		legend.Size = &size
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		legend.Layout = &struct {
			Truncate *struct {
				Enabled  *bool    `json:"enabled,omitempty"`
				MaxLines *float32 `json:"max_lines,omitempty"`
			} `json:"truncate,omitempty"`
			Type kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalLayoutType `json:"type"`
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
		statsAPI := make([]kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalStatistics, 0, len(stats))
		for _, s := range stats {
			statsAPI = append(statsAPI, kbapi.KibanaHTTPAPIsXyLegendOutsideVerticalStatistics(s))
		}
		legend.Statistics = &statsAPI
	}
	if err := result.FromKibanaHTTPAPIsXyLegendOutsideVertical(legend); err != nil {
		diags.AddError("Failed to create outside vertical legend", err.Error())
	}
	return result, diags
}

func xyChartConfigXyUsesESQL(m *models.XYChartConfigModel) bool {
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

func xyChartConfigStylingToAPI(m *models.XYChartConfigModel) *kbapi.KibanaHTTPAPIsXyStyling {
	areas := kbapi.KibanaHTTPAPIsXyStylingAreas{}
	bars := kbapi.KibanaHTTPAPIsXyStylingBars{}
	overlays := kbapi.KibanaHTTPAPIsXyStylingOverlays{}
	points := kbapi.KibanaHTTPAPIsXyStylingPoints{}
	fit := kbapi.KibanaHTTPAPIsXyFitting{Type: kbapi.KibanaHTTPAPIsXyFittingTypeNone}
	if m.Fitting != nil {
		fit = xyFittingToAPI(m.Fitting)
	}
	s := &kbapi.KibanaHTTPAPIsXyStyling{
		Areas:    &areas,
		Bars:     &bars,
		Fitting:  &fit,
		Overlays: &overlays,
		Points:   &points,
	}
	if m.Decorations != nil {
		xyDecorationsWriteToStyling(m.Decorations, s)
	}
	return s
}

func xyChartConfigToAPINoESQL(m *models.XYChartConfigModel) (kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanelTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = m.Description.ValueStringPointer()
	}

	if m.Axis != nil {
		axis, axisDiags := xyAxisToAPI(m.Axis)
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = xyChartConfigStylingToAPI(m)

	if len(m.Layers) > 0 {
		layers := make([]kbapi.KibanaHTTPAPIsXyLayersNoESQL, 0, len(m.Layers))
		for _, layer := range m.Layers {
			apiLayer, layerDiags := xyLayerToAPILayersNoESQL(&layer)
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
		legend, legendDiags := xyLegendToAPI(m.Legend)
		diags.Append(legendDiags...)
		if !legendDiags.HasError() {
			chart.Legend = &legend
		}
	}

	if m.Query != nil {
		chart.Query = lenscommon.FilterSimpleToAPI(m.Query)
	}

	chart.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return chart, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel_Drilldowns_Item](
		writes, &chart.TimeRange, &chart.HideTitle, &chart.HideBorder, &chart.References, &chart.Drilldowns,
	)...)

	return chart, diags
}

// toAPIESQL converts the XY chart config model to an ES|QL API payload.
func xyChartConfigToAPIESQL(m *models.XYChartConfigModel) (kbapi.KibanaHTTPAPIsXyChartESQLByValuePanel, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.KibanaHTTPAPIsXyChartESQLByValuePanel{Type: kbapi.KibanaHTTPAPIsXyChartESQLByValuePanelTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = m.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = m.Description.ValueStringPointer()
	}

	if m.Axis != nil {
		axis, axisDiags := xyAxisToAPI(m.Axis)
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = xyChartConfigStylingToAPI(m)

	if len(m.Layers) > 0 {
		layers := make([]kbapi.KibanaHTTPAPIsXyLayerESQL, 0, len(m.Layers))
		for _, layer := range m.Layers {
			apiLayer, layerDiags := xyLayerToAPILayerESQL(&layer)
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
		legend, legendDiags := xyLegendToAPI(m.Legend)
		diags.Append(legendDiags...)
		if !legendDiags.HasError() {
			chart.Legend = &legend
		}
	}

	chart.Filters = lenscommon.BuildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lenscommon.LensChartPresentationWritesFor(m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return chart, diags
	}

	diags.Append(lenscommon.ApplyLensChartPresentationWrites[kbapi.KibanaHTTPAPIsXyChartESQLByValuePanel_Drilldowns_Item](
		writes, &chart.TimeRange, &chart.HideTitle, &chart.HideBorder, &chart.References, &chart.Drilldowns,
	)...)

	return chart, diags
}

func xyChartConfigFromAPINoESQL(
	ctx context.Context,
	m *models.XYChartConfigModel,
	prior *models.XYChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsXyChartNoESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	if len(apiChart.Layers) > 0 {
		priorLayers := []models.XYLayerModel(nil)
		if prior != nil {
			priorLayers = prior.Layers
		}
		m.Layers = make([]models.XYLayerModel, 0, len(apiChart.Layers))
		for i, apiLayer := range apiChart.Layers {
			layer := models.XYLayerModel{}
			if i < len(priorLayers) {
				layer = priorLayers[i]
			}
			layerDiags := xyLayerFromAPILayersNoESQL(ctx, &layer, apiLayer)
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				m.Layers = append(m.Layers, layer)
			}
		}
	}

	m.Axis = &models.XYAxisModel{}
	axisDiags := xyAxisFromAPI(m.Axis, apiChart.Axis)
	diags.Append(axisDiags...)

	m.Decorations = &models.XYDecorationsModel{}
	xyDecorationsReadFromStyling(m.Decorations, apiChart.Styling)

	m.Fitting = &models.XYFittingModel{}
	if apiChart.Styling != nil {
		xyFittingFromAPI(m.Fitting, apiChart.Styling.Fitting)
	}

	if apiChart.Legend == nil {
		m.Legend = nil
	} else {
		m.Legend = &models.XYLegendModel{}
		legendDiags := xyLegendFromAPI(ctx, m.Legend, apiChart.Legend)
		diags.Append(legendDiags...)
	}

	// Preserve nil query when prior state omitted it (query is optional in schema).
	if prior != nil && prior.Query == nil {
		m.Query = nil
	} else {
		m.Query = &models.FilterSimpleModel{}
		lenscommon.FilterSimpleFromAPI(m.Query, apiChart.Query)
	}

	m.Filters = lenscommon.PopulateFiltersFromAPI(apiChart.Filters, &diags)

	if !lenscommon.PopulateLensChartPresentation(
		ctx, &m.LensChartPresentationTFModel, prior, apiChart.TimeRange,
		apiChart.HideTitle, apiChart.HideBorder, apiChart.References, apiChart.Drilldowns, &diags,
	) {
		return diags
	}

	return diags
}

func xyChartConfigFromAPIESQL(
	ctx context.Context,
	m *models.XYChartConfigModel,
	prior *models.XYChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsXyChartESQLByValuePanel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Title = types.StringPointerValue(apiChart.Title)
	m.Description = types.StringPointerValue(apiChart.Description)

	if len(apiChart.Layers) > 0 {
		priorLayers := []models.XYLayerModel(nil)
		if prior != nil {
			priorLayers = prior.Layers
		}
		m.Layers = make([]models.XYLayerModel, 0, len(apiChart.Layers))
		for i, apiLayer := range apiChart.Layers {
			layer := models.XYLayerModel{}
			if i < len(priorLayers) {
				layer = priorLayers[i]
			}
			layerDiags := xyLayerFromAPILayerESQL(ctx, &layer, apiLayer)
			diags.Append(layerDiags...)
			if !layerDiags.HasError() {
				m.Layers = append(m.Layers, layer)
			}
		}
	}

	m.Axis = &models.XYAxisModel{}
	axisDiags := xyAxisFromAPI(m.Axis, apiChart.Axis)
	diags.Append(axisDiags...)

	m.Decorations = &models.XYDecorationsModel{}
	xyDecorationsReadFromStyling(m.Decorations, apiChart.Styling)

	m.Fitting = &models.XYFittingModel{}
	if apiChart.Styling != nil {
		xyFittingFromAPI(m.Fitting, apiChart.Styling.Fitting)
	}

	if apiChart.Legend == nil {
		m.Legend = nil
	} else {
		m.Legend = &models.XYLegendModel{}
		legendDiags := xyLegendFromAPI(ctx, m.Legend, apiChart.Legend)
		diags.Append(legendDiags...)
	}

	m.Query = nil

	m.Filters = lenscommon.PopulateFiltersFromAPI(apiChart.Filters, &diags)

	if !lenscommon.PopulateLensChartPresentation(
		ctx, &m.LensChartPresentationTFModel, prior, apiChart.TimeRange,
		apiChart.HideTitle, apiChart.HideBorder, apiChart.References, apiChart.Drilldowns, &diags,
	) {
		return diags
	}

	return diags
}

func xyChartConfigToAPI(m *models.XYChartConfigModel) (lenscommon.VisByValueConfig0, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.VisByValueConfig0
	if m == nil {
		return attrs, diags
	}
	configModel := *m

	if xyChartConfigXyUsesESQL(&configModel) {
		chart, xyDiags := xyChartConfigToAPIESQL(&configModel)
		diags.Append(xyDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		if err := attrs.FromKibanaHTTPAPIsXyChartESQLByValuePanel(chart); err != nil {
			diags.AddError("Failed to convert XY chart ES|QL config", err.Error())
			return attrs, diags
		}
		return attrs, diags
	}

	chart, xyDiags := xyChartConfigToAPINoESQL(&configModel)
	diags.Append(xyDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	if err := attrs.FromKibanaHTTPAPIsXyChartNoESQLByValuePanel(chart); err != nil {
		diags.AddError("Failed to convert XY chart non-ES|QL config", err.Error())
		return attrs, diags
	}
	return attrs, diags
}
