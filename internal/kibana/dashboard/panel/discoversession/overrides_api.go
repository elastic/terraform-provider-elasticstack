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

package discoversession

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

// discoverSessionOverridesAPI mirrors the anonymous overrides struct generated for the
// by-reference discover session config.
type discoverSessionOverridesAPI = struct {
	ColumnOrder    *[]string `json:"column_order,omitempty"`
	ColumnSettings *map[string]struct {
		Width *float32 `json:"width,omitempty"`
	} `json:"column_settings,omitempty"`
	DefaultRenderedNodes *float32                                                                                      `json:"default_rendered_nodes,omitempty"`
	Density              *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesDensity              `json:"density,omitempty"`
	DocumentsDisplayMode *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesDocumentsDisplayMode `json:"documents_display_mode,omitempty"`
	HeaderRowHeight      *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_HeaderRowHeight  `json:"header_row_height,omitempty"`
	HideNulls            *bool                                                                                         `json:"hide_nulls,omitempty"`
	RowHeight            *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSession_Config_1_Overrides_RowHeight        `json:"row_height,omitempty"`
	RowsPerPage          *float32                                                                                      `json:"rows_per_page,omitempty"`
	SampleSize           *float32                                                                                      `json:"sample_size,omitempty"`
	Sort                 *[]struct {
		Direction kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeDiscoverSessionConfig1OverridesSortDirection `json:"direction"`
		Name      string                                                                                `json:"name"`
	} `json:"sort,omitempty"`
	WrapLines *bool `json:"wrap_lines,omitempty"`
}
