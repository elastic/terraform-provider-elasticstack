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

package osquerypack

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type osqueryPackBaseModel struct {
	entitycore.KibanaConnectionField
	ID          types.String `tfsdk:"id"`
	PackID      types.String `tfsdk:"pack_id"`
	SpaceID     types.String `tfsdk:"space_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	PolicyIDs   types.List   `tfsdk:"policy_ids"`
	Shards      types.Map    `tfsdk:"shards"`
	Queries     types.Map    `tfsdk:"queries"`
}

type osqueryPackModel struct {
	entitycore.ResourceTimeoutsField
	osqueryPackBaseModel
}

type queryModel struct {
	Query        types.String `tfsdk:"query"`
	Platform     types.Set    `tfsdk:"platform"`
	Version      types.String `tfsdk:"version"`
	Snapshot     types.Bool   `tfsdk:"snapshot"`
	Removed      types.Bool   `tfsdk:"removed"`
	SavedQueryID types.String `tfsdk:"saved_query_id"`
	EcsMapping   types.Map    `tfsdk:"ecs_mapping"`
}

type ecsMappingModel = osquery.ECSMapping

var (
	_ entitycore.KibanaResourceModel     = osqueryPackModel{}
	_ entitycore.WithVersionRequirements = osqueryPackModel{}

	osqueryPackMinVersion = version.Must(version.NewVersion("8.5.0"))
)

func (m osqueryPackBaseModel) GetID() types.String         { return m.ID }
func (m osqueryPackBaseModel) GetResourceID() types.String { return m.PackID }
func (m osqueryPackBaseModel) GetSpaceID() types.String    { return m.SpaceID }

func (m *osqueryPackBaseModel) setCompositeIdentity(spaceID, packID string) {
	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: packID}).String())
	m.PackID = types.StringValue(packID)
	m.SpaceID = types.StringValue(spaceID)
}

func (osqueryPackBaseModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *osqueryPackMinVersion,
			ErrorMessage: fmt.Sprintf("Osquery packs require Elastic Stack v%s or later.", osqueryPackMinVersion),
		},
	}, nil
}

func (m *osqueryPackBaseModel) populateFromAPI(ctx context.Context, spaceID string, data *kibanaoapi.OsqueryPackDetail) diag.Diagnostics {
	if data == nil {
		return nil
	}

	if spaceID == "" {
		spaceID = clients.DefaultSpaceID
	}

	var diags diag.Diagnostics

	m.setCompositeIdentity(spaceID, data.SavedObjectID)
	m.Name = types.StringValue(data.Name)
	m.Description = typeutils.StringishPointerValue(data.Description)
	m.Enabled = types.BoolPointerValue(data.Enabled)

	if data.PolicyIDs != nil && len(*data.PolicyIDs) > 0 {
		policyIDs, d := types.ListValueFrom(ctx, types.StringType, *data.PolicyIDs)
		diags.Append(d...)
		m.PolicyIDs = policyIDs
	} else {
		m.PolicyIDs = types.ListNull(types.StringType)
	}

	m.Shards = shardsMapFromAPI(data.Shards)

	queries, d := queriesMapFromAPI(ctx, data.Queries)
	diags.Append(d...)
	m.Queries = queries

	return diags
}

func shardsMapFromAPI(shards kibanaoapi.OsqueryPackShards) types.Map {
	if len(shards) == 0 {
		return types.MapNull(types.Float64Type)
	}

	elems := make(map[string]attr.Value, len(shards))
	for policyID, percent := range shards {
		elems[policyID] = types.Float64Value(percent)
	}

	result, _ := types.MapValue(types.Float64Type, elems)
	return result
}

func queriesMapFromAPI(ctx context.Context, queries *kbapi.SecurityOsqueryAPIObjectQueries) (types.Map, diag.Diagnostics) {
	if queries == nil || len(*queries) == 0 {
		return types.MapNull(queryMapElemType()), nil
	}

	var diags diag.Diagnostics
	elems := make(map[string]attr.Value, len(*queries))

	for name, item := range *queries {
		var q queryModel
		diags.Append(q.fromAPIType(item)...)
		if diags.HasError() {
			return types.MapNull(queryMapElemType()), diags
		}

		obj, d := types.ObjectValueFrom(ctx, queryAttrTypes(), q)
		diags.Append(d...)
		if diags.HasError() {
			return types.MapNull(queryMapElemType()), diags
		}
		elems[name] = obj
	}

	result, d := types.MapValue(queryMapElemType(), elems)
	diags.Append(d...)
	return result, diags
}

func (m *queryModel) fromAPIType(item kbapi.SecurityOsqueryAPIObjectQueriesItem) diag.Diagnostics {
	var diags diag.Diagnostics

	if item.Query != nil {
		m.Query = types.StringValue(*item.Query)
	} else {
		m.Query = types.StringNull()
	}

	m.Platform = osquery.PlatformSetFromAPI(item.Platform)

	if item.Version != nil {
		m.Version = types.StringValue(*item.Version)
	} else {
		m.Version = types.StringNull()
	}

	m.Snapshot = types.BoolPointerValue(item.Snapshot)
	m.Removed = types.BoolPointerValue(item.Removed)

	if item.SavedQueryId != nil {
		m.SavedQueryID = types.StringValue(*item.SavedQueryId)
	} else {
		m.SavedQueryID = types.StringNull()
	}

	ecsMapping, d := osquery.ECSMappingMapFromAPI(item.EcsMapping)
	diags.Append(d...)
	m.EcsMapping = ecsMapping

	return diags
}

func (m queryModel) toAPIType(ctx context.Context) (kbapi.SecurityOsqueryAPIObjectQueriesItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	item := kbapi.SecurityOsqueryAPIObjectQueriesItem{}

	if typeutils.IsKnown(m.Query) {
		q := m.Query.ValueString()
		item.Query = &q
	}

	platform, d := osquery.PlatformToAPI(ctx, m.Platform)
	diags.Append(d...)
	item.Platform = platform

	if typeutils.IsKnown(m.Version) {
		v := m.Version.ValueString()
		item.Version = &v
	}

	item.Snapshot = m.Snapshot.ValueBoolPointer()
	item.Removed = m.Removed.ValueBoolPointer()

	if typeutils.IsKnown(m.SavedQueryID) {
		id := m.SavedQueryID.ValueString()
		item.SavedQueryId = &id
	}

	ecsMapping, d := osquery.ECSMappingMapToAPI(ctx, m.EcsMapping)
	diags.Append(d...)
	item.EcsMapping = ecsMapping

	return item, diags
}

func platformSetFromAPI(_ context.Context, platform *kbapi.SecurityOsqueryAPIPlatform) types.Set {
	return osquery.PlatformSetFromAPI(platform)
}

func platformCommaStringFromSet(ctx context.Context, platform types.Set) (*kbapi.SecurityOsqueryAPIPlatform, diag.Diagnostics) {
	return osquery.PlatformToAPI(ctx, platform)
}

func ecsMappingMapFromAPI(_ context.Context, mapping *kbapi.SecurityOsqueryAPIECSMapping) (types.Map, diag.Diagnostics) {
	return osquery.ECSMappingMapFromAPI(mapping)
}

func ecsMappingMapToAPI(ctx context.Context, mapping types.Map) (*kbapi.SecurityOsqueryAPIECSMapping, diag.Diagnostics) {
	return osquery.ECSMappingMapToAPI(ctx, mapping)
}

func queryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrQuery:        types.StringType,
		attrPlatform:     types.SetType{ElemType: types.StringType},
		attrVersion:      types.StringType,
		attrSnapshot:     types.BoolType,
		attrRemoved:      types.BoolType,
		attrSavedQueryID: types.StringType,
		attrEcsMapping:   types.MapType{ElemType: osquery.ECSMappingElemType()},
	}
}

func queryMapElemType() attr.Type {
	return types.ObjectType{AttrTypes: queryAttrTypes()}
}

func ecsMappingAttrTypes() map[string]attr.Type {
	return osquery.ECSMappingAttrTypes
}

func ecsMappingMapElemType() attr.Type {
	return osquery.ECSMappingElemType()
}
