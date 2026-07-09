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

package panelkit

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// URLDrilldownWire is the canonical JSON shape for a URL drilldown entry sent to the Kibana API.
// Trigger and Type are always the same string constants across all panel types.
type URLDrilldownWire struct {
	EncodeURL    *bool  `json:"encode_url,omitempty"`
	Label        string `json:"label"`
	OpenInNewTab *bool  `json:"open_in_new_tab,omitempty"`
	Trigger      string `json:"trigger"`
	Type         string `json:"type"`
	URL          string `json:"url"`
}

// BuildURLDrilldownsWire converts Terraform URLDrilldownModel entries into a wire slice ready
// for JSON-marshaling to the Kibana API.
func BuildURLDrilldownsWire(drilldowns []models.URLDrilldownModel) []URLDrilldownWire {
	result := make([]URLDrilldownWire, len(drilldowns))
	for i, dd := range drilldowns {
		result[i] = URLDrilldownWire{
			URL:     dd.URL.ValueString(),
			Label:   dd.Label.ValueString(),
			Trigger: "on_open_panel_menu",
			Type:    attrURLDrilldown,
		}
		if typeutils.IsKnown(dd.EncodeURL) {
			result[i].EncodeURL = dd.EncodeURL.ValueBoolPointer()
		}
		if typeutils.IsKnown(dd.OpenInNewTab) {
			result[i].OpenInNewTab = dd.OpenInNewTab.ValueBoolPointer()
		}
	}
	return result
}

// InjectDrilldownsJSON marshals drilldowns as a JSON array and injects them into the api struct
// via a round-trip marshal/unmarshal so that panel-specific kbapi enum types are not required.
// api must be a non-nil pointer to a struct that has a JSON "drilldowns" field.
func InjectDrilldownsJSON(api any, drilldowns []models.URLDrilldownModel) diag.Diagnostics {
	ddsJSON, err := json.Marshal(BuildURLDrilldownsWire(drilldowns))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal drilldowns", err.Error())}
	}
	base, err := json.Marshal(api)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to marshal panel config", err.Error())}
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal panel config", err.Error())}
	}
	m["drilldowns"] = ddsJSON
	merged, err := json.Marshal(m)
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to re-marshal panel config", err.Error())}
	}
	if err := json.Unmarshal(merged, api); err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to apply drilldowns to panel config", err.Error())}
	}
	return nil
}

// DrilldownBoolImportPreserving maps an API *bool to a Terraform Bool, returning null when the
// API omits the field or when the value equals the server default (so practitioners can omit it).
func DrilldownBoolImportPreserving(api *bool, serverDefault bool) types.Bool {
	if api == nil {
		return types.BoolNull()
	}
	if *api == serverDefault {
		return types.BoolNull()
	}
	return types.BoolValue(*api)
}

// ReadDrilldownsFromWireJSON unmarshals a JSON-encoded array of drilldown wire objects and maps
// them into URLDrilldownModel entries, null-preserving encode_url and open_in_new_tab against
// priorDrilldowns. If priorDrilldowns is nil (e.g. on import), API values are used directly.
// Returns nil when b cannot be unmarshaled or the resulting slice is empty.
func ReadDrilldownsFromWireJSON(b []byte, priorDrilldowns []models.URLDrilldownModel) []models.URLDrilldownModel {
	var wire []URLDrilldownWire
	if err := json.Unmarshal(b, &wire); err != nil || len(wire) == 0 {
		return nil
	}
	result := make([]models.URLDrilldownModel, len(wire))
	for i, dd := range wire {
		result[i] = models.URLDrilldownModel{
			URL:   types.StringValue(dd.URL),
			Label: types.StringValue(dd.Label),
		}

		var prior *models.URLDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		switch {
		case prior != nil && prior.EncodeURL.IsNull():
			result[i].EncodeURL = types.BoolNull()
		case dd.EncodeURL != nil:
			result[i].EncodeURL = types.BoolValue(*dd.EncodeURL)
		default:
			result[i].EncodeURL = types.BoolNull()
		}

		switch {
		case prior != nil && prior.OpenInNewTab.IsNull():
			result[i].OpenInNewTab = types.BoolNull()
		case dd.OpenInNewTab != nil:
			result[i].OpenInNewTab = types.BoolValue(*dd.OpenInNewTab)
		default:
			result[i].OpenInNewTab = types.BoolNull()
		}
	}
	return result
}
