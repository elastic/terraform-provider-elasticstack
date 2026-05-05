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

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/clusterprivilege"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/indexprivilege"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Data struct {
	ID                      types.String                                      `tfsdk:"id"`
	ElasticsearchConnection types.List                                        `tfsdk:"elasticsearch_connection"`
	Name                    types.String                                      `tfsdk:"name"`
	Description             types.String                                      `tfsdk:"description"`
	Applications            types.Set                                         `tfsdk:"applications"`
	Global                  customtypes.JSONWithDefaultsValue[map[string]any] `tfsdk:"global"`
	Cluster                 types.Set                                         `tfsdk:"cluster"`
	Indices                 types.Set                                         `tfsdk:"indices"`
	RemoteIndices           types.Set                                         `tfsdk:"remote_indices"`
	Metadata                jsontypes.Normalized                              `tfsdk:"metadata"`
	RunAs                   types.Set                                         `tfsdk:"run_as"`
}

type ApplicationData struct {
	Application types.String `tfsdk:"application"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Resources   types.Set    `tfsdk:"resources"`
}

type CommonIndexPermsData struct {
	FieldSecurity types.Object         `tfsdk:"field_security"`
	Names         types.Set            `tfsdk:"names"`
	Privileges    types.Set            `tfsdk:"privileges"`
	Query         jsontypes.Normalized `tfsdk:"query"`
}

type IndexPermsData struct {
	CommonIndexPermsData
	AllowRestrictedIndices types.Bool `tfsdk:"allow_restricted_indices"`
}

type RemoteIndexPermsData struct {
	CommonIndexPermsData
	Clusters types.Set `tfsdk:"clusters"`
}

type FieldSecurityData struct {
	Grant  types.Set `tfsdk:"grant"`
	Except types.Set `tfsdk:"except"`
}

// toAPIModel converts the Terraform model to the API model
func (data *Data) toAPIModel(ctx context.Context) (*estypes.Role, diag.Diagnostics) {
	var diags diag.Diagnostics
	var role estypes.Role

	// Description
	if typeutils.IsKnown(data.Description) {
		description := data.Description.ValueString()
		role.Description = &description
	}

	// Applications
	if typeutils.IsKnown(data.Applications) {
		var applicationsList []ApplicationData
		diags.Append(data.Applications.ElementsAs(ctx, &applicationsList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		applications := make([]estypes.ApplicationPrivileges, len(applicationsList))
		for i, app := range applicationsList {
			var privileges, resources []string
			diags.Append(app.Privileges.ElementsAs(ctx, &privileges, false)...)
			diags.Append(app.Resources.ElementsAs(ctx, &resources, false)...)
			if diags.HasError() {
				return nil, diags
			}

			applications[i] = estypes.ApplicationPrivileges{
				Application: app.Application.ValueString(),
				Privileges:  privileges,
				Resources:   resources,
			}
		}
		role.Applications = applications
	}

	// Global
	if typeutils.IsKnown(data.Global) {
		var global map[string]map[string]map[string][]string
		if err := json.Unmarshal([]byte(data.Global.ValueString()), &global); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing global JSON: %s", err))
			return nil, diags
		}
		role.Global = global
	}

	// Cluster
	if typeutils.IsKnown(data.Cluster) {
		var cluster []string
		diags.Append(data.Cluster.ElementsAs(ctx, &cluster, false)...)
		if diags.HasError() {
			return nil, diags
		}
		role.Cluster = make([]clusterprivilege.ClusterPrivilege, len(cluster))
		for i, s := range cluster {
			role.Cluster[i] = clusterprivilege.ClusterPrivilege{Name: s}
		}
	}

	// Indices
	if typeutils.IsKnown(data.Indices) {
		var indicesList []IndexPermsData
		diags.Append(data.Indices.ElementsAs(ctx, &indicesList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		indices := make([]estypes.IndicesPrivileges, len(indicesList))
		for i, idx := range indicesList {
			newIndex, d := indexPermissionsToAPIModel(ctx, idx.CommonIndexPermsData)
			if d.HasError() {
				return nil, d
			}

			if typeutils.IsKnown(idx.AllowRestrictedIndices) {
				newIndex.AllowRestrictedIndices = idx.AllowRestrictedIndices.ValueBoolPointer()
			}
			indices[i] = newIndex
		}
		role.Indices = indices
	}

	// Remote Indices
	if typeutils.IsKnown(data.RemoteIndices) {
		var remoteIndicesList []RemoteIndexPermsData
		diags.Append(data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		remoteIndices := make([]estypes.RemoteIndicesPrivileges, len(remoteIndicesList))
		for i, remoteIdx := range remoteIndicesList {
			idx, d := indexPermissionsToAPIModel(ctx, remoteIdx.CommonIndexPermsData)
			if d.HasError() {
				return nil, d
			}
			var clusters []string
			diags.Append(remoteIdx.Clusters.ElementsAs(ctx, &clusters, false)...)
			if diags.HasError() {
				return nil, diags
			}

			remoteIndices[i] = estypes.RemoteIndicesPrivileges{
				Names:                  idx.Names,
				Privileges:             idx.Privileges,
				Query:                  idx.Query,
				FieldSecurity:          idx.FieldSecurity,
				AllowRestrictedIndices: idx.AllowRestrictedIndices,
				Clusters:               clusters,
			}
		}
		role.RemoteIndices = remoteIndices
	}

	// Metadata
	if typeutils.IsKnown(data.Metadata) {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing metadata JSON: %s", err))
			return nil, diags
		}
		role.Metadata = make(estypes.Metadata, len(metadata))
		for k, v := range metadata {
			raw, err := json.Marshal(v)
			if err != nil {
				diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata value: %s", err))
				return nil, diags
			}
			role.Metadata[k] = raw
		}
	}

	// Run As
	if typeutils.IsKnown(data.RunAs) {
		var runAs []string
		diags.Append(data.RunAs.ElementsAs(ctx, &runAs, false)...)
		if diags.HasError() {
			return nil, diags
		}
		role.RunAs = runAs
	}

	return &role, diags
}

func indexPermissionsToAPIModel(ctx context.Context, data CommonIndexPermsData) (estypes.IndicesPrivileges, diag.Diagnostics) {
	var names, privileges []string
	diags := data.Names.ElementsAs(ctx, &names, false)
	diags.Append(data.Privileges.ElementsAs(ctx, &privileges, false)...)
	if diags.HasError() {
		return estypes.IndicesPrivileges{}, diags
	}

	newIndex := estypes.IndicesPrivileges{
		Names: names,
	}
	if len(privileges) > 0 {
		newIndex.Privileges = make([]indexprivilege.IndexPrivilege, len(privileges))
		for i, p := range privileges {
			newIndex.Privileges[i] = indexprivilege.IndexPrivilege{Name: p}
		}
	}

	if typeutils.IsKnown(data.Query) {
		query := data.Query.ValueString()
		newIndex.Query = &query
	}

	// Field Security
	if typeutils.IsKnown(data.FieldSecurity) {
		fieldSec, d := fieldSecurityToAPIModel(ctx, data.FieldSecurity)
		if d.HasError() {
			return estypes.IndicesPrivileges{}, d
		}
		newIndex.FieldSecurity = fieldSec
	}

	return newIndex, diags
}

func fieldSecurityToAPIModel(ctx context.Context, data types.Object) (*estypes.FieldSecurity, diag.Diagnostics) {
	var fieldSec FieldSecurityData
	diags := data.As(ctx, &fieldSec, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	fieldSecurity := estypes.FieldSecurity{}
	if typeutils.IsKnown(fieldSec.Grant) {
		var grants []string
		diags.Append(fieldSec.Grant.ElementsAs(ctx, &grants, false)...)
		if diags.HasError() {
			return nil, diags
		}
		fieldSecurity.Grant = grants
	}

	if typeutils.IsKnown(fieldSec.Except) {
		var excepts []string
		diags.Append(fieldSec.Except.ElementsAs(ctx, &excepts, false)...)
		if diags.HasError() {
			return nil, diags
		}
		fieldSecurity.Except = excepts
	}
	return &fieldSecurity, diags
}

// fromAPIModel converts the API model to the Terraform model
func (data *Data) fromAPIModel(ctx context.Context, role *estypes.Role) diag.Diagnostics {
	var diags diag.Diagnostics
	// Preserve original null values for optional attributes to distinguish between:
	// - User doesn't set attribute (null) - should remain null even if API returns empty array
	// - User explicitly sets empty array ([]) - should become empty set
	originalCluster := data.Cluster
	originalRunAs := data.RunAs
	originalDescription := data.Description

	// Description
	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	} else {
		// If the API omits/returns null for description, preserve a configured empty string ("")
		// to avoid post-apply state consistency issues.
		switch {
		case typeutils.IsKnown(originalDescription) && originalDescription.ValueString() == "":
			data.Description = originalDescription
		default:
			data.Description = types.StringNull()
		}
	}

	// Applications
	if len(role.Applications) > 0 {
		appElements := make([]attr.Value, len(role.Applications))
		for i, app := range role.Applications {
			privSet, d := types.SetValueFrom(ctx, types.StringType, app.Privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			resSet, d := types.SetValueFrom(ctx, types.StringType, app.Resources)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			appObj, d := types.ObjectValue(getApplicationAttrTypes(), map[string]attr.Value{
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

		appSet, d := types.SetValue(types.ObjectType{AttrTypes: getApplicationAttrTypes()}, appElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Applications = appSet
	} else {
		data.Applications = types.SetNull(types.ObjectType{AttrTypes: getApplicationAttrTypes()})
	}

	// Cluster
	clusterStrings := make([]string, len(role.Cluster))
	for i, cp := range role.Cluster {
		clusterStrings[i] = cp.String()
	}
	var clusterDiags diag.Diagnostics
	data.Cluster, clusterDiags = typeutils.NonEmptySetOrDefault(ctx, originalCluster, types.StringType, clusterStrings)
	diags.Append(clusterDiags...)
	if diags.HasError() {
		return diags
	}

	// Global
	if role.Global != nil {
		global, err := json.Marshal(role.Global)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling global JSON: %s", err))
			return diags
		}
		data.Global = customtypes.NewJSONWithDefaultsValue(string(global), populateGlobalPrivilegesDefaults)
	} else {
		data.Global = customtypes.NewJSONWithDefaultsNull(populateGlobalPrivilegesDefaults)
	}

	// Indices
	if len(role.Indices) > 0 {
		indicesElements := make([]attr.Value, len(role.Indices))
		for i, index := range role.Indices {
			namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
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

			var fieldSecObj types.Object
			if index.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, schemautil.NonNilSlice(index.FieldSecurity.Grant))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, schemautil.NonNilSlice(index.FieldSecurity.Except))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d = types.ObjectValue(getFieldSecurityAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecObj = types.ObjectNull(getFieldSecurityAttrTypes())
			}

			indexObj, d := types.ObjectValue(getIndexPermsAttrTypes(), map[string]attr.Value{
				"field_security":           fieldSecObj,
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

		indicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getIndexPermsAttrTypes()}, indicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Indices = indicesSet
	} else {
		data.Indices = types.SetNull(types.ObjectType{AttrTypes: getIndexPermsAttrTypes()})
	}

	// Remote Indices
	if len(role.RemoteIndices) > 0 {
		remoteIndicesElements := make([]attr.Value, len(role.RemoteIndices))
		for i, remoteIndex := range role.RemoteIndices {
			clustersSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.Clusters)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			namesSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.Names)
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

			var fieldSecObj types.Object
			if remoteIndex.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, schemautil.NonNilSlice(remoteIndex.FieldSecurity.Grant))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, schemautil.NonNilSlice(remoteIndex.FieldSecurity.Except))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d = types.ObjectValue(getRemoteFieldSecurityAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecObj = types.ObjectNull(getRemoteFieldSecurityAttrTypes())
			}

			remoteIndexObj, d := types.ObjectValue(getRemoteIndexPermsAttrTypes(), map[string]attr.Value{
				"clusters":       clustersSet,
				"field_security": fieldSecObj,
				"query":          queryVal,
				"names":          namesSet,
				"privileges":     privSet,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			remoteIndicesElements[i] = remoteIndexObj
		}

		remoteIndicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getRemoteIndexPermsAttrTypes()}, remoteIndicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.RemoteIndices = remoteIndicesSet
	} else {
		data.RemoteIndices = types.SetNull(types.ObjectType{AttrTypes: getRemoteIndexPermsAttrTypes()})
	}

	// Metadata
	if role.Metadata != nil {
		metadata, err := json.Marshal(role.Metadata)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
			return diags
		}
		data.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	// Run As
	var runAsDiags diag.Diagnostics
	data.RunAs, runAsDiags = typeutils.NonEmptySetOrDefault(ctx, originalRunAs, types.StringType, role.RunAs)
	diags.Append(runAsDiags...)

	return diags
}

func (data Data) GetID() types.String                    { return data.ID }
func (data Data) GetResourceID() types.String            { return data.Name }
func (data Data) GetElasticsearchConnection() types.List { return data.ElasticsearchConnection }
