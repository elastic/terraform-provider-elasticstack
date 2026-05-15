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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func xyAxisFromAPI(m *models.XYAxisModel, apiAxis kbapi.VisApiXyAxisConfig) diag.Diagnostics {
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
		m.X = &models.XYAxisConfigModel{}
		xDiags := xyAxisConfigFromAPI(m.X, &xView)
		diags.Append(xDiags...)
		if xyAxisConfigIsEmpty(m.X) {
			m.X = nil
		}
	}

	if apiAxis.Y != nil {
		m.Y = &models.YAxisConfigModel{}
		yDiags := YAxisConfigFromAPIY(m.Y, apiAxis.Y)
		diags.Append(yDiags...)
		if YAxisConfigIsEmpty(m.Y) {
			m.Y = nil
		}
	}

	if apiAxis.Y2 != nil {
		m.Y2 = &models.YAxisConfigModel{}
		y2Diags := YAxisConfigFromAPIY2(m.Y2, apiAxis.Y2)
		diags.Append(y2Diags...)
		if YAxisConfigIsEmpty(m.Y2) {
			m.Y2 = nil
		}
	}

	return diags
}

func xyAxisToAPI(m *models.XYAxisModel) (kbapi.VisApiXyAxisConfig, diag.Diagnostics) {
	if m == nil {
		return kbapi.VisApiXyAxisConfig{}, nil
	}

	var diags diag.Diagnostics
	var axis kbapi.VisApiXyAxisConfig

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
		yAxis, yDiags := YAxisConfigToAPIY(m.Y)
		diags.Append(yDiags...)
		axis.Y = yAxis
	}

	if m.Y2 != nil {
		y2Axis, y2Diags := YAxisConfigToAPIY2(m.Y2)
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
	if apiAxis.Labels != nil {
		m.LabelOrientation = types.StringValue(string(apiAxis.Labels.Orientation))
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
		xAxis.Labels = &struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		}{Orientation: kbapi.VisApiOrientation(m.LabelOrientation.ValueString())}
	}
	if typeutils.IsKnown(m.Scale) {
		scale := kbapi.VisApiXyAxisConfigXScale(m.Scale.ValueString())
		xAxis.Scale = &scale
	}
	if m.Title != nil {
		xAxis.Title = lenscommon.AxisTitleToAPI(m.Title)
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

func YAxisConfigIsEmpty(m *models.YAxisConfigModel) bool {
	if m == nil {
		return true
	}
	if typeutils.IsKnown(m.Ticks) || typeutils.IsKnown(m.Grid) || typeutils.IsKnown(m.LabelOrientation) || typeutils.IsKnown(m.Scale) || typeutils.IsKnown(m.DomainJSON) {
		return false
	}
	return axisTitleIsDefault(m.Title)
}

func YAxisConfigFromAPIY(m *models.YAxisConfigModel, apiAxis *struct {
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
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func YAxisConfigToAPIY(m *models.YAxisConfigModel) (*struct {
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
		yAxis.Title = lenscommon.AxisTitleToAPI(m.Title)
	}
	if typeutils.IsKnown(m.DomainJSON) {
		domainDiags := m.DomainJSON.Unmarshal(&yAxis.Domain)
		diags.Append(domainDiags...)
	}

	return yAxis, diags
}

func YAxisConfigFromAPIY2(m *models.YAxisConfigModel, apiAxis *struct {
	Domain kbapi.VisApiXyAxisConfig_Y2_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
		m.Title = &models.AxisTitleModel{}
		lenscommon.AxisTitleFromAPI(m.Title, apiAxis.Title)
	}

	domainJSON, err := json.Marshal(apiAxis.Domain)
	if err == nil {
		m.DomainJSON = jsontypes.NewNormalizedValue(string(domainJSON))
	}

	return diags
}

func YAxisConfigToAPIY2(m *models.YAxisConfigModel) (*struct {
	Domain kbapi.VisApiXyAxisConfig_Y2_Domain `json:"domain"`
	Grid   *struct {
		Visible bool `json:"visible"`
	} `json:"grid,omitempty"`
	Labels *struct {
		Orientation kbapi.VisApiOrientation `json:"orientation"`
	} `json:"labels,omitempty"`
	Scale *kbapi.VisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
		Domain kbapi.VisApiXyAxisConfig_Y2_Domain `json:"domain"`
		Grid   *struct {
			Visible bool `json:"visible"`
		} `json:"grid,omitempty"`
		Labels *struct {
			Orientation kbapi.VisApiOrientation `json:"orientation"`
		} `json:"labels,omitempty"`
		Scale *kbapi.VisApiXyAxisConfigY2Scale `json:"scale,omitempty"`
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
		scale := kbapi.VisApiXyAxisConfigY2Scale(m.Scale.ValueString())
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

func xyDecorationsReadFromStyling(m *models.XYDecorationsModel, s kbapi.XyStyling) {
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

func xyDecorationsWriteToStyling(m *models.XYDecorationsModel, s *kbapi.XyStyling) {
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

func xyFittingFromAPI(m *models.XYFittingModel, apiFitting kbapi.XyFitting) {
	m.Type = typeutils.StringishValue(apiFitting.Type)
	m.Dotted = types.BoolPointerValue(apiFitting.Emphasize)
	if apiFitting.Extend != nil {
		m.EndValue = types.StringValue(string(*apiFitting.Extend))
	} else {
		m.EndValue = types.StringNull()
	}
}

func xyFittingToAPI(m *models.XYFittingModel) kbapi.XyFitting {
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

func xyLegendFromAPI(ctx context.Context, m *models.XYLegendModel, apiLegend kbapi.XyLegend) diag.Diagnostics {
	var diags diag.Diagnostics
	m.Position = types.StringNull()
	m.Size = types.StringNull()
	m.Columns = types.Int64Null()
	m.TruncateAfterLines = types.Int64Null()
	m.Alignment = types.StringNull()
	m.Statistics = types.ListNull(types.StringType)

	// Try inside legend first
	legendInside, err := apiLegend.AsXyLegendInside()
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
	legendOutsideVertical, err := apiLegend.AsXyLegendOutsideVertical()
	if err == nil &&
		legendOutsideVertical.Placement != nil &&
		*legendOutsideVertical.Placement == kbapi.XyLegendOutsideVerticalPlacementOutside &&
		(legendOutsideVertical.Position == nil ||
			*legendOutsideVertical.Position == kbapi.XyLegendOutsideVerticalPositionLeft ||
			*legendOutsideVertical.Position == kbapi.XyLegendOutsideVerticalPositionRight) &&
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

func xyLegendToAPI(m *models.XYLegendModel) (kbapi.XyLegend, diag.Diagnostics) {
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
		legend.Placement = kbapi.Inside
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

func xyChartConfigStylingToAPI(m *models.XYChartConfigModel) kbapi.XyStyling {
	fit := kbapi.XyFitting{Type: kbapi.XyFittingTypeNone}
	if m.Fitting != nil {
		fit = xyFittingToAPI(m.Fitting)
	}
	s := kbapi.XyStyling{
		Areas:    kbapi.XyStylingAreas{},
		Bars:     kbapi.XyStylingBars{},
		Fitting:  fit,
		Overlays: kbapi.XyStylingOverlays{},
		Points:   kbapi.XyStylingPoints{},
	}
	if m.Decorations != nil {
		xyDecorationsWriteToStyling(m.Decorations, &s)
	}
	return s
}

// toAPINoESQL converts the XY chart config model to a non-ES|QL API payload.
func xyChartConfigToAPINoESQL(m *models.XYChartConfigModel, dashboard *models.DashboardModel) (kbapi.XyChartNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.XyChartNoESQL{Type: kbapi.XyChartNoESQLTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = new(m.Description.ValueString())
	}

	if m.Axis != nil {
		axis, axisDiags := xyAxisToAPI(m.Axis)
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = xyChartConfigStylingToAPI(m)

	if len(m.Layers) > 0 {
		layers := make([]kbapi.XyLayersNoESQL, 0, len(m.Layers))
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
			chart.Legend = legend
		}
	}

	if m.Query != nil {
		chart.Query = filterSimpleToAPI(m.Query)
	}

	chart.Filters = buildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return chart, diags
	}

	chart.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		chart.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		chart.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		chart.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.XyChartNoESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			chart.Drilldowns = &items
		}
	}

	return chart, diags
}

// toAPIESQL converts the XY chart config model to an ES|QL API payload.
func xyChartConfigToAPIESQL(m *models.XYChartConfigModel, dashboard *models.DashboardModel) (kbapi.XyChartESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	chart := kbapi.XyChartESQL{Type: kbapi.XyChartESQLTypeXy}

	if typeutils.IsKnown(m.Title) {
		chart.Title = new(m.Title.ValueString())
	}
	if typeutils.IsKnown(m.Description) {
		chart.Description = new(m.Description.ValueString())
	}

	if m.Axis != nil {
		axis, axisDiags := xyAxisToAPI(m.Axis)
		diags.Append(axisDiags...)
		chart.Axis = axis
	}

	chart.Styling = xyChartConfigStylingToAPI(m)

	if len(m.Layers) > 0 {
		layers := make([]kbapi.XyLayerESQL, 0, len(m.Layers))
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
			chart.Legend = legend
		}
	}

	chart.Filters = buildFiltersForAPI(m.Filters, &diags)

	writes, presDiags := lensChartPresentationWritesFor(dashboard, m.LensChartPresentationTFModel)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return chart, diags
	}

	chart.TimeRange = writes.TimeRange
	if writes.HideTitle != nil {
		chart.HideTitle = writes.HideTitle
	}
	if writes.HideBorder != nil {
		chart.HideBorder = writes.HideBorder
	}
	if writes.References != nil {
		chart.References = writes.References
	}
	if len(writes.DrilldownsRaw) > 0 {
		items, ddDiags := decodeLensDrilldownSlice[kbapi.XyChartESQL_Drilldowns_Item](writes.DrilldownsRaw)
		diags.Append(ddDiags...)
		if !ddDiags.HasError() {
			chart.Drilldowns = &items
		}
	}

	return chart, diags
}

func xyChartConfigFromAPINoESQL(ctx context.Context, m *models.XYChartConfigModel, dashboard *models.DashboardModel, prior *models.XYChartConfigModel, apiChart kbapi.XyChartNoESQL) diag.Diagnostics {
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
	xyFittingFromAPI(m.Fitting, apiChart.Styling.Fitting)

	m.Legend = &models.XYLegendModel{}
	legendDiags := xyLegendFromAPI(ctx, m.Legend, apiChart.Legend)
	diags.Append(legendDiags...)

	// Preserve nil query when prior state omitted it (query is optional in schema).
	if prior != nil && prior.Query == nil {
		m.Query = nil
	} else {
		m.Query = &models.FilterSimpleModel{}
		filterSimpleFromAPI(m.Query, apiChart.Query)
	}

	m.Filters = populateFiltersFromAPI(apiChart.Filters, &diags)

	var priorLens *models.LensChartPresentationTFModel
	if prior != nil {
		p := prior.LensChartPresentationTFModel
		priorLens = &p
	}
	ddWire, ddOmit, ddWireDiags := lensDrilldownsAPIToWire(apiChart.Drilldowns)
	diags.Append(ddWireDiags...)
	if ddWireDiags.HasError() {
		return diags
	}
	pres, presDiags := lensChartPresentationReadsFor(ctx, dashboard, priorLens, apiChart.TimeRange, apiChart.HideTitle, apiChart.HideBorder, apiChart.References, ddWire, ddOmit)
	diags.Append(presDiags...)
	if presDiags.HasError() {
		return diags
	}
	m.LensChartPresentationTFModel = pres

	return diags
}
