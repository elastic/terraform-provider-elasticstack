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
	entitycore.SpaceIDField
	Query       types.String `tfsdk:"query"`
	Description types.String `tfsdk:"description"`
	Platform    types.Set    `tfsdk:"platform"`
	Interval    types.Int64  `tfsdk:"interval"`
	Version     types.String `tfsdk:"version"`
	Snapshot    types.Bool   `tfsdk:"snapshot"`
	Removed     types.Bool   `tfsdk:"removed"`
	EcsMapping  types.Map    `tfsdk:"ecs_mapping"`
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

func (m osquerySavedQueryBaseModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return entitycore.SingleVersionRequirement(
		*MinSupportedVersion,
		fmt.Sprintf("Osquery saved queries require Elastic Stack v%s or later.", MinSupportedVersion),
	), nil
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
	spaceID := clients.EffectiveSpaceIDFromValue(m.SpaceID)

	id, space := clients.CompositeIdentityValues(spaceID, savedQueryID)
	m.ID = id
	m.SavedQueryID = types.StringValue(savedQueryID)

	// Populate computed space_id when absent, but preserve unknown plan/state values.
	if m.SpaceID.IsNull() || (typeutils.IsKnown(m.SpaceID) && m.SpaceID.ValueString() == "") {
		m.SpaceID = space
	}
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

func intervalFromCreateAPI(interval *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	return intervalFromOneOf(interval.AsSecurityOsqueryAPICreateSavedQueryResponseDataInterval0, interval.AsSecurityOsqueryAPICreateSavedQueryResponseDataInterval1, "Create")
}

func intervalFromGetAPI(interval *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	return intervalFromOneOf(interval.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval0, interval.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval1, "Get")
}

func intervalFromUpdateAPI(interval *kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse_Data_Interval) (types.Int64, diag.Diagnostics) {
	if interval == nil {
		return types.Int64Null(), nil
	}

	return intervalFromOneOf(interval.AsSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval0, interval.AsSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval1, "Update")
}

// intervalFromOneOf unwraps a generated kbapi oneOf(int, string) interval response into types.Int64.
// asInt and asString are the generated type's AsXxx0/AsXxx1 accessors; context labels the error message.
func intervalFromOneOf(asInt func() (int, error), asString func() (string, error), context string) (types.Int64, diag.Diagnostics) {
	if value, err := asInt(); err == nil {
		return types.Int64Value(int64(value)), nil
	}

	if value, err := asString(); err == nil {
		return parseIntervalString(value)
	}

	return types.Int64Null(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid interval value", fmt.Sprintf("%s response interval is neither integer nor string.", context)),
	}
}

func versionFromCreateAPI(version *kbapi.SecurityOsqueryAPICreateSavedQueryResponse_Data_Version) (types.String, diag.Diagnostics) {
	if version == nil {
		return types.StringNull(), nil
	}

	return versionFromOneOf(version.AsSecurityOsqueryAPICreateSavedQueryResponseDataVersion1, version.AsSecurityOsqueryAPICreateSavedQueryResponseDataVersion0, "Create")
}

func versionFromGetAPI(version *kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse_Data_Version) (types.String, diag.Diagnostics) {
	if version == nil {
		return types.StringNull(), nil
	}

	return versionFromOneOf(version.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion1, version.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion0, "Get")
}

// versionFromOneOf unwraps a generated kbapi oneOf(string, int) version response into types.String.
// asString and asInt are the generated type's AsXxx1/AsXxx0 accessors; context labels the error message.
func versionFromOneOf(asString func() (string, error), asInt func() (int, error), context string) (types.String, diag.Diagnostics) {
	if value, err := asString(); err == nil {
		return typeutils.TrimmedStringishValue(value), nil
	}

	if value, err := asInt(); err == nil {
		return types.StringValue(strconv.Itoa(value)), nil
	}

	return types.StringNull(), diag.Diagnostics{
		diag.NewErrorDiagnostic("Invalid version value", fmt.Sprintf("%s response version is neither integer nor string.", context)),
	}
}

func versionFromUpdateAPI(version *string) types.String {
	if version == nil {
		return types.StringNull()
	}

	return typeutils.TrimmedStringishValue(*version)
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
