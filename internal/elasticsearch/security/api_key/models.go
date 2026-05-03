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

	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updateapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updatecrossclusterapikey"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type searchModel struct {
	Names                  types.List           `tfsdk:"names"`
	FieldSecurity          jsontypes.Normalized `tfsdk:"field_security"`
	Query                  jsontypes.Normalized `tfsdk:"query"`
	AllowRestrictedIndices types.Bool           `tfsdk:"allow_restricted_indices"`
}

type replicationModel struct {
	Names types.List `tfsdk:"names"`
}

type accessModel struct {
	Search      types.List `tfsdk:"search"`
	Replication types.List `tfsdk:"replication"`
}

type tfModel struct {
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

func (model tfModel) GetID() (*clients.CompositeID, diag.Diagnostics) {
	compID, sdkDiags := clients.CompositeIDFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compID, nil
}

func (model tfModel) toAPICreateRequest() (*createapikey.Request, diag.Diagnostics) {
	req := createapikey.NewRequest()

	if model.Name.ValueString() != "" {
		req.Name = model.Name.ValueStringPointer()
	}
	if model.Expiration.ValueString() != "" {
		req.Expiration = model.Expiration.ValueString()
	}

	if typeutils.IsKnown(model.Metadata) {
		var metadata map[string]any
		diags := model.Metadata.Unmarshal(&metadata)
		if diags.HasError() {
			return nil, diags
		}
		typedMetadata, err := toTypedMetadata(metadata)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Metadata = typedMetadata
	}

	var roleDescriptors map[string]models.APIKeyRoleDescriptor
	diags := model.RoleDescriptors.Unmarshal(&roleDescriptors)
	if diags.HasError() {
		return nil, diags
	}

	if len(roleDescriptors) > 0 {
		typedDescriptors, err := toTypedRoleDescriptors(roleDescriptors)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.RoleDescriptors = typedDescriptors
	}

	return req, nil
}

func (model tfModel) toUpdateAPIRequest() (*updateapikey.Request, diag.Diagnostics) {
	req := updateapikey.NewRequest()

	// Note: the Update API Key endpoint does not accept expiration.
	// The old code explicitly zeroed it out before sending.

	if typeutils.IsKnown(model.Metadata) {
		var metadata map[string]any
		diags := model.Metadata.Unmarshal(&metadata)
		if diags.HasError() {
			return nil, diags
		}
		typedMetadata, err := toTypedMetadata(metadata)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Metadata = typedMetadata
	}

	var roleDescriptors map[string]models.APIKeyRoleDescriptor
	diags := model.RoleDescriptors.Unmarshal(&roleDescriptors)
	if diags.HasError() {
		return nil, diags
	}

	if len(roleDescriptors) > 0 {
		typedDescriptors, err := toTypedRoleDescriptors(roleDescriptors)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.RoleDescriptors = typedDescriptors
	}

	return req, nil
}

func (model tfModel) toCrossClusterAPICreateRequest(ctx context.Context) (*createcrossclusterapikey.Request, diag.Diagnostics) {
	req := createcrossclusterapikey.NewRequest()
	req.Name = model.Name.ValueString()

	if model.Expiration.ValueString() != "" {
		req.Expiration = model.Expiration.ValueString()
	}

	if typeutils.IsKnown(model.Metadata) {
		var metadata map[string]any
		diags := model.Metadata.Unmarshal(&metadata)
		if diags.HasError() {
			return nil, diags
		}
		typedMetadata, err := toTypedMetadata(metadata)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Metadata = typedMetadata
	}

	// Build the access configuration
	access := &models.CrossClusterAPIKeyAccess{}

	if typeutils.IsKnown(model.Access) {
		var accessData accessModel
		diags := model.Access.As(ctx, &accessData, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		if typeutils.IsKnown(accessData.Search) {
			var searchObjects []searchModel
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
					diags := json.Unmarshal([]byte(searchObj.FieldSecurity.ValueString()), &fieldSecurity)
					if diags != nil {
						return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal field_security", diags.Error())}
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
			var replicationObjects []replicationModel
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

		if access.Search != nil || access.Replication != nil {
			typedAccess, err := toTypedAccess(*access)
			if err != nil {
				return nil, diagutil.FrameworkDiagFromError(err)
			}
			req.Access = typedAccess
		}
	}

	return req, nil
}

func (model tfModel) toUpdateCrossClusterAPIRequest(ctx context.Context) (*updatecrossclusterapikey.Request, diag.Diagnostics) {
	req := updatecrossclusterapikey.NewRequest()

	// Note: the Update Cross-Cluster API Key endpoint does not accept expiration.
	// The old code explicitly zeroed it out before sending.

	if typeutils.IsKnown(model.Metadata) {
		var metadata map[string]any
		diags := model.Metadata.Unmarshal(&metadata)
		if diags.HasError() {
			return nil, diags
		}
		typedMetadata, err := toTypedMetadata(metadata)
		if err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		req.Metadata = typedMetadata
	}

	// Build the access configuration
	access := &models.CrossClusterAPIKeyAccess{}

	if typeutils.IsKnown(model.Access) {
		var accessData accessModel
		diags := model.Access.As(ctx, &accessData, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		if typeutils.IsKnown(accessData.Search) {
			var searchObjects []searchModel
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
					diags := json.Unmarshal([]byte(searchObj.FieldSecurity.ValueString()), &fieldSecurity)
					if diags != nil {
						return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal field_security", diags.Error())}
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
			var replicationObjects []replicationModel
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

		if access.Search != nil || access.Replication != nil {
			typedAccess, err := toTypedAccess(*access)
			if err != nil {
				return nil, diagutil.FrameworkDiagFromError(err)
			}
			req.Access = typedAccess
		}
	}

	return req, nil
}

func (model *tfModel) populateFromCreate(apiKey *createapikey.Response) {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.APIKey = basetypes.NewStringValue(apiKey.ApiKey)
	model.Encoded = basetypes.NewStringValue(apiKey.Encoded)
}

func (model *tfModel) populateFromCrossClusterCreate(apiKey *createcrossclusterapikey.Response) {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.APIKey = basetypes.NewStringValue(apiKey.ApiKey)
	model.Encoded = basetypes.NewStringValue(apiKey.Encoded)
	if apiKey.Expiration != nil && *apiKey.Expiration > 0 {
		model.ExpirationTimestamp = basetypes.NewInt64Value(*apiKey.Expiration)
	}
}

func (model *tfModel) populateFromAPI(apiKey *estypes.ApiKey, serverVersion *version.Version) diag.Diagnostics {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.ExpirationTimestamp = basetypes.NewInt64Value(0)
	if apiKey.Expiration != nil {
		model.ExpirationTimestamp = basetypes.NewInt64Value(*apiKey.Expiration)
	}
	model.Metadata = jsontypes.NewNormalizedNull()

	if serverVersion.GreaterThanOrEqual(MinVersionReturningRoleDescriptors) {
		model.RoleDescriptors = customtypes.NewJSONWithDefaultsNull(populateRoleDescriptorsDefaults)

		if apiKey.RoleDescriptors != nil {
			modelDescriptors, err := toModelRoleDescriptors(apiKey.RoleDescriptors)
			if err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}

			descriptors, diags := marshalNormalizedJSONValue(modelDescriptors)
			if diags.HasError() {
				return diags
			}

			model.RoleDescriptors = customtypes.NewJSONWithDefaultsValue(descriptors.ValueString(), populateRoleDescriptorsDefaults)
		}
	}

	if apiKey.Metadata != nil {
		metadata, err := toModelMetadata(apiKey.Metadata)
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		metadataJSON, diags := marshalNormalizedJSONValue(metadata)
		if diags.HasError() {
			return diags
		}
		model.Metadata = metadataJSON
	}

	return nil
}

func marshalNormalizedJSONValue(item any) (jsontypes.Normalized, diag.Diagnostics) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return jsontypes.Normalized{}, diagutil.FrameworkDiagFromError(err)
	}

	return jsontypes.NewNormalizedValue(string(jsonBytes)), nil
}
