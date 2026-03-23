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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mapOptionalBoolWithSnapshotDefault maps an optional API bool to a Terraform Bool,
// preserving snapshot defaults (e.g. when the API returns false and the user hasn't set it).
//
//nolint:unparam // snapshotDefault is a parameter for API flexibility; callers use false for partition charts
func mapOptionalBoolWithSnapshotDefault(current types.Bool, apiValue *bool, snapshotDefault bool) types.Bool {
	switch {
	case apiValue == nil:
		if typeutils.IsKnown(current) {
			return current
		}
		return types.BoolNull()
	case typeutils.IsKnown(current) && *apiValue == snapshotDefault && current.ValueBool() != *apiValue:
		return current
	case !typeutils.IsKnown(current) && *apiValue == snapshotDefault:
		return types.BoolNull()
	default:
		return types.BoolValue(*apiValue)
	}
}

// mapOptionalFloatWithSnapshotDefault maps an optional API float to a Terraform Float64,
// preserving snapshot defaults.
//
//nolint:unparam // snapshotDefault is a parameter for API flexibility; callers use 1 for partition charts
func mapOptionalFloatWithSnapshotDefault(current types.Float64, apiValue *float32, snapshotDefault float64) types.Float64 {
	switch {
	case apiValue == nil:
		if typeutils.IsKnown(current) {
			return current
		}
		return types.Float64Null()
	case typeutils.IsKnown(current) && float64(*apiValue) == snapshotDefault && current.ValueFloat64() != float64(*apiValue):
		return current
	case !typeutils.IsKnown(current) && float64(*apiValue) == snapshotDefault:
		return types.Float64Null()
	default:
		return types.Float64Value(float64(*apiValue))
	}
}

// partitionLegendModel is the shared Terraform model for partition chart legends
// (treemap, mosaic). Used by both treemap and mosaic config models.
type partitionLegendModel struct {
	Nested            types.Bool    `tfsdk:"nested"`
	Size              types.String  `tfsdk:"size"`
	TruncateAfterLine types.Float64 `tfsdk:"truncate_after_lines"`
	Visible           types.String  `tfsdk:"visible"`
}

func (m *partitionLegendModel) fromTreemapLegend(api kbapi.TreemapLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Float64Value(float64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Float64Null()
	}
	if api.Visible != nil {
		m.Visible = types.StringValue(string(*api.Visible))
	} else {
		m.Visible = types.StringNull()
	}
}

func (m *partitionLegendModel) fromMosaicLegend(api kbapi.MosaicLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Float64Value(float64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Float64Null()
	}
	if api.Visible != nil {
		m.Visible = types.StringValue(string(*api.Visible))
	} else {
		m.Visible = types.StringNull()
	}
}

func (m *partitionLegendModel) toTreemapLegend() kbapi.TreemapLegend {
	legend := kbapi.TreemapLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueFloat64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.TreemapLegendVisible(m.Visible.ValueString())
		legend.Visible = &v
	}
	return legend
}

func (m *partitionLegendModel) toMosaicLegend() kbapi.MosaicLegend {
	legend := kbapi.MosaicLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueFloat64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.MosaicLegendVisible(m.Visible.ValueString())
		legend.Visible = &v
	}
	return legend
}

// partitionValueDisplay is the shared Terraform model for partition chart value display
// (treemap, mosaic). Used by both treemap and mosaic config models.
type partitionValueDisplay struct {
	Mode            types.String  `tfsdk:"mode"`
	PercentDecimals types.Float64 `tfsdk:"percent_decimals"`
}

func (m *partitionValueDisplay) fromTreemapNoESQL(api *struct {
	Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

func (m *partitionValueDisplay) fromTreemapESQL(api *struct {
	Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

func (m *partitionValueDisplay) fromMosaicNoESQL(api *struct {
	Mode            kbapi.MosaicNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

func (m *partitionValueDisplay) fromMosaicESQL(api *struct {
	Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
}) {
	if api == nil {
		return
	}
	m.Mode = types.StringValue(string(api.Mode))
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

func (m *partitionValueDisplay) toTreemapNoESQL() struct {
	Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.TreemapNoESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                            `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.TreemapNoESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}

func (m *partitionValueDisplay) toTreemapESQL() struct {
	Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.TreemapESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                          `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.TreemapESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}

func (m *partitionValueDisplay) toMosaicNoESQL() struct {
	Mode            kbapi.MosaicNoESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.MosaicNoESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                           `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.MosaicNoESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}

func (m *partitionValueDisplay) toMosaicESQL() struct {
	Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
	PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
} {
	vd := struct {
		Mode            kbapi.MosaicESQLValueDisplayMode `json:"mode"`
		PercentDecimals *float32                         `json:"percent_decimals,omitempty"`
	}{
		Mode: kbapi.MosaicESQLValueDisplayMode(m.Mode.ValueString()),
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}
