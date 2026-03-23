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

package exportsavedobjects

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export Kibana saved objects. This data source allows you to export saved objects from Kibana and store the result in the Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID for the export.",
				Computed:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"objects": schema.ListNestedAttribute{
				Description: "List of objects to export.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of the saved object.",
							Required:    true,
						},
						"id": schema.StringAttribute{
							Description: "The ID of the saved object.",
							Required:    true,
						},
					},
				},
			},
			"exclude_export_details": schema.BoolAttribute{
				Description: "Do not add export details. Defaults to true.",
				Optional:    true,
			},
			"include_references_deep": schema.BoolAttribute{
				Description: "Include references to other saved objects recursively. Defaults to true.",
				Optional:    true,
			},
			"exported_objects": schema.StringAttribute{
				Description: "The exported objects in NDJSON format.",
				Computed:    true,
			},
		},
	}
}

type objectModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	Objects               types.List   `tfsdk:"objects"`
	ExcludeExportDetails  types.Bool   `tfsdk:"exclude_export_details"`
	IncludeReferencesDeep types.Bool   `tfsdk:"include_references_deep"`
	ExportedObjects       types.String `tfsdk:"exported_objects"`
}
