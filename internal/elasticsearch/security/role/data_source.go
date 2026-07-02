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

package role

import (
	"context"
	"encoding/json"
	"fmt"

	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Retrieves roles in the native realm. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			attrName: dsschema.StringAttribute{
				MarkdownDescription: "The name of the role.",
				Required:            true,
			},
			attrDescription: dsschema.StringAttribute{
				MarkdownDescription: "The description of the role.",
				Computed:            true,
			},
			attrCluster: dsschema.SetAttribute{
				MarkdownDescription: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"run_as": dsschema.SetAttribute{
				MarkdownDescription: "A list of users that the owners of this role can impersonate.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			attrGlobal: dsschema.StringAttribute{
				MarkdownDescription: "An object defining global privileges.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			attrMetadata: dsschema.StringAttribute{
				MarkdownDescription: "Optional meta-data.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			blockApplications: dsschema.SetNestedAttribute{
				MarkdownDescription: "A list of application privilege entries.",
				Computed:            true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						attrApplication: dsschema.StringAttribute{
							MarkdownDescription: "The name of the application to which this entry applies.",
							Computed:            true,
						},
						attrPrivileges: dsschema.SetAttribute{
							MarkdownDescription: "A list of strings, where each element is the name of an application privilege or action.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrResources: dsschema.SetAttribute{
							MarkdownDescription: "A list resources to which the privileges are applied.",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
			blockIndices: dsschema.SetNestedAttribute{
				MarkdownDescription: "A list of indices permissions entries.",
				Computed:            true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						attrFieldSecurity: dsschema.ListNestedAttribute{
							MarkdownDescription: fieldSecurityDescription,
							Computed:            true,
							NestedObject: dsschema.NestedAttributeObject{
								Attributes: map[string]dsschema.Attribute{
									attrGrant: dsschema.SetAttribute{
										MarkdownDescription: fieldGrantDescription,
										ElementType:         types.StringType,
										Computed:            true,
									},
									attrExcept: dsschema.SetAttribute{
										MarkdownDescription: fieldExceptDescription,
										ElementType:         types.StringType,
										Computed:            true,
									},
								},
							},
						},
						attrNames: dsschema.SetAttribute{
							MarkdownDescription: indicesNamesDescription,
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrPrivileges: dsschema.SetAttribute{
							MarkdownDescription: indicesPrivilegesDescription,
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrQuery: dsschema.StringAttribute{
							MarkdownDescription: indicesQueryDescription,
							Computed:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						attrAllowRestrictedIndices: dsschema.BoolAttribute{
							MarkdownDescription: allowRestrictedIndicesDescription,
							Computed:            true,
						},
					},
				},
			},
			blockRemoteIndices: dsschema.SetNestedAttribute{
				MarkdownDescription: remoteIndicesDescription,
				Computed:            true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						attrClusters: dsschema.SetAttribute{
							MarkdownDescription: "A list of cluster aliases to which the permissions in this entry apply.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrFieldSecurity: dsschema.ListNestedAttribute{
							MarkdownDescription: fieldSecurityDescription,
							Computed:            true,
							NestedObject: dsschema.NestedAttributeObject{
								Attributes: map[string]dsschema.Attribute{
									attrGrant: dsschema.SetAttribute{
										MarkdownDescription: fieldGrantDescription,
										ElementType:         types.StringType,
										Computed:            true,
									},
									attrExcept: dsschema.SetAttribute{
										MarkdownDescription: fieldExceptDescription,
										ElementType:         types.StringType,
										Computed:            true,
									},
								},
							},
						},
						attrNames: dsschema.SetAttribute{
							MarkdownDescription: indicesNamesDescription,
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrPrivileges: dsschema.SetAttribute{
							MarkdownDescription: indicesPrivilegesDescription,
							ElementType:         types.StringType,
							Computed:            true,
						},
						attrQuery: dsschema.StringAttribute{
							MarkdownDescription: indicesQueryDescription,
							Computed:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						attrAllowRestrictedIndices: dsschema.BoolAttribute{
							MarkdownDescription: allowRestrictedIndicesDescription,
							Computed:            true,
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
	id, idDiags := esClient.ID(ctx, roleName)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	// Call GetRole
	result, roleDiags := elasticsearch.GetRole(ctx, esClient, roleName)
	diags.Append(roleDiags...)
	if diags.HasError() {
		return config, diags
	}

	// Not-found: return empty ID, keep name, no diagnostics
	if result == nil {
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
	diags.Append(config.fromAPIModel(ctx, result.Role, result.Global)...)
	if diags.HasError() {
		return config, diags
	}

	// Ensure name is set to the role name we looked up
	config.Name = types.StringValue(roleName)

	return config, diags
}

func (config *roleDataSourceModel) fromAPIModel(ctx context.Context, role *esTypes.Role, rawGlobal json.RawMessage) diag.Diagnostics {
	var diags diag.Diagnostics

	// Description
	config.Description = typeutils.StringishPointerValue(role.Description)

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
	if len(rawGlobal) > 0 {
		config.Global = jsontypes.NewNormalizedValue(string(rawGlobal))
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
				attrApplication: types.StringValue(app.Application),
				attrPrivileges:  privSet,
				attrResources:   resSet,
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

			queryVal, d := marshalIndexQuery(index.Query)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			allowRestrictedVal := types.BoolPointerValue(index.AllowRestrictedIndices)

			fieldSecList, d := buildFieldSecurityDSList(ctx, index.FieldSecurity)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			indexObj, d := types.ObjectValue(getIndexPermsDSAttrTypes(), map[string]attr.Value{
				attrFieldSecurity:          fieldSecList,
				attrNames:                  namesSet,
				attrPrivileges:             privSet,
				attrQuery:                  queryVal,
				attrAllowRestrictedIndices: allowRestrictedVal,
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

			queryVal, d := marshalIndexQuery(remoteIndex.Query)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			var allowRestrictedVal types.Bool
			if remoteIndex.AllowRestrictedIndices != nil {
				allowRestrictedVal = types.BoolValue(*remoteIndex.AllowRestrictedIndices)
			} else {
				allowRestrictedVal = types.BoolNull()
			}

			fieldSecList, d := buildFieldSecurityDSList(ctx, remoteIndex.FieldSecurity)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			remoteIndexObj, d := types.ObjectValue(getRemoteIndexPermsDSAttrTypes(), map[string]attr.Value{
				attrAllowRestrictedIndices: allowRestrictedVal,
				attrClusters:               clustersSet,
				attrFieldSecurity:          fieldSecList,
				attrNames:                  namesSet,
				attrPrivileges:             privSet,
				attrQuery:                  queryVal,
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

func buildFieldSecurityDSList(ctx context.Context, fs *esTypes.FieldSecurity) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := getFieldSecurityDSAttrTypes()
	if fs == nil {
		return types.ListValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{}), diags
	}
	grantSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(fs.Grant))
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}
	exceptSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(fs.Except))
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}
	obj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
		attrGrant:  grantSet,
		attrExcept: exceptSet,
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

// Data source attribute type helpers (mirror data source schema structure)
func getApplicationDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrApplication: types.StringType,
		attrPrivileges:  types.SetType{ElemType: types.StringType},
		attrResources:   types.SetType{ElemType: types.StringType},
	}
}

func getFieldSecurityDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrGrant:  types.SetType{ElemType: types.StringType},
		attrExcept: types.SetType{ElemType: types.StringType},
	}
}

func getIndexPermsDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrFieldSecurity: types.ListType{
			ElemType: types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()},
		},
		attrNames:                  types.SetType{ElemType: types.StringType},
		attrPrivileges:             types.SetType{ElemType: types.StringType},
		attrQuery:                  jsontypes.NormalizedType{},
		attrAllowRestrictedIndices: types.BoolType,
	}
}

func getRemoteIndexPermsDSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAllowRestrictedIndices: types.BoolType,
		attrClusters:               types.SetType{ElemType: types.StringType},
		attrFieldSecurity: types.ListType{
			ElemType: types.ObjectType{AttrTypes: getFieldSecurityDSAttrTypes()},
		},
		attrNames:      types.SetType{ElemType: types.StringType},
		attrPrivileges: types.SetType{ElemType: types.StringType},
		attrQuery:      jsontypes.NormalizedType{},
	}
}
