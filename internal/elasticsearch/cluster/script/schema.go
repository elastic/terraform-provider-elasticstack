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

package script

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *scriptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"script_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the stored script. Must be unique within the cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lang": schema.StringAttribute{
				MarkdownDescription: "Script language. For search templates, use `mustache`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("painless", "expression", "mustache", "java"),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "For scripts, a string containing the script. For search templates, an object containing the search template.",
				Required:            true,
			},
			"params": schema.StringAttribute{
				MarkdownDescription: "Parameters for the script or search template.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"context": schema.StringAttribute{
				MarkdownDescription: "Context in which the script or search template should run.",
				Optional:            true,
			},
		},
	}
}
