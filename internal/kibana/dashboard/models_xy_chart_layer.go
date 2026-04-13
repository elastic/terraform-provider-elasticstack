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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// xyLayerModel represents a layer in an XY chart
type xyLayerModel struct {
	Type               types.String             `tfsdk:"type"`
	DataLayer          *dataLayerModel          `tfsdk:"data_layer"`
	ReferenceLineLayer *referenceLineLayerModel `tfsdk:"reference_line_layer"`
}

// dataLayerModel represents a data layer (NoESQL or ESQL)
type dataLayerModel struct {
	DataSourceJSON      jsontypes.Normalized `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	XJSON               jsontypes.Normalized `tfsdk:"x_json"`
	Y                   []yMetricModel       `tfsdk:"y"`
	BreakdownByJSON     jsontypes.Normalized `tfsdk:"breakdown_by_json"`
}

// yMetricModel represents a Y-axis metric
type yMetricModel struct {
	ConfigJSON jsontypes.Normalized `tfsdk:"config_json"`
}

// referenceLineLayerModel represents a reference line layer
type referenceLineLayerModel struct {
	DataSourceJSON      jsontypes.Normalized `tfsdk:"data_source_json"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	Thresholds          []thresholdModel     `tfsdk:"thresholds"`
}

// thresholdModel represents a reference line threshold
type thresholdModel struct {
	Axis        types.String         `tfsdk:"axis"`
	ColorJSON   jsontypes.Normalized `tfsdk:"color_json"`
	Column      types.String         `tfsdk:"column"`
	ValueJSON   jsontypes.Normalized `tfsdk:"value_json"`
	Fill        types.String         `tfsdk:"fill"`
	Icon        types.String         `tfsdk:"icon"`
	Operation   types.String         `tfsdk:"operation"`
	StrokeDash  types.String         `tfsdk:"stroke_dash"`
	StrokeWidth types.Float64        `tfsdk:"stroke_width"`
	Text        types.String         `tfsdk:"text"`
}

func xyReferenceLineLayerTypeFromTF(tfType string) kbapi.XyReferenceLineLayerNoESQLType {
	return kbapi.XyReferenceLineLayerNoESQLType(tfType)
}

// fromAPILayersNoESQL populates the layer model from a DSL (non-ES|QL) XY layer union value.
func (m *xyLayerModel) fromAPILayersNoESQL(apiLayer kbapi.XyLayersNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	layerJSON, err := apiLayer.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal layer", err.Error())
		return diags
	}

	var layerType struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(layerJSON, &layerType); err != nil {
		diags.AddError("Failed to determine layer type", err.Error())
		return diags
	}
	m.Type = types.StringValue(layerType.Type)

	isReferenceLine := layerType.Type == "referenceLines" || layerType.Type == string(kbapi.ReferenceLines)
	if isReferenceLine {
		refLine, err := apiLayer.AsXyReferenceLineLayerNoESQL()
		if err != nil {
			diags.AddError("Failed to parse reference line layer", err.Error())
			return diags
		}
		m.ReferenceLineLayer = &referenceLineLayerModel{}
		return m.ReferenceLineLayer.fromAPINoESQL(refLine)
	}

	if layerType.Type == string(kbapi.Annotations) {
		diags.AddError("Unsupported XY layer type", "annotation layers are not supported by this resource")
		return diags
	}

	dl, err := apiLayer.AsXyLayerNoESQL()
	if err != nil {
		diags.AddError("Failed to parse data layer", err.Error())
		return diags
	}
	m.DataLayer = &dataLayerModel{}
	return m.DataLayer.fromAPINoESQL(dl)
}

// fromAPILayerESQL populates the layer model from an ES|QL XY data layer.
func (m *xyLayerModel) fromAPILayerESQL(apiLayer kbapi.XyLayerESQL) diag.Diagnostics {
	m.Type = types.StringValue(string(apiLayer.Type))
	m.DataLayer = &dataLayerModel{}
	return m.DataLayer.fromAPIESql(apiLayer)
}

// toAPILayersNoESQL converts the layer model to the DSL layer union type.
func (m *xyLayerModel) toAPILayersNoESQL() (kbapi.XyLayersNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out kbapi.XyLayersNoESQL

	if m.ReferenceLineLayer != nil {
		ref, refDiags := m.ReferenceLineLayer.toAPIXyReferenceLineLayerNoESQL(m.Type.ValueString())
		diags.Append(refDiags...)
		if diags.HasError() {
			return out, diags
		}
		if err := out.FromXyReferenceLineLayerNoESQL(ref); err != nil {
			diags.AddError("Failed to build reference line layer", err.Error())
		}
		return out, diags
	}

	if m.DataLayer != nil {
		dl, dataDiags := m.DataLayer.toAPIXyLayerNoESQL(m.Type.ValueString())
		diags.Append(dataDiags...)
		if diags.HasError() {
			return out, diags
		}
		if err := out.FromXyLayerNoESQL(dl); err != nil {
			diags.AddError("Failed to build data layer", err.Error())
		}
		return out, diags
	}

	diags.AddError("Invalid layer", "Layer must have either data_layer or reference_line_layer configured")
	return out, diags
}

// toAPILayerESQL converts a configured data layer to the ES|QL API layer type.
func (m *xyLayerModel) toAPILayerESQL() (kbapi.XyLayerESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.XyLayerESQL
	if m.DataLayer == nil {
		diags.AddError("Invalid layer", "ES|QL XY charts require a data_layer")
		return zero, diags
	}
	return m.DataLayer.toAPIXyLayerESQL(m.Type.ValueString())
}

// fromAPINoESQL populates data layer from NoESQL API response
func (m *dataLayerModel) fromAPINoESQL(apiLayer kbapi.XyLayerNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.DataSource)
	if err == nil {
		m.DataSourceJSON = jsontypes.NewNormalizedValue(string(datasetJSON))
	} else {
		diags.AddError("Failed to marshal dataset", err.Error())
	}

	m.IgnoreGlobalFilters = types.BoolPointerValue(apiLayer.IgnoreGlobalFilters)
	if apiLayer.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiLayer.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	if apiLayer.X != nil {
		xJSON, err := json.Marshal(apiLayer.X)
		if err == nil {
			m.XJSON = jsontypes.NewNormalizedValue(string(xJSON))
		}
	}

	if apiLayer.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiLayer.BreakdownBy)
		if err == nil {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	}

	// Convert Y metrics
	if len(apiLayer.Y) > 0 {
		m.Y = make([]yMetricModel, 0, len(apiLayer.Y))
		for _, y := range apiLayer.Y {
			yJSON, err := json.Marshal(y)
			if err == nil {
				m.Y = append(m.Y, yMetricModel{
					ConfigJSON: jsontypes.NewNormalizedValue(string(yJSON)),
				})
			}
		}
	}

	return diags
}

// fromAPIESql populates data layer from ESQL API response
func (m *dataLayerModel) fromAPIESql(apiLayer kbapi.XyLayerESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.DataSource)
	if err == nil {
		m.DataSourceJSON = jsontypes.NewNormalizedValue(string(datasetJSON))
	} else {
		diags.AddError("Failed to marshal dataset", err.Error())
	}

	m.IgnoreGlobalFilters = types.BoolPointerValue(apiLayer.IgnoreGlobalFilters)
	if apiLayer.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiLayer.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	if apiLayer.X != nil {
		xJSON, err := json.Marshal(apiLayer.X)
		if err == nil {
			m.XJSON = jsontypes.NewNormalizedValue(string(xJSON))
		}
	}

	if apiLayer.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiLayer.BreakdownBy)
		if err == nil {
			m.BreakdownByJSON = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	}

	// Convert Y metrics
	if len(apiLayer.Y) > 0 {
		m.Y = make([]yMetricModel, 0, len(apiLayer.Y))
		for _, y := range apiLayer.Y {
			yJSON, err := json.Marshal(y)
			if err == nil {
				m.Y = append(m.Y, yMetricModel{
					ConfigJSON: jsontypes.NewNormalizedValue(string(yJSON)),
				})
			}
		}
	}

	return diags
}

// toAPIXyLayerNoESQL converts a data layer model to the typed non-ES|QL API layer.
func (m *dataLayerModel) toAPIXyLayerNoESQL(layerType string) (kbapi.XyLayerNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	layer := kbapi.XyLayerNoESQL{Type: kbapi.XyLayerNoESQLType(layerType)}

	if typeutils.IsKnown(m.DataSourceJSON) {
		diags.Append(m.DataSourceJSON.Unmarshal(&layer.DataSource)...)
	}

	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		layer.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		s := float32(m.Sampling.ValueFloat64())
		layer.Sampling = &s
	}

	if typeutils.IsKnown(m.XJSON) {
		var x kbapi.XyLayerNoESQL_X
		diags.Append(m.XJSON.Unmarshal(&x)...)
		if !diags.HasError() {
			layer.X = &x
		}
	}

	if typeutils.IsKnown(m.BreakdownByJSON) {
		var bb kbapi.XyLayerNoESQL_BreakdownBy
		diags.Append(m.BreakdownByJSON.Unmarshal(&bb)...)
		if !diags.HasError() {
			layer.BreakdownBy = &bb
		}
	}

	if len(m.Y) > 0 {
		layer.Y = make([]kbapi.XyLayerNoESQL_Y_Item, 0, len(m.Y))
		for _, y := range m.Y {
			if !typeutils.IsKnown(y.ConfigJSON) {
				continue
			}
			var item kbapi.XyLayerNoESQL_Y_Item
			diags.Append(y.ConfigJSON.Unmarshal(&item)...)
			layer.Y = append(layer.Y, item)
		}
	}

	return layer, diags
}

// toAPIXyLayerESQL converts a data layer model to the typed ES|QL API layer.
func (m *dataLayerModel) toAPIXyLayerESQL(layerType string) (kbapi.XyLayerESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.XyLayerESQL

	layer := map[string]any{
		"type": layerType,
	}
	if typeutils.IsKnown(m.DataSourceJSON) {
		var ds any
		diags.Append(m.DataSourceJSON.Unmarshal(&ds)...)
		layer["data_source"] = ds
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		layer["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBool()
	}
	if typeutils.IsKnown(m.Sampling) {
		layer["sampling"] = m.Sampling.ValueFloat64()
	}
	if typeutils.IsKnown(m.XJSON) {
		var x any
		diags.Append(m.XJSON.Unmarshal(&x)...)
		layer["x"] = x
	}
	if typeutils.IsKnown(m.BreakdownByJSON) {
		var bb any
		diags.Append(m.BreakdownByJSON.Unmarshal(&bb)...)
		layer["breakdown_by"] = bb
	}
	if len(m.Y) > 0 {
		yMetrics := make([]any, 0, len(m.Y))
		for _, y := range m.Y {
			if !typeutils.IsKnown(y.ConfigJSON) {
				continue
			}
			var yc any
			diags.Append(y.ConfigJSON.Unmarshal(&yc)...)
			yMetrics = append(yMetrics, yc)
		}
		layer["y"] = yMetrics
	}

	if diags.HasError() {
		return zero, diags
	}

	layerJSON, err := json.Marshal(layer)
	if err != nil {
		diags.AddError("Failed to marshal ES|QL data layer", err.Error())
		return zero, diags
	}

	var out kbapi.XyLayerESQL
	if err := json.Unmarshal(layerJSON, &out); err != nil {
		diags.AddError("Failed to decode ES|QL data layer", err.Error())
		return zero, diags
	}
	return out, diags
}

// fromAPINoESQL populates reference line layer from NoESQL API response
func (m *referenceLineLayerModel) fromAPINoESQL(apiLayer kbapi.XyReferenceLineLayerNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.DataSource)
	if err == nil {
		m.DataSourceJSON = jsontypes.NewNormalizedValue(string(datasetJSON))
	} else {
		diags.AddError("Failed to marshal dataset", err.Error())
	}

	m.IgnoreGlobalFilters = types.BoolPointerValue(apiLayer.IgnoreGlobalFilters)
	if apiLayer.Sampling != nil {
		m.Sampling = types.Float64Value(float64(*apiLayer.Sampling))
	} else {
		m.Sampling = types.Float64Null()
	}

	// Convert thresholds
	if len(apiLayer.Thresholds) > 0 {
		m.Thresholds = make([]thresholdModel, 0, len(apiLayer.Thresholds))
		for _, t := range apiLayer.Thresholds {
			thresholdJSON, err := json.Marshal(t)
			if err != nil {
				continue
			}

			var probe map[string]any
			_ = json.Unmarshal(thresholdJSON, &probe)
			if _, hasAxis := probe["axis"].(string); hasAxis || probe["column"] != nil {
				var tm thresholdModel
				tmDiags := tm.fromAPIJSON(thresholdJSON)
				diags.Append(tmDiags...)
				if !tmDiags.HasError() {
					m.Thresholds = append(m.Thresholds, tm)
					continue
				}
			}

			// NoESQL reference line thresholds are operation definitions (count, sum, static_value, formula, etc)
			// rather than the richer object shape used by ES|QL reference lines.
			m.Thresholds = append(m.Thresholds, thresholdModel{
				Axis:        types.StringNull(),
				ColorJSON:   jsontypes.NewNormalizedNull(),
				Column:      types.StringNull(),
				ValueJSON:   jsontypes.NewNormalizedValue(string(thresholdJSON)),
				Fill:        types.StringNull(),
				Icon:        types.StringNull(),
				Operation:   types.StringNull(),
				StrokeDash:  types.StringNull(),
				StrokeWidth: types.Float64Null(),
				Text:        types.StringNull(),
			})
		}
	}

	return diags
}

// toAPIXyReferenceLineLayerNoESQL converts a reference line layer model to the typed API layer.
func (m *referenceLineLayerModel) toAPIXyReferenceLineLayerNoESQL(layerType string) (kbapi.XyReferenceLineLayerNoESQL, diag.Diagnostics) {
	var diags diag.Diagnostics
	layer := kbapi.XyReferenceLineLayerNoESQL{
		Type: xyReferenceLineLayerTypeFromTF(layerType),
	}

	if typeutils.IsKnown(m.DataSourceJSON) {
		diags.Append(m.DataSourceJSON.Unmarshal(&layer.DataSource)...)
	}
	if typeutils.IsKnown(m.IgnoreGlobalFilters) {
		layer.IgnoreGlobalFilters = new(m.IgnoreGlobalFilters.ValueBool())
	}
	if typeutils.IsKnown(m.Sampling) {
		s := float32(m.Sampling.ValueFloat64())
		layer.Sampling = &s
	}

	if len(m.Thresholds) > 0 {
		items := make([]kbapi.XyReferenceLineLayerNoESQL_Thresholds_Item, 0, len(m.Thresholds))
		for _, t := range m.Thresholds {
			if typeutils.IsKnown(t.ValueJSON) {
				var op any
				valueDiags := t.ValueJSON.Unmarshal(&op)
				diags.Append(valueDiags...)
				if valueDiags.HasError() {
					continue
				}
				opBytes, err := json.Marshal(op)
				if err != nil {
					diags.AddError("Failed to marshal reference line threshold", err.Error())
					continue
				}
				var item kbapi.XyReferenceLineLayerNoESQL_Thresholds_Item
				if err := item.UnmarshalJSON(opBytes); err != nil {
					diags.AddError("Failed to decode reference line threshold", err.Error())
					continue
				}
				items = append(items, item)
				continue
			}

			thresholdMap, tDiags := t.toAPI()
			diags.Append(tDiags...)
			if tDiags.HasError() {
				continue
			}
			thBytes, err := json.Marshal(thresholdMap)
			if err != nil {
				diags.AddError("Failed to marshal reference line threshold", err.Error())
				continue
			}
			var item kbapi.XyReferenceLineLayerNoESQL_Thresholds_Item
			if err := item.UnmarshalJSON(thBytes); err != nil {
				diags.AddError("Failed to decode reference line threshold", err.Error())
				continue
			}
			items = append(items, item)
		}
		layer.Thresholds = items
	}

	return layer, diags
}

// fromAPIJSON populates threshold from JSON
func (m *thresholdModel) fromAPIJSON(jsonData []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	var thresholdData map[string]any
	if err := json.Unmarshal(jsonData, &thresholdData); err != nil {
		diags.AddError("Failed to unmarshal threshold", err.Error())
		return diags
	}

	if axis, ok := thresholdData["axis"].(string); ok {
		m.Axis = types.StringValue(axis)
	} else {
		m.Axis = types.StringNull()
	}

	if color, ok := thresholdData["color"]; ok {
		colorJSON, err := json.Marshal(color)
		if err == nil {
			m.ColorJSON = jsontypes.NewNormalizedValue(string(colorJSON))
		}
	}

	if column, ok := thresholdData["column"].(string); ok {
		m.Column = types.StringValue(column)
	} else {
		m.Column = types.StringNull()
	}

	if value, ok := thresholdData["value"]; ok {
		valueJSON, err := json.Marshal(value)
		if err == nil {
			m.ValueJSON = jsontypes.NewNormalizedValue(string(valueJSON))
		}
	}

	if fill, ok := thresholdData["fill"].(string); ok {
		m.Fill = types.StringValue(fill)
	} else {
		m.Fill = types.StringNull()
	}

	if icon, ok := thresholdData["icon"].(string); ok {
		m.Icon = types.StringValue(icon)
	} else {
		m.Icon = types.StringNull()
	}

	if operation, ok := thresholdData["operation"].(string); ok {
		m.Operation = types.StringValue(operation)
	} else {
		m.Operation = types.StringNull()
	}

	if strokeDash, ok := thresholdData["stroke_dash"].(string); ok {
		m.StrokeDash = types.StringValue(strokeDash)
	} else {
		m.StrokeDash = types.StringNull()
	}

	if strokeWidth, ok := thresholdData["stroke_width"].(float64); ok {
		m.StrokeWidth = types.Float64Value(strokeWidth)
	} else {
		m.StrokeWidth = types.Float64Null()
	}

	if text, ok := thresholdData["text"].(string); ok {
		m.Text = types.StringValue(text)
	} else {
		m.Text = types.StringNull()
	}

	return diags
}

// toAPI converts threshold to API map
func (m *thresholdModel) toAPI() (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	threshold := make(map[string]any)

	if typeutils.IsKnown(m.Axis) {
		threshold["axis"] = m.Axis.ValueString()
	}

	if typeutils.IsKnown(m.ColorJSON) {
		var color any
		diags.Append(m.ColorJSON.Unmarshal(&color)...)
		threshold["color"] = color
	}

	if typeutils.IsKnown(m.Column) {
		threshold["column"] = m.Column.ValueString()
	}

	if typeutils.IsKnown(m.ValueJSON) {
		var value any
		diags.Append(m.ValueJSON.Unmarshal(&value)...)
		threshold["value"] = value
	}

	if typeutils.IsKnown(m.Fill) {
		threshold["fill"] = m.Fill.ValueString()
	}

	if typeutils.IsKnown(m.Icon) {
		threshold["icon"] = m.Icon.ValueString()
	}

	if typeutils.IsKnown(m.Operation) {
		threshold["operation"] = m.Operation.ValueString()
	}

	if typeutils.IsKnown(m.StrokeDash) {
		threshold["stroke_dash"] = m.StrokeDash.ValueString()
	}

	if typeutils.IsKnown(m.StrokeWidth) {
		threshold["stroke_width"] = m.StrokeWidth.ValueFloat64()
	}

	if typeutils.IsKnown(m.Text) {
		threshold["text"] = m.Text.ValueString()
	}

	return threshold, diags
}
