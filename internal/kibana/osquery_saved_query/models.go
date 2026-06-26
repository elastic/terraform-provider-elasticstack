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
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ entitycore.KibanaResourceModel     = osquerySavedQueryModel{}
	_ entitycore.WithVersionRequirements = osquerySavedQueryModel{}
)

var (
	attrEcsMappingField  = "field"
	attrEcsMappingValue  = "value"
	attrEcsMappingValues = "values"

	ecsMappingAttrTypes = osquery.ECSMappingAttrTypes
)

type osquerySavedQueryBaseModel struct {
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

type osquerySavedQueryModel struct {
	entitycore.ResourceTimeoutsField
	osquerySavedQueryBaseModel
}

type ecsMapping osquery.ECSMapping

func getEcsMappingElemType() attr.Type {
	return osquery.ECSMappingElemType()
}

func (m osquerySavedQueryBaseModel) GetID() types.String         { return m.ID }
func (m osquerySavedQueryBaseModel) GetResourceID() types.String { return m.SavedQueryID }
func (m osquerySavedQueryBaseModel) GetSpaceID() types.String    { return m.SpaceID }

func (m osquerySavedQueryBaseModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *MinSupportedVersion,
			ErrorMessage: fmt.Sprintf("Osquery saved queries require Elastic Stack v%s or later.", MinSupportedVersion),
		},
	}, nil
}

func (m *osquerySavedQueryBaseModel) populateFromCreateAPI(_ context.Context, entity *kibanaoapi.OsquerySavedQueryCreateEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromCreateAPI(entity.Interval)
	version, versionDiags := versionFromCreateAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(versionDiags...)
	diags.Append(m.populateSharedFields(entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryBaseModel) populateFromGetAPI(_ context.Context, entity *kibanaoapi.OsquerySavedQueryGetEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromGetAPI(entity.Interval)
	version, versionDiags := versionFromGetAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(versionDiags...)
	diags.Append(m.populateSharedFields(entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryBaseModel) populateFromUpdateAPI(_ context.Context, entity *kibanaoapi.OsquerySavedQueryUpdateEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	interval, intervalDiags := intervalFromUpdateAPI(entity.Interval)
	version := versionFromUpdateAPI(entity.Version)

	diags := diag.Diagnostics{}
	diags.Append(intervalDiags...)
	diags.Append(m.populateSharedFields(entity.ID, entity.SavedObjectID, entity.Query, entity.Description, entity.Platform, entity.EcsMapping, entity.Snapshot, entity.Removed, interval, version)...)

	return diags
}

func (m *osquerySavedQueryBaseModel) populateSharedFields(
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
	m.Description = typeutils.TrimmedStringishPointerValue(description)
	m.Platform = osquery.PlatformSetFromAPI(platform)
	m.Interval = interval
	m.Version = version
	m.Snapshot = types.BoolPointerValue(snapshot)
	m.Removed = types.BoolPointerValue(removed)

	ecsMapping, ecsDiags := osquery.ECSMappingMapFromAPI(ecsMappingAPI)
	diags.Append(ecsDiags...)
	m.EcsMapping = ecsMapping

	return diags
}

func (m *osquerySavedQueryBaseModel) setCompositeIdentity(savedQueryID kbapi.SecurityOsqueryAPISavedQueryId) {
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
	if typeutils.IsKnown(spaceID) {
		return clients.EffectiveSpaceID(spaceID.ValueString())
	}

	return clients.DefaultSpaceID
}

func knownSavedObjectID(savedObjectID types.String) (string, bool) {
	if typeutils.IsKnown(savedObjectID) && savedObjectID.ValueString() != "" {
		return savedObjectID.ValueString(), true
	}

	return "", false
}

func (e ecsMapping) toAPIType() (kbapi.SecurityOsqueryAPIECSMappingItem, diag.Diagnostics) {
	return osquery.ECSMapping(e).ToAPIType()
}

func ecsMappingFromAPIType(item kbapi.SecurityOsqueryAPIECSMappingItem) (ecsMapping, diag.Diagnostics) {
	result, diags := osquery.ECSMappingFromAPIType("", item)
	return ecsMapping(result), diags
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

func parseIntervalString(value string) (types.Int64, diag.Diagnostics) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return types.Int64Null(), diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid interval value", fmt.Sprintf("Failed to parse interval %q as int64: %s", value, err)),
		}
	}

	return types.Int64Value(parsed), nil
}

func stringSetValue(values []string) types.Set {
	return osquery.StringSetValue(values)
}
