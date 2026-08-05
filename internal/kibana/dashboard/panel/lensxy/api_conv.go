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
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	legendPlacementInside = "inside"
	legendLayoutGrid      = "grid"
)

func lensScaleString(scale any) types.String {
	if scale == nil {
		return types.StringNull()
	}
	raw, err := json.Marshal(scale)
	if err != nil {
		return types.StringNull()
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func lensUnionStringValue(value any) (string, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", false
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", false
	}
	return result, true
}

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
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Scale `json:"scale,omitempty"`
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
	m.Scale = lensScaleString(apiAxis.Scale)

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
		var scale kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Scale
		if json.Unmarshal([]byte(strconv.Quote(m.Scale.ValueString())), &scale) == nil {
			xAxis.Scale = &scale
		}
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
	Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Scale `json:"scale,omitempty"`
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
	m.Scale = lensScaleString(apiAxis.Scale)

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
	Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Scale `json:"scale,omitempty"`
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
		Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain `json:"domain,omitempty"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		} `json:"labels,omitempty"`
		Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Scale `json:"scale,omitempty"`
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
		var scale kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Scale
		if json.Unmarshal([]byte(strconv.Quote(m.Scale.ValueString())), &scale) == nil {
			yAxis.Scale = &scale
		}
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
	Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Scale `json:"scale,omitempty"`
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
	m.Scale = lensScaleString(apiAxis.Scale)

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
	Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain,omitempty"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
	} `json:"labels,omitempty"`
	Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Scale `json:"scale,omitempty"`
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
		Domain *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain `json:"domain,omitempty"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		} `json:"labels,omitempty"`
		Scale *kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Scale `json:"scale,omitempty"`
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
		var scale kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Scale
		if json.Unmarshal([]byte(strconv.Quote(m.Scale.ValueString())), &scale) == nil {
			yAxis.Scale = &scale
		}
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
		switch value, _ := lensUnionStringValue(s.Points.Visibility); value {
		case "hidden":
			m.PointVisibility = types.StringValue("never")
		case "visible":
			m.PointVisibility = types.StringValue("always")
		default:
			m.PointVisibility = types.StringValue("auto")
		}
	} else {
		m.PointVisibility = types.StringNull()
	}
	if s.Interpolation != nil {
		if value, ok := lensUnionStringValue(s.Interpolation); ok {
			m.LineInterpolation = types.StringValue(value)
		} else {
			m.LineInterpolation = types.StringNull()
		}
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
		value := "auto"
		switch m.PointVisibility.ValueString() {
		case "never":
			value = "hidden"
		case "always":
			value = "visible"
		}
		var visibility kbapi.KibanaHTTPAPIsXyStylingPoints_Visibility
		if json.Unmarshal([]byte(strconv.Quote(value)), &visibility) == nil {
			s.Points.Visibility = &visibility
		}
	}
	if typeutils.IsKnown(m.LineInterpolation) {
		var interpolation kbapi.KibanaHTTPAPIsXyStyling_Interpolation
		if json.Unmarshal([]byte(strconv.Quote(m.LineInterpolation.ValueString())), &interpolation) == nil {
			s.Interpolation = &interpolation
		}
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

	var fitting struct {
		Type   string  `json:"type"`
		Extend *string `json:"extend"`
	}
	if raw, err := json.Marshal(apiFitting); err == nil {
		_ = json.Unmarshal(raw, &fitting)
	}
	if fitting.Type == "" {
		m.Type = types.StringNull()
	} else {
		m.Type = types.StringValue(fitting.Type)
	}
	m.Dotted = types.BoolPointerValue(apiFitting.Emphasize)
	if fitting.Extend != nil {
		m.EndValue = types.StringValue(*fitting.Extend)
	} else {
		m.EndValue = types.StringNull()
	}
}

func xyFittingToAPI(m *models.XYFittingModel) kbapi.KibanaHTTPAPIsXyFitting {
	out := kbapi.KibanaHTTPAPIsXyFitting{}
	fitting := struct {
		Type      string  `json:"type"`
		Emphasize *bool   `json:"emphasize,omitempty"`
		Extend    *string `json:"extend,omitempty"`
	}{Type: "none"}
	if m == nil {
		_ = json.Unmarshal([]byte(`{"type":"none"}`), &out)
		return out
	}
	if typeutils.IsKnown(m.Type) {
		fitting.Type = m.Type.ValueString()
	}
	if typeutils.IsKnown(m.Dotted) {
		fitting.Emphasize = new(m.Dotted.ValueBool())
	}
	if typeutils.IsKnown(m.EndValue) {
		fitting.Extend = new(m.EndValue.ValueString())
	}
	raw, _ := json.Marshal(fitting)
	_ = json.Unmarshal(raw, &out)
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

	raw, err := json.Marshal(apiLegend)
	if err != nil {
		return diags
	}
	var legend struct {
		Visibility *string  `json:"visibility"`
		Placement  *string  `json:"placement"`
		Size       *string  `json:"size"`
		Position   *string  `json:"position"`
		Columns    *float32 `json:"columns"`
		Statistics []string `json:"statistics"`
		Layout     *struct {
			Truncate *struct {
				MaxLines *float32 `json:"max_lines"`
			} `json:"truncate"`
		} `json:"layout"`
	}
	if err := json.Unmarshal(raw, &legend); err != nil {
		return diags
	}
	m.Inside = types.BoolValue(legend.Placement != nil && *legend.Placement == legendPlacementInside)
	m.Visibility = typeutils.StringishPointerValue(legend.Visibility)
	if m.Inside.ValueBool() {
		m.Alignment = typeutils.StringishPointerValue(legend.Position)
		if legend.Columns != nil {
			m.Columns = types.Int64Value(int64(*legend.Columns))
		}
	} else {
		m.Position = typeutils.StringishPointerValue(legend.Position)
		m.Size = typeutils.StringishPointerValue(legend.Size)
	}
	if legend.Layout != nil && legend.Layout.Truncate != nil && legend.Layout.Truncate.MaxLines != nil {
		m.TruncateAfterLines = types.Int64Value(int64(*legend.Layout.Truncate.MaxLines))
	}
	if legend.Statistics != nil {
		stats := make([]types.String, 0, len(legend.Statistics))
		for _, stat := range legend.Statistics {
			stats = append(stats, types.StringValue(stat))
		}
		var statsDiags diag.Diagnostics
		m.Statistics, statsDiags = types.ListValueFrom(ctx, types.StringType, stats)
		diags.Append(statsDiags...)
	}
	return diags
}

func xyLegendToAPI(m *models.XYLegendModel) (kbapi.KibanaHTTPAPIsXyLegend, diag.Diagnostics) {
	if m == nil {
		return kbapi.KibanaHTTPAPIsXyLegend{}, nil
	}

	var diags diag.Diagnostics
	legend := map[string]any{"visibility": "auto"}
	isInside := typeutils.IsKnown(m.Inside) && m.Inside.ValueBool()
	if isInside {
		legend["placement"] = legendPlacementInside
	} else {
		legend["placement"] = "outside"
	}
	if typeutils.IsKnown(m.Visibility) {
		legend["visibility"] = m.Visibility.ValueString()
	}
	if typeutils.IsKnown(m.TruncateAfterLines) {
		legend["layout"] = map[string]any{
			"type": legendLayoutGrid,
			"truncate": map[string]any{
				"max_lines": m.TruncateAfterLines.ValueInt64(),
			},
		}
	}
	if isInside && typeutils.IsKnown(m.Columns) {
		legend["columns"] = m.Columns.ValueInt64()
	}
	if isInside && typeutils.IsKnown(m.Alignment) {
		legend["position"] = m.Alignment.ValueString()
	}
	if !isInside && typeutils.IsKnown(m.Position) {
		legend["position"] = m.Position.ValueString()
	}
	if !isInside {
		size := "m"
		if typeutils.IsKnown(m.Size) {
			size = m.Size.ValueString()
		}
		if size != "" {
			legend["size"] = size
		}
	}
	statsElemsToStrings := func() {
		if !typeutils.IsKnown(m.Statistics) {
			return
		}

		elems := m.Statistics.Elements()
		if len(elems) == 0 {
			return
		}

		stats := make([]string, 0, len(elems))
		for _, elem := range elems {
			strVal, ok := elem.(types.String)
			if !ok {
				diags.AddError("Invalid legend statistic value", "Expected statistics element to be a string")
				return
			}
			if !typeutils.IsKnown(strVal) {
				diags.AddError("Invalid legend statistic value", "Statistics element must be known")
				return
			}
			stats = append(stats, strVal.ValueString())
		}
		legend["statistics"] = stats
	}
	statsElemsToStrings()
	var result kbapi.KibanaHTTPAPIsXyLegend
	if raw, err := json.Marshal(legend); err != nil {
		diags.AddError("Failed to encode legend", err.Error())
	} else if err := json.Unmarshal(raw, &result); err != nil {
		diags.AddError("Failed to create XY legend", err.Error())
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
	fit := xyFittingToAPI(nil)
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

func xyChartConfigToAPINoESQL(m *models.XYChartConfigModel) (kbapi.KibanaHTTPAPIsXyChartNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.KibanaHTTPAPIsXyChartNoESQL{Type: kbapi.KibanaHTTPAPIsXyChartNoESQLTypeXy}

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

	return chart, diags
}

// toAPIESQL converts the XY chart config model to an ES|QL API payload.
func xyChartConfigToAPIESQL(m *models.XYChartConfigModel) (kbapi.KibanaHTTPAPIsXyChartESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.KibanaHTTPAPIsXyChartESQL{Type: kbapi.KibanaHTTPAPIsXyChartESQLTypeXy}

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

	return chart, diags
}

func xyChartConfigFromAPINoESQL(
	ctx context.Context,
	m *models.XYChartConfigModel,
	prior *models.XYChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsXyChartNoESQL,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
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

	if !lenscommon.PopulateLensChartPresentationFromAPI(
		ctx, &m.LensChartPresentationTFModel, prior, presentation, &diags,
	) {
		return diags
	}

	return diags
}

func xyChartConfigFromAPIESQL(
	ctx context.Context,
	m *models.XYChartConfigModel,
	prior *models.XYChartConfigModel,
	apiChart kbapi.KibanaHTTPAPIsXyChartESQL,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
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

	if !lenscommon.PopulateLensChartPresentationFromAPI(
		ctx, &m.LensChartPresentationTFModel, prior, presentation, &diags,
	) {
		return diags
	}

	return diags
}

func xyChartConfigToAPI(m *models.XYChartConfigModel) (lenscommon.LensByValueConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrs lenscommon.LensByValueConfig
	if m == nil {
		return attrs, diags
	}
	configModel := *m
	presentation, presDiags := lenscommon.LensChartPresentationToAPI(configModel.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return attrs, diags
	}
	attrs.Presentation = presentation

	if xyChartConfigXyUsesESQL(&configModel) {
		chart, xyDiags := xyChartConfigToAPIESQL(&configModel)
		diags.Append(xyDiags...)
		if diags.HasError() {
			return attrs, diags
		}
		var xyChart kbapi.KibanaHTTPAPIsXyChart
		if err := xyChart.FromKibanaHTTPAPIsXyChartESQL(chart); err != nil {
			diags.AddError("Failed to convert XY chart ES|QL config", err.Error())
			return attrs, diags
		}
		if err := attrs.Chart.FromKibanaHTTPAPIsXyChart(xyChart); err != nil {
			diags.AddError("Failed to encode XY chart ES|QL config", err.Error())
		}
		return attrs, diags
	}

	chart, xyDiags := xyChartConfigToAPINoESQL(&configModel)
	diags.Append(xyDiags...)
	if diags.HasError() {
		return attrs, diags
	}
	var xyChart kbapi.KibanaHTTPAPIsXyChart
	if err := xyChart.FromKibanaHTTPAPIsXyChartNoESQL(chart); err != nil {
		diags.AddError("Failed to convert XY chart non-ES|QL config", err.Error())
		return attrs, diags
	}
	if err := attrs.Chart.FromKibanaHTTPAPIsXyChart(xyChart); err != nil {
		diags.AddError("Failed to encode XY chart non-ES|QL config", err.Error())
	}
	return attrs, diags
}
