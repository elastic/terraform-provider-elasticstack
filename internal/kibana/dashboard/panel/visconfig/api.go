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
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const panelType = "vis"

// Handler implements iface.Handler for Kibana `vis` dashboard panels (`vis_config` / panel `config_json`).
type Handler struct{}

func (Handler) PanelType() string                 { return panelType }
func (Handler) SchemaAttribute() schema.Attribute { return SchemaAttribute() }

func (Handler) ClassifyJSON(map[string]any) bool { return false }

func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any { return config }

func (Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (Handler) AlignStateFromPlan(context.Context, *models.PanelModel, *models.PanelModel) {}

func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	cfgJSON := attrs["config_json"]
	visCfg := attrs["vis_config"]
	configJSONSet := panelkit.AttrConcreteSet(cfgJSON)
	configJSONUnk := panelkit.AttrUnknown(cfgJSON)
	visSet := panelkit.AttrConcreteSet(visCfg)
	visUnk := panelkit.AttrUnknown(visCfg)

	setCount := 0
	hasUnknown := configJSONUnk || visUnk
	if configJSONSet {
		setCount++
	}
	if visSet {
		setCount++
	}
	if setCount == 1 {
		return diags
	}
	if setCount == 0 && hasUnknown {
		return diags
	}

	detail := fmt.Sprintf("Panels with `type = \"vis\"` require exactly one of %s.", "`vis_config`, panel-level `config_json`")
	if setCount == 0 {
		diags.AddAttributeError(attrPath, "Missing vis panel configuration", detail)
		return diags
	}
	diags.AddAttributeError(attrPath, "Invalid vis panel configuration", detail)
	return diags
}

// FromAPI maps a kbapi vis panel into Terraform panel models (mirrors legacy dashboard.models_panels vis branch).
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	visPanel, err := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeVis()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var diags diag.Diagnostics
	dashboardModel := iface.EnclosingDashboard(ctx)
	pm.Grid = panelkit.GridFromAPI(visPanel.Grid.X, visPanel.Grid.Y, visPanel.Grid.W, visPanel.Grid.H)
	pm.ID = panelkit.IDFromAPI(visPanel.Id)

	configBytes, err := visPanel.Config.MarshalJSON()
	if err == nil {
		configJSON := customtypes.NewJSONWithDefaultsValue(string(configBytes), panelkit.PanelJSONDefaultsFunc())
		if prior != nil {
			configJSON = panelkit.PreservePriorPanelConfigJSON(ctx, prior.ConfigJSON, configJSON, &diags)
		}
		pm.ConfigJSON = configJSON
	}

	if priorPanelUsesConfigJSONOnly(prior) {
		return diags
	}

	var root map[string]any
	if len(configBytes) == 0 || json.Unmarshal(configBytes, &root) != nil {
		return diags
	}

	visPrior := configPriorForVisRead(prior, pm)

	switch classifyLensConfigFromRoot(root) {
	case lensConfigClassByReference:
		cfg1, err1 := visPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1()
		if err1 != nil {
			diags.AddError("Invalid visualization panel configuration on read", err1.Error())
			break
		}
		diags.Append(populateVisByReferenceFromAPI(ctx, visPrior, pm, cfg1)...)

	case lensConfigClassByValueChart:
		config0, err0 := visPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0()
		if err0 != nil {
			diags.AddError("Invalid visualization panel configuration on read", err0.Error())
			break
		}
		pm.VisConfig = &models.VisConfigModel{
			ByValue: &models.VisByValueModel{},
		}
		diags.Append(populateLensVisByValueFromTypedChartAPI(ctx, dashboardModel, prior, &pm.VisConfig.ByValue.LensByValueChartBlocks, config0, true)...)

	default:
		if visPrior != nil && visPrior.ByReference != nil {
			break
		}
		config0, err0 := visPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0()
		if err0 != nil {
			break
		}
		visType := lenscommon.DetectVizType(config0)
		if visType == "" {
			break
		}
		conv := lenscommon.ForType(visType)
		if conv == nil {
			diags.AddError(
				"Unsupported visualization chart type",
				fmt.Sprintf(
					"The dashboard returned Lens visualization discriminator %q which this provider does not support as typed `vis_config.by_value`. "+
						"Use panel-level `config_json` as the escape hatch to manage this panel until support is added.",
					visType,
				),
			)
			break
		}
		pm.VisConfig = &models.VisConfigModel{
			ByValue: &models.VisByValueModel{},
		}
		seedWaffleLensByValueChartFromPriorPanel(&pm.VisConfig.ByValue.LensByValueChartBlocks, prior)
		seedLensChartPriorIntoBlocks(prior, &pm.VisConfig.ByValue.LensByValueChartBlocks, visType)
		diags.Append(conv.PopulateFromAttributes(ctx, lensChartResolver(dashboardModel), &pm.VisConfig.ByValue.LensByValueChartBlocks, config0)...)
	}

	return diags
}

// ToAPI serializes Terraform vis panel state into kbapi (mirrors legacy visConfigToAPI / visByReferenceToAPI).
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	grid := panelkit.GridToAPI(pm.Grid)
	id := panelkit.IDToAPI(pm.ID)

	var diags diag.Diagnostics
	cfg := pm.VisConfig
	if cfg == nil {
		if typeutils.IsKnown(pm.ConfigJSON) && !pm.ConfigJSON.IsNull() {
			configJSON := []byte(pm.ConfigJSON.ValueString())
			var config kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config
			if err := config.UnmarshalJSON(configJSON); err != nil {
				diags.AddError("Failed to unmarshal visualization panel config", err.Error())
				return kbapi.DashboardPanelItem{}, diags
			}
			visPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis{
				Config: config,
				Grid:   grid,
				Id:     id,
				Type:   kbapi.Vis,
			}
			var panelItem kbapi.DashboardPanelItem
			if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(visPanel); err != nil {
				diags.AddError("Failed to create visualization panel", err.Error())
			}
			return panelItem, diags
		}
		diags.AddError("Missing `vis_config`", "The `vis_config` block is required for typed `vis` panels.")
		return kbapi.DashboardPanelItem{}, diags
	}
	switch {
	case cfg.ByReference != nil:
		return visByReferenceToAPI(*cfg.ByReference, grid, id)
	case cfg.ByValue != nil:
		blocks := &cfg.ByValue.LensByValueChartBlocks
		conv, okConv := lenscommon.FirstForBlocks(blocks)
		if !okConv {
			diags.AddError("Invalid `vis_config.by_value`", "The typed chart block could not be resolved to a Lens visualization converter.")
			return kbapi.DashboardPanelItem{}, diags
		}
		config0, d := conv.BuildAttributes(blocks, lensChartResolver(dashboard))
		diags.Append(d...)
		if d.HasError() {
			return kbapi.DashboardPanelItem{}, diags
		}
		var config kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config
		if err := config.FromKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig0(config0); err != nil {
			diags.AddError("Failed to create visualization panel config", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
		visPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis{
			Config: config,
			Grid:   grid,
			Id:     id,
			Type:   kbapi.Vis,
		}
		var panelItem kbapi.DashboardPanelItem
		if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(visPanel); err != nil {
			diags.AddError("Failed to create visualization panel", err.Error())
		}
		return panelItem, diags
	default:
		diags.AddError("Invalid `vis_config`", "Exactly one of `by_value` or `by_reference` must be set inside `vis_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}
}

func populateVisByReferenceFromAPI(
	ctx context.Context,
	prior *models.VisConfigModel,
	pm *models.PanelModel,
	cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1,
) diag.Diagnostics {
	var priorBR *models.VisByReferenceModel
	if prior != nil {
		priorBR = prior.ByReference
	}

	by, diags := lenscommon.PopulateVisByReferenceTFModelFromAPIConfig1(ctx, cfg1, priorBR)
	if diags.HasError() {
		return diags
	}

	brCopy := by
	pm.VisConfig = &models.VisConfigModel{
		ByReference: &brCopy,
	}
	return diags
}

func visByReferenceToAPI(
	byRef models.VisByReferenceModel,
	grid struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	},
	panelID *string,
) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	api1, d := lenscommon.VisByReferenceModelToAPIConfig1(byRef, "vis_config.by_reference.references_json")
	diags.Append(d...)
	if d.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}
	var config kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis_Config
	if err := config.FromKibanaHTTPAPIsKbnDashboardPanelTypeVisConfig1(api1); err != nil {
		diags.AddError("Failed to set vis by_reference config", err.Error())
		return kbapi.DashboardPanelItem{}, diags
	}
	visPanel := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeVis{
		Config: config,
		Grid:   grid,
		Id:     panelID,
		Type:   kbapi.Vis,
	}
	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKibanaHTTPAPIsKbnDashboardPanelTypeVis(visPanel); err != nil {
		diags.AddError("Failed to create visualization panel", err.Error())
	}
	return panelItem, diags
}
