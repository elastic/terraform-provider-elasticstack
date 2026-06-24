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

package osquerysavedquery

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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
)

var (
	_ entitycore.KibanaResourceModel     = osquerySavedQueryModel{}
	_ entitycore.WithVersionRequirements = osquerySavedQueryModel{}
)

var (
	osquerySavedQueryMinVersion = version.Must(version.NewVersion("8.5.0"))

	attrEcsMappingField  = "field"
	attrEcsMappingValue  = "value"
	attrEcsMappingValues = "values"

	ecsMappingAttrTypes = map[string]attr.Type{
		attrEcsMappingField:  types.StringType,
		attrEcsMappingValue:  types.StringType,
		attrEcsMappingValues: types.SetType{ElemType: types.StringType},
	}
)

type osquerySavedQueryModel struct {
	entitycore.ResourceTimeoutsField
	entitycore.KibanaConnectionField

	ID            types.String `tfsdk:"id"`
	SavedObjectID types.String `tfsdk:"saved_object_id"`
	SavedQueryID  types.String `tfsdk:"saved_query_id"`
	SpaceID       types.String `tfsdk:"space_id"`
	Query         types.String `tfsdk:"query"`
	Description   types.String `tfsdk:"description"`
	Platform      types.Set    `tfsdk:"platform"`
	Interval      types.Int64  `tfsdk:"interval"`
	Version       types.String `tfsdk:"version"`
	Snapshot      types.Bool   `tfsdk:"snapshot"`
	Removed       types.Bool   `tfsdk:"removed"`
	EcsMapping    types.Map    `tfsdk:"ecs_mapping"`
}

type ecsMapping struct {
	Field  types.String `tfsdk:"field"`
	Value  types.String `tfsdk:"value"`
	Values types.Set    `tfsdk:"values"`
}

func getEcsMappingElemType() attr.Type {
	return types.ObjectType{AttrTypes: ecsMappingAttrTypes}
}

func (m osquerySavedQueryModel) GetID() types.String         { return m.ID }
func (m osquerySavedQueryModel) GetResourceID() types.String { return m.SavedQueryID }
func (m osquerySavedQueryModel) GetSpaceID() types.String    { return m.SpaceID }

func (m osquerySavedQueryModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *osquerySavedQueryMinVersion,
			ErrorMessage: fmt.Sprintf("Osquery saved queries require Elastic Stack v%s or later.", osquerySavedQueryMinVersion),
		},
	}, nil
}

func (m *osquerySavedQueryModel) populateFromCreateAPI(ctx context.Context, entity *kibanaoapi.OsquerySavedQueryCreateEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromCreateAPI(entity.Interval)
	version, versionDiags := versionFromCreateAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(versionDiags...)
	diags.Append(m.populateSharedFields(ctx, entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryModel) populateFromGetAPI(ctx context.Context, entity *kibanaoapi.OsquerySavedQueryGetEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromGetAPI(entity.Interval)
	version, versionDiags := versionFromGetAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(versionDiags...)
	diags.Append(m.populateSharedFields(ctx, entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryModel) populateFromUpdateAPI(ctx context.Context, entity *kibanaoapi.OsquerySavedQueryUpdateEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromUpdateAPI(entity.Interval)
	version := versionFromUpdateAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(m.populateSharedFields(ctx, entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryModel) populateSharedFields(
	ctx context.Context,
	savedQueryID kbapi.SecurityOsqueryAPISavedQueryId,
	savedObjectID string,
	query *kbapi.SecurityOsqueryAPIQuery,
	description *kbapi.SecurityOsqueryAPISavedQueryDescription,
	platform *kbapi.SecurityOsqueryAPIPlatform,
	ecsMappingAPI *kbapi.SecurityOsqueryAPIECSMapping,
	snapshot *kbapi.SecurityOsqueryAPISnapshot,
	removed *kbapi.SecurityOsqueryAPIRemoved,
	interval types.Int64,
	version types.String,
) diag.Diagnostics {
	var diags diag.Diagnostics

	m.setCompositeIdentity(savedQueryID)
	if savedObjectID == "" {
		m.SavedObjectID = types.StringNull()
	} else {
		m.SavedObjectID = types.StringValue(savedObjectID)
	}
	m.Query = typeutils.StringishPointerValue(query)
	m.Description = optionalStringPointerValue(description)
	m.Platform = platformSetFromAPI(platform)
	m.Interval = interval
	m.Version = version
	m.Snapshot = boolPointerValue(snapshot)
	m.Removed = boolPointerValue(removed)

	ecsMapping, ecsDiags := ecsMappingMapFromAPI(ctx, ecsMappingAPI)
	diags.Append(ecsDiags...)
	m.EcsMapping = ecsMapping

	return diags
}

func (m *osquerySavedQueryModel) setCompositeIdentity(savedQueryID kbapi.SecurityOsqueryAPISavedQueryId) {
	spaceID := compositeSpaceID(m.SpaceID)

	compID := clients.CompositeID{
		ClusterID:  spaceID,
		ResourceID: savedQueryID,
	}

	m.ID = types.StringValue(compID.String())
	m.SavedQueryID = types.StringValue(savedQueryID)

	// Populate computed space_id when absent, but preserve unknown plan/state values.
	if m.SpaceID.IsNull() || (typeutils.IsKnown(m.SpaceID) && m.SpaceID.ValueString() == "") {
		m.SpaceID = types.StringValue(spaceID)
	}
}

// compositeSpaceID returns the space segment for composite IDs. Unknown space_id
// falls back to clients.DefaultSpaceID for ID composition without overwriting unknown state.
func compositeSpaceID(spaceID types.String) string {
	if typeutils.IsKnown(spaceID) && spaceID.ValueString() != "" {
		return spaceID.ValueString()
	}

	return clients.DefaultSpaceID
}

func (e ecsMapping) toAPIType() (kbapi.SecurityOsqueryAPIECSMappingItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	fieldSet := typeutils.IsKnown(e.Field)
	valueSet := typeutils.IsKnown(e.Value)
	valuesSet := typeutils.IsKnown(e.Values)

	setCount := 0
	if fieldSet {
		setCount++
	}
	if valueSet {
		setCount++
	}
	if valuesSet {
		setCount++
	}

	if setCount != 1 {
		diags.AddError(
			"Invalid ecs_mapping element",
			"Exactly one of field, value, or values must be set per ecs_mapping element.",
		)
		return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
	}

	item := kbapi.SecurityOsqueryAPIECSMappingItem{}

	switch {
	case fieldSet:
		item.Field = e.Field.ValueStringPointer()
	case valueSet:
		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := value.FromSecurityOsqueryAPIECSMappingItemValue0(e.Value.ValueString()); err != nil {
			diags.AddError("Invalid ecs_mapping element", fmt.Sprintf("Failed to encode scalar value: %s", err))
			return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
		}
		item.Value = &value
	case valuesSet:
		var values []string
		for _, element := range e.Values.Elements() {
			if str, ok := element.(types.String); ok && typeutils.IsKnown(str) {
				values = append(values, str.ValueString())
			}
		}
		sort.Strings(values)

		var value kbapi.SecurityOsqueryAPIECSMappingItem_Value
		if err := value.FromSecurityOsqueryAPIECSMappingItemValue1(values); err != nil {
			diags.AddError("Invalid ecs_mapping element", fmt.Sprintf("Failed to encode array values: %s", err))
			return kbapi.SecurityOsqueryAPIECSMappingItem{}, diags
		}
		item.Value = &value
	}

	return item, diags
}

func ecsMappingFromAPIType(item kbapi.SecurityOsqueryAPIECSMappingItem) (ecsMapping, diag.Diagnostics) {
	result := ecsMapping{
		Field:  types.StringNull(),
		Value:  types.StringNull(),
		Values: types.SetNull(types.StringType),
	}

	if item.Value != nil {
		if scalar, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue0(); err == nil {
			result.Value = types.StringValue(scalar)
			return result, nil
		}

		if values, err := item.Value.AsSecurityOsqueryAPIECSMappingItemValue1(); err == nil {
			sorted := append([]string(nil), values...)
			sort.Strings(sorted)
			result.Values = stringSetValue(sorted)
			return result, nil
		}

		return result, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid ecs_mapping value", "API ecs_mapping value is neither string nor string array."),
		}
	}

	if item.Field != nil {
		result.Field = types.StringValue(*item.Field)
	}

	return result, nil
}

func ecsMappingMapFromAPI(_ context.Context, api *kbapi.SecurityOsqueryAPIECSMapping) (types.Map, diag.Diagnostics) {
	if api == nil || len(*api) == 0 {
		return types.MapNull(getEcsMappingElemType()), nil
	}

	elems := make(map[string]attr.Value, len(*api))
	var diags diag.Diagnostics
	for key, item := range *api {
		mapping, mappingDiags := ecsMappingFromAPIType(item)
		diags.Append(mappingDiags...)
		if diags.HasError() {
			return types.MapNull(getEcsMappingElemType()), diags
		}

		obj, objDiags := types.ObjectValue(ecsMappingAttrTypes, map[string]attr.Value{
			attrEcsMappingField:  mapping.Field,
			attrEcsMappingValue:  mapping.Value,
			attrEcsMappingValues: mapping.Values,
		})
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.MapNull(getEcsMappingElemType()), diags
		}
		elems[key] = obj
	}

	m, mapDiags := types.MapValue(getEcsMappingElemType(), elems)
	diags.Append(mapDiags...)
	return m, diags
}

func platformSetFromAPI(platform *kbapi.SecurityOsqueryAPIPlatform) types.Set {
	if platform == nil || strings.TrimSpace(*platform) == "" {
		return types.SetNull(types.StringType)
	}

	parts := strings.Split(*platform, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}

	sort.Strings(values)
	return stringSetValue(values)
}

func platformToAPI(ctx context.Context, platform types.Set) (*kbapi.SecurityOsqueryAPIPlatform, diag.Diagnostics) {
	if !typeutils.IsKnown(platform) || platform.IsNull() {
		return nil, nil
	}

	var diags diag.Diagnostics
	var values []string
	diags.Append(platform.ElementsAs(ctx, &values, false)...)
	if diags.HasError() {
		return nil, diags
	}

	if len(values) == 0 {
		return nil, diags
	}

	sort.Strings(values)
	joined := strings.Join(values, ",")
	return &joined, diags
}

func intervalFromCreateAPI(interval *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	if value, err := interval.AsSecurityOsqueryAPICreateSavedQueryResponseDataInterval0(); err == nil {
		return types.Int64Value(int64(value)), nil
	}

	if value, err := interval.AsSecurityOsqueryAPICreateSavedQueryResponseDataInterval1(); err == nil {
		return parseIntervalString(value)
	}

	return types.Int64Null(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid interval value", "Create response interval is neither integer nor string."),
	}
}

func intervalFromGetAPI(interval *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	if value, err := interval.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval0(); err == nil {
		return types.Int64Value(int64(value)), nil
	}

	if value, err := interval.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval1(); err == nil {
		return parseIntervalString(value)
	}

	return types.Int64Null(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid interval value", "Get response interval is neither integer nor string."),
	}
}

func intervalFromUpdateAPI(interval *kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	if value, err := interval.AsSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval0(); err == nil {
		return types.Int64Value(int64(value)), nil
	}

	if value, err := interval.AsSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval1(); err == nil {
		return parseIntervalString(value)
	}

	return types.Int64Null(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid interval value", "Update response interval is neither integer nor string."),
	}
}

func versionFromCreateAPI(version *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version) (types.String, diag.Diagnostics) {
	if version == nil {
		return types.StringNull(), nil
	}

	if value, err := version.AsSecurityOsqueryAPICreateSavedQueryResponseDataVersion1(); err == nil {
		return versionStringValue(value), nil
	}

	if value, err := version.AsSecurityOsqueryAPICreateSavedQueryResponseDataVersion0(); err == nil {
		return types.StringValue(strconv.Itoa(value)), nil
	}

	return types.StringNull(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid version value", "Create response version is neither integer nor string."),
	}
}

func versionFromGetAPI(version *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version) (types.String, diag.Diagnostics) {
	if version == nil {
		return types.StringNull(), nil
	}

	if value, err := version.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion1(); err == nil {
		return versionStringValue(value), nil
	}

	if value, err := version.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion0(); err == nil {
		return types.StringValue(strconv.Itoa(value)), nil
	}

	return types.StringNull(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid version value", "Get response version is neither integer nor string."),
	}
}

func versionFromUpdateAPI(version *string) types.String {
	if version == nil {
		return types.StringNull()
	}

	return versionStringValue(*version)
}

func versionStringValue(value string) types.String {
	if strings.TrimSpace(value) == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func optionalStringPointerValue[T ~string](value *T) types.String {
	if value == nil || strings.TrimSpace(string(*value)) == "" {
		return types.StringNull()
	}

	return types.StringValue(string(*value))
}

func parseIntervalString(value string) (types.Int64, diag.Diagnostics) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return types.Int64Null(), diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid interval value", fmt.Sprintf("Failed to parse interval %q as int64: %s", value, err)),
		}
	}

	return types.Int64Value(parsed), nil
}

func boolPointerValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}

	return types.BoolValue(*value)
}

func stringSetValue(values []string) types.Set {
	if len(values) == 0 {
		return types.SetNull(types.StringType)
	}

	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}

	return types.SetValueMust(types.StringType, elements)
}
