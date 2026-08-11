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

package sloburnrate

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BuildConfig writes Terraform state from pm into panel's typed API config.
func BuildConfig(pm models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate) diag.Diagnostics {
	cfg := pm.SloBurnRateConfig
	if cfg == nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Missing SLO burn rate panel configuration",
			"SLO burn rate panels require `slo_burn_rate_config`.",
		)
		return diags
	}

	embeddable := kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable{
		SloId:    cfg.SloID.ValueString(),
		Duration: cfg.Duration.ValueString(),
	}

	if typeutils.IsKnown(cfg.SloInstanceID) {
		embeddable.SloInstanceId = cfg.SloInstanceID.ValueStringPointer()
	}
	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&embeddable.Title, &embeddable.Description, &embeddable.HideTitle, &embeddable.HideBorder)

	var diags diag.Diagnostics
	if len(cfg.Drilldowns) > 0 {
		diags.Append(panelkit.InjectDrilldownsJSON(&embeddable, cfg.Drilldowns)...)
	}

	panel.Config = embeddable
	return diags
}

// PopulateFromAPI maps Kibana SLO burn rate embeddable config into Terraform panel state while
// preserving prior null intent (REQ-009). prior is the prior TF state/plan panel, or nil on import.
//
// pm always arrives with SloBurnRateConfig unset (callers build state from a zero-valued
// PanelModel to avoid aliasing plan pointers), so that field cannot be used to detect whether this
// panel was previously this same type. prior.SloBurnRateConfig is the only reliable signal:
// non-nil means the panel was already this type and its null intent must be honored; nil means
// there is no prior null intent for this config block (creation, import, or a type change).
func PopulateFromAPI(pm *models.PanelModel, prior *models.PanelModel, apiConfig kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable) diag.Diagnostics {
	if prior == nil || prior.SloBurnRateConfig == nil {
		pm.SloBurnRateConfig = sloBurnRateConfigFromAPIImport(apiConfig)
		return nil
	}

	// Same-type update: rebuild from the API, then reapply the prior config's null intent for any
	// optional field the plan/state had not set (REQ-009 null-preservation).
	existing := sloBurnRateConfigFromAPIImport(apiConfig)
	existing.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, prior.SloBurnRateConfig.Drilldowns)
	// sloBurnRateConfigFromAPIImport computes slo_instance_id assuming there is no prior value to
	// consult (creation/import). On an update there is a prior value, so recompute it using that
	// knowledge instead: this is the only way a practitioner who explicitly set slo_instance_id =
	// "*" sees it round-trip as "*" rather than be silently nulled.
	existing.SloInstanceID = panelkit.PreserveSloInstanceID(apiConfig.SloInstanceId, true, prior.SloBurnRateConfig.SloInstanceID)
	sloBurnRatePreserveNullIntentFromPrior(prior.SloBurnRateConfig, existing)
	pm.SloBurnRateConfig = existing
	return nil
}

func sloBurnRateConfigFromAPIImport(apiConfig kbapi.KibanaHTTPAPIsSloBurnRateEmbeddable) *models.SloBurnRateConfigModel {
	cfg := &models.SloBurnRateConfigModel{
		SloID:    types.StringValue(apiConfig.SloId),
		Duration: types.StringValue(apiConfig.Duration),
	}
	cfg.SloInstanceID = panelkit.PreserveSloInstanceID(apiConfig.SloInstanceId, false, types.StringNull())
	panelkit.CopyPresentationFromAPI(&cfg.Title, &cfg.Description, &cfg.HideTitle, &cfg.HideBorder,
		apiConfig.Title, apiConfig.Description, apiConfig.HideTitle, apiConfig.HideBorder)
	cfg.Drilldowns = readSloBurnRateDrilldownsFromAPI(apiConfig.Drilldowns, nil)
	return cfg
}

func sloBurnRatePreserveNullIntentFromPrior(prior, existing *models.SloBurnRateConfigModel) {
	if prior == nil || existing == nil {
		return
	}
	// SloInstanceID's null intent is already applied by panelkit.PreserveSloInstanceID above.
	panelkit.NullPreservePresentationFromPrior(prior.Title, prior.Description, prior.HideTitle, prior.HideBorder,
		&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder)
	if len(prior.Drilldowns) == 0 {
		existing.Drilldowns = nil
	}
}

type sloBurnRateAPIDrilldown = struct {
	EncodeUrl    *bool                                                      `json:"encode_url,omitempty"` //nolint:revive
	Label        string                                                     `json:"label"`
	OpenInNewTab *bool                                                      `json:"open_in_new_tab,omitempty"`
	Trigger      kbapi.KibanaHTTPAPIsSloBurnRateEmbeddableDrilldownsTrigger `json:"trigger"`
	Type         kbapi.KibanaHTTPAPIsSloBurnRateEmbeddableDrilldownsType    `json:"type"`
	Url          string                                                     `json:"url"` //nolint:revive
}

func readSloBurnRateDrilldownsFromAPI(
	apiDrilldowns *[]sloBurnRateAPIDrilldown,
	priorDrilldowns []models.URLDrilldownModel,
) []models.URLDrilldownModel {
	items := panelkit.BuildURLDrilldownItems(apiDrilldowns, func(d sloBurnRateAPIDrilldown) panelkit.URLDrilldownAPIItemData {
		return panelkit.URLDrilldownAPIItemData{
			URL:          d.Url,
			Label:        d.Label,
			EncodeUrl:    d.EncodeUrl,
			OpenInNewTab: d.OpenInNewTab,
		}
	})
	return panelkit.ReadURLDrilldownsFromAPI(items, priorDrilldowns)
}
