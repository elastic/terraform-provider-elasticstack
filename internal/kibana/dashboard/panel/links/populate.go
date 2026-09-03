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

package links

import (
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// linksPanelAPIConfigLooksByReference distinguishes inline vs linked links panel configs.
// kbapi's Config union unmarshals successfully into both generated structs, so we key off
// JSON `ref_id` (present only on the by-reference branch).
func linksPanelAPIConfigLooksByReference(apiCfg kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config) bool {
	raw, err := apiCfg.MarshalJSON()
	if err != nil {
		return false
	}
	var probe struct {
		RefID string `json:"ref_id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return strings.TrimSpace(probe.RefID) != ""
}

// linksPanelPriorTFBranchMismatchesAPI reports out-of-band branch changes. Prior Terraform
// state used exclusively one branch while the API payload uses the other.
func linksPanelPriorTFBranchMismatchesAPI(apiLooksByRef bool, prior *models.LinksPanelConfigModel) bool {
	if prior == nil {
		return false
	}
	hasValue := prior.ByValue != nil
	hasRef := prior.ByReference != nil
	if apiLooksByRef && hasValue && !hasRef {
		return true
	}
	if !apiLooksByRef && hasRef && !hasValue {
		return true
	}
	return false
}

func populateLinksPanelFromAPI(pm, prior *models.PanelModel, apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks) diag.Diagnostics {
	var diags diag.Diagnostics

	apiByRef := linksPanelAPIConfigLooksByReference(apiPanel.Config)

	if pm.LinksConfig == nil {
		cfg, d := linksPanelConfigFromAPIImport(apiPanel, apiByRef)
		diags.Append(d...)
		pm.LinksConfig = cfg
		if prior == nil {
			return diags
		}
	}

	existing := pm.LinksConfig
	if existing == nil {
		return diags
	}

	if prior != nil && linksPanelPriorTFBranchMismatchesAPI(apiByRef, prior.LinksConfig) {
		cfg, d := linksPanelConfigFromAPIImport(apiPanel, apiByRef)
		diags.Append(d...)
		if cfg != nil {
			*existing = *cfg
		}
		return diags
	}

	if apiByRef {
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1()
		if err != nil {
			diags.AddError("Failed to decode links by-reference API config", err.Error())
			return diags
		}
		return linksPanelMergeConfig1FromAPI(existing, prior, cfg1)
	}

	cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0()
	if err != nil {
		diags.AddError("Failed to decode links by-value API config", err.Error())
		return diags
	}
	return linksPanelMergeConfig0FromAPI(existing, prior, cfg0)
}

func linksPanelConfigFromAPIImport(
	apiPanel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks,
	apiByRef bool,
) (*models.LinksPanelConfigModel, diag.Diagnostics) {
	if apiByRef {
		cfg1, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1()
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to decode links by-reference config", err.Error())}
		}
		return linksConfig1FromAPIImport(cfg1), nil
	}

	cfg0, err := apiPanel.Config.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0()
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to decode links by-value config", err.Error())}
	}
	return linksConfig0FromAPIImport(cfg0), nil
}

func linksConfig0FromAPIImport(cfg0 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0) *models.LinksPanelConfigModel {
	byValue := &models.LinksPanelByValueModel{
		Title:       types.StringPointerValue(cfg0.Title),
		Description: types.StringPointerValue(cfg0.Description),
		HideTitle:   types.BoolPointerValue(cfg0.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg0.HideBorder),
		Links:       make([]models.LinkItemModel, len(cfg0.Links)),
	}
	if cfg0.Layout != nil {
		byValue.Layout = types.StringValue(string(*cfg0.Layout))
	} else {
		byValue.Layout = types.StringNull()
	}
	for i, item := range cfg0.Links {
		byValue.Links[i] = linkItemFromAPI(item, nil)
	}

	return &models.LinksPanelConfigModel{
		ByValue: byValue,
	}
}

func linksConfig1FromAPIImport(cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1) *models.LinksPanelConfigModel {
	byReference := &models.LinksPanelByReferenceModel{
		RefID:       types.StringValue(cfg1.RefId),
		Title:       types.StringPointerValue(cfg1.Title),
		Description: types.StringPointerValue(cfg1.Description),
		HideTitle:   types.BoolPointerValue(cfg1.HideTitle),
		HideBorder:  types.BoolPointerValue(cfg1.HideBorder),
	}

	return &models.LinksPanelConfigModel{
		ByReference: byReference,
	}
}

func linksPanelMergeConfig0FromAPI(
	existing *models.LinksPanelConfigModel,
	prior *models.PanelModel,
	cfg0 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior == nil || prior.LinksConfig == nil || prior.LinksConfig.ByValue == nil {
		return diags
	}

	priorBV := prior.LinksConfig.ByValue

	mergeDisplayString(&existing.ByValue.Title, priorBV.Title, cfg0.Title)
	mergeDisplayString(&existing.ByValue.Description, priorBV.Description, cfg0.Description)
	mergeDisplayBool(&existing.ByValue.HideTitle, priorBV.HideTitle, cfg0.HideTitle)
	mergeDisplayBool(&existing.ByValue.HideBorder, priorBV.HideBorder, cfg0.HideBorder)

	if cfg0.Layout != nil {
		existing.ByValue.Layout = types.StringValue(string(*cfg0.Layout))
	}

	existing.ByValue.Links = make([]models.LinkItemModel, len(cfg0.Links))
	for i, apiItem := range cfg0.Links {
		var priorItem *models.LinkItemModel
		if i < len(priorBV.Links) {
			p := priorBV.Links[i]
			priorItem = &p
		}
		existing.ByValue.Links[i] = linkItemFromAPI(apiItem, priorItem)
	}

	return diags
}

func linksPanelMergeConfig1FromAPI(
	existing *models.LinksPanelConfigModel,
	prior *models.PanelModel,
	cfg1 kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1,
) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior == nil || prior.LinksConfig == nil || prior.LinksConfig.ByReference == nil {
		return diags
	}

	priorBR := prior.LinksConfig.ByReference

	mergeDisplayString(&existing.ByReference.Title, priorBR.Title, cfg1.Title)
	mergeDisplayString(&existing.ByReference.Description, priorBR.Description, cfg1.Description)
	mergeDisplayBool(&existing.ByReference.HideTitle, priorBR.HideTitle, cfg1.HideTitle)
	mergeDisplayBool(&existing.ByReference.HideBorder, priorBR.HideBorder, cfg1.HideBorder)

	if typeutils.IsKnown(priorBR.RefID) {
		existing.ByReference.RefID = types.StringValue(cfg1.RefId)
	}

	return diags
}

// mergeDisplayString merges an optional display string from the API into state. Unlike
// panelkit.ApplyPresentationFromAPI (which resets to null when the API omits the field),
// this keeps the prior known value when the API elides the field, so server-side omission
// of optional presentation fields does not surface as drift against user-authored config.
func mergeDisplayString(existing *types.String, prior types.String, api *string) {
	if !typeutils.IsKnown(prior) {
		*existing = prior
		return
	}
	if api != nil {
		*existing = types.StringValue(*api)
		return
	}
	*existing = prior
}

// mergeDisplayBool is the bool analogue of mergeDisplayString; see its rationale.
func mergeDisplayBool(existing *types.Bool, prior types.Bool, api *bool) {
	if !typeutils.IsKnown(prior) {
		*existing = prior
		return
	}
	if api != nil {
		*existing = types.BoolValue(*api)
		return
	}
	*existing = prior
}

func linkItemFromAPI(
	apiItem kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config_0_Links_Item,
	priorItem *models.LinkItemModel,
) models.LinkItemModel {
	// The generated AsXxx methods unmarshal without checking the discriminator,
	// and both branches have identical JSON shapes, so classify by the wire type.
	switch discriminator, _ := apiItem.Discriminator(); discriminator {
	case string(kbapi.DashboardLink):
		if dashboardLink, err := apiItem.AsKibanaHTTPAPIsKbnLinkPanelTypeDashboardLink(); err == nil {
			return dashboardLinkFromAPI(dashboardLink, priorItem)
		}
	case string(kbapi.ExternalLink):
		if externalLink, err := apiItem.AsKibanaHTTPAPIsKbnLinkTypeExternalLink(); err == nil {
			return externalLinkFromAPI(externalLink, priorItem)
		}
	}
	return models.LinkItemModel{}
}

func dashboardLinkFromAPI(
	api kbapi.KibanaHTTPAPIsKbnLinkPanelTypeDashboardLink,
	prior *models.LinkItemModel,
) models.LinkItemModel {
	m := models.LinkItemModel{
		Type:        types.StringValue(linkTypeDashboard),
		Destination: types.StringValue(api.Destination),
	}

	if prior == nil {
		m.Label = types.StringPointerValue(api.Label)
		if api.Options != nil {
			m.OpenInNewTab = types.BoolPointerValue(api.Options.OpenInNewTab)
			m.UseFilters = types.BoolPointerValue(api.Options.UseFilters)
			m.UseTimeRange = types.BoolPointerValue(api.Options.UseTimeRange)
		}
		return m
	}

	m.Label = linkStringFromAPI(prior.Label, api.Label)
	m.OpenInNewTab = prior.OpenInNewTab
	m.UseFilters = prior.UseFilters
	m.UseTimeRange = prior.UseTimeRange
	if api.Options != nil {
		if api.Options.OpenInNewTab != nil {
			m.OpenInNewTab = types.BoolValue(*api.Options.OpenInNewTab)
		}
		if api.Options.UseFilters != nil {
			m.UseFilters = types.BoolValue(*api.Options.UseFilters)
		}
		if api.Options.UseTimeRange != nil {
			m.UseTimeRange = types.BoolValue(*api.Options.UseTimeRange)
		}
	}

	return m
}

func externalLinkFromAPI(
	api kbapi.KibanaHTTPAPIsKbnLinkTypeExternalLink,
	prior *models.LinkItemModel,
) models.LinkItemModel {
	m := models.LinkItemModel{
		Type:        types.StringValue(linkTypeExternal),
		Destination: types.StringValue(api.Destination),
	}

	if prior == nil {
		m.Label = types.StringPointerValue(api.Label)
		if api.Options != nil {
			m.OpenInNewTab = types.BoolPointerValue(api.Options.OpenInNewTab)
			m.EncodeURL = types.BoolPointerValue(api.Options.EncodeUrl)
		}
		return m
	}

	m.Label = linkStringFromAPI(prior.Label, api.Label)
	m.OpenInNewTab = prior.OpenInNewTab
	m.EncodeURL = prior.EncodeURL
	if api.Options != nil {
		if api.Options.OpenInNewTab != nil {
			m.OpenInNewTab = types.BoolValue(*api.Options.OpenInNewTab)
		}
		if api.Options.EncodeUrl != nil {
			m.EncodeURL = types.BoolValue(*api.Options.EncodeUrl)
		}
	}

	return m
}

func linkStringFromAPI(prior types.String, api *string) types.String {
	if api != nil {
		return types.StringValue(*api)
	}
	return prior
}
