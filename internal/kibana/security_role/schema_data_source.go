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
			attrName: schema.StringAttribute{
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
					attrCluster: schema.SetAttribute{
						Description: "List of the cluster privileges.",
						Computed:    true,
						ElementType: types.StringType,
					},
					attrRunAs: schema.SetAttribute{
						Description: "A list of usernames the owners of this role can impersonate.",
						Computed:    true,
						ElementType: types.StringType,
					},
					attrIndices: schema.SetNestedAttribute{
						Description: "A list of indices permissions entries.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								attrFieldSecurity: schema.SingleNestedAttribute{
									Description: descFieldSecurityBlock,
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										attrGrant: schema.SetAttribute{
											Description: descFieldSecurityGrant,
											Computed:    true,
											ElementType: types.StringType,
										},
										attrExcept: schema.SetAttribute{
											Description: descFieldSecurityExcept,
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								attrQuery: schema.StringAttribute{
									Description: descIndexQuery,
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								attrNames: schema.SetAttribute{
									Description: descIndexNames,
									Computed:    true,
									ElementType: types.StringType,
								},
								attrPrivileges: schema.SetAttribute{
									Description: descIndexPrivileges,
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
					attrRemoteIndices: schema.SetNestedAttribute{
						Description: remoteIndicesPermissionsDescription,
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								attrClusters: schema.SetAttribute{
									Description: "A list of cluster aliases to which the permissions in this entry apply.",
									Computed:    true,
									ElementType: types.StringType,
								},
								attrFieldSecurity: schema.SingleNestedAttribute{
									Description: descFieldSecurityBlock,
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										attrGrant: schema.SetAttribute{
											Description: descFieldSecurityGrant,
											Computed:    true,
											ElementType: types.StringType,
										},
										attrExcept: schema.SetAttribute{
											Description: descFieldSecurityExcept,
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
								attrQuery: schema.StringAttribute{
									Description: descIndexQuery,
									Computed:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								attrNames: schema.SetAttribute{
									Description: descIndexNames,
									Computed:    true,
									ElementType: types.StringType,
								},
								attrPrivileges: schema.SetAttribute{
									Description: descIndexPrivileges,
									Computed:    true,
									ElementType: types.StringType,
								},
								attrAllowRestrictedIndices: schema.BoolAttribute{
									Description: allowRestrictedIndicesDescription,
									Computed:    true,
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
						attrSpaces: schema.SetAttribute{
							Description: "The spaces to apply the privileges to. To grant access to all spaces, set to [\"*\"], or omit the value.",
							Computed:    true,
							ElementType: types.StringType,
						},
						attrBase: schema.SetAttribute{
							Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"].",
							Computed:    true,
							ElementType: types.StringType,
						},
						attrFeature: schema.SetNestedAttribute{
							Description: "List of privileges for specific features.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									attrName: schema.StringAttribute{
										Description: "Feature name.",
										Computed:    true,
									},
									attrPrivileges: schema.SetAttribute{
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
