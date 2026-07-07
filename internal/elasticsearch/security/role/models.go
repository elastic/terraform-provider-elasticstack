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

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/clusterprivilege"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/indexprivilege"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type Data struct {
	entitycore.ResourceTimeoutsField
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
	AllowRestrictedIndices types.Bool `tfsdk:"allow_restricted_indices"`
	Clusters               types.Set  `tfsdk:"clusters"`
}

type FieldSecurityData struct {
	Grant  types.Set `tfsdk:"grant"`
	Except types.Set `tfsdk:"except"`
}

// toAPIModel converts the Terraform model to the API model
func (data *Data) toAPIModel(ctx context.Context) (*elasticsearch.Role, diag.Diagnostics) {
	var diags diag.Diagnostics
	var role elasticsearch.Role

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
		raw := []byte(data.Global.ValueString())
		var probe any
		if err := json.Unmarshal(raw, &probe); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing global JSON: %s", err))
			return nil, diags
		}
		role.Global = raw
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

			remoteEntry := estypes.RemoteIndicesPrivileges{
				Names:         idx.Names,
				Privileges:    idx.Privileges,
				Query:         idx.Query,
				FieldSecurity: idx.FieldSecurity,
				Clusters:      clusters,
			}
			if typeutils.IsKnown(remoteIdx.AllowRestrictedIndices) {
				remoteEntry.AllowRestrictedIndices = remoteIdx.AllowRestrictedIndices.ValueBoolPointer()
			}
			remoteIndices[i] = remoteEntry
		}
		role.RemoteIndices = remoteIndices
	}

	// Metadata
	if metadata := typeutils.NormalizedTypeToMap[any](data.Metadata, path.Root("metadata"), &diags); metadata != nil {
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

// fromAPIModel converts the API model to the Terraform model.
func (data *Data) fromAPIModel(ctx context.Context, role *elasticsearch.Role) diag.Diagnostics {
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
	if len(role.Global) > 0 {
		data.Global = customtypes.NewJSONWithDefaultsValue(string(role.Global), populateGlobalPrivilegesDefaults)
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

			queryVal, d := marshalIndexQuery(index.Query)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			allowRestrictedVal := types.BoolPointerValue(index.AllowRestrictedIndices)

			var fieldSecObj types.Object
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

				fieldSecObj, d = types.ObjectValue(getFieldSecurityAttrTypes(), map[string]attr.Value{
					attrGrant:  grantSet,
					attrExcept: exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecObj = types.ObjectNull(getFieldSecurityAttrTypes())
			}

			indexObj, d := types.ObjectValue(getIndexPermsAttrTypes(), map[string]attr.Value{
				attrFieldSecurity:          fieldSecObj,
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

			var fieldSecObj types.Object
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

				fieldSecObj, d = types.ObjectValue(getFieldSecurityAttrTypes(), map[string]attr.Value{
					attrGrant:  grantSet,
					attrExcept: exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecObj = types.ObjectNull(getFieldSecurityAttrTypes())
			}

			remoteIndexObj, d := types.ObjectValue(getRemoteIndexPermsAttrTypes(), map[string]attr.Value{
				attrAllowRestrictedIndices: allowRestrictedVal,
				attrClusters:               clustersSet,
				attrFieldSecurity:          fieldSecObj,
				attrQuery:                  queryVal,
				attrNames:                  namesSet,
				attrPrivileges:             privSet,
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

var _ entitycore.WithVersionRequirements = Data{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements] and declares
// conditional minimum Elasticsearch versions for configured description and remote_indices.
func (data Data) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement

	if typeutils.IsKnown(data.Description) {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion: *MinSupportedDescriptionVersion,
			ErrorMessage: fmt.Sprintf(
				"'description' is supported only for Elasticsearch v%s and above",
				MinSupportedDescriptionVersion.String(),
			),
		})
	}

	if typeutils.IsKnown(data.RemoteIndices) && len(data.RemoteIndices.Elements()) > 0 {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion: *MinSupportedRemoteIndicesVersion,
			ErrorMessage: fmt.Sprintf(
				"'remote_indices' is supported only for Elasticsearch v%s and above",
				MinSupportedRemoteIndicesVersion.String(),
			),
		})
	}

	return reqs, nil
}
