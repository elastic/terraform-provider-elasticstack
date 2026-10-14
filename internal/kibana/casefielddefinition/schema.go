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

package casefielddefinition

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed resource-description.md
var resourceDescription string

func getResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the field definition: `<space_id>/<field_definition_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": kbschema.ResourceSpaceIDAttribute(),
			"field_definition_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The server-assigned UUID of the field definition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The solution that owns the field definition: `cases` (Stack), `securitySolution`, or `observability`. Changing this forces a new resource.",
				Validators: []validator.String{
					stringvalidator.OneOf("cases", "securitySolution", "observability"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The field name, unique per owner (case-insensitive). Must match the `name` key in `definition`. " +
					"When omitted, the server extracts it from `definition`. Immutable: changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition": schema.StringAttribute{
				Required:   true,
				CustomType: customtypes.NormalizedYamlType{},
				MarkdownDescription: "The field definition as a YAML string describing a single field (type, label, control, metadata, validation). " +
					"Known limitation: edits that only reformat the YAML (key order, whitespace) may still plan an in-place update " +
					"even though the value is semantically unchanged. Limited to 30000 characters.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(30000),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Optional human-readable description of the field's purpose. The API preserves the stored value when " +
					"the attribute is omitted, so removing it from configuration does not clear it; set it to an empty string (`\"\"`) to clear it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_global": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "When true, the field is rendered in every case for this owner regardless of the template used. " +
					"A global field cannot be set back to false while it is linked to an active custom field. " +
					"The API preserves the stored value when the attribute is omitted, so removing it from configuration does not reset it; set it to `false` explicitly.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"display_order": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Position of a global field in the case details view, assigned by the server when the field is first made global; it is unset for fields that have never been global.",
			},
		},
	}
}
