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

package enrollmenttokens

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *enrollmentTokensDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: dataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"policy_id": schema.StringAttribute{
				Description: policyIDDescription,
				Optional:    true,
			},
			"space_id": schema.StringAttribute{
				Description: spaceIDDescription,
				Optional:    true,
			},
			"tokens": schema.ListNestedAttribute{
				Description: "A list of enrollment tokens.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key_id": schema.StringAttribute{
							Description: "The unique identifier of the enrollment token.",
							Computed:    true,
						},
						"api_key": schema.StringAttribute{
							Description: "The API key.",
							Computed:    true,
							Sensitive:   true,
						},
						"api_key_id": schema.StringAttribute{
							Description: "The API key identifier.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The time at which the enrollment token was created.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the enrollment token.",
							Computed:    true,
						},
						"active": schema.BoolAttribute{
							Description: "Indicates if the enrollment token is active.",
							Computed:    true,
						},
						"policy_id": schema.StringAttribute{
							Description: "The identifier of the associated agent policy.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func getTokenType() attr.Type {
	return getSchema().Attributes["tokens"].GetType().(attr.TypeWithElementType).ElementType()
}
