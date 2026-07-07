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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

// URLDrilldownAPIEntry is the minimal set of fields ReadURLDrilldownsFromAPI needs from any
// panel-specific anonymous drilldown struct. Panel packages map their kbapi types to this.
type URLDrilldownAPIEntry struct {
	URL          string
	Label        string
	EncodeURL    *bool
	OpenInNewTab *bool
}

// ReadURLDrilldownsFromAPI maps a normalized slice of API drilldown entries into
// []models.URLDrilldownModel with null-intent preservation from prior state.
// encodeURLDefault and openInNewTabDefault govern import-mode (prior==nil) behaviour via
// DrilldownBoolImportPreserving: when the API value equals the default, null is stored so
// practitioners can omit the field in their config.
func ReadURLDrilldownsFromAPI(
	entries []URLDrilldownAPIEntry,
	priorDrilldowns []models.URLDrilldownModel,
	encodeURLDefault, openInNewTabDefault bool,
) []models.URLDrilldownModel {
	if len(entries) == 0 {
		return nil
	}

	result := make([]models.URLDrilldownModel, len(entries))
	for i, d := range entries {
		result[i] = models.URLDrilldownModel{
			URL:   types.StringValue(d.URL),
			Label: types.StringValue(d.Label),
		}

		var prior *models.URLDrilldownModel
		if i < len(priorDrilldowns) {
			prior = &priorDrilldowns[i]
		}

		if prior == nil {
			result[i].EncodeURL = DrilldownBoolImportPreserving(d.EncodeURL, encodeURLDefault)
			result[i].OpenInNewTab = DrilldownBoolImportPreserving(d.OpenInNewTab, openInNewTabDefault)
			continue
		}

		switch {
		case prior.EncodeURL.IsNull():
			result[i].EncodeURL = types.BoolNull()
		case d.EncodeURL != nil:
			result[i].EncodeURL = types.BoolValue(*d.EncodeURL)
		default:
			result[i].EncodeURL = types.BoolNull()
		}

		switch {
		case prior.OpenInNewTab.IsNull():
			result[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			result[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			result[i].OpenInNewTab = types.BoolNull()
		}
	}

	return result
}
