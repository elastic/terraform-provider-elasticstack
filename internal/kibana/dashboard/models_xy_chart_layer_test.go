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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dataLayerModel_fromAPINoESQL_toAPI(t *testing.T) {
	layerJSON := `{
		"type": "area",
		"dataset": {"type":"dataView","id":"logs-*"},
		"ignore_global_filters": true,
		"sampling": 0.5,
		"y": [{"operation":"count","color":"#68BC00","axis":"left"}]
	}`
	var apiLayer kbapi.XyLayerNoESQL
	require.NoError(t, json.Unmarshal([]byte(layerJSON), &apiLayer))

	model := &dataLayerModel{}
	diags := model.fromAPINoESQL(apiLayer)
	require.False(t, diags.HasError())

	assert.False(t, model.DatasetJSON.IsNull())
	assert.True(t, model.IgnoreGlobalFilters.ValueBool())
	assert.InDelta(t, 0.5, model.Sampling.ValueFloat64(), 0.001)
	assert.Len(t, model.Y, 1)
	assert.False(t, model.Y[0].ConfigJSON.IsNull())

	raw, diags := model.toAPI("area")
	require.False(t, diags.HasError())
	require.NotNil(t, raw)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(raw, &roundTrip))
	assert.Equal(t, "area", roundTrip["type"])
	assert.Equal(t, true, roundTrip["ignore_global_filters"])
	assert.InDelta(t, 0.5, roundTrip["sampling"], 0.001)
}

func Test_dataLayerModel_fromAPIESql_toAPI(t *testing.T) {
	layerJSON := `{
		"type": "area",
		"dataset": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"ignore_global_filters": false,
		"sampling": 1,
		"y": [{"operation":"value","column":"bytes","color":{"type":"static","color":"#68BC00"},"axis":"left"}]
	}`
	var apiLayer kbapi.XyLayerESQL
	require.NoError(t, json.Unmarshal([]byte(layerJSON), &apiLayer))

	model := &dataLayerModel{}
	diags := model.fromAPIESql(apiLayer)
	require.False(t, diags.HasError())

	assert.False(t, model.DatasetJSON.IsNull())
	assert.False(t, model.IgnoreGlobalFilters.ValueBool())
	assert.InDelta(t, 1, model.Sampling.ValueFloat64(), 0.001)
	assert.Len(t, model.Y, 1)

	raw, diags := model.toAPI("area")
	require.False(t, diags.HasError())
	require.NotNil(t, raw)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(raw, &roundTrip))
	assert.Equal(t, "area", roundTrip["type"])
}

func Test_referenceLineLayerModel_fromAPINoESQL_toAPI(t *testing.T) {
	layerJSON := `{
		"type": "referenceLines",
		"dataset": {"type":"dataView","id":"logs-*"},
		"ignore_global_filters": false,
		"sampling": 1,
		"thresholds": [{"type":"static","value":100}]
	}`
	var apiLayer kbapi.XyReferenceLineLayerNoESQL
	require.NoError(t, json.Unmarshal([]byte(layerJSON), &apiLayer))

	model := &referenceLineLayerModel{}
	diags := model.fromAPINoESQL(apiLayer)
	require.False(t, diags.HasError())

	assert.False(t, model.DatasetJSON.IsNull())
	assert.Len(t, model.Thresholds, 1)
	assert.False(t, model.Thresholds[0].ValueJSON.IsNull())

	raw, diags := model.toAPI("referenceLines")
	require.False(t, diags.HasError())
	require.NotNil(t, raw)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(raw, &roundTrip))
	assert.Equal(t, "referenceLines", roundTrip["type"])
	assert.NotNil(t, roundTrip["thresholds"])
}

func Test_referenceLineLayerModel_fromAPIESql_toAPI(t *testing.T) {
	layerJSON := `{
		"type": "referenceLines",
		"dataset": {"type":"esql","query":"FROM logs-* | LIMIT 10"},
		"ignore_global_filters": true,
		"sampling": 0.5,
		"thresholds": [{
			"axis": "left",
			"column": "bytes",
			"value": 1000,
			"color": {"type":"static","color":"#54B399"},
			"fill": "above",
			"operation": "value"
		}]
	}`
	var apiLayer kbapi.XyReferenceLineLayerESQL
	require.NoError(t, json.Unmarshal([]byte(layerJSON), &apiLayer))

	model := &referenceLineLayerModel{}
	diags := model.fromAPIESql(apiLayer)
	require.False(t, diags.HasError())

	assert.False(t, model.DatasetJSON.IsNull())
	assert.Len(t, model.Thresholds, 1)
	assert.Equal(t, types.StringValue("left"), model.Thresholds[0].Axis)
	assert.Equal(t, types.StringValue("bytes"), model.Thresholds[0].Column)
	assert.Equal(t, types.StringValue("above"), model.Thresholds[0].Fill)
	assert.Equal(t, types.StringValue("value"), model.Thresholds[0].Operation)

	raw, diags := model.toAPI("referenceLines")
	require.False(t, diags.HasError())
	require.NotNil(t, raw)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(raw, &roundTrip))
	assert.Equal(t, "referenceLines", roundTrip["type"])
}

func Test_thresholdModel_fromAPIJSON_toAPI(t *testing.T) {
	thresholdJSON := []byte(`{
		"axis": "left",
		"column": "bytes",
		"value": 1000,
		"color": {"type":"static","color":"#54B399"},
		"fill": "above",
		"icon": "alert",
		"operation": "value",
		"stroke_dash": "solid",
		"stroke_width": 2,
		"text": "Threshold"
	}`)

	model := &thresholdModel{}
	diags := model.fromAPIJSON(thresholdJSON)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("left"), model.Axis)
	assert.Equal(t, types.StringValue("bytes"), model.Column)
	assert.Equal(t, types.StringValue("above"), model.Fill)
	assert.Equal(t, types.StringValue("alert"), model.Icon)
	assert.Equal(t, types.StringValue("value"), model.Operation)
	assert.Equal(t, types.StringValue("solid"), model.StrokeDash)
	assert.Equal(t, types.Float64Value(2), model.StrokeWidth)
	assert.Equal(t, types.StringValue("Threshold"), model.Text)

	thresholdMap, diags := model.toAPI()
	require.False(t, diags.HasError())
	assert.Equal(t, "left", thresholdMap["axis"])
	assert.Equal(t, "bytes", thresholdMap["column"])
	assert.Equal(t, "above", thresholdMap["fill"])
	assert.Equal(t, "value", thresholdMap["operation"])
	assert.InDelta(t, 2.0, thresholdMap["stroke_width"], 0.001)
	assert.Equal(t, "Threshold", thresholdMap["text"])
}

func Test_xyLayerModel_fromAPI_toAPI_referenceLine(t *testing.T) {
	// Build a reference line layer via JSON (type referenceLines)
	layerJSON := `{
		"type": "referenceLines",
		"dataset": {"type":"dataView","id":"logs-*"},
		"ignore_global_filters": false,
		"sampling": 1,
		"thresholds": [{"type":"static","value":50}]
	}`
	var apiLayer kbapi.XyChart_Layers_Item
	require.NoError(t, apiLayer.UnmarshalJSON([]byte(layerJSON)))

	model := &xyLayerModel{}
	diags := model.fromAPI(apiLayer)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("referenceLines"), model.Type)
	require.NotNil(t, model.ReferenceLineLayer)
	assert.Nil(t, model.DataLayer)

	result, diags := model.toAPI()
	require.False(t, diags.HasError())

	resultJSON, err := result.MarshalJSON()
	require.NoError(t, err)
	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(resultJSON, &roundTrip))
	assert.Equal(t, "referenceLines", roundTrip["type"])
}

func Test_xyLayerModel_fromAPI_toAPI_dataLayer(t *testing.T) {
	layerJSON := `{
		"type": "area",
		"dataset": {"type":"dataView","id":"logs-*"},
		"y": [{"operation":"count","color":"#68BC00","axis":"left"}]
	}`
	var apiLayer kbapi.XyChart_Layers_Item
	require.NoError(t, apiLayer.UnmarshalJSON([]byte(layerJSON)))

	model := &xyLayerModel{}
	diags := model.fromAPI(apiLayer)
	require.False(t, diags.HasError())

	assert.Equal(t, types.StringValue("area"), model.Type)
	require.NotNil(t, model.DataLayer)
	assert.Nil(t, model.ReferenceLineLayer)
	assert.Len(t, model.DataLayer.Y, 1)

	result, diags := model.toAPI()
	require.False(t, diags.HasError())

	resultJSON, err := result.MarshalJSON()
	require.NoError(t, err)
	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(resultJSON, &roundTrip))
	assert.Equal(t, "area", roundTrip["type"])
}
