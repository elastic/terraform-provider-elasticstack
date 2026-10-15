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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataLayerFromAPINoESQL_preservesPriorYWhenKibanaInjectsAxis(t *testing.T) {
	t.Parallel()

	prior := `{"empty_as_null":true,"operation":"count"}`
	m := &models.DataLayerModel{
		Y: []models.YMetricModel{{
			ConfigJSON: jsontypes.NewNormalizedValue(prior),
		}},
	}

	var yItem kbapi.KibanaHTTPAPIsVisXyLayerNoESQL_Y_Item
	require.NoError(t, json.Unmarshal([]byte(`{"operation":"count","empty_as_null":true,"axis":"y","color":{"type":"auto"}}`), &yItem))

	diags := dataLayerFromAPINoESQL(t.Context(), m, kbapi.KibanaHTTPAPIsVisXyLayerNoESQL{
		Y: []kbapi.KibanaHTTPAPIsVisXyLayerNoESQL_Y_Item{yItem},
	})
	require.False(t, diags.HasError(), "%v", diags)
	require.Len(t, m.Y, 1)
	assert.JSONEq(t, prior, m.Y[0].ConfigJSON.ValueString())
}

func TestReferenceLineLayerToAPI_thresholdAxisHandling(t *testing.T) {
	t.Parallel()

	const staticOp = `{"operation":"static_value","value":42}`
	tests := []struct {
		name      string
		axes      []types.String
		values    []string
		wantAxes  []any // nil = key absent
		wantError bool
	}{
		{name: "explicit axis is sent", axes: []types.String{types.StringValue("y2")}, values: []string{staticOp}, wantAxes: []any{"y2"}},
		{name: "null axis is omitted", axes: []types.String{types.StringNull()}, values: []string{staticOp}, wantAxes: []any{nil}},
		{name: "unknown axis is omitted", axes: []types.String{types.StringUnknown()}, values: []string{staticOp}, wantAxes: []any{nil}},
		{name: "equal axis in value_json is fine", axes: []types.String{types.StringValue("x")}, values: []string{`{"operation":"static_value","value":42,"axis":"x"}`}, wantAxes: []any{"x"}},
		{name: "conflicting axis in value_json is an error", axes: []types.String{types.StringValue("y2")}, values: []string{`{"operation":"static_value","value":42,"axis":"x"}`}, wantError: true},
		{name: "non-object value_json with axis is an error", axes: []types.String{types.StringValue("y")}, values: []string{`42`}, wantError: true},
		{name: "null value_json with axis is an error", axes: []types.String{types.StringValue("y")}, values: []string{`null`}, wantError: true},
		{name: "non-object value_json without axis is untouched", axes: []types.String{types.StringNull()}, values: []string{`42`}, wantAxes: []any{nil}},
		{
			name:     "mixed axes across thresholds",
			axes:     []types.String{types.StringValue("x"), types.StringNull(), types.StringValue("y2")},
			values:   []string{staticOp, staticOp, `{"operation":"count"}`},
			wantAxes: []any{"x", nil, "y2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			thresholds := make([]models.ThresholdModel, len(tt.axes))
			for i := range tt.axes {
				thresholds[i] = models.ThresholdModel{Axis: tt.axes[i], ValueJSON: jsontypes.NewNormalizedValue(tt.values[i])}
			}
			layer, diags := referenceLineLayerToAPIXyReferenceLineLayerNoESQL(&models.ReferenceLineLayerModel{
				DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
				Thresholds:     thresholds,
			}, "reference_lines")
			if tt.wantError {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError(), "%v", diags)
			require.Len(t, layer.Thresholds, len(tt.wantAxes))
			for i, want := range tt.wantAxes {
				raw, err := layer.Thresholds[i].MarshalJSON()
				require.NoError(t, err)
				var got map[string]any
				if json.Unmarshal(raw, &got) != nil {
					assert.Nil(t, want)
					continue
				}
				assert.Equal(t, want, got["axis"])
			}
		})
	}
}

func TestReferenceLineLayerFromAPI_thresholdAxisRoundTrip(t *testing.T) {
	t.Parallel()

	const (
		staticPlan = `{"operation":"static_value","value":42}`
		countPlan  = `{"operation":"count"}`

		staticAxisX = `{"operation":"static_value","value":42,"axis":"x"}`
		countAxisY2 = `{"operation":"count","axis":"y2"}`
	)
	tests := []struct {
		name         string
		planAxis     types.String
		planValue    string
		apiThreshold string
		wantAxis     types.String
	}{
		{name: "static x echoed", planAxis: types.StringValue("x"), planValue: staticPlan, apiThreshold: `{"operation":"static_value","value":42,"axis":"x"}`, wantAxis: types.StringValue("x")},
		{name: "static y2 echoed", planAxis: types.StringValue("y2"), planValue: staticPlan, apiThreshold: `{"operation":"static_value","value":42,"axis":"y2"}`, wantAxis: types.StringValue("y2")},
		{name: "static y stripped", planAxis: types.StringValue("y"), planValue: staticPlan, apiThreshold: `{"operation":"static_value","value":42}`, wantAxis: types.StringValue("y")},
		{name: "static omitted with injected y", planAxis: types.StringNull(), planValue: staticPlan, apiThreshold: `{"operation":"static_value","value":42,"axis":"y"}`, wantAxis: types.StringNull()},
		{name: "static axis inside value_json", planAxis: types.StringNull(), planValue: staticAxisX, apiThreshold: staticAxisX, wantAxis: types.StringNull()},
		{name: "count x echoed", planAxis: types.StringValue("x"), planValue: countPlan, apiThreshold: `{"operation":"count","axis":"x"}`, wantAxis: types.StringValue("x")},
		{name: "count y2 echoed", planAxis: types.StringValue("y2"), planValue: countPlan, apiThreshold: `{"operation":"count","axis":"y2"}`, wantAxis: types.StringValue("y2")},
		{name: "count y echoed", planAxis: types.StringValue("y"), planValue: countPlan, apiThreshold: `{"operation":"count","axis":"y"}`, wantAxis: types.StringValue("y")},
		{name: "count y stripped", planAxis: types.StringValue("y"), planValue: countPlan, apiThreshold: `{"operation":"count"}`, wantAxis: types.StringValue("y")},
		{name: "count omitted with injected y", planAxis: types.StringNull(), planValue: countPlan, apiThreshold: `{"operation":"count","axis":"y"}`, wantAxis: types.StringNull()},
		{name: "count axis inside value_json echoed", planAxis: types.StringNull(), planValue: countAxisY2, apiThreshold: countAxisY2, wantAxis: types.StringNull()},
		{name: "count axis y inside value_json stripped", planAxis: types.StringNull(), planValue: `{"operation":"count","axis":"y"}`, apiThreshold: `{"operation":"count"}`, wantAxis: types.StringNull()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var item kbapi.KibanaHTTPAPIsVisXyReferenceLineLayerNoESQL_Thresholds_Item
			require.NoError(t, json.Unmarshal([]byte(tt.apiThreshold), &item))
			state := &models.ReferenceLineLayerModel{}
			diags := referenceLineLayerFromAPINoESQL(state, kbapi.KibanaHTTPAPIsVisXyReferenceLineLayerNoESQL{
				Thresholds: []kbapi.KibanaHTTPAPIsVisXyReferenceLineLayerNoESQL_Thresholds_Item{item},
			})
			require.False(t, diags.HasError(), "%v", diags)
			require.Len(t, state.Thresholds, 1)

			plan := &models.ReferenceLineLayerModel{Thresholds: []models.ThresholdModel{{
				Axis:      tt.planAxis,
				ValueJSON: jsontypes.NewNormalizedValue(tt.planValue),
			}}}
			alignXYChartStateFromPlan(
				&models.XYChartConfigModel{Layers: []models.XYLayerModel{{ReferenceLineLayer: plan}}},
				&models.XYChartConfigModel{Layers: []models.XYLayerModel{{ReferenceLineLayer: state}}},
			)

			got := state.Thresholds[0]
			assert.True(t, tt.wantAxis.Equal(got.Axis), "axis: want %s got %s", tt.wantAxis, got.Axis)
			assert.JSONEq(t, tt.planValue, got.ValueJSON.ValueString())
		})
	}
}
