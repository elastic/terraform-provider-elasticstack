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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Handler implements iface.Handler for Kibana `links` dashboard panels.
type Handler struct{}

func (Handler) PanelType() string                  { return panelType }
func (Handler) SchemaAttribute() schema.Attribute  { return SchemaAttribute() }
func (Handler) ClassifyJSON(_ map[string]any) bool { return false }
func (Handler) PopulateJSONDefaults(config map[string]any) map[string]any {
	return config
}

func (Handler) PinnedHandler() iface.PinnedHandler { return nil }

func (Handler) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}

// ValidatePanelConfig ensures that `links_config` is present for links panels.
func (Handler) ValidatePanelConfig(_ context.Context, attrs map[string]attr.Value, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if panelkit.AttrConcreteSet(attrs["links_config"]) || panelkit.AttrUnknown(attrs["links_config"]) {
		return diags
	}

	diags.AddAttributeError(
		attrPath.AtName("links_config"),
		"Missing links panel configuration",
		"Links panels require `links_config`.",
	)
	return diags
}

// FromAPI populates Terraform panel state from a links panel API item.
func (Handler) FromAPI(ctx context.Context, pm, prior *models.PanelModel, item kbapi.DashboardPanelItem) diag.Diagnostics {
	return panelkit.SimpleFromAPI(ctx, pm, prior,
		item.AsKibanaHTTPAPIsKbnDashboardPanelTypeLinks,
		func(p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks) (kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, *string) {
			return p.Grid, p.Id
		},
		func(pm *models.PanelModel, prior *models.PanelModel, p kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks) diag.Diagnostics {
			return populateLinksPanelFromAPI(ctx, pm, prior, p)
		},
	)
}

// ToAPI serializes Terraform links panel state into kbapi.
func (Handler) ToAPI(pm models.PanelModel, dashboard *models.DashboardModel) (kbapi.DashboardPanelItem, diag.Diagnostics) {
	_ = dashboard

	var diags diag.Diagnostics

	diags.Append(panelkit.RejectConfigJSON(pm, panelType)...)
	if diags.HasError() {
		return kbapi.DashboardPanelItem{}, diags
	}

	if pm.LinksConfig == nil {
		diags.AddError(
			"Missing links panel configuration",
			"Links panels require `links_config`.",
		)
		return kbapi.DashboardPanelItem{}, diags
	}

	return panelkit.SimpleToAPI(pm,
		func(grid kbapi.KibanaHTTPAPIsKbnDashboardPanelGrid, id *string) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks, diag.Diagnostics) {
			config, d := linksConfigToAPI(*pm.LinksConfig)
			return kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks{
				Config: config,
				Grid:   grid,
				Id:     id,
				Type:   kbapi.Links,
			}, d
		},
		func(item *kbapi.DashboardPanelItem, panel kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks) error {
			return item.FromKibanaHTTPAPIsKbnDashboardPanelTypeLinks(panel)
		},
		"Failed to create links panel",
	)
}

func linksConfigToAPI(cfg models.LinksPanelConfigModel) (kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config, diag.Diagnostics) {
	var config kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config

	switch {
	case cfg.ByValue != nil:
		c0 := linksByValueConfigToAPI(*cfg.ByValue)
		if err := config.FromKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0(c0); err != nil {
			return config, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to build links by-value config", err.Error())}
		}

	case cfg.ByReference != nil:
		c1 := linksByReferenceConfigToAPI(*cfg.ByReference)
		if err := config.FromKibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1(c1); err != nil {
			return config, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to build links by-reference config", err.Error())}
		}

	default:
		return config, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid links_config",
			"Exactly one of `by_value` or `by_reference` must be set inside `links_config`.",
		)}
	}

	return config, nil
}

func linksByValueConfigToAPI(bv models.LinksPanelByValueModel) kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0 {
	config := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0{}

	if typeutils.IsKnown(bv.Layout) {
		layout := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0Layout(bv.Layout.ValueString())
		config.Layout = &layout
	}

	panelkit.BuildPresentationConfig(
		bv.Title, bv.Description, bv.HideTitle, bv.HideBorder,
		&config.Title, &config.Description, &config.HideTitle, &config.HideBorder,
	)

	config.Links = make([]kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config_0_Links_Item, len(bv.Links))
	for i, link := range bv.Links {
		config.Links[i] = linkItemToAPI(link)
	}

	return config
}

func linksByReferenceConfigToAPI(br models.LinksPanelByReferenceModel) kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1 {
	config := kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1{
		RefId: br.RefID.ValueString(),
	}

	panelkit.BuildPresentationConfig(
		br.Title, br.Description, br.HideTitle, br.HideBorder,
		&config.Title, &config.Description, &config.HideTitle, &config.HideBorder,
	)

	return config
}

func linkItemToAPI(item models.LinkItemModel) kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config_0_Links_Item {
	var out kbapi.KibanaHTTPAPIsKbnDashboardPanelTypeLinks_Config_0_Links_Item

	switch item.Type.ValueString() {
	case "dashboard":
		link := kbapi.KibanaHTTPAPIsKbnLinkPanelTypeDashboardLink{
			Destination: item.Destination.ValueString(),
			Label:       typeutils.OptionalString(item.Label),
			Type:        kbapi.DashboardLink,
		}

		opts := &struct {
			OpenInNewTab *bool `json:"open_in_new_tab,omitempty"`
			UseFilters   *bool `json:"use_filters,omitempty"`
			UseTimeRange *bool `json:"use_time_range,omitempty"`
		}{
			OpenInNewTab: typeutils.OptionalBool(item.OpenInNewTab),
			UseFilters:   typeutils.OptionalBool(item.UseFilters),
			UseTimeRange: typeutils.OptionalBool(item.UseTimeRange),
		}
		if opts.OpenInNewTab != nil || opts.UseFilters != nil || opts.UseTimeRange != nil {
			link.Options = opts
		}

		_ = out.FromKibanaHTTPAPIsKbnLinkPanelTypeDashboardLink(link)

	case "external":
		link := kbapi.KibanaHTTPAPIsKbnLinkTypeExternalLink{
			Destination: item.Destination.ValueString(),
			Label:       typeutils.OptionalString(item.Label),
			Type:        kbapi.ExternalLink,
		}

		//nolint:revive // EncodeUrl must match the generated kbapi field name.
		opts := &struct {
			EncodeUrl    *bool `json:"encode_url,omitempty"`
			OpenInNewTab *bool `json:"open_in_new_tab,omitempty"`
		}{
			EncodeUrl:    typeutils.OptionalBool(item.EncodeURL),
			OpenInNewTab: typeutils.OptionalBool(item.OpenInNewTab),
		}
		if opts.EncodeUrl != nil || opts.OpenInNewTab != nil {
			link.Options = opts
		}

		_ = out.FromKibanaHTTPAPIsKbnLinkTypeExternalLink(link)
	}

	return out
}
