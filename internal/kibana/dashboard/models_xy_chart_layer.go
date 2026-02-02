package dashboard

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	Dataset             jsontypes.Normalized `tfsdk:"dataset"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	X                   jsontypes.Normalized `tfsdk:"x"`
	Y                   []yMetricModel       `tfsdk:"y"`
	BreakdownBy         jsontypes.Normalized `tfsdk:"breakdown_by"`
}

// yMetricModel represents a Y-axis metric
type yMetricModel struct {
	Config jsontypes.Normalized `tfsdk:"config"`
}

// referenceLineLayerModel represents a reference line layer
type referenceLineLayerModel struct {
	Dataset             jsontypes.Normalized `tfsdk:"dataset"`
	IgnoreGlobalFilters types.Bool           `tfsdk:"ignore_global_filters"`
	Sampling            types.Float64        `tfsdk:"sampling"`
	Thresholds          []thresholdModel     `tfsdk:"thresholds"`
}

// thresholdModel represents a reference line threshold
type thresholdModel struct {
	Axis        types.String         `tfsdk:"axis"`
	Color       jsontypes.Normalized `tfsdk:"color"`
	Column      types.String         `tfsdk:"column"`
	Value       jsontypes.Normalized `tfsdk:"value"`
	Fill        types.String         `tfsdk:"fill"`
	Icon        types.String         `tfsdk:"icon"`
	Operation   types.String         `tfsdk:"operation"`
	StrokeDash  types.String         `tfsdk:"stroke_dash"`
	StrokeWidth types.Float64        `tfsdk:"stroke_width"`
	Text        types.String         `tfsdk:"text"`
}

// fromAPI populates the layer model from API response
func (m *xyLayerModel) fromAPI(apiLayer kbapi.XyChartSchema_Layers_Item) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to determine which layer type it is by marshaling and unmarshaling
	var layerType struct {
		Type string `json:"type"`
	}
	layerJSON, err := apiLayer.MarshalJSON()
	if err != nil {
		diags.AddError("Failed to marshal layer", err.Error())
		return diags
	}
	if err := json.Unmarshal(layerJSON, &layerType); err != nil {
		diags.AddError("Failed to determine layer type", err.Error())
		return diags
	}

	m.Type = types.StringValue(layerType.Type)

	// Check if it's a reference line layer
	isReferenceLine := layerType.Type == "reference_line"

	if isReferenceLine {
		// Try NoESQL reference line first
		refLineNoESql, err := apiLayer.AsXyReferenceLineLayerNoESQL()
		if err == nil {
			m.ReferenceLineLayer = &referenceLineLayerModel{}
			return m.ReferenceLineLayer.fromAPINoESQL(refLineNoESql)
		}

		// Try ESQL reference line
		refLineESql, err := apiLayer.AsXyReferenceLineLayerESQL()
		if err == nil {
			m.ReferenceLineLayer = &referenceLineLayerModel{}
			return m.ReferenceLineLayer.fromAPIESql(refLineESql)
		}

		diags.AddError("Failed to parse reference line layer", "Unable to parse as NoESQL or ESQL reference line")
		return diags
	}

	// It's a data layer - try NoESQL first
	dataLayerNoESql, err := apiLayer.AsXyLayerNoESQL()
	if err == nil {
		m.DataLayer = &dataLayerModel{}
		return m.DataLayer.fromAPINoESQL(dataLayerNoESql)
	}

	// Try ESQL data layer
	dataLayerESql, err := apiLayer.AsXyLayerESQL()
	if err == nil {
		m.DataLayer = &dataLayerModel{}
		return m.DataLayer.fromAPIESql(dataLayerESql)
	}

	diags.AddError("Failed to parse data layer", "Unable to parse as NoESQL or ESQL data layer")
	return diags
}

// toAPI converts the layer model to API format
func (m *xyLayerModel) toAPI() (kbapi.XyChartSchema_Layers_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result kbapi.XyChartSchema_Layers_Item

	if m.ReferenceLineLayer != nil {
		// Convert reference line layer
		refLineJSON, refDiags := m.ReferenceLineLayer.toAPI(m.Type.ValueString())
		diags.Append(refDiags...)
		if !refDiags.HasError() {
			if err := result.UnmarshalJSON(refLineJSON); err != nil {
				diags.AddError("Failed to unmarshal reference line layer", err.Error())
			}
		}
		return result, diags
	}

	if m.DataLayer != nil {
		// Convert data layer
		dataJSON, dataDiags := m.DataLayer.toAPI(m.Type.ValueString())
		diags.Append(dataDiags...)
		if !dataDiags.HasError() {
			if err := result.UnmarshalJSON(dataJSON); err != nil {
				diags.AddError("Failed to unmarshal data layer", err.Error())
			}
		}
		return result, diags
	}

	diags.AddError("Invalid layer", "Layer must have either data_layer or reference_line_layer configured")
	return result, diags
}

// fromAPINoESQL populates data layer from NoESQL API response
func (m *dataLayerModel) fromAPINoESQL(apiLayer kbapi.XyLayerNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.Dataset)
	if err == nil {
		m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))
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
			m.X = jsontypes.NewNormalizedValue(string(xJSON))
		}
	}

	if apiLayer.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiLayer.BreakdownBy)
		if err == nil {
			m.BreakdownBy = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	}

	// Convert Y metrics
	if len(apiLayer.Y) > 0 {
		m.Y = make([]yMetricModel, 0, len(apiLayer.Y))
		for _, y := range apiLayer.Y {
			yJSON, err := json.Marshal(y)
			if err == nil {
				m.Y = append(m.Y, yMetricModel{
					Config: jsontypes.NewNormalizedValue(string(yJSON)),
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
	datasetJSON, err := json.Marshal(apiLayer.Dataset)
	if err == nil {
		m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))
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
			m.X = jsontypes.NewNormalizedValue(string(xJSON))
		}
	}

	if apiLayer.BreakdownBy != nil {
		breakdownJSON, err := json.Marshal(apiLayer.BreakdownBy)
		if err == nil {
			m.BreakdownBy = jsontypes.NewNormalizedValue(string(breakdownJSON))
		}
	}

	// Convert Y metrics
	if len(apiLayer.Y) > 0 {
		m.Y = make([]yMetricModel, 0, len(apiLayer.Y))
		for _, y := range apiLayer.Y {
			yJSON, err := json.Marshal(y)
			if err == nil {
				m.Y = append(m.Y, yMetricModel{
					Config: jsontypes.NewNormalizedValue(string(yJSON)),
				})
			}
		}
	}

	return diags
}

// toAPI converts data layer to API JSON
func (m *dataLayerModel) toAPI(layerType string) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Build a map with all the fields
	layer := map[string]interface{}{
		"type": layerType,
	}

	if utils.IsKnown(m.Dataset) {
		var dataset interface{}
		diags.Append(m.Dataset.Unmarshal(&dataset)...)
		layer["dataset"] = dataset
	}

	if utils.IsKnown(m.IgnoreGlobalFilters) {
		layer["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBool()
	}

	if utils.IsKnown(m.Sampling) {
		layer["sampling"] = m.Sampling.ValueFloat64()
	}

	if utils.IsKnown(m.X) {
		var x interface{}
		diags.Append(m.X.Unmarshal(&x)...)
		layer["x"] = x
	}

	if utils.IsKnown(m.BreakdownBy) {
		var breakdownBy interface{}
		diags.Append(m.BreakdownBy.Unmarshal(&breakdownBy)...)
		layer["breakdown_by"] = breakdownBy
	}

	// Convert Y metrics
	if len(m.Y) > 0 {
		yMetrics := make([]interface{}, 0, len(m.Y))
		for _, y := range m.Y {
			if utils.IsKnown(y.Config) {
				var yConfig interface{}
				diags.Append(y.Config.Unmarshal(&yConfig)...)
				yMetrics = append(yMetrics, yConfig)
			}
		}
		layer["y"] = yMetrics
	}

	if diags.HasError() {
		return nil, diags
	}

	layerJSON, err := json.Marshal(layer)
	if err != nil {
		diags.AddError("Failed to marshal layer", err.Error())
		return nil, diags
	}

	return layerJSON, diags
}

// fromAPINoESQL populates reference line layer from NoESQL API response
func (m *referenceLineLayerModel) fromAPINoESQL(apiLayer kbapi.XyReferenceLineLayerNoESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.Dataset)
	if err == nil {
		m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))
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

			var threshold thresholdModel
			thresholdDiags := threshold.fromAPIJSON(thresholdJSON)
			diags.Append(thresholdDiags...)
			if !thresholdDiags.HasError() {
				m.Thresholds = append(m.Thresholds, threshold)
			}
		}
	}

	return diags
}

// fromAPIESql populates reference line layer from ESQL API response
func (m *referenceLineLayerModel) fromAPIESql(apiLayer kbapi.XyReferenceLineLayerESQL) diag.Diagnostics {
	var diags diag.Diagnostics

	// Marshal to JSON to preserve the exact structure
	datasetJSON, err := json.Marshal(apiLayer.Dataset)
	if err == nil {
		m.Dataset = jsontypes.NewNormalizedValue(string(datasetJSON))
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

			var threshold thresholdModel
			thresholdDiags := threshold.fromAPIJSON(thresholdJSON)
			diags.Append(thresholdDiags...)
			if !thresholdDiags.HasError() {
				m.Thresholds = append(m.Thresholds, threshold)
			}
		}
	}

	return diags
}

// toAPI converts reference line layer to API JSON
func (m *referenceLineLayerModel) toAPI(layerType string) (json.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Build a map with all the fields
	layer := map[string]interface{}{
		"type": layerType,
	}

	if utils.IsKnown(m.Dataset) {
		var dataset interface{}
		diags.Append(m.Dataset.Unmarshal(&dataset)...)
		layer["dataset"] = dataset
	}

	if utils.IsKnown(m.IgnoreGlobalFilters) {
		layer["ignore_global_filters"] = m.IgnoreGlobalFilters.ValueBool()
	}

	if utils.IsKnown(m.Sampling) {
		layer["sampling"] = m.Sampling.ValueFloat64()
	}

	// Convert thresholds
	if len(m.Thresholds) > 0 {
		thresholds := make([]interface{}, 0, len(m.Thresholds))
		for _, t := range m.Thresholds {
			thresholdMap, tDiags := t.toAPI()
			diags.Append(tDiags...)
			if !tDiags.HasError() {
				thresholds = append(thresholds, thresholdMap)
			}
		}
		layer["thresholds"] = thresholds
	}

	if diags.HasError() {
		return nil, diags
	}

	layerJSON, err := json.Marshal(layer)
	if err != nil {
		diags.AddError("Failed to marshal reference line layer", err.Error())
		return nil, diags
	}

	return layerJSON, diags
}

// fromAPIJSON populates threshold from JSON
func (m *thresholdModel) fromAPIJSON(jsonData []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	var thresholdData map[string]interface{}
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
			m.Color = jsontypes.NewNormalizedValue(string(colorJSON))
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
			m.Value = jsontypes.NewNormalizedValue(string(valueJSON))
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
func (m *thresholdModel) toAPI() (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	threshold := make(map[string]interface{})

	if utils.IsKnown(m.Axis) {
		threshold["axis"] = m.Axis.ValueString()
	}

	if utils.IsKnown(m.Color) {
		var color interface{}
		diags.Append(m.Color.Unmarshal(&color)...)
		threshold["color"] = color
	}

	if utils.IsKnown(m.Column) {
		threshold["column"] = m.Column.ValueString()
	}

	if utils.IsKnown(m.Value) {
		var value interface{}
		diags.Append(m.Value.Unmarshal(&value)...)
		threshold["value"] = value
	}

	if utils.IsKnown(m.Fill) {
		threshold["fill"] = m.Fill.ValueString()
	}

	if utils.IsKnown(m.Icon) {
		threshold["icon"] = m.Icon.ValueString()
	}

	if utils.IsKnown(m.Operation) {
		threshold["operation"] = m.Operation.ValueString()
	}

	if utils.IsKnown(m.StrokeDash) {
		threshold["stroke_dash"] = m.StrokeDash.ValueString()
	}

	if utils.IsKnown(m.StrokeWidth) {
		threshold["stroke_width"] = m.StrokeWidth.ValueFloat64()
	}

	if utils.IsKnown(m.Text) {
		threshold["text"] = m.Text.ValueString()
	}

	return threshold, diags
}
