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

// DetectVizType returns the Kibana Lens chart discriminator string from vis_config.by_value
// union payload attrs (same strings as VizConverter.VizType / kbapi chart Type fields).
// Empty string means the union could not be decoded to a known handled chart variant.
//
// Implementation mirrors the former dashboard.detectLensVisType loop over kbapi.As*
// helpers so lens packages stay free of dashboard imports.
func DetectVizType(attrs VisByValueConfig0) string {
	if chart, err := attrs.AsKibanaHTTPAPIsXyChartNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsXyChartESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTreemapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTreemapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMosaicNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMosaicESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsDatatableNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsDatatableESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTagcloudNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTagcloudESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsHeatmapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsHeatmapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsRegionMapNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsRegionMapESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsLegacyMetricNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMetricNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMetricESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsPieNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsPieESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsGaugeNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsGaugeESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsWaffleNoESQL(); err == nil {
		return string(chart.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsWaffleESQL(); err == nil {
		return string(chart.Type)
	}
	return ""
}
