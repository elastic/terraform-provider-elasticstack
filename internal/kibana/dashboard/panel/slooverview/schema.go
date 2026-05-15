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

package slooverview

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "slo_overview"

// SchemaAttribute returns the slo_overview_config SingleNestedAttribute definition.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for an SLO overview panel. Use either `single` (for a single SLO) or `groups` (for grouped SLO overview).",
		BlockName:   "slo_overview_config",
		PanelType:   panelType,
		Attributes:  nestedAttributes(),
		ExtraValidators: []validator.Object{
			panelkit.ExactlyOneOfNestedAttrsValidator(panelkit.ExactlyOneOfNestedAttrsOpts{
				AttrNames:     []string{"single", "groups"},
				Summary:       "Invalid slo_overview_config",
				MissingDetail: "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`.",
				TooManyDetail: "Exactly one of `single` or `groups` must be configured inside `slo_overview_config`, not both.",
				Description:   "Ensures exactly one of `single` or `groups` is configured inside `slo_overview_config`.",
			}),
		},
	})
}

func nestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"single": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a single-SLO overview panel. Mutually exclusive with `groups`.",
			Optional:            true,
			Attributes:          singleAttributes(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("groups")),
			},
		},
		"groups": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for a grouped SLO overview panel. Mutually exclusive with `single`.",
			Optional:            true,
			Attributes:          groupsAttributes(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("single")),
			},
		},
	}
}

func sharedDisplayAttributes() map[string]schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["drilldowns"] = panelkit.URLDrilldownListAttribute(
		"URL drilldowns attached to the panel. The trigger (`on_open_panel_menu`) and type (`url_drilldown`) are set automatically.",
		panelkit.URLDrilldownOptions{
			URLMarkdownDescription:          "The URL template for the drilldown. Variables are documented at https://www.elastic.co/docs/explore-analyze/dashboards/drilldowns#url-template-variable.",
			LabelMarkdownDescription:        "The display label for the drilldown link.",
			EncodeURLMarkdownDescription:    "When true, the URL is percent-encoded.",
			OpenInNewTabMarkdownDescription: "When true, the drilldown URL opens in a new browser tab.",
		},
	)
	return attrs
}

func singleAttributes() map[string]schema.Attribute {
	attrs := sharedDisplayAttributes()
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

func groupsAttributes() map[string]schema.Attribute {
	attrs := sharedDisplayAttributes()
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
