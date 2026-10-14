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

package casetemplate

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed resource-description.md
var resourceDescription string

func getResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the template: `<space_id>/<template_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": kbschema.ResourceSpaceIDAttribute(),
			"template_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The server-assigned ID of the template, stable across versions.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The solution that owns the template: `cases` (Stack), `securitySolution`, or `observability`. Changing this forces a new resource.",
				Validators: []validator.String{
					stringvalidator.OneOf("cases", "securitySolution", "observability"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The template name, 1-50 characters, unique per owner (case-insensitive).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
			"definition_yaml": schema.StringAttribute{
				Required:   true,
				CustomType: customtypes.NormalizedYamlType{},
				MarkdownDescription: "The template definition as a YAML string: case defaults (`name`, `severity`, `category`, `tags`, `assignees`, `connector`, `settings`) " +
					"and a `fields` array of inline field definitions and/or `$ref` entries into the field library. " +
					"Edits that only change whitespace, comments, or mapping-key order are ignored at plan time. Limited to 30000 characters.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(30000),
				},
				PlanModifiers: []planmodifier.String{
					suppressSemanticYAMLDiff(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Optional description of the template, at most 1000 characters. " +
					"Removing it from configuration clears the stored value; an empty string is kept as an empty string.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1000),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Optional tags (at most 200, each at most 256 characters). Order and duplicates are preserved. " +
					"Removing the attribute from configuration clears the stored tags; an empty list is kept as an empty list.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(200),
					listvalidator.ValueStringsAre(stringvalidator.LengthAtMost(256)),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Whether the template is enabled. Disabled templates are hidden from the case creation flow but remain manageable. " +
					"The server default of `true` applies only on create; omitting the attribute on a later update keeps the current value, " +
					"so set it to `true` explicitly to re-enable a disabled template.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"template_version": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The current version of the template. Increments on every successful update.",
				PlanModifiers: []planmodifier.Int64{
					keepComputedWhenOnlyFormattingChanged(),
				},
			},
			"field_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of fields defined in the template, when reported by the API.",
				PlanModifiers: []planmodifier.Int64{
					keepComputedWhenOnlyFormattingChanged(),
				},
			},
			"usage_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of times the template has been used to create a case, when reported by the API.",
				PlanModifiers: []planmodifier.Int64{
					keepComputedWhenOnlyFormattingChanged(),
				},
			},
			"is_default": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this is the default template for its owner, when reported by the API.",
				PlanModifiers: []planmodifier.Bool{
					keepComputedWhenOnlyFormattingChanged(),
				},
			},
		},
	}
}
