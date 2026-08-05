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

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

// IsNoESQLCandidateActuallyESQL returns true when a panel decoded as NoESQL actually
// carries an ES|QL or table data source. All NoESQLByValuePanel DataSource fields
// implement json.Marshaler, so a single interface covers every panel type.
func IsNoESQLCandidateActuallyESQL(dataSource interface{ MarshalJSON() ([]byte, error) }) bool {
	return LensDataSourceIsESQLOrTable(dataSource.MarshalJSON())
}

// DetectVizType returns the Kibana Lens chart discriminator string from a LensApiConfig
// payload (same strings as VizConverter.VizType / kbapi chart Type fields).
// Empty string means the union could not be decoded to a known handled chart variant.
func DetectVizType(attrs kbapi.KibanaHTTPAPIsLensApiConfig) string {
	if chart, err := attrs.AsKibanaHTTPAPIsXyChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsXyChartNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsXyChartESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTreemapChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsTreemapNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsTreemapESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMosaicChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsMosaicNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsMosaicESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsDatatableChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsDatatableNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsDatatableESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsTagcloudChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsTagcloudNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsTagcloudESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsHeatmapChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsHeatmapNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsHeatmapESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsRegionMapChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsRegionMapNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsRegionMapESQL(); err == nil {
			return string(value.Type)
		}
	}
	if value, err := attrs.AsKibanaHTTPAPIsLegacyMetricNoESQL(); err == nil {
		return string(value.Type)
	}
	if chart, err := attrs.AsKibanaHTTPAPIsMetricChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsMetricNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsMetricESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsPieChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsPieNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsPieESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsGaugeChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsGaugeNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsGaugeESQL(); err == nil {
			return string(value.Type)
		}
	}
	if chart, err := attrs.AsKibanaHTTPAPIsWaffleChart(); err == nil {
		if value, err := chart.AsKibanaHTTPAPIsWaffleNoESQL(); err == nil {
			return string(value.Type)
		}
		if value, err := chart.AsKibanaHTTPAPIsWaffleESQL(); err == nil {
			return string(value.Type)
		}
	}
	return ""
}
