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
	"fmt"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	roleName := config.Name.ValueString()

	// Resolve the composite ID
	id, sdkDiags := esClient.ID(ctx, roleName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	// Call GetRole (returns SDK v2 diagnostics)
	role, sdkDiags := elasticsearch.GetRole(ctx, esClient, roleName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	// Not-found: return empty ID, keep name, no diagnostics
	if role == nil {
		config.ID = types.StringValue("")
		config.Description = types.StringNull()
		config.Cluster = types.SetNull(types.StringType)
		config.RunAs = types.SetNull(types.StringType)
		config.Global = jsontypes.NewNormalizedNull()
		config.Metadata = jsontypes.NewNormalizedNull()
		config.Applications = types.SetNull(types.ObjectType{AttrTypes: getApplicationDSAttrTypes()})
		config.Indices = types.SetNull(types.ObjectType{AttrTypes: getIndexPermsDSAttrTypes()})
		config.RemoteIndices = types.SetNull(types.ObjectType{AttrTypes: getRemoteIndexPermsDSAttrTypes()})
		return config, diags
	}

	// Map API response to model
	diags.Append(config.fromAPIModel(ctx, role)...)
	if diags.HasError() {
		return config, diags
	}

	// Ensure name is set to the role name we looked up
	config.Name = types.StringValue(roleName)

	return config, diags
}

func (config *roleDataSourceModel) fromAPIModel(ctx context.Context, role *esTypes.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	// Description
	if role.Description != nil {
		config.Description = types.StringValue(*role.Description)
	} else {
		config.Description = types.StringNull()
	}

	// Cluster
	clusterStrings := make([]string, len(role.Cluster))
	for i, cp := range role.Cluster {
		clusterStrings[i] = cp.String()
	}
	clusterSet, d := types.SetValueFrom(ctx, types.StringType, clusterStrings)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	config.Cluster = clusterSet

	// RunAs
	runAsSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(role.RunAs))
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	config.RunAs = runAsSet

	// Global
	if role.Global != nil {
		globalBytes, err := json.Marshal(role.Global)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling global JSON: %s", err))
			return diags
		}
		config.Global = jsontypes.NewNormalizedValue(string(globalBytes))
	} else {
		config.Global = jsontypes.NewNormalizedNull()
	}

	// Metadata
	if role.Metadata != nil {
		metadataBytes, err := json.Marshal(role.Metadata)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
			return diags
		}
		config.Metadata = jsontypes.NewNormalizedValue(string(metadataBytes))
	} else {
		config.Metadata = jsontypes.NewNormalizedNull()
	}

	// Applications
	if len(role.Applications) > 0 {
		appElements := make([]attr.Value, len(role.Applications))
		for i, app := range role.Applications {
			privSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(app.Privileges))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			resSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(app.Resources))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			appObj, d := types.ObjectValue(getApplicationDSAttrTypes(), map[string]attr.Value{
				"application": types.StringValue(app.Application),
				"privileges":  privSet,
				"resources":   resSet,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			appElements[i] = appObj
		}

		appSet, d := types.SetValue(types.ObjectType{AttrTypes: getApplicationDSAttrTypes()}, appElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		config.Applications = appSet
	} else {
		config.Applications = types.SetNull(types.ObjectType{AttrTypes: getApplicationDSAttrTypes()})
	}

	// Indices
	if len(role.Indices) > 0 {
		indicesElements := make([]attr.Value, len(role.Indices))
		for i, index := range role.Indices {
			namesSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(index.Names))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			privileges := make([]string, len(index.Privileges))
			for j, p := range index.Privileges {
				privileges[j] = p.String()
			}
			privSet, d := types.SetValueFrom(ctx, types.StringType, privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			var queryVal jsontypes.Normalized
			if index.Query != nil {
				switch q := index.Query.(type) {
				case string:
					queryVal = jsontypes.NewNormalizedValue(q)
				default:
					b, err := json.Marshal(index.Query)
					if err != nil {
						diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling query: %s", err))
						return diags
					}
					queryVal = jsontypes.NewNormalizedValue(string(b))
				}
			} else {
				queryVal = jsontypes.NewNormalizedNull()
			}

			var allowRestrictedVal types.Bool
			if index.AllowRestrictedIndices != nil {
				allowRestrictedVal = types.BoolValue(*index.AllowRestrictedIndices)
			} else {
				allowRestrictedVal = types.BoolNull()
			}

			// Build field_security as a types.List with 0 or 1 elements (ListNestedAttribute)
			var fieldSecList types.List
			if index.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(index.FieldSecurity.Grant))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(index.FieldSecurity.Except))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d := types.ObjectValue(getFieldSecurityDSAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecList, d = types.ListValue(types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()}, []attr.Value{fieldSecObj})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecList = types.ListValueMust(types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()}, []attr.Value{})
			}

			indexObj, d := types.ObjectValue(getIndexPermsDSAttrTypes(), map[string]attr.Value{
				"field_security":           fieldSecList,
				"names":                    namesSet,
				"privileges":               privSet,
				"query":                    queryVal,
				"allow_restricted_indices": allowRestrictedVal,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			indicesElements[i] = indexObj
		}

		indicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getIndexPermsDSAttrTypes()}, indicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		config.Indices = indicesSet
	} else {
		config.Indices = types.SetNull(types.ObjectType{AttrTypes: getIndexPermsDSAttrTypes()})
	}

	// Remote Indices
	if len(role.RemoteIndices) > 0 {
		remoteIndicesElements := make([]attr.Value, len(role.RemoteIndices))
		for i, remoteIndex := range role.RemoteIndices {
			clustersSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(remoteIndex.Clusters))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			namesSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(remoteIndex.Names))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			privileges := make([]string, len(remoteIndex.Privileges))
			for j, p := range remoteIndex.Privileges {
				privileges[j] = p.String()
			}
			privSet, d := types.SetValueFrom(ctx, types.StringType, privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			var queryVal jsontypes.Normalized
			if remoteIndex.Query != nil {
				switch q := remoteIndex.Query.(type) {
				case string:
					queryVal = jsontypes.NewNormalizedValue(q)
				default:
					b, err := json.Marshal(remoteIndex.Query)
					if err != nil {
						diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling query: %s", err))
						return diags
					}
					queryVal = jsontypes.NewNormalizedValue(string(b))
				}
			} else {
				queryVal = jsontypes.NewNormalizedNull()
			}

			// Build field_security as a types.List with 0 or 1 elements (ListNestedAttribute)
			var fieldSecList types.List
			if remoteIndex.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(remoteIndex.FieldSecurity.Grant))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(remoteIndex.FieldSecurity.Except))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d := types.ObjectValue(getFieldSecurityDSAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecList, d = types.ListValue(types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()}, []attr.Value{fieldSecObj})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecList = types.ListValueMust(types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()}, []attr.Value{})
			}

			remoteIndexObj, d := types.ObjectValue(getRemoteIndexPermsDSAttrTypes(), map[string]attr.Value{
				"clusters":       clustersSet,
				"field_security": fieldSecList,
				"names":          namesSet,
				"privileges":     privSet,
				"query":          queryVal,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			remoteIndicesElements[i] = remoteIndexObj
		}

		remoteIndicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getRemoteIndexPermsDSAttrTypes()}, remoteIndicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		config.RemoteIndices = remoteIndicesSet
	} else {
		config.RemoteIndices = types.SetNull(types.ObjectType{AttrTypes: getRemoteIndexPermsDSAttrTypes()})
	}

	return diags
}

// Data source attribute type helpers (mirror data source schema structure)
func getApplicationDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"application": types.StringType,
		"privileges":  types.SetType{ElemType: types.StringType},
		"resources":   types.SetType{ElemType: types.StringType},
	}
}

func getFieldSecurityDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"grant":  types.SetType{ElemType: types.StringType},
		"except": types.SetType{ElemType: types.StringType},
	}
}

func getIndexPermsDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"field_security": types.ListType{
			ElemType: types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()},
		},
		"names":                    types.SetType{ElemType: types.StringType},
		"privileges":               types.SetType{ElemType: types.StringType},
		"query":                    jsontypes.NormalizedType{},
		"allow_restricted_indices": types.BoolType,
	}
}

func getRemoteIndexPermsDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"clusters": types.SetType{ElemType: types.StringType},
		"field_security": types.ListType{
			ElemType: types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()},
		},
		"names":      types.SetType{ElemType: types.StringType},
		"privileges": types.SetType{ElemType: types.StringType},
		"query":      jsontypes.NormalizedType{},
	}
}
