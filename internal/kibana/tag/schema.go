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

package tag

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Kibana tag. Requires Kibana 9.5.0 or later. " +
			"Tags managed by Kibana cannot be controlled by this resource; use the " +
			"`elasticstack_kibana_tags` data source to read them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<space_id>/<tag_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrTagID: schema.StringAttribute{
				MarkdownDescription: "Client-specified UUID for the tag. When set, the provider uses PUT semantics to " +
					"create the tag. When omitted, the server mints the ID on POST. Changing this value forces replacement.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrSpaceID: kbschema.ResourceSpaceIDAttribute(),
			attrName: schema.StringAttribute{
				MarkdownDescription: "Display name of the tag.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
			attrColor: schema.StringAttribute{
				MarkdownDescription: "Hex color for the tag (for example `#772299`). When omitted, the server generates a random color.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(tagHexColorPattern, "must be a six-digit hex color in the form `#RRGGBB`"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrDescription: schema.StringAttribute{
				MarkdownDescription: "Optional description of the tag.",
				Optional:            true,
			},
			attrCreatedAt: schema.StringAttribute{
				MarkdownDescription: "ISO 8601 timestamp when the tag was created.",
				Computed:            true,
			},
			attrUpdatedAt: schema.StringAttribute{
				MarkdownDescription: "ISO 8601 timestamp when the tag was last updated.",
				Computed:            true,
			},
		},
	}
}
