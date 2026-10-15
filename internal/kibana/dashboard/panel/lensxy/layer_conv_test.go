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

func TestReferenceLineLayerToAPI_thresholdAxisIsMergedIntoValueJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		axis       types.String
		valueJSON  string
		wantAxis   any
		wantHasKey bool
	}{
		{name: "explicit axis is sent", axis: types.StringValue("y2"), valueJSON: `{"operation":"static_value","value":42}`, wantAxis: "y2", wantHasKey: true},
		{name: "null axis is omitted", axis: types.StringNull(), valueJSON: `{"operation":"static_value","value":42}`},
		{name: "axis already inside value_json wins", axis: types.StringValue("y2"), valueJSON: `{"operation":"static_value","value":42,"axis":"x"}`, wantAxis: "x", wantHasKey: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			layer, diags := referenceLineLayerToAPIXyReferenceLineLayerNoESQL(&models.ReferenceLineLayerModel{
				DataSourceJSON: jsontypes.NewNormalizedValue(`{"type":"data_view_spec","index_pattern":"metrics-*"}`),
				Thresholds: []models.ThresholdModel{{
					Axis:      tt.axis,
					ValueJSON: jsontypes.NewNormalizedValue(tt.valueJSON),
				}},
			}, "reference_lines")
			require.False(t, diags.HasError(), "%v", diags)
			require.Len(t, layer.Thresholds, 1)

			raw, err := layer.Thresholds[0].MarshalJSON()
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, json.Unmarshal(raw, &got))
			axis, has := got["axis"]
			assert.Equal(t, tt.wantHasKey, has)
			assert.Equal(t, tt.wantAxis, axis)
		})
	}
}
