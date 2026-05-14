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

package dashboard

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func fillUnknownDashboardPanelFromAPI(ctx context.Context, tfPanel *models.PanelModel, pm *models.PanelModel, panelItem kbapi.DashboardPanelItem) {
	pm.ID = types.StringNull()
	pm.ConfigJSON = customtypes.NewJSONWithDefaultsNull(populatePanelConfigJSONDefaults)
	pm.Grid = models.PanelGridModel{}

	rawBytes, err := panelItem.MarshalJSON()
	if err != nil {
		return
	}
	var rawObj map[string]any
	if json.Unmarshal(rawBytes, &rawObj) != nil {
		return
	}
	if grid, ok := rawObj["grid"].(map[string]any); ok {
		x, _ := grid["x"].(float64)
		y, _ := grid["y"].(float64)
		var wPtr, hPtr *float32
		if wVal, ok := grid["w"].(float64); ok {
			wPtr = typeutils.Float32Ptr(wVal)
		}
		if hVal, ok := grid["h"].(float64); ok {
			hPtr = typeutils.Float32Ptr(hVal)
		}
		pm.Grid = panelkit.GridFromAPI(float32(x), float32(y), wPtr, hPtr)
	}
	if id, ok := rawObj["id"].(string); ok && id != "" {
		pm.ID = types.StringValue(id)
	}
	if config, ok := rawObj["config"]; ok {
		configBytes, mErr := json.Marshal(config)
		if mErr == nil {
			configJSON := customtypes.NewJSONWithDefaultsValue(string(configBytes), populatePanelConfigJSONDefaults)
			if tfPanel != nil {
				var wrap diag.Diagnostics
				configJSON = panelkit.PreservePriorJSONWithDefaultsIfEquivalent(ctx, tfPanel.ConfigJSON, configJSON, &wrap)
			}
			pm.ConfigJSON = configJSON
		}
	}
	clearPanelConfigBlocks(pm)
}
