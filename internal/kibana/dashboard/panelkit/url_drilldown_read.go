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

// URLDrilldownAPIItemData is a concrete intermediate type that callers populate
// from their panel-specific anonymous drilldown structs before calling
// ReadURLDrilldownsFromAPI. This avoids duplicating the null-preservation
// iteration logic across panel packages.
type URLDrilldownAPIItemData struct {
	URL          string
	Label        string
	EncodeUrl    *bool //nolint:revive
	OpenInNewTab *bool
}

// drilldownURLEncodeURLDefault and drilldownURLOpenInNewTabDefault are the Kibana
// server defaults used when null-preserving on import (prior == nil).
const (
	drilldownURLEncodeURLDefault    = true
	drilldownURLOpenInNewTabDefault = false
)

// ReadURLDrilldownsFromAPI builds a []models.URLDrilldownModel from a slice of
// URLDrilldownAPIItemData, applying null-preservation against the prior Terraform
// state. When prior is nil (import), DrilldownBoolImportPreserving is used so that
// fields matching Kibana server defaults are returned as null.
func ReadURLDrilldownsFromAPI(
	apiItems []URLDrilldownAPIItemData,
	prior []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	if len(apiItems) == 0 {
		return nil
	}

	out := make([]models.URLDrilldownModel, len(apiItems))
	for i, d := range apiItems {
		out[i].URL = types.StringValue(d.URL)
		out[i].Label = types.StringValue(d.Label)

		var p *models.URLDrilldownModel
		if i < len(prior) {
			p = &prior[i]
		}

		if p == nil {
			out[i].EncodeURL = DrilldownBoolImportPreserving(d.EncodeUrl, drilldownURLEncodeURLDefault)
			out[i].OpenInNewTab = DrilldownBoolImportPreserving(d.OpenInNewTab, drilldownURLOpenInNewTabDefault)
			continue
		}

		switch {
		case p.EncodeURL.IsNull():
			out[i].EncodeURL = types.BoolNull()
		case d.EncodeUrl != nil:
			out[i].EncodeURL = types.BoolValue(*d.EncodeUrl)
		default:
			out[i].EncodeURL = types.BoolNull()
		}

		switch {
		case p.OpenInNewTab.IsNull():
			out[i].OpenInNewTab = types.BoolNull()
		case d.OpenInNewTab != nil:
			out[i].OpenInNewTab = types.BoolValue(*d.OpenInNewTab)
		default:
			out[i].OpenInNewTab = types.BoolNull()
		}
	}
	return out
}

// ReadDiscoverSessionDrilldownsFromAPI is a thin wrapper around ReadURLDrilldownsFromAPI
// for the discoversession panel, which uses models.DiscoverSessionPanelDrilldown (identical
// fields to models.URLDrilldownModel but a distinct type).
func ReadDiscoverSessionDrilldownsFromAPI(
	apiItems []URLDrilldownAPIItemData,
	prior []models.DiscoverSessionPanelDrilldown,
) []models.DiscoverSessionPanelDrilldown {
	priorURL := make([]models.URLDrilldownModel, len(prior))
	for i, p := range prior {
		priorURL[i] = models.URLDrilldownModel(p)
	}
	result := ReadURLDrilldownsFromAPI(apiItems, priorURL)
	if result == nil {
		return nil
	}
	out := make([]models.DiscoverSessionPanelDrilldown, len(result))
	for i, r := range result {
		out[i] = models.DiscoverSessionPanelDrilldown(r)
	}
	return out
}
