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

// IsNoESQLCandidateActuallyESQL returns true when a panel decoded as NoESQL actually
// carries an ES|QL or table data source. All NoESQLByValuePanel DataSource fields
// implement json.Marshaler, so a single interface covers every panel type.
func IsNoESQLCandidateActuallyESQL(dataSource interface{ MarshalJSON() ([]byte, error) }) bool {
	return LensDataSourceIsESQLOrTable(dataSource.MarshalJSON())
}

// DetectVizType returns the Kibana Lens chart discriminator string from vis_config.by_value
// union payload attrs (same strings as VizConverter.VizType / kbapi chart Type fields).
// Empty string means the union could not be decoded to a known handled chart variant.
//
// Implementation mirrors the former dashboard.detectLensVisType loop over kbapi.As*
// helpers so lens packages stay free of dashboard imports.
func DetectVizType(attrs VisByValueConfig0) string {
	if chart, err := attrs.AsKibanaHTTPAPIsXyChartNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsXyChartESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTreemapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTreemapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMosaicNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMosaicESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsDatatableNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsDatatableESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTagcloudNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTagcloudESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsHeatmapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsHeatmapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsRegionMapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsRegionMapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsLegacyMetricNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMetricNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMetricESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsPieNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsPieESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsGaugeNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsGaugeESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsWaffleNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsWaffleESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	return ""
}
