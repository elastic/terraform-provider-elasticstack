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
	if chart, err := attrs.AsKibanaHTTPAPIsVisXyChartNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisXyChartESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisTreemapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisTreemapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisMosaicNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisMosaicESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisDatatableNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisDatatableESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisTagcloudNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisTagcloudESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisHeatmapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisHeatmapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisRegionMapNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisRegionMapESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisLegacyMetricNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisMetricNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisMetricESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisPieNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisPieESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisGaugeNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisGaugeESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisWaffleNoESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsVisWaffleESQLByValuePanel(); err == nil {
		return string(chart.Type)
	}
	return ""
}
