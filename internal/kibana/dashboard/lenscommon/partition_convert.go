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

package lenscommon

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PartitionLegendFromPieLegend maps API pie legend into Terraform partition legend model.
func PartitionLegendFromPieLegend(m *models.PartitionLegendModel, api kbapi.PieLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	if api.Size != "" {
		m.Size = types.StringValue(string(api.Size))
	} else {
		m.Size = types.StringValue(string(kbapi.LegendSizeAuto))
	}
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Int64Null()
	}
	if api.Visibility != nil {
		m.Visible = types.StringValue(string(*api.Visibility))
	} else {
		// Align with pie_chart_config.legend schema default (visible = auto) when Kibana omits the field.
		m.Visible = types.StringValue(string(kbapi.PieLegendVisibilityAuto))
	}
}

// PartitionLegendToPieLegend maps Terraform partition legend model to API pie legend.
func PartitionLegendToPieLegend(m *models.PartitionLegendModel) kbapi.PieLegend {
	legend := kbapi.PieLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueInt64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.PieLegendVisibility(m.Visible.ValueString())
		legend.Visibility = &v
	}
	return legend
}

// PartitionLegendFromTreemapLegend maps API treemap legend into Terraform partition legend model.
func PartitionLegendFromTreemapLegend(m *models.PartitionLegendModel, api kbapi.TreemapLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Int64Null()
	}
	if api.Visibility != nil {
		m.Visible = types.StringValue(string(*api.Visibility))
	} else {
		m.Visible = types.StringNull()
	}
}

// PartitionLegendFromMosaicLegend maps API mosaic legend into Terraform partition legend model.
func PartitionLegendFromMosaicLegend(m *models.PartitionLegendModel, api kbapi.MosaicLegend) {
	m.Nested = types.BoolPointerValue(api.Nested)
	m.Size = types.StringValue(string(api.Size))
	if api.TruncateAfterLines != nil {
		m.TruncateAfterLine = types.Int64Value(int64(*api.TruncateAfterLines))
	} else {
		m.TruncateAfterLine = types.Int64Null()
	}
	if api.Visibility != nil {
		m.Visible = types.StringValue(string(*api.Visibility))
	} else {
		m.Visible = types.StringNull()
	}
}

// PartitionLegendToTreemapLegend maps Terraform partition legend model to API treemap legend.
func PartitionLegendToTreemapLegend(m *models.PartitionLegendModel) kbapi.TreemapLegend {
	legend := kbapi.TreemapLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueInt64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.TreemapLegendVisibility(m.Visible.ValueString())
		legend.Visibility = &v
	}
	return legend
}

// PartitionLegendToMosaicLegend maps Terraform partition legend model to API mosaic legend.
func PartitionLegendToMosaicLegend(m *models.PartitionLegendModel) kbapi.MosaicLegend {
	legend := kbapi.MosaicLegend{Size: kbapi.LegendSize(m.Size.ValueString())}
	if typeutils.IsKnown(m.Nested) {
		legend.Nested = new(m.Nested.ValueBool())
	}
	if typeutils.IsKnown(m.TruncateAfterLine) {
		legend.TruncateAfterLines = new(float32(m.TruncateAfterLine.ValueInt64()))
	}
	if typeutils.IsKnown(m.Visible) {
		v := kbapi.MosaicLegendVisibility(m.Visible.ValueString())
		legend.Visibility = &v
	}
	return legend
}

// PartitionValueDisplayFromAPI maps API value display styling into Terraform model.
func PartitionValueDisplayFromAPI(m *models.PartitionValueDisplay, api kbapi.ValueDisplay) {
	m.Mode = typeutils.StringishPointerValue(api.Mode)
	if api.PercentDecimals != nil {
		m.PercentDecimals = types.Float64Value(float64(*api.PercentDecimals))
	} else {
		m.PercentDecimals = types.Float64Null()
	}
}

// PartitionValueDisplayToAPI maps Terraform partition value display model to API.
func PartitionValueDisplayToAPI(m *models.PartitionValueDisplay) kbapi.ValueDisplay {
	vd := kbapi.ValueDisplay{}
	if typeutils.IsKnown(m.Mode) {
		mode := kbapi.ValueDisplayMode(m.Mode.ValueString())
		vd.Mode = &mode
	}
	if typeutils.IsKnown(m.PercentDecimals) {
		vd.PercentDecimals = new(float32(m.PercentDecimals.ValueFloat64()))
	}
	return vd
}
