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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/require"
)

func TestTerraformChartBlockKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		vizType string
		want    string
	}{
		"xy":            {string(kbapi.KibanaHTTPAPIsXyChartNoESQLTypeXy), "xy_chart_config"},
		"datatable":     {string(kbapi.KibanaHTTPAPIsDatatableNoESQLTypeDataTable), "datatable_config"},
		"tagcloud":      {string(kbapi.KibanaHTTPAPIsTagcloudNoESQLTypeTagCloud), "tagcloud_config"},
		"region map":    {string(kbapi.KibanaHTTPAPIsRegionMapNoESQLTypeRegionMap), "region_map_config"},
		"pie":           {string(kbapi.KibanaHTTPAPIsPieNoESQLTypePie), "pie_chart_config"},
		"metric":        {string(kbapi.KibanaHTTPAPIsMetricNoESQLTypeMetric), "metric_chart_config"},
		"legacy metric": {string(kbapi.LegacyMetric), "legacy_metric_config"},
		"gauge":         {string(kbapi.KibanaHTTPAPIsGaugeNoESQLTypeGauge), "gauge_config"},
		"heatmap":       {string(kbapi.KibanaHTTPAPIsHeatmapNoESQLTypeHeatmap), "heatmap_config"},
		"mosaic":        {string(kbapi.KibanaHTTPAPIsMosaicNoESQLTypeMosaic), "mosaic_config"},
		"treemap":       {string(kbapi.KibanaHTTPAPIsTreemapNoESQLTypeTreemap), "treemap_config"},
		"waffle":        {string(kbapi.KibanaHTTPAPIsWaffleNoESQLTypeWaffle), "waffle_config"},
		"unknown":       {"unknown", ""},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, testCase.want, TerraformChartBlockKey(testCase.vizType))
		})
	}
}
