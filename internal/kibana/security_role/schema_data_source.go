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

package security_role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieve a specific role. See the [role management API documentation](https://www.elastic.co/guide/en/kibana/current/role-management-specific-api-get.html) for more details.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name for the role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for the role",
				Optional:    true,
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Optional meta-data.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"elasticsearch": schema.SingleNestedAttribute{
				Description: "Elasticsearch cluster and index privileges.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"cluster": schema.SetAttribute{
						Description: "List of the cluster privileges.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"run_as": schema.SetAttribute{
						Description: "A list of usernames the owners of this role can impersonate.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"indices": schema.SetNestedAttribute{
						Description: "A list of indices permissions entries.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"field_security": schema.SingleNestedAttribute{
									Description: "The document fields that the owners of the role have read access to.",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"grant": schema.SetAttribute{
											Description: "List of the fields to grant the access to.",
											Computed:    true,
											ElementType: types.StringType,
										},
										"except": schema.SetAttribute{
											Description: "List of the fields to which the grants will not be applied.",
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								"query": schema.StringAttribute{
									Description: "A search query that defines the documents the owners of the role have read access to.",
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								"names": schema.SetAttribute{
									Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
									Computed:    true,
									ElementType: types.StringType,
								},
								"privileges": schema.SetAttribute{
									Description: "The index level privileges that the owners of the role have on the specified indices.",
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
					"remote_indices": schema.SetNestedAttribute{
						Description: remoteIndicesPermissionsDescription,
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"clusters": schema.SetAttribute{
									Description: "A list of cluster aliases to which the permissions in this entry apply.",
									Computed:    true,
									ElementType: types.StringType,
								},
								"field_security": schema.SingleNestedAttribute{
									Description: "The document fields that the owners of the role have read access to.",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"grant": schema.SetAttribute{
											Description: "List of the fields to grant the access to.",
											Computed:    true,
											ElementType: types.StringType,
										},
										"except": schema.SetAttribute{
											Description: "List of the fields to which the grants will not be applied.",
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								"query": schema.StringAttribute{
									Description: "A search query that defines the documents the owners of the role have read access to.",
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								"names": schema.SetAttribute{
									Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
									Computed:    true,
									ElementType: types.StringType,
								},
								"privileges": schema.SetAttribute{
									Description: "The index level privileges that the owners of the role have on the specified indices.",
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
			"kibana": schema.SetNestedAttribute{
				Description: "The list of objects that specify the Kibana privileges for the role.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"spaces": schema.SetAttribute{
							Description: "The spaces to apply the privileges to. To grant access to all spaces, set to [\"*\"], or omit the value.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"base": schema.SetAttribute{
							Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"].",
							Computed:    true,
							ElementType: types.StringType,
						},
						"feature": schema.SetNestedAttribute{
							Description: "List of privileges for specific features.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Feature name.",
										Computed:    true,
									},
									"privileges": schema.SetAttribute{
										Description: "Feature privileges.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
