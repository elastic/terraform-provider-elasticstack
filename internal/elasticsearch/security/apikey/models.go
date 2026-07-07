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

package apikey

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/updateapikey"
	"github.com/elastic/go-elasticsearch/v9/typedapi/security/updatecrossclusterapikey"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type SearchModel struct {
	Names                  types.List           `tfsdk:"names"`
	FieldSecurity          jsontypes.Normalized `tfsdk:"field_security"`
	Query                  jsontypes.Normalized `tfsdk:"query"`
	AllowRestrictedIndices types.Bool           `tfsdk:"allow_restricted_indices"`
}

type ReplicationModel struct {
	Names types.List `tfsdk:"names"`
}

type AccessModel struct {
	Search      types.List `tfsdk:"search"`
	Replication types.List `tfsdk:"replication"`
}

type TfModel struct {
	entitycore.ResourceTimeoutsField
	ID                      types.String                                                              `tfsdk:"id"`
	ElasticsearchConnection types.List                                                                `tfsdk:"elasticsearch_connection"`
	KeyID                   types.String                                                              `tfsdk:"key_id"`
	Name                    types.String                                                              `tfsdk:"name"`
	Type                    types.String                                                              `tfsdk:"type"`
	RoleDescriptors         customtypes.JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor] `tfsdk:"role_descriptors"`
	Expiration              types.String                                                              `tfsdk:"expiration"`
	ExpirationTimestamp     types.Int64                                                               `tfsdk:"expiration_timestamp"`
	Metadata                jsontypes.Normalized                                                      `tfsdk:"metadata"`
	Access                  types.Object                                                              `tfsdk:"access"`
	APIKey                  types.String                                                              `tfsdk:"api_key"`
	Encoded                 types.String                                                              `tfsdk:"encoded"`
}

func (model TfModel) GetID() types.String {
	return model.ID
}

func (model TfModel) GetResourceID() types.String {
	return model.Name
}

func (model TfModel) GetElasticsearchConnection() types.List {
	return model.ElasticsearchConnection
}

// GetReadResourceID satisfies entitycore.WithReadResourceID: the API key read
// identity is the immutable key_id (not the user-supplied Name) because the
// Elasticsearch Get/Update API key APIs are keyed by id.
func (model TfModel) GetReadResourceID() string {
	if typeutils.IsKnown(model.KeyID) && model.KeyID.ValueString() != "" {
		return model.KeyID.ValueString()
	}
	if typeutils.IsKnown(model.ID) && model.ID.ValueString() != "" {
		compID, diags := clients.CompositeIDFromStr(model.ID.ValueString())
		if !diags.HasError() && compID != nil {
			return compID.ResourceID
		}
	}
	return ""
}

var _ entitycore.WithReadResourceID = TfModel{}

var _ entitycore.WithVersionRequirements = TfModel{}

// GetVersionRequirements declares the conditional Elasticsearch version
// requirements implied by the planned model: cross-cluster API keys require
// MinVersionWithCrossCluster, and any role descriptor carrying a `restriction`
// block requires MinVersionWithRestriction.
func (model TfModel) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var diags diag.Diagnostics
	var reqs []entitycore.VersionRequirement

	if model.Type.ValueString() == CrossClusterAPIKeyType {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *MinVersionWithCrossCluster,
			ErrorMessage: fmt.Sprintf("Cross-cluster API keys are only supported in Elasticsearch version %s and above.", MinVersionWithCrossCluster.String()),
		})
	}

	if typeutils.IsKnown(model.RoleDescriptors) {
		var roleDescriptors map[string]models.APIKeyRoleDescriptor
		unmarshalDiags := model.RoleDescriptors.Unmarshal(&roleDescriptors)
		if unmarshalDiags.HasError() {
			diags.Append(unmarshalDiags...)
			return reqs, diags
		}

		var keysWithRestrictions []string
		for key, descriptor := range roleDescriptors {
			if descriptor.Restriction != nil {
				keysWithRestrictions = append(keysWithRestrictions, key)
			}
		}
		if len(keysWithRestrictions) > 0 {
			sort.Strings(keysWithRestrictions)
			reqs = append(reqs, entitycore.VersionRequirement{
				MinVersion: *MinVersionWithRestriction,
				ErrorMessage: fmt.Sprintf(
					"Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor(s) %s",
					strings.Join(keysWithRestrictions, ", "),
				),
			})
		}
	}

	return reqs, diags
}

func (model TfModel) buildTypedRoleDescriptors() (map[string]estypes.RoleDescriptor, diag.Diagnostics) {
	if !typeutils.IsKnown(model.RoleDescriptors) {
		return nil, nil
	}

	var roleDescriptors map[string]models.APIKeyRoleDescriptor
	diags := model.RoleDescriptors.Unmarshal(&roleDescriptors)
	if diags.HasError() {
		return nil, diags
	}

	if len(roleDescriptors) == 0 {
		return nil, nil
	}

	typedDescriptors, err := toTypedRoleDescriptors(roleDescriptors)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return typedDescriptors, nil
}

func (model TfModel) buildTypedMetadata() (estypes.Metadata, diag.Diagnostics) {
	if !typeutils.IsKnown(model.Metadata) {
		return nil, nil
	}
	var metadata map[string]any
	diags := model.Metadata.Unmarshal(&metadata)
	if diags.HasError() {
		return nil, diags
	}
	typedMetadata, err := toTypedMetadata(metadata)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return typedMetadata, nil
}

func (model TfModel) toAPICreateModel() (*createapikey.Request, diag.Diagnostics) {
	req := createapikey.NewRequest()

	if model.Name.ValueString() != "" {
		req.Name = model.Name.ValueStringPointer()
	}
	if model.Expiration.ValueString() != "" {
		req.Expiration = model.Expiration.ValueString()
	}

	typedMetadata, diags := model.buildTypedMetadata()
	if diags.HasError() {
		return nil, diags
	}
	req.Metadata = typedMetadata

	typedDescriptors, diags := model.buildTypedRoleDescriptors()
	if diags.HasError() {
		return nil, diags
	}
	req.RoleDescriptors = typedDescriptors

	return req, nil
}

func (model TfModel) ToUpdateAPIRequest() (*updateapikey.Request, diag.Diagnostics) {
	req := updateapikey.NewRequest()

	// Note: the Update API Key endpoint does not accept expiration.
	// The old code explicitly zeroed it out before sending.

	typedMetadata, diags := model.buildTypedMetadata()
	if diags.HasError() {
		return nil, diags
	}
	req.Metadata = typedMetadata

	typedDescriptors, diags := model.buildTypedRoleDescriptors()
	if diags.HasError() {
		return nil, diags
	}
	req.RoleDescriptors = typedDescriptors

	return req, nil
}

func (model TfModel) buildCrossClusterAccess(ctx context.Context) (*models.CrossClusterAPIKeyAccess, diag.Diagnostics) {
	if !typeutils.IsKnown(model.Access) {
		return nil, nil
	}

	access := &models.CrossClusterAPIKeyAccess{}

	var accessData AccessModel
	diags := model.Access.As(ctx, &accessData, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(accessData.Search) {
		var searchObjects []SearchModel
		diags := accessData.Search.ElementsAs(ctx, &searchObjects, false)
		if diags.HasError() {
			return nil, diags
		}

		var searchEntries []models.CrossClusterAPIKeyAccessEntry
		for _, searchObj := range searchObjects {
			entry := models.CrossClusterAPIKeyAccessEntry{}

			if typeutils.IsKnown(searchObj.Names) {
				var names []string
				diags := searchObj.Names.ElementsAs(ctx, &names, false)
				if diags.HasError() {
					return nil, diags
				}
				entry.Names = names
			}

			if typeutils.IsKnown(searchObj.FieldSecurity) && !searchObj.FieldSecurity.IsNull() {
				var fieldSecurity models.FieldSecurity
				err := json.Unmarshal([]byte(searchObj.FieldSecurity.ValueString()), &fieldSecurity)
				if err != nil {
					return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal field_security", err.Error())}
				}
				entry.FieldSecurity = &fieldSecurity
			}

			if typeutils.IsKnown(searchObj.Query) && !searchObj.Query.IsNull() {
				query := searchObj.Query.ValueString()
				entry.Query = &query
			}

			if typeutils.IsKnown(searchObj.AllowRestrictedIndices) {
				allowRestricted := searchObj.AllowRestrictedIndices.ValueBool()
				entry.AllowRestrictedIndices = &allowRestricted
			}

			searchEntries = append(searchEntries, entry)
		}
		if len(searchEntries) > 0 {
			access.Search = searchEntries
		}
	}

	if typeutils.IsKnown(accessData.Replication) {
		var replicationObjects []ReplicationModel
		diags := accessData.Replication.ElementsAs(ctx, &replicationObjects, false)
		if diags.HasError() {
			return nil, diags
		}

		var replicationEntries []models.CrossClusterAPIKeyAccessEntry
		for _, replicationObj := range replicationObjects {
			if typeutils.IsKnown(replicationObj.Names) {
				var names []string
				diags := replicationObj.Names.ElementsAs(ctx, &names, false)
				if diags.HasError() {
					return nil, diags
				}
				if len(names) > 0 {
					replicationEntries = append(replicationEntries, models.CrossClusterAPIKeyAccessEntry{
						Names: names,
					})
				}
			}
		}
		if len(replicationEntries) > 0 {
			access.Replication = replicationEntries
		}
	}

	return access, nil
}

func (model TfModel) toCrossClusterAPICreateModel(ctx context.Context) (*createcrossclusterapikey.Request, diag.Diagnostics) {
	req := createcrossclusterapikey.NewRequest()
	req.Name = model.Name.ValueString()

	if model.Expiration.ValueString() != "" {
		req.Expiration = model.Expiration.ValueString()
	}

	typedMetadata, diags := model.buildTypedMetadata()
	if diags.HasError() {
		return nil, diags
	}
	req.Metadata = typedMetadata

	access, diags := model.buildCrossClusterAccess(ctx)
	if diags.HasError() {
		return nil, diags
	}
	if access != nil && (access.Search != nil || access.Replication != nil) {
		typedAccess, err := toTypedAccess(*access)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Access = typedAccess
	}

	return req, nil
}

func (model TfModel) ToUpdateCrossClusterAPIRequest(ctx context.Context) (*updatecrossclusterapikey.Request, diag.Diagnostics) {
	req := updatecrossclusterapikey.NewRequest()

	// Note: the Update Cross-Cluster API Key endpoint does not accept expiration.
	// The old code explicitly zeroed it out before sending.

	typedMetadata, diags := model.buildTypedMetadata()
	if diags.HasError() {
		return nil, diags
	}
	req.Metadata = typedMetadata

	access, diags := model.buildCrossClusterAccess(ctx)
	if diags.HasError() {
		return nil, diags
	}
	if access != nil && (access.Search != nil || access.Replication != nil) {
		typedAccess, err := toTypedAccess(*access)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Access = typedAccess
	}

	return req, nil
}

func (model *TfModel) populateFromCreate(apiKey *createapikey.Response) {
	model.populateCommonCreateFields(apiKey.Id, apiKey.Name, apiKey.ApiKey, apiKey.Encoded, apiKey.Expiration)
}

func (model *TfModel) populateFromCrossClusterCreate(apiKey *createcrossclusterapikey.Response) {
	model.populateCommonCreateFields(apiKey.Id, apiKey.Name, apiKey.ApiKey, apiKey.Encoded, apiKey.Expiration)
}

// populateCommonCreateFields writes the fields returned by both the REST and
// cross-cluster create API key responses into the model. ExpirationTimestamp is
// set to zero when the upstream API returns no expiration so the field is
// always known after a successful create; the resource read path subsequently
// overwrites it from the Get API Key response.
func (model *TfModel) populateCommonCreateFields(id, name, apiKey, encoded string, expiration *int64) {
	model.KeyID = basetypes.NewStringValue(id)
	model.Name = basetypes.NewStringValue(name)
	model.APIKey = basetypes.NewStringValue(apiKey)
	model.Encoded = basetypes.NewStringValue(encoded)
	model.ExpirationTimestamp = basetypes.NewInt64Value(0)
	if expiration != nil && *expiration > 0 {
		model.ExpirationTimestamp = basetypes.NewInt64Value(*expiration)
	}
}

func (model *TfModel) PopulateFromAPI(apiKey *estypes.ApiKey, caps APIKeyCapabilities) diag.Diagnostics {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.ExpirationTimestamp = basetypes.NewInt64Value(0)
	if apiKey.Expiration != nil {
		model.ExpirationTimestamp = basetypes.NewInt64Value(*apiKey.Expiration)
	}
	model.Metadata = jsontypes.NewNormalizedNull()

	if caps.SupportsRoleDescriptors {
		model.RoleDescriptors = customtypes.NewJSONWithDefaultsNull(PopulateRoleDescriptorsDefaults)

		if apiKey.RoleDescriptors != nil {
			modelDescriptors, err := toModelRoleDescriptors(apiKey.RoleDescriptors)
			if err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}

			var marshalDiags diag.Diagnostics
			descriptors := typeutils.MarshalToNormalized(modelDescriptors, path.Root("role_descriptors"), &marshalDiags)
			if marshalDiags.HasError() {
				return marshalDiags
			}

			model.RoleDescriptors = customtypes.NewJSONWithDefaultsValue(descriptors.ValueString(), PopulateRoleDescriptorsDefaults)
		}
	} else if !typeutils.IsKnown(model.RoleDescriptors) {
		// The Get API Key endpoint does not return role_descriptors prior to 8.5.0.
		// If the value is unknown (e.g. not specified in config during Create), set
		// it to null so Terraform receives a known value after apply.
		model.RoleDescriptors = customtypes.NewJSONWithDefaultsNull(PopulateRoleDescriptorsDefaults)
	}

	if apiKey.Metadata != nil {
		metadata, err := toModelMetadata(apiKey.Metadata)
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		var marshalDiags diag.Diagnostics
		model.Metadata = typeutils.MarshalToNormalized(metadata, path.Root("metadata"), &marshalDiags)
		if marshalDiags.HasError() {
			return marshalDiags
		}
	}

	return nil
}
