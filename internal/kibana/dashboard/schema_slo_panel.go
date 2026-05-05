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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// getSloOverviewSchema returns the schema for the slo_overview_config block.
func getSloOverviewSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"single": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a single-SLO overview panel. Mutually exclusive with `groups`.",
			Optional:            true,
			Attributes:          getSloSingleSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("groups")),
			},
		},
		"groups": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a grouped SLO overview panel. Mutually exclusive with `single`.",
			Optional:            true,
			Attributes:          getSloGroupsSchema(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("single")),
			},
		},
	}
}

// getSloSharedDisplaySchema returns display attributes shared by both single and groups modes.
func getSloSharedDisplaySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "The title displayed on the panel.",
			Optional:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description displayed on the panel.",
			Optional:            true,
		},
		"hide_title": schema.BoolAttribute{
			MarkdownDescription: "When true, the panel title is hidden.",
			Optional:            true,
		},
		"hide_border": schema.BoolAttribute{
			MarkdownDescription: "When true, the panel border is hidden.",
			Optional:            true,
		},
		"drilldowns": schema.ListNestedAttribute{
			MarkdownDescription: "URL drilldowns attached to the panel. The trigger (`on_open_panel_menu`) and type (`url_drilldown`) are set automatically.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "The URL template for the drilldown. Variables are documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable.",
						Required:            true,
					},
					"label": schema.StringAttribute{
						MarkdownDescription: "The display label for the drilldown link.",
						Required:            true,
					},
					"encode_url": schema.BoolAttribute{
						MarkdownDescription: "When true, the URL is percent-encoded.",
						Optional:            true,
					},
					"open_in_new_tab": schema.BoolAttribute{
						MarkdownDescription: "When true, the drilldown URL opens in a new browser tab.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// getSloSingleSchema returns the attributes for the single sub-block.
func getSloSingleSchema() map[string]schema.Attribute {
	attrs := getSloSharedDisplaySchema()
	attrs["slo_id"] = schema.StringAttribute{
		MarkdownDescription: "The unique identifier of the SLO to display.",
		Required:            true,
	}
	attrs["slo_instance_id"] = schema.StringAttribute{
		MarkdownDescription: "The SLO instance ID. Set when the SLO uses group_by; identifies which instance to display. Defaults to `*` (all instances) when omitted.",
		Optional:            true,
	}
	attrs["remote_name"] = schema.StringAttribute{
		MarkdownDescription: "The name of the remote cluster where the SLO is defined.",
		Optional:            true,
	}
	return attrs
}

// getSloGroupsSchema returns the attributes for the groups sub-block.
func getSloGroupsSchema() map[string]schema.Attribute {
	attrs := getSloSharedDisplaySchema()
	attrs["group_filters"] = schema.SingleNestedAttribute{
		MarkdownDescription: "Optional filters for grouped SLO overview mode.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"group_by": schema.StringAttribute{
				MarkdownDescription: "Group SLOs by this field. Valid values are `slo.tags`, `status`, `slo.indicator.type`, `_index`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("slo.tags", "status", "slo.indicator.type", "_index"),
				},
			},
			"groups": schema.ListAttribute{
				MarkdownDescription: "List of group values to include (maximum 100).",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
			},
			"kql_query": schema.StringAttribute{
				MarkdownDescription: "KQL query string to filter the SLOs shown in the group overview.",
				Optional:            true,
			},
			"filters_json": schema.StringAttribute{
				MarkdownDescription: "AS-code filter array as a JSON string. Accepts the polymorphic filter schema (condition, group, DSL, spatial).",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
			},
		},
	}
	return attrs
}

// sloOverviewConfigModeValidator ensures exactly one of single or groups is set.
var _ validator.Object = sloOverviewConfigModeValidator{}

type sloOverviewConfigModeValidator struct{}

func (v sloOverviewConfigModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `single` or `groups` is configured inside `slo_overview_config`."
}

func (v sloOverviewConfigModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sloOverviewConfigModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	singleVal := attrs["single"]
	groupsVal := attrs["groups"]

	singleSet := singleVal != nil && !singleVal.IsNull() && !singleVal.IsUnknown()
	groupsSet := groupsVal != nil && !groupsVal.IsNull() && !groupsVal.IsUnknown()

	if singleSet && groupsSet {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid slo_overview_config", "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`, not both.")
		return
	}
	if !singleSet && !groupsSet {
		// Both unknown is acceptable (during planning with computed resources).
		singleUnknown := singleVal != nil && singleVal.IsUnknown()
		groupsUnknown := groupsVal != nil && groupsVal.IsUnknown()
		if singleUnknown || groupsUnknown {
			return
		}
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid slo_overview_config", "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`.")
	}
}
