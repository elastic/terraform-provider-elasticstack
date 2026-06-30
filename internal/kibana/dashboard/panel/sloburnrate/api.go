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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "slo_burn_rate"

const panelConfigAttrsKeyPrefix = panelType + "_config"

// Handler implements iface.Handler for the slo_burn_rate dashboard panel discriminator.
type Handler struct {
	panelkit.NoopHandlerBase
}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

// FromAPI fills pm from kbapi DashboardPanelItem for this panel discriminator.
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm *models.PanelModel, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate) diag.Diagnostics {
			return PopulateFromAPI(pm, prior, p.Config)
		},
	)
}

// ToAPI serializes Terraform panel state into a kbapi union item.
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
		var diags diag.Diagnostics
		diags.AddError(
			"Unsupported panel type for config_json",
			"Panel-level `config_json` is not supported for `slo_burn_rate` panels. Use `slo_burn_rate_config` instead.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}
	if pm.SloBurnRateConfig == nil {
		var diags diag.Diagnostics
		diags.AddError("Missing SLO burn rate panel configuration", "SLO burn rate panels require `slo_burn_rate_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate{
		Grid: grid,
		Id:   id,
		Type: kbapi.SloBurnRate,
	}
	diags := BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeSloBurnRate(panel); err != nil {
		diags.AddError("Failed to create SLO burn rate panel", err.Error())
	}
	return panelItem, diags
}

// ValidatePanelConfig returns diagnostics only for this panel type when required fields inside
// slo_burn_rate_config are absent and the validated attribute map targets that nested object.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, shaped := panelkit.ResolvePanelAttrsShape(attrs, panelConfigAttrsKeyPrefix, "slo_id", "duration")
	if !shaped {
		out.AddAttributeError(attrPath.AtName(panelConfigAttrsKeyPrefix), "Missing SLO burn rate panel configuration", "SLO burn rate panels require `slo_burn_rate_config`.")
		return out
	}

	cfgPath := attrPath
	if !flat {
		cfgPath = attrPath.AtName(panelConfigAttrsKeyPrefix)
		nestedRaw := attrs[panelConfigAttrsKeyPrefix]
		if nestedRaw != nil {
			switch {
			case nestedRaw.IsUnknown():
				return out
			case nestedRaw.IsNull():
				out.AddAttributeError(cfgPath, "Missing SLO burn rate panel configuration", "SLO burn rate panels require `slo_burn_rate_config`.")
				return out
			}
		}
	}

	var sloVal, durVal attr.Value

	switch {
	case flat:
		sloVal, durVal = attrs["slo_id"], attrs["duration"]
	default:
		objAttrs := obj.Attributes()
		sloVal = objAttrs["slo_id"]
		durVal = objAttrs["duration"]
	}

	deferSLO, missSLO := panelkit.StringAttrDeferOrMissing(sloVal)
	if !deferSLO && missSLO {
		out.AddAttributeError(cfgPath.AtName("slo_id"), `Invalid SLO burn rate configuration`, "`slo_id` is required.")
	}

	deferDur, missDur := panelkit.StringAttrDeferOrMissing(durVal)
	switch {
	case deferDur:
	case missDur:
		out.AddAttributeError(cfgPath.AtName("duration"), `Invalid SLO burn rate configuration`, "`duration` is required.")
	default:
		durStr := durVal.(types.String)
		if !sloBurnRateDurationRegexp.MatchString(durStr.ValueString()) {
			out.AddAttributeError(
				cfgPath.AtName("duration"),
				`Invalid SLO burn rate configuration`,
				"`duration` must match the pattern `^\\d+[mhd]$` (a positive integer followed by m, h, or d).",
			)
		}
	}

	return out
}
