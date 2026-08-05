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

package visconfig

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// lensConfigJSON separates a by-value Lens config into the two generated API
// models that describe the same JSON object: the chart payload and dashboard
// presentation fields.
func lensConfigJSON(raw []byte) (
	kbapi.KibanaHTTPAPIsLensApiConfig,
	kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
	error,
) {
	var chart kbapi.KibanaHTTPAPIsLensApiConfig
	if err := json.Unmarshal(raw, &chart); err != nil {
		return kbapi.KibanaHTTPAPIsLensApiConfig{}, kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0{}, err
	}

	var presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0
	if err := json.Unmarshal(raw, &presentation); err != nil {
		return kbapi.KibanaHTTPAPIsLensApiConfig{}, kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0{}, err
	}

	return chart, presentation, nil
}

// composeLensConfigJSON merges a typed Lens chart with presentation fields into
// the JSON accepted by the dashboard vis config union. The chart wins for any
// overlapping fields because it is the source of the Lens chart discriminator.
func composeLensConfigJSON(
	chart kbapi.KibanaHTTPAPIsLensApiConfig,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
) ([]byte, error) {
	chartJSON, err := json.Marshal(chart)
	if err != nil {
		return nil, fmt.Errorf("marshal Lens chart config: %w", err)
	}
	presentationJSON, err := json.Marshal(presentation)
	if err != nil {
		return nil, fmt.Errorf("marshal Lens presentation config: %w", err)
	}

	var composed map[string]json.RawMessage
	if err := json.Unmarshal(presentationJSON, &composed); err != nil {
		return nil, fmt.Errorf("decode Lens presentation config: %w", err)
	}
	var chartFields map[string]json.RawMessage
	if err := json.Unmarshal(chartJSON, &chartFields); err != nil {
		return nil, fmt.Errorf("decode Lens chart config: %w", err)
	}
	maps.Copy(composed, chartFields)

	raw, err := json.Marshal(composed)
	if err != nil {
		return nil, fmt.Errorf("marshal composed Lens config: %w", err)
	}
	return raw, nil
}

func composeLensVisConfig(
	chart kbapi.KibanaHTTPAPIsLensApiConfig,
	presentation kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0,
) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config, error) {
	raw, err := composeLensConfigJSON(chart, presentation)
	if err != nil {
		return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config{}, err
	}

	var config kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config{}, err
	}
	return config, nil
}
