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

package security

import (
	"context"
	"encoding/json"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Description   types.String         `tfsdk:"description"`
	Cluster       types.Set            `tfsdk:"cluster"`
	RunAs         types.Set            `tfsdk:"run_as"`
	Global        jsontypes.Normalized `tfsdk:"global"`
	Metadata      jsontypes.Normalized `tfsdk:"metadata"`
	Applications  types.Set            `tfsdk:"applications"`
	Indices       types.Set            `tfsdk:"indices"`
	RemoteIndices types.Set            `tfsdk:"remote_indices"`
}

type applicationDataSourceModel struct {
	Application types.String `tfsdk:"application"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Resources   types.Set    `tfsdk:"resources"`
}

type fieldSecurityDataSourceModel struct {
	Grant  types.Set `tfsdk:"grant"`
	Except types.Set `tfsdk:"except"`
}

type indexDataSourceModel struct {
	FieldSecurity          types.Object         `tfsdk:"field_security"`
	Names                  types.Set            `tfsdk:"names"`
	Privileges             types.Set            `tfsdk:"privileges"`
	Query                  jsontypes.Normalized `tfsdk:"query"`
	AllowRestrictedIndices types.Bool           `tfsdk:"allow_restricted_indices"`
}

type remoteIndexDataSourceModel struct {
	Clusters      types.Set            `tfsdk:"clusters"`
	FieldSecurity types.Object         `tfsdk:"field_security"`
	Names         types.Set            `tfsdk:"names"`
	Privileges    types.Set            `tfsdk:"privileges"`
	Query         jsontypes.Normalized `tfsdk:"query"`
}

func NewRoleDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[roleDataSourceModel](
		entitycore.ComponentElasticsearch,
		"security_role",
		getDataSourceSchema,
		readDataSource,
	)
}

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves roles in the native realm. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the role.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the role.",
				Computed:            true,
			},
			"cluster": schema.SetAttribute{
				MarkdownDescription: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"run_as": schema.SetAttribute{
				MarkdownDescription: "A list of users that the owners of this role can impersonate.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"global": schema.StringAttribute{
				MarkdownDescription: "An object defining global privileges.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional meta-data.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"applications": schema.SetNestedAttribute{
				MarkdownDescription: "A list of application privilege entries.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"application": schema.StringAttribute{
							MarkdownDescription: "The name of the application to which this entry applies.",
							Computed:            true,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "A list of strings, where each element is the name of an application privilege or action.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"resources": schema.SetAttribute{
							MarkdownDescription: "A list resources to which the privileges are applied.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
			"indices": schema.SetNestedAttribute{
				MarkdownDescription: "A list of indices permissions entries.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_security": schema.ListNestedAttribute{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"grant": schema.SetAttribute{
										MarkdownDescription: "List of the fields to grant the access to.",
										ElementType:         types.StringType,
										Computed:            true,
									},
									"except": schema.SetAttribute{
										MarkdownDescription: "List of the fields to which the grants will not be applied.",
										ElementType:         types.StringType,
										Computed:            true,
									},
								},
							},
						},
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Computed:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						"allow_restricted_indices": schema.BoolAttribute{
							MarkdownDescription: roleAllowRestrictedIndicesDescription,
							Computed:            true,
						},
					},
				},
			},
			"remote_indices": schema.SetNestedAttribute{
				MarkdownDescription: roleRemoteIndicesDescription,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"clusters": schema.SetAttribute{
							MarkdownDescription: "A list of cluster aliases to which the permissions in this entry apply.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"field_security": schema.ListNestedAttribute{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"grant": schema.SetAttribute{
										MarkdownDescription: "List of the fields to grant the access to.",
										ElementType:         types.StringType,
										Computed:            true,
									},
									"except": schema.SetAttribute{
										MarkdownDescription: "List of the fields to which the grants will not be applied.",
										ElementType:         types.StringType,
										Computed:            true,
									},
								},
							},
						},
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Computed:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
					},
				},
			},
		},
	}
}

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config roleDataSourceModel) (roleDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	return config, diags
}

func flattenApplicationsData(apps []esTypes.ApplicationPrivileges) []any {
	if len(apps) > 0 {
		oapps := make([]any, len(apps))
		for i, app := range apps {
			oa := make(map[string]any)
			oa["application"] = app.Application
			oa["privileges"] = app.Privileges
			oa["resources"] = app.Resources
			oapps[i] = oa
		}
		return oapps
	}
	return make([]any, 0)
}

func flattenIndicesData(indices []esTypes.IndicesPrivileges) []any {
	oindx := make([]any, len(indices))

	for i, index := range indices {
		oi := make(map[string]any)
		oi["names"] = index.Names

		privileges := make([]string, len(index.Privileges))
		for j, p := range index.Privileges {
			privileges[j] = p.String()
		}
		oi["privileges"] = privileges

		var queryStr *string
		if index.Query != nil {
			switch q := index.Query.(type) {
			case string:
				queryStr = &q
			default:
				b, err := json.Marshal(index.Query)
				if err != nil {
					b = []byte("null")
				}
				s := string(b)
				queryStr = &s
			}
		}
		oi["query"] = queryStr
		oi["allow_restricted_indices"] = index.AllowRestrictedIndices

		if index.FieldSecurity != nil {
			fsec := make(map[string]any)
			fsec["grant"] = index.FieldSecurity.Grant
			fsec["except"] = index.FieldSecurity.Except
			oi["field_security"] = []any{fsec}
		}
		oindx[i] = oi
	}
	return oindx
}

func flattenRemoteIndicesData(remoteIndices []esTypes.RemoteIndicesPrivileges) []any {
	oRemoteIndx := make([]any, len(remoteIndices))

	for i, remoteIndex := range remoteIndices {
		oi := make(map[string]any)
		oi["names"] = remoteIndex.Names
		oi["clusters"] = remoteIndex.Clusters

		privileges := make([]string, len(remoteIndex.Privileges))
		for j, p := range remoteIndex.Privileges {
			privileges[j] = p.String()
		}
		oi["privileges"] = privileges

		var queryStr *string
		if remoteIndex.Query != nil {
			switch q := remoteIndex.Query.(type) {
			case string:
				queryStr = &q
			default:
				b, err := json.Marshal(remoteIndex.Query)
				if err != nil {
					b = []byte("null")
				}
				s := string(b)
				queryStr = &s
			}
		}
		oi["query"] = queryStr

		if remoteIndex.FieldSecurity != nil {
			fsec := make(map[string]any)
			fsec["grant"] = remoteIndex.FieldSecurity.Grant
			fsec["except"] = remoteIndex.FieldSecurity.Except
			oi["field_security"] = []any{fsec}
		}
		oRemoteIndx[i] = oi
	}
	return oRemoteIndx
}