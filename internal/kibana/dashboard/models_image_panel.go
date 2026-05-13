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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// imagePanelConfigModel is the Terraform model for `image_config`.
type imagePanelConfigModel struct {
	Src             imagePanelSrcModel         `tfsdk:"src"`
	AltText         types.String               `tfsdk:"alt_text"`
	ObjectFit       types.String               `tfsdk:"object_fit"`
	BackgroundColor types.String               `tfsdk:"background_color"`
	Title           types.String               `tfsdk:"title"`
	Description     types.String               `tfsdk:"description"`
	HideTitle       types.Bool                 `tfsdk:"hide_title"`
	HideBorder      types.Bool                 `tfsdk:"hide_border"`
	Drilldowns      []imagePanelDrilldownModel `tfsdk:"drilldowns"`
}

type imagePanelSrcModel struct {
	File *imagePanelSrcFileModel `tfsdk:"file"`
	URL  *imagePanelSrcURLModel  `tfsdk:"url"`
}

type imagePanelSrcFileModel struct {
	FileID types.String `tfsdk:"file_id"`
}

type imagePanelSrcURLModel struct {
	URL types.String `tfsdk:"url"`
}

type imagePanelDrilldownModel struct {
	DashboardDrilldown *imagePanelDashboardDrilldownModel `tfsdk:"dashboard_drilldown"`
	URLDrilldown       *imagePanelURLDrilldownModel       `tfsdk:"url_drilldown"`
}

type imagePanelDashboardDrilldownModel struct {
	DashboardID  types.String `tfsdk:"dashboard_id"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	UseFilters   types.Bool   `tfsdk:"use_filters"`
	UseTimeRange types.Bool   `tfsdk:"use_time_range"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

type imagePanelURLDrilldownModel struct {
	URL          types.String `tfsdk:"url"`
	Label        types.String `tfsdk:"label"`
	Trigger      types.String `tfsdk:"trigger"`
	EncodeURL    types.Bool   `tfsdk:"encode_url"`
	OpenInNewTab types.Bool   `tfsdk:"open_in_new_tab"`
}

func imagePanelToAPI(pm panelModel, grid struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
}, panelID *string) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := pm.ImageConfig
	if cfg == nil {
		diags.AddError("Missing image panel configuration", "Image panels require `image_config`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	out := kbapi.KbnDashboardPanelTypeImage{
		Grid: grid,
		Id:   panelID,
		Type: kbapi.Image,
	}

	if typeutils.IsKnown(cfg.Title) {
		out.Config.Title = cfg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.Description) {
		out.Config.Description = cfg.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.HideTitle) {
		out.Config.HideTitle = cfg.HideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfg.HideBorder) {
		out.Config.HideBorder = cfg.HideBorder.ValueBoolPointer()
	}

	img := &out.Config.ImageConfig
	if typeutils.IsKnown(cfg.AltText) {
		img.AltText = cfg.AltText.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.BackgroundColor) {
		img.BackgroundColor = cfg.BackgroundColor.ValueStringPointer()
	}
	if typeutils.IsKnown(cfg.ObjectFit) {
		fit := kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFit(cfg.ObjectFit.ValueString())
		img.ObjectFit = &fit
	}

	switch {
	case cfg.Src.File != nil:
		src0 := kbapi.KbnDashboardPanelTypeImageConfigImageConfigSrc0{
			Type:   kbapi.File,
			FileId: cfg.Src.File.FileID.ValueString(),
		}
		if err := img.Src.FromKbnDashboardPanelTypeImageConfigImageConfigSrc0(src0); err != nil {
			diags.AddError("Invalid image src", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	case cfg.Src.URL != nil:
		src1 := kbapi.KbnDashboardPanelTypeImageConfigImageConfigSrc1{
			Type: kbapi.Url,
			Url:  cfg.Src.URL.URL.ValueString(),
		}
		if err := img.Src.FromKbnDashboardPanelTypeImageConfigImageConfigSrc1(src1); err != nil {
			diags.AddError("Invalid image src", err.Error())
			return kbapi.DashboardPanelItem{}, diags
		}
	default:
		diags.AddError("Invalid image src", "Exactly one of `file` or `url` must be set inside `src`.")
		return kbapi.DashboardPanelItem{}, diags
	}

	if len(cfg.Drilldowns) > 0 {
		items := make([]kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item, 0, len(cfg.Drilldowns))
		for _, d := range cfg.Drilldowns {
			item, dDiags := imagePanelDrilldownToAPI(d)
			diags.Append(dDiags...)
			if dDiags.HasError() {
				return kbapi.DashboardPanelItem{}, diags
			}
			items = append(items, item)
		}
		out.Config.Drilldowns = &items
	}

	var panelItem kbapi.DashboardPanelItem
	if err := panelItem.FromKbnDashboardPanelTypeImage(out); err != nil {
		diags.AddError("Failed to create image panel", err.Error())
	}
	return panelItem, diags
}

func imagePanelDrilldownToAPI(d imagePanelDrilldownModel) (kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item, diag.Diagnostics) {
	var diags diag.Diagnostics
	var item kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item

	switch {
	case d.DashboardDrilldown != nil:
		dd := d.DashboardDrilldown
		wire := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0{
			DashboardId: dd.DashboardID.ValueString(),
			Label:       dd.Label.ValueString(),
			Trigger:     kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0Trigger(dd.Trigger.ValueString()),
			Type:        kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown,
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
		if err := item.FromKbnDashboardPanelTypeImageConfigDrilldowns0(wire); err != nil {
			diags.AddError("Invalid dashboard drilldown", err.Error())
		}
	case d.URLDrilldown != nil:
		ud := d.URLDrilldown
		wire := kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1{
			Url:     ud.URL.ValueString(),
			Label:   ud.Label.ValueString(),
			Trigger: kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1Trigger(ud.Trigger.ValueString()),
			Type:    kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown,
		}
		if typeutils.IsKnown(ud.EncodeURL) {
			wire.EncodeUrl = ud.EncodeURL.ValueBoolPointer()
		}
		if typeutils.IsKnown(ud.OpenInNewTab) {
			wire.OpenInNewTab = ud.OpenInNewTab.ValueBoolPointer()
		}
		if err := item.FromKbnDashboardPanelTypeImageConfigDrilldowns1(wire); err != nil {
			diags.AddError("Invalid URL drilldown", err.Error())
		}
	default:
		diags.AddError("Invalid drilldown", "Each drilldown must set either `dashboard_drilldown` or `url_drilldown`.")
	}
	return item, diags
}

func populateImagePanelFromAPI(pm *panelModel, tfPanel *panelModel, apiPanel kbapi.KbnDashboardPanelTypeImage) {
	apiCfg := apiPanel.Config

	if tfPanel == nil {
		pm.ImageConfig = imagePanelConfigFromAPIImport(apiPanel)
		return
	}

	existing := pm.ImageConfig
	if existing == nil {
		return
	}

	if typeutils.IsKnown(existing.Title) {
		existing.Title = types.StringPointerValue(apiCfg.Title)
	}
	if typeutils.IsKnown(existing.Description) {
		existing.Description = types.StringPointerValue(apiCfg.Description)
	}
	if typeutils.IsKnown(existing.HideTitle) {
		existing.HideTitle = types.BoolPointerValue(apiCfg.HideTitle)
	}
	if typeutils.IsKnown(existing.HideBorder) {
		existing.HideBorder = types.BoolPointerValue(apiCfg.HideBorder)
	}

	if typeutils.IsKnown(existing.AltText) {
		existing.AltText = types.StringPointerValue(apiCfg.ImageConfig.AltText)
	}
	if typeutils.IsKnown(existing.BackgroundColor) {
		existing.BackgroundColor = types.StringPointerValue(apiCfg.ImageConfig.BackgroundColor)
	}

	existing.ObjectFit = nullPreservingImageObjectFit(existing.ObjectFit, apiCfg.ImageConfig.ObjectFit)
	existing.Src = imagePanelSrcFromAPI(apiCfg.ImageConfig.Src)
	existing.Drilldowns = readImageDrilldownsFromAPI(apiCfg.Drilldowns, existing.Drilldowns)
}

func nullPreservingImageObjectFit(prior types.String, api *kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFit) types.String {
	if api == nil {
		return types.StringNull()
	}
	v := string(*api)
	if prior.IsNull() || !typeutils.IsKnown(prior) {
		if v == string(kbapi.KbnDashboardPanelTypeImageConfigImageConfigObjectFitContain) {
			return types.StringNull()
		}
		return types.StringValue(v)
	}
	return types.StringValue(v)
}

func imagePanelConfigFromAPIImport(apiPanel kbapi.KbnDashboardPanelTypeImage) *imagePanelConfigModel {
	apiCfg := apiPanel.Config
	cfg := &imagePanelConfigModel{
		Src:             imagePanelSrcFromAPI(apiCfg.ImageConfig.Src),
		AltText:         types.StringPointerValue(apiCfg.ImageConfig.AltText),
		BackgroundColor: types.StringPointerValue(apiCfg.ImageConfig.BackgroundColor),
		Title:           types.StringPointerValue(apiCfg.Title),
		Description:     types.StringPointerValue(apiCfg.Description),
		HideTitle:       types.BoolPointerValue(apiCfg.HideTitle),
		HideBorder:      types.BoolPointerValue(apiCfg.HideBorder),
		Drilldowns:      readImageDrilldownsFromAPI(apiCfg.Drilldowns, nil),
		ObjectFit:       nullPreservingImageObjectFit(types.StringNull(), apiCfg.ImageConfig.ObjectFit),
	}
	return cfg
}

func imagePanelSrcFromAPI(src kbapi.KbnDashboardPanelTypeImage_Config_ImageConfig_Src) imagePanelSrcModel {
	var out imagePanelSrcModel
	src0, err := src.AsKbnDashboardPanelTypeImageConfigImageConfigSrc0()
	if err == nil && src0.Type == kbapi.File {
		out.File = &imagePanelSrcFileModel{
			FileID: types.StringValue(src0.FileId),
		}
		return out
	}
	src1, err := src.AsKbnDashboardPanelTypeImageConfigImageConfigSrc1()
	if err == nil && src1.Type == kbapi.Url {
		out.URL = &imagePanelSrcURLModel{
			URL: types.StringValue(src1.Url),
		}
	}
	return out
}

func readImageDrilldownsFromAPI(
	api *[]kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item,
	prior []imagePanelDrilldownModel,
) []imagePanelDrilldownModel {
	if api == nil || len(*api) == 0 {
		return nil
	}
	out := make([]imagePanelDrilldownModel, len(*api))
	for i, item := range *api {
		var p *imagePanelDrilldownModel
		if i < len(prior) {
			p = &prior[i]
		}
		out[i] = readImageDrilldownFromAPI(item, p)
	}
	return out
}

func readImageDrilldownFromAPI(item kbapi.KbnDashboardPanelTypeImage_Config_Drilldowns_Item, prior *imagePanelDrilldownModel) imagePanelDrilldownModel {
	dd0, err0 := item.AsKbnDashboardPanelTypeImageConfigDrilldowns0()
	if err0 == nil && dd0.Type == kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0TypeDashboardDrilldown {
		var priorDash *imagePanelDashboardDrilldownModel
		if prior != nil {
			priorDash = prior.DashboardDrilldown
		}
		return imagePanelDrilldownModel{
			DashboardDrilldown: readImageDashboardDrilldownFromAPI(dd0, priorDash),
		}
	}

	dd1, err1 := item.AsKbnDashboardPanelTypeImageConfigDrilldowns1()
	if err1 == nil && dd1.Type == kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1TypeUrlDrilldown {
		var priorURL *imagePanelURLDrilldownModel
		if prior != nil {
			priorURL = prior.URLDrilldown
		}
		return imagePanelDrilldownModel{
			URLDrilldown: readImageURLDrilldownFromAPI(dd1, priorURL),
		}
	}

	return imagePanelDrilldownModel{}
}

func readImageDashboardDrilldownFromAPI(
	api kbapi.KbnDashboardPanelTypeImageConfigDrilldowns0,
	prior *imagePanelDashboardDrilldownModel,
) *imagePanelDashboardDrilldownModel {
	m := &imagePanelDashboardDrilldownModel{
		DashboardID: types.StringValue(api.DashboardId),
		Label:       types.StringValue(api.Label),
		Trigger:     types.StringValue(string(api.Trigger)),
	}

	// Import (no prior practitioner state for this drilldown): omit API values that match Kibana defaults so
	// omitted HCL stays aligned with imported state (REQ-040).
	if prior == nil {
		m.UseFilters = panelDrilldownBoolImportPreserving(api.UseFilters, drilldownDashboardBoolDefault)
		m.UseTimeRange = panelDrilldownBoolImportPreserving(api.UseTimeRange, drilldownDashboardBoolDefault)
		m.OpenInNewTab = panelDrilldownBoolImportPreserving(api.OpenInNewTab, drilldownDashboardBoolDefault)
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

func readImageURLDrilldownFromAPI(api kbapi.KbnDashboardPanelTypeImageConfigDrilldowns1, prior *imagePanelURLDrilldownModel) *imagePanelURLDrilldownModel {
	m := &imagePanelURLDrilldownModel{
		URL:     types.StringValue(api.Url),
		Label:   types.StringValue(api.Label),
		Trigger: types.StringValue(string(api.Trigger)),
	}

	if prior == nil {
		m.EncodeURL = panelDrilldownBoolImportPreserving(api.EncodeUrl, drilldownURLEncodeURLDefault)
		m.OpenInNewTab = panelDrilldownBoolImportPreserving(api.OpenInNewTab, drilldownURLOpenInNewTabDefault)
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
