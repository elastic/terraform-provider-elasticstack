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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{
		Description: dataSourceDescription,
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"policy_id": dsschema.StringAttribute{
				Description: policyIDDescription,
				Optional:    true,
			},
			"space_id": dsschema.StringAttribute{
				Description: spaceIDDescription,
				Optional:    true,
			},
			"tokens": dsschema.ListNestedAttribute{
				Description: "A list of enrollment tokens.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"key_id": dsschema.StringAttribute{
							Description: "The unique identifier of the enrollment token.",
							Computed:    true,
						},
						"api_key": dsschema.StringAttribute{
							Description: "The API key.",
							Computed:    true,
							Sensitive:   true,
						},
						"api_key_id": dsschema.StringAttribute{
							Description: "The API key identifier.",
							Computed:    true,
						},
						"created_at": dsschema.StringAttribute{
							Description: "The time at which the enrollment token was created.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the enrollment token.",
							Computed:    true,
						},
						"active": dsschema.BoolAttribute{
							Description: "Indicates if the enrollment token is active.",
							Computed:    true,
						},
						"policy_id": dsschema.StringAttribute{
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
	return getDataSourceSchema().Attributes["tokens"].GetType().(attr.TypeWithElementType).ElementType()
}
