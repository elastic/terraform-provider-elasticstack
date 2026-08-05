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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lensAPIConfigFromJSON(t *testing.T, raw string) kbapi.KibanaHTTPAPIsLensApiConfig {
	t.Helper()

	var config kbapi.KibanaHTTPAPIsLensApiConfig
	require.NoError(t, json.Unmarshal([]byte(raw), &config))
	return config
}

func TestHasLensByReferenceShapeAtRoot_refIDOnly(t *testing.T) {
	t.Parallel()
	assert.True(t, HasLensByReferenceShapeAtRoot(map[string]any{"ref_id": "panel_0"}))
	assert.False(t, HasLensByReferenceShapeAtRoot(map[string]any{"ref_id": ""}))
	assert.False(t, HasLensByReferenceShapeAtRoot(map[string]any{"time_range": map[string]any{"from": "now-7d", "to": "now"}}))
}

func TestDetectVizType_chartKindsPerArm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty_union", raw: `{}`, want: ""},
		{name: "xy", raw: `{"type":"xy"}`, want: "xy"},
		{name: "treemap", raw: `{"type":"treemap"}`, want: "treemap"},
		{name: "mosaic", raw: `{"type":"mosaic"}`, want: "mosaic"},
		{name: "datatable", raw: `{"type":"data_table"}`, want: "data_table"},
		{name: "tagcloud", raw: `{"type":"tag_cloud"}`, want: "tag_cloud"},
		{name: "heatmap", raw: `{"type":"heatmap"}`, want: "heatmap"},
		{name: "region_map", raw: `{"type":"region_map"}`, want: "region_map"},
		{name: "legacy_metric", raw: `{"type":"legacy_metric"}`, want: "legacy_metric"},
		{name: "metric", raw: `{"type":"metric"}`, want: "metric"},
		{name: "pie", raw: `{"type":"pie"}`, want: "pie"},
		{name: "gauge", raw: `{"type":"gauge"}`, want: "gauge"},
		{name: "waffle", raw: `{"type":"waffle"}`, want: "waffle"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, DetectVizType(lensAPIConfigFromJSON(t, tc.raw)))
		})
	}
}
