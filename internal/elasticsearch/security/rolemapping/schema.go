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

package rolemapping

import (
	"context"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *roleMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The distinct name that identifies the role mapping, used solely as an identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"rules": schema.StringAttribute{
				MarkdownDescription: "The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.",
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "A list of role names that are granted to the users that match the role mapping rules.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ExactlyOneOf(path.MatchRoot("role_templates")),
				},
			},
			"role_templates": schema.StringAttribute{
				MarkdownDescription: "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("roles")),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				Default:             stringdefault.StaticString("{}"),
			},
		},
	}
}
