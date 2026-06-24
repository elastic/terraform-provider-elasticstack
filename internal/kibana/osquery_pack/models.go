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
	"sort"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type osqueryPackModel struct {
	entitycore.ResourceTimeoutsField
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

type queryModel struct {
	Query        types.String `tfsdk:"query"`
	Platform     types.Set    `tfsdk:"platform"`
	Version      types.String `tfsdk:"version"`
	Snapshot     types.Bool   `tfsdk:"snapshot"`
	Removed      types.Bool   `tfsdk:"removed"`
	SavedQueryID types.String `tfsdk:"saved_query_id"`
	EcsMapping   types.Map    `tfsdk:"ecs_mapping"`
}

type ecsMappingModel struct {
	Field  types.String `tfsdk:"field"`
	Value  types.String `tfsdk:"value"`
	Values types.Set    `tfsdk:"values"`
}

var (
	_ entitycore.KibanaResourceModel  = osqueryPackModel{}
	_ entitycore.WithVersionRequirements = osqueryPackModel{}

	osqueryPackMinVersion = version.Must(version.NewVersion("8.5.0"))
)

func (m osqueryPackModel) GetID() types.String         { return m.ID }
func (m osqueryPackModel) GetResourceID() types.String { return m.PackID }
func (m osqueryPackModel) GetSpaceID() types.String    { return m.SpaceID }

func (osqueryPackModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *osqueryPackMinVersion,
			ErrorMessage: fmt.Sprintf("Osquery packs require Elastic Stack v%s or later.", osqueryPackMinVersion),
		},
	}, nil
}

func (m *osqueryPackModel) populateFromAPI(ctx context.Context, spaceID string, data *kibanaoapi.OsqueryPackDetail) diag.Diagnostics {
	if data == nil {
		return nil
	}

	if spaceID == "" {
		spaceID = clients.DefaultSpaceID
	}

	var diags diag.Diagnostics

	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.SavedObjectId}).String())
	m.PackID = types.StringValue(data.SavedObjectId)
	m.SpaceID = types.StringValue(spaceID)
	m.Name = types.StringValue(string(data.Name))
	m.Description = types.StringPointerValue(descriptionString(data.Description))
	m.Enabled = types.BoolPointerValue(enabledBool(data.Enabled))

	if data.PolicyIds != nil && len(*data.PolicyIds) > 0 {
		policyIDs, d := types.ListValueFrom(ctx, types.StringType, *data.PolicyIds)
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

func descriptionString(v *kbapi.SecurityOsqueryAPIPackDescription) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func enabledBool(v *kbapi.SecurityOsqueryAPIEnabled) *bool {
	if v == nil {
		return nil
	}
	b := bool(*v)
	return &b
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
		diags.Append(q.fromAPIType(ctx, item)...)
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

func (m *queryModel) fromAPIType(ctx context.Context, item kbapi.SecurityOsqueryAPIObjectQueriesItem) diag.Diagnostics {
	var diags diag.Diagnostics

	if item.Query != nil {
		m.Query = types.StringValue(string(*item.Query))
	} else {
		m.Query = types.StringNull()
	}

	m.Platform = platformSetFromAPI(item.Platform)

	if item.Version != nil {
		m.Version = types.StringValue(string(*item.Version))
	} else {
		m.Version = types.StringNull()
	}

	m.Snapshot = types.BoolPointerValue(snapshotBool(item.Snapshot))
	m.Removed = types.BoolPointerValue(removedBool(item.Removed))

	if item.SavedQueryId != nil {
		m.SavedQueryID = types.StringValue(string(*item.SavedQueryId))
	} else {
		m.SavedQueryID = types.StringNull()
	}

	ecsMapping, d := ecsMappingMapFromAPI(ctx, item.EcsMapping)
	diags.Append(d...)
	m.EcsMapping = ecsMapping

	return diags
}

func (m queryModel) toAPIType(ctx context.Context) (kbapi.SecurityOsqueryAPIObjectQueriesItem, diag.Diagnostics) {
	var diags diag.Diagnostics
	item := kbapi.SecurityOsqueryAPIObjectQueriesItem{}

	if typeutils.IsKnown(m.Query) {
		q := kbapi.SecurityOsqueryAPIQuery(m.Query.ValueString())
		item.Query = &q
	}

	platform, d := platformCommaStringFromSet(ctx, m.Platform)
	diags.Append(d...)
	item.Platform = platform

	if typeutils.IsKnown(m.Version) {
		v := kbapi.SecurityOsqueryAPIVersion(m.Version.ValueString())
		item.Version = &v
	}

	item.Snapshot = m.Snapshot.ValueBoolPointer()
	item.Removed = m.Removed.ValueBoolPointer()

	if typeutils.IsKnown(m.SavedQueryID) {
		id := kbapi.SecurityOsqueryAPISavedQueryId(m.SavedQueryID.ValueString())
		item.SavedQueryId = &id
	}

	ecsMapping, d := ecsMappingMapToAPI(ctx, m.EcsMapping)
	diags.Append(d...)
	item.EcsMapping = ecsMapping

	return item, diags
}

func snapshotBool(v *kbapi.SecurityOsqueryAPISnapshot) *bool {
	if v == nil {
		return nil
	}
	b := bool(*v)
	return &b
}

func removedBool(v *kbapi.SecurityOsqueryAPIRemoved) *bool {
	if v == nil {
		return nil
	}
	b := bool(*v)
	return &b
}

func platformSetFromAPI(platform *kbapi.SecurityOsqueryAPIPlatform) types.Set {
	if platform == nil || *platform == "" {
		return types.SetNull(types.StringType)
	}

	parts := strings.Split(string(*platform), ",")
	platforms := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			platforms = append(platforms, part)
		}
	}

	if len(platforms) == 0 {
		return types.SetNull(types.StringType)
	}

	sort.Strings(platforms)
	set, _ := types.SetValueFrom(context.Background(), types.StringType, platforms)
	return set
}

func platformCommaStringFromSet(ctx context.Context, platform types.Set) (*kbapi.SecurityOsqueryAPIPlatform, diag.Diagnostics) {
	if !typeutils.IsKnown(platform) || platform.IsNull() {
		return nil, nil
	}

	var platforms []string
	diags := platform.ElementsAs(ctx, &platforms, false)
	if diags.HasError() {
		return nil, diags
	}

	if len(platforms) == 0 {
		return nil, nil
	}

	sort.Strings(platforms)
	s := kbapi.SecurityOsqueryAPIPlatform(strings.Join(platforms, ","))
	return &s, nil
}

func ecsMappingMapFromAPI(ctx context.Context, mapping *kbapi.SecurityOsqueryAPIECSMapping) (types.Map, diag.Diagnostics) {
	if mapping == nil || len(*mapping) == 0 {
		return types.MapNull(ecsMappingMapElemType()), nil
	}

	var diags diag.Diagnostics
	elems := make(map[string]attr.Value, len(*mapping))

	for key, item := range *mapping {
		var m ecsMappingModel
		m.fromAPIType(item)

		obj, d := types.ObjectValueFrom(ctx, ecsMappingAttrTypes(), m)
		diags.Append(d...)
		if diags.HasError() {
			return types.MapNull(ecsMappingMapElemType()), diags
		}
		elems[key] = obj
	}

	result, d := types.MapValue(ecsMappingMapElemType(), elems)
	diags.Append(d...)
	return result, diags
}

func ecsMappingMapToAPI(ctx context.Context, mapping types.Map) (*kbapi.SecurityOsqueryAPIECSMapping, diag.Diagnostics) {
	if !typeutils.IsKnown(mapping) || mapping.IsNull() {
		return nil, nil
	}

	var diags diag.Diagnostics
	elems := make(kbapi.SecurityOsqueryAPIECSMapping, len(mapping.Elements()))

	for key, av := range mapping.Elements() {
		var m ecsMappingModel
		d := av.(basetypes.ObjectValue).As(ctx, &m, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		item, d := m.toAPIType()
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		elems[key] = item
	}

	if len(elems) == 0 {
		return nil, nil
	}

	return &elems, diags
}

func (m *ecsMappingModel) fromAPIType(item kbapi.SecurityOsqueryAPIECSMappingItem) {
	m.Field = types.StringNull()
	m.Value = types.StringNull()
	m.Values = types.SetNull(types.StringType)

	if item.Field != nil {
		m.Field = types.StringValue(*item.Field)
		return
	}

	if item.Value == nil {
		return
	}

	if str, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue0(); err == nil {
		m.Value = types.StringValue(str)
		return
	}

	if arr, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue1(); err == nil {
		set, _ := types.SetValueFrom(context.Background(), types.StringType, arr)
		m.Values = set
	}
}

func (m ecsMappingModel) toAPIType() (kbapi.SecurityOsqueryAPIECSMappingItem, diag.Diagnostics) {
	item := kbapi.SecurityOsqueryAPIECSMappingItem{}

	if typeutils.IsKnown(m.Field) {
		item.Field = m.Field.ValueStringPointer()
		return item, nil
	}

	if typeutils.IsKnown(m.Value) {
		var val kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := val.FromSecurityOsqueryAPIECSMappingItemValue0(m.Value.ValueString()); err != nil {
			return item, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to encode ECS mapping value", err.Error()),
			}
		}
		item.Value = &val
		return item, nil
	}

	if typeutils.IsKnown(m.Values) {
		var values []string
		diags := m.Values.ElementsAs(context.Background(), &values, false)
		if diags.HasError() {
			return item, diags
		}

		var val kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := val.FromSecurityOsqueryAPIECSMappingItemValue1(values); err != nil {
			return item, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to encode ECS mapping values", err.Error()),
			}
		}
		item.Value = &val
		return item, nil
	}

	return item, nil
}

func queryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"query":          types.StringType,
		"platform":       types.SetType{ElemType: types.StringType},
		"version":        types.StringType,
		"snapshot":       types.BoolType,
		"removed":        types.BoolType,
		"saved_query_id": types.StringType,
		"ecs_mapping":    types.MapType{ElemType: ecsMappingMapElemType()},
	}
}

func queryMapElemType() attr.Type {
	return types.ObjectType{AttrTypes: queryAttrTypes()}
}

func ecsMappingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"field":  types.StringType,
		"value":  types.StringType,
		"values": types.SetType{ElemType: types.StringType},
	}
}

func ecsMappingMapElemType() attr.Type {
	return types.ObjectType{AttrTypes: ecsMappingAttrTypes()}
}
