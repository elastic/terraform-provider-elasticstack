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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Registry panel type key (`links` + `_config` => `links_config` on PanelModel).
// Terraform schema attribute keys reused across the links panel schema and validators.
const (
	panelType        = "links"
	linkTypeDashboard = "dashboard"
	linkTypeExternal  = "external"
	attrType         = "type"
	attrDestination  = "destination"
	attrLabel        = "label"
	attrOpenInNewTab = "open_in_new_tab"
	attrUseFilters   = "use_filters"
	attrUseTimeRange = "use_time_range"
	attrEncodeURL    = "encode_url"
)

var _ validator.Object = linksConfigModeValidator{}

// linksConfigModeValidator ensures exactly one of by_value or by_reference is set.
type linksConfigModeValidator struct{}

func (linksConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `by_value` or `by_reference` is set inside `links_config`."
}

func (v linksConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (linksConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	byValueSet := panelkit.AttrConcreteSet(attrs["by_value"])
	byRefSet := panelkit.AttrConcreteSet(attrs["by_reference"])

	if byValueSet && byRefSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid links_config", "Exactly one of `by_value` or `by_reference` must be set inside `links_config`, not both.")
		return
	}

	if !byValueSet && !byRefSet {
		if byValueAttr := attrs["by_value"]; byValueAttr != nil && byValueAttr.IsUnknown() {
			return
		}
		if byRefAttr := attrs["by_reference"]; byRefAttr != nil && byRefAttr.IsUnknown() {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid links_config", "Exactly one of `by_value` or `by_reference` must be set inside `links_config`.")
	}
}

var _ validator.Object = linksItemTypeValidator{}

// linksItemTypeValidator ensures type-specific attributes are only set for the matching link item type.
type linksItemTypeValidator struct{}

func (linksItemTypeValidator) Description(_ context.Context) string {
	return "Ensures type-specific attributes are only set for the matching link item `type`."
}

func (v linksItemTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (linksItemTypeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	typeAttr, ok := attrs[attrType]
	if !ok || typeAttr == nil || typeAttr.IsNull() || typeAttr.IsUnknown() {
		return
	}

	typ, ok := typeAttr.(types.String)
	if !ok {
		return
	}

	switch typ.ValueString() {
	case linkTypeDashboard:
		if panelkit.AttrConcreteSet(attrs[attrEncodeURL]) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(attrEncodeURL),
				"Invalid link item",
				"`encode_url` is not valid for `type = \"dashboard\"`; it is only valid for `type = \"external\"`.",
			)
		}
	case linkTypeExternal:
		if panelkit.AttrConcreteSet(attrs[attrUseFilters]) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(attrUseFilters),
				"Invalid link item",
				"`use_filters` is not valid for `type = \"external\"`; it is only valid for `type = \"dashboard\"`.",
			)
		}
		if panelkit.AttrConcreteSet(attrs[attrUseTimeRange]) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(attrUseTimeRange),
				"Invalid link item",
				"`use_time_range` is not valid for `type = \"external\"`; it is only valid for `type = \"dashboard\"`.",
			)
		}
	}
}

// SchemaAttribute returns the Terraform schema for `links_config`.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a `links` panel (`kbn-dashboard-panel-type-links`). " +
			"Set exactly one of `by_value` or `by_reference`.",
		BlockName:       "links_config",
		PanelType:       panelType,
		Required:        true,
		Attributes:      linksSchemaInnerAttributes(),
		ExtraValidators: []validator.Object{linksConfigModeValidator{}},
	})
}

func linksSchemaInnerAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"by_value": schema.SingleNestedAttribute{
			MarkdownDescription: "Inline links panel configuration.",
			Optional:            true,
			Attributes:          linksByValueAttributes(),
		},
		"by_reference": schema.SingleNestedAttribute{
			MarkdownDescription: "Reference a Kibana Links library saved object.",
			Optional:            true,
			Attributes:          linksByReferenceAttributes(),
		},
	}
}

func linksByValueAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{}
	maps.Copy(attrs, panelkit.PanelPresentationAttributes())
	maps.Copy(attrs, map[string]schema.Attribute{
		"layout": schema.StringAttribute{
			MarkdownDescription: "Layout direction for the links panel.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("horizontal", "vertical"),
			},
		},
		"links": schema.ListNestedAttribute{
			MarkdownDescription: "List of links to display in the panel.",
			Required:            true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: linksItemAttributes(),
				Validators: []validator.Object{linksItemTypeValidator{}},
			},
		},
	})
	return attrs
}

func linksByReferenceAttributes() map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{}
	maps.Copy(attrs, panelkit.PanelPresentationAttributes())
	maps.Copy(attrs, map[string]schema.Attribute{
		"ref_id": schema.StringAttribute{
			MarkdownDescription: "Reference id of a Kibana Links library saved object.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	})
	return attrs
}

func linksItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrType: schema.StringAttribute{
			MarkdownDescription: "Type of link: `dashboard` for an internal Kibana dashboard link, or `external` for an arbitrary URL.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(linkTypeDashboard, linkTypeExternal),
			},
		},
		attrDestination: schema.StringAttribute{
			MarkdownDescription: "Destination of the link: dashboard saved-object id for `dashboard` links, or a URL for `external` links.",
			Required:            true,
		},
		attrLabel: schema.StringAttribute{
			MarkdownDescription: "Optional display label for the link.",
			Optional:            true,
		},
		attrOpenInNewTab: schema.BoolAttribute{
			MarkdownDescription: "When true, opens the link in a new browser tab.",
			Optional:            true,
		},
		attrUseFilters: schema.BoolAttribute{
			MarkdownDescription: "When true, the dashboard link applies the current filters.",
			Optional:            true,
		},
		attrUseTimeRange: schema.BoolAttribute{
			MarkdownDescription: "When true, the dashboard link applies the current time range.",
			Optional:            true,
		},
		attrEncodeURL: schema.BoolAttribute{
			MarkdownDescription: "When true, the external URL is percent-encoded.",
			Optional:            true,
		},
	}
}
