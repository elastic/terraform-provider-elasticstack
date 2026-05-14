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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "slo_burn_rate"

const panelConfigAttrsKeyPrefix = panelType + "_config"

// Handler implements iface.Handler for the slo_burn_rate dashboard panel discriminator.
type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}
func (Handler) PinnedHandler() iface.PinnedHandler { return nil }
func (Handler) AlignStateFromPlan(ctx context.Context, plan, state *models.PanelModel) {
	_, _, _ = ctx, plan, state
}

// FromAPI fills pm from kbapi DashboardPanelItem for this panel discriminator.
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	apiPanel, err := item.AsKbnDashboardPanelTypeSloBurnRate()
	if err != nil {
		var d diag.Diagnostics
		d.AddError("Dashboard panel decode", err.Error())
		return d
	}

	pm.Grid = panelkit.GridFromAPI(apiPanel.Grid.X, apiPanel.Grid.Y, apiPanel.Grid.W, apiPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(apiPanel.Id)
	pm.ConfigJSON = panelConfigJSONNull()
	diags := PopulateFromAPI(pm, prior, apiPanel.Config)
	_ = ctx
	return diags
}

// ToAPI serializes Terraform panel state into a kbapi union item.
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)
	panel := kbapi.KbnDashboardPanelTypeSloBurnRate{
		Grid: grid,
		Id:   id,
		Type: kbapi.SloBurnRate,
	}
	diags := BuildConfig(pm, &panel)
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeSloBurnRate(panel); err != nil {
		diags.AddError("Failed to create SLO burn rate panel", err.Error())
	}
	return panelItem, diags
}

// sloBurnRateAttrsShape detects the attrs map convention used by contracttest.ValidatePanelConfig
// (flattened slo_id + duration keys) versus a full-panel attribute map keyed by slo_burn_rate_config.
func sloBurnRateAttrsShape(attrs map[string]attr.Value) (flat bool, obj types.Object, ok bool) {
	if attrs == nil {
		return false, types.Object{}, false
	}

	if _, slo := attrs["slo_id"]; slo {
		if _, dur := attrs["duration"]; dur {
			return true, types.Object{}, true
		}
		return false, types.Object{}, false
	}

	if raw, nested := attrs[panelConfigAttrsKeyPrefix]; nested {
		if obj, ok := raw.(types.Object); ok {
			return false, obj, ok
		}
		return false, types.Object{}, false
	}

	return false, types.Object{}, false
}

// ValidatePanelConfig returns diagnostics only for this panel type when required fields inside
// slo_burn_rate_config are absent and the validated attribute map targets that nested object.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, shaped := sloBurnRateAttrsShape(attrs)
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
