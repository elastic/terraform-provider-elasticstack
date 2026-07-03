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

package image

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	drilldownDashboardBoolDefault   = false
	drilldownURLEncodeURLDefault    = true
	drilldownURLOpenInNewTabDefault = false
)

// BuildConfig writes Terraform image panel state into the API panel's config (Grid/Id set separately).
func BuildConfig(pm *models.PanelModel, panel *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage, diags *diag.Diagnostics) {
	cfg := pm.ImageConfig
	if cfg == nil {
		diags.AddError("Missing image panel configuration", "Image panels require `image_config`.")
		return
	}

	panelkit.BuildPresentationConfig(cfg.Title, cfg.Description, cfg.HideTitle, cfg.HideBorder,
		&panel.Config.Title, &panel.Config.Description, &panel.Config.HideTitle, &panel.Config.HideBorder)

	img := &panel.Config.ImageConfig
	if typeutils.IsKnown(cfg.AltText) {
		img.AltText = cfg.AltText.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.BackgroundColor) {
		img.BackgroundColor = cfg.BackgroundColor.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.ObjectFit) {
		fit := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigObjectFit(cfg.ObjectFit.ValueString())
		img.ObjectFit = &fit
	}

	switch {
	case cfg.Src.File != nil:
		src0 := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc0{
			Type:   kbapi.File,
			FileId: cfg.Src.File.FileID.ValueString(),
		}
		if err := img.Src.FromKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc0(src0); err != nil {
			diags.AddError("Invalid image src", err.Error())
			return
		}
	case cfg.Src.URL != nil:
		src1 := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc1{
			Type: kbapi.Url,
			Url:  cfg.Src.URL.URL.ValueString(),
		}
		if err := img.Src.FromKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc1(src1); err != nil {
			diags.AddError("Invalid image src", err.Error())
			return
		}
	default:
		diags.AddError("Invalid image src", "Exactly one of `file` or `url` must be set inside `src`.")
		return
	}

	if len(cfg.Drilldowns) > 0 {
		items := make([]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_Drilldowns_Item, 0, len(cfg.Drilldowns))
		for _, d := range cfg.Drilldowns {
			item, dDiags := drilldownItemToAPI(d)
			diags.Append(dDiags...)
			if dDiags.HasError() {
				return
			}
			items = append(items, item)
		}
		panel.Config.Drilldowns = &items
	}
}

func drilldownItemToAPI(d models.ImagePanelDrilldownModel) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_Drilldowns_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	var item kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_Drilldowns_Item

	switch {
	case d.DashboardDrilldown != nil:
		dd := d.DashboardDrilldown
		wire := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0{
			DashboardId: dd.DashboardID.ValueString(),
			Label:       dd.Label.ValueString(),
			Trigger:     kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0Trigger(dd.Trigger.ValueString()),
			Type:        kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown,
		}
		if typeutils.IsKnown(dd.UseFilters) {
			wire.UseFilters = dd.UseFilters.ValueBoolPointer()
		}
		if typeutils.IsKnown(dd.UseTimeRange) {
			wire.UseTimeRange = dd.UseTimeRange.ValueBoolPointer()
		}
		if typeutils.IsKnown(dd.OpenInNewTab) {
			wire.OpenInNewTab = dd.OpenInNewTab.ValueBoolPointer()
		}
		if err := item.FromKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0(wire); err != nil {
			diags.AddError("Invalid dashboard drilldown", err.Error())
		}
	case d.URLDrilldown != nil:
		ud := d.URLDrilldown
		wire := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1{
			Url:     ud.URL.ValueString(),
			Label:   ud.Label.ValueString(),
			Trigger: kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1Trigger(ud.Trigger.ValueString()),
			Type:    kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown,
		}
		if typeutils.IsKnown(ud.EncodeURL) {
			wire.EncodeUrl = ud.EncodeURL.ValueBoolPointer()
		}
		if typeutils.IsKnown(ud.OpenInNewTab) {
			wire.OpenInNewTab = ud.OpenInNewTab.ValueBoolPointer()
		}
		if err := item.FromKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1(wire); err != nil {
			diags.AddError("Invalid URL drilldown", err.Error())
		}
	default:
		diags.AddError("Invalid drilldown", "Each drilldown must set either `dashboard_drilldown` or `url_drilldown`.")
	}
	return item, diags
}

// PopulateFromAPI maps API panel into practitioner state seeded from tfPanel.
func PopulateFromAPI(pm *models.PanelModel, tfPanel *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage) {
	apiCfg := apiPanel.Config

	if pm.ImageConfig == nil {
		pm.ImageConfig = imagePanelConfigFromAPIImport(apiPanel)
	}

	if tfPanel == nil {
		return
	}

	existing := pm.ImageConfig
	if existing == nil {
		return
	}

	panelkit.ApplyPresentationFromAPI(&existing.Title, &existing.Description, &existing.HideTitle, &existing.HideBorder,
		apiCfg.Title, apiCfg.Description, apiCfg.HideTitle, apiCfg.HideBorder)

	if typeutils.IsKnown(existing.AltText) {
		existing.AltText = types.StringPointerValue(apiCfg.ImageConfig.AltText)
	}
	if typeutils.IsKnown(existing.BackgroundColor) {
		existing.BackgroundColor = types.StringPointerValue(apiCfg.ImageConfig.BackgroundColor)
	}

	existing.ObjectFit = nullPreservingImageObjectFit(existing.ObjectFit, apiCfg.ImageConfig.ObjectFit)
	existing.Src = panelSrcFromAPI(apiCfg.ImageConfig.Src)

	var priorDrilldowns []models.ImagePanelDrilldownModel
	if tfPanel.ImageConfig != nil {
		priorDrilldowns = tfPanel.ImageConfig.Drilldowns
	}
	existing.Drilldowns = readImageDrilldownsFromAPI(apiCfg.Drilldowns, priorDrilldowns)
}

func nullPreservingImageObjectFit(prior types.String, api *kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigObjectFit) types.String {
	if api == nil {
		return types.StringNull()
	}
	v := string(*api)
	if prior.IsNull() || !typeutils.IsKnown(prior) {
		if v == string(kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigObjectFitContain) {
			return types.StringNull()
		}
		return types.StringValue(v)
	}
	return types.StringValue(v)
}

func imagePanelConfigFromAPIImport(apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage) *models.ImagePanelConfigModel {
	apiCfg := apiPanel.Config
	return &models.ImagePanelConfigModel{
		Src:             panelSrcFromAPI(apiCfg.ImageConfig.Src),
		AltText:         types.StringPointerValue(apiCfg.ImageConfig.AltText),
		BackgroundColor: types.StringPointerValue(apiCfg.ImageConfig.BackgroundColor),
		Title:           types.StringPointerValue(apiCfg.Title),
		Description:     types.StringPointerValue(apiCfg.Description),
		HideTitle:       types.BoolPointerValue(apiCfg.HideTitle),
		HideBorder:      types.BoolPointerValue(apiCfg.HideBorder),
		Drilldowns:      readImageDrilldownsFromAPI(apiCfg.Drilldowns, nil),
		ObjectFit:       nullPreservingImageObjectFit(types.StringNull(), apiCfg.ImageConfig.ObjectFit),
	}
}

func panelSrcFromAPI(src kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_ImageConfig_Src) models.ImagePanelSrcModel {
	var out models.ImagePanelSrcModel
	src0, err := src.AsKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc0()
	if err == nil && src0.Type == kbapi.File {
		out.File = &models.ImagePanelSrcFileModel{
			FileID: types.StringValue(src0.FileId),
		}
		return out
	}
	src1, err := src.AsKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigImageConfigSrc1()
	if err == nil && src1.Type == kbapi.Url {
		out.URL = &models.ImagePanelSrcURLModel{
			URL: types.StringValue(src1.Url),
		}
	}
	return out
}

func readImageDrilldownsFromAPI(
	api *[]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_Drilldowns_Item,
	prior []models.ImagePanelDrilldownModel,
) []models.ImagePanelDrilldownModel {
	if api == nil || len(*api) == 0 {
		return nil
	}
	out := make([]models.ImagePanelDrilldownModel, len(*api))
	for i, item := range *api {
		var p *models.ImagePanelDrilldownModel
		if i < len(prior) {
			p = &prior[i]
		}
		out[i] = readImageDrilldownFromAPI(item, p)
	}
	return out
}

func readImageDrilldownFromAPI(item kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImage_Config_Drilldowns_Item, prior *models.ImagePanelDrilldownModel) models.ImagePanelDrilldownModel {
	dd0, err0 := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0()
	if err0 == nil && dd0.Type == kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown {
		var priorDash *models.ImagePanelDashboardDrilldownModel
		if prior != nil {
			priorDash = prior.DashboardDrilldown
		}
		return models.ImagePanelDrilldownModel{
			DashboardDrilldown: readImageDashboardDrilldownFromAPI(dd0, priorDash),
		}
	}

	dd1, err1 := item.AsKibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1()
	if err1 == nil && dd1.Type == kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown {
		var priorURL *models.ImagePanelURLDrilldownModel
		if prior != nil {
			priorURL = prior.URLDrilldown
		}
		return models.ImagePanelDrilldownModel{
			URLDrilldown: readImageURLDrilldownFromAPI(dd1, priorURL),
		}
	}

	return models.ImagePanelDrilldownModel{}
}

func readImageDashboardDrilldownFromAPI(
	api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns0,
	prior *models.ImagePanelDashboardDrilldownModel,
) *models.ImagePanelDashboardDrilldownModel {
	m := &models.ImagePanelDashboardDrilldownModel{
		DashboardID: types.StringValue(api.DashboardId),
		Label:       types.StringValue(api.Label),
		Trigger:     types.StringValue(string(api.Trigger)),
	}

	if prior == nil {
		m.UseFilters = panelkit.DrilldownBoolImportPreserving(api.UseFilters, drilldownDashboardBoolDefault)
		m.UseTimeRange = panelkit.DrilldownBoolImportPreserving(api.UseTimeRange, drilldownDashboardBoolDefault)
		m.OpenInNewTab = panelkit.DrilldownBoolImportPreserving(api.OpenInNewTab, drilldownDashboardBoolDefault)
		return m
	}

	switch {
	case prior.UseFilters.IsNull():
		m.UseFilters = types.BoolNull()
	case api.UseFilters != nil:
		m.UseFilters = types.BoolValue(*api.UseFilters)
	default:
		m.UseFilters = types.BoolNull()
	}

	switch {
	case prior.UseTimeRange.IsNull():
		m.UseTimeRange = types.BoolNull()
	case api.UseTimeRange != nil:
		m.UseTimeRange = types.BoolValue(*api.UseTimeRange)
	default:
		m.UseTimeRange = types.BoolNull()
	}

	switch {
	case prior.OpenInNewTab.IsNull():
		m.OpenInNewTab = types.BoolNull()
	case api.OpenInNewTab != nil:
		m.OpenInNewTab = types.BoolValue(*api.OpenInNewTab)
	default:
		m.OpenInNewTab = types.BoolNull()
	}

	return m
}

func readImageURLDrilldownFromAPI(api kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeImageConfigDrilldowns1, prior *models.ImagePanelURLDrilldownModel) *models.ImagePanelURLDrilldownModel {
	m := &models.ImagePanelURLDrilldownModel{
		URL:     types.StringValue(api.Url),
		Label:   types.StringValue(api.Label),
		Trigger: types.StringValue(string(api.Trigger)),
	}

	if prior == nil {
		m.EncodeURL = panelkit.DrilldownBoolImportPreserving(api.EncodeUrl, drilldownURLEncodeURLDefault)
		m.OpenInNewTab = panelkit.DrilldownBoolImportPreserving(api.OpenInNewTab, drilldownURLOpenInNewTabDefault)
		return m
	}

	switch {
	case prior.EncodeURL.IsNull():
		m.EncodeURL = types.BoolNull()
	case api.EncodeUrl != nil:
		m.EncodeURL = types.BoolValue(*api.EncodeUrl)
	default:
		m.EncodeURL = types.BoolNull()
	}

	switch {
	case prior.OpenInNewTab.IsNull():
		m.OpenInNewTab = types.BoolNull()
	case api.OpenInNewTab != nil:
		m.OpenInNewTab = types.BoolValue(*api.OpenInNewTab)
	default:
		m.OpenInNewTab = types.BoolNull()
	}

	return m
}
