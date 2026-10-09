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

package alertingrules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ruleElementAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Rule ID.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Rule name.",
			Computed:            true,
		},
		"rule_type_id": schema.StringAttribute{
			MarkdownDescription: "Rule type identifier.",
			Computed:            true,
		},
		"consumer": schema.StringAttribute{
			MarkdownDescription: "Kibana application that owns the rule.",
			Computed:            true,
		},
		"enabled": schema.BoolAttribute{
			MarkdownDescription: "Whether the rule is enabled.",
			Computed:            true,
		},
		"tags": schema.SetAttribute{
			MarkdownDescription: "Tags on the rule. An empty set when the rule has no tags.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"scheduled_task_id": schema.StringAttribute{
			MarkdownDescription: "Identifier of the rule's scheduled task. Null when Kibana omits it.",
			Computed:            true,
		},
		"last_execution_status": schema.StringAttribute{
			MarkdownDescription: "Status of the rule's last execution, such as `ok` or `active`. Null when Kibana reports no execution status.",
			Computed:            true,
		},
		"last_execution_date": schema.StringAttribute{
			MarkdownDescription: "UTC RFC3339 timestamp of the rule's last execution, with exactly three fractional-second digits. " +
				"Null when Kibana reports no last execution time or the time cannot be parsed.",
			Computed: true,
		},
	}
}

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Reads Kibana alerting rules in a space, either one rule by ID or the rules selected by a KQL filter.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				MarkdownDescription: "Kibana space identifier. When omitted, the default space is used and stored in state.",
				Optional:            true,
				Computed:            true,
			},
			"rule_id": schema.StringAttribute{
				MarkdownDescription: "Plain Kibana rule ID to look up in `space_id`. Mutually exclusive with `filter`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("filter")),
				},
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "KQL filter passed to Kibana unmodified. Mutually exclusive with `rule_id`. " +
					"When both `rule_id` and `filter` are omitted, every rule in the space is returned. Set `filter` in spaces with many rules.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("rule_id")),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Space and, for a single-rule lookup, rule ID. `<space_id>/<rule_id>` when `rule_id` is set, otherwise `<space_id>`.",
				Computed:            true,
			},
			"rules": schema.ListNestedAttribute{
				MarkdownDescription: "Rules returned by the lookup or search. `last_execution_status`, `last_execution_date`, and `scheduled_task_id` are null when Kibana does not report them.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: ruleElementAttributes(),
				},
			},
		},
	}
}
