package api_key

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	KeyID                   types.String         `tfsdk:"key_id"`
	Name                    types.String         `tfsdk:"name"`
	Type                    types.String         `tfsdk:"type"`
	RoleDescriptors         jsontypes.Normalized `tfsdk:"role_descriptors"`
	Expiration              types.String         `tfsdk:"expiration"`
	ExpirationTimestamp     types.Int64          `tfsdk:"expiration_timestamp"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Access                  types.Object         `tfsdk:"access"`
	APIKey                  types.String         `tfsdk:"api_key"`
	Encoded                 types.String         `tfsdk:"encoded"`
}

func (model tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, utils.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compId, nil
}

func (model tfModel) toAPIModel() (models.ApiKey, diag.Diagnostics) {
	apiModel := models.ApiKey{
		ID:         model.KeyID.ValueString(),
		Name:       model.Name.ValueString(),
		Expiration: model.Expiration.ValueString(),
	}

	if utils.IsKnown(model.Metadata) {
		diags := model.Metadata.Unmarshal(&apiModel.Metadata)
		if diags.HasError() {
			return models.ApiKey{}, diags
		}
	}

	diags := model.RoleDescriptors.Unmarshal(&apiModel.RolesDescriptors)
	if diags.HasError() {
		return models.ApiKey{}, diags
	}

	return apiModel, nil
}

func (model tfModel) toCrossClusterAPIModel(ctx context.Context) (models.CrossClusterApiKey, diag.Diagnostics) {
	apiModel := models.CrossClusterApiKey{
		ID:         model.KeyID.ValueString(),
		Name:       model.Name.ValueString(),
		Expiration: model.Expiration.ValueString(),
	}

	if utils.IsKnown(model.Metadata) {
		diags := model.Metadata.Unmarshal(&apiModel.Metadata)
		if diags.HasError() {
			return models.CrossClusterApiKey{}, diags
		}
	}

	// Build the access configuration
	access := &models.CrossClusterApiKeyAccess{}

	if utils.IsKnown(model.Access) {
		var accessData accessModel
		diags := model.Access.As(ctx, &accessData, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return models.CrossClusterApiKey{}, diags
		}

		if utils.IsKnown(accessData.Search) {
			var searchObjects []searchModel
			diags := accessData.Search.ElementsAs(ctx, &searchObjects, false)
			if diags.HasError() {
				return models.CrossClusterApiKey{}, diags
			}

			var searchEntries []models.CrossClusterApiKeyAccessEntry
			for _, searchObj := range searchObjects {
				entry := models.CrossClusterApiKeyAccessEntry{}

				if utils.IsKnown(searchObj.Names) {
					var names []string
					diags := searchObj.Names.ElementsAs(ctx, &names, false)
					if diags.HasError() {
						return models.CrossClusterApiKey{}, diags
					}
					entry.Names = names
				}

				if utils.IsKnown(searchObj.FieldSecurity) && !searchObj.FieldSecurity.IsNull() {
					var fieldSecurity models.FieldSecurity
					diags := json.Unmarshal([]byte(searchObj.FieldSecurity.ValueString()), &fieldSecurity)
					if diags != nil {
						return models.CrossClusterApiKey{}, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to unmarshal field_security", diags.Error())}
					}
					entry.FieldSecurity = &fieldSecurity
				}

				if utils.IsKnown(searchObj.Query) && !searchObj.Query.IsNull() {
					query := searchObj.Query.ValueString()
					entry.Query = &query
				}

				if utils.IsKnown(searchObj.AllowRestrictedIndices) {
					allowRestricted := searchObj.AllowRestrictedIndices.ValueBool()
					entry.AllowRestrictedIndices = &allowRestricted
				}

				searchEntries = append(searchEntries, entry)
			}
			if len(searchEntries) > 0 {
				access.Search = searchEntries
			}
		}

		if utils.IsKnown(accessData.Replication) {
			var replicationObjects []replicationModel
			diags := accessData.Replication.ElementsAs(ctx, &replicationObjects, false)
			if diags.HasError() {
				return models.CrossClusterApiKey{}, diags
			}

			var replicationEntries []models.CrossClusterApiKeyAccessEntry
			for _, replicationObj := range replicationObjects {
				if utils.IsKnown(replicationObj.Names) {
					var names []string
					diags := replicationObj.Names.ElementsAs(ctx, &names, false)
					if diags.HasError() {
						return models.CrossClusterApiKey{}, diags
					}
					if len(names) > 0 {
						replicationEntries = append(replicationEntries, models.CrossClusterApiKeyAccessEntry{
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
			apiModel.Access = access
		}
	}

	return apiModel, nil
}

func (model *tfModel) populateFromCreate(apiKey models.ApiKeyCreateResponse) {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.APIKey = basetypes.NewStringValue(apiKey.Key)
	model.Encoded = basetypes.NewStringValue(apiKey.EncodedKey)
}

func (model *tfModel) populateFromCrossClusterCreate(apiKey models.CrossClusterApiKeyCreateResponse) {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.APIKey = basetypes.NewStringValue(apiKey.Key)
	model.Encoded = basetypes.NewStringValue(apiKey.EncodedKey)
	if apiKey.Expiration > 0 {
		model.ExpirationTimestamp = basetypes.NewInt64Value(apiKey.Expiration)
	}
}

func (model *tfModel) populateFromAPI(apiKey models.ApiKeyResponse, serverVersion *version.Version) diag.Diagnostics {
	model.KeyID = basetypes.NewStringValue(apiKey.Id)
	model.Name = basetypes.NewStringValue(apiKey.Name)
	model.ExpirationTimestamp = basetypes.NewInt64Value(apiKey.Expiration)
	model.Metadata = jsontypes.NewNormalizedNull()

	if serverVersion.GreaterThanOrEqual(MinVersionReturningRoleDescriptors) {
		model.RoleDescriptors = jsontypes.NewNormalizedNull()

		if apiKey.RolesDescriptors != nil {
			descriptors, diags := marshalNormalizedJsonValue(apiKey.RolesDescriptors)
			if diags.HasError() {
				return diags
			}

			model.RoleDescriptors = descriptors
		}
	}

	if apiKey.Metadata != nil {
		metadata, diags := marshalNormalizedJsonValue(apiKey.Metadata)
		if diags.HasError() {
			return diags
		}

		model.Metadata = metadata
	}

	return nil
}

func marshalNormalizedJsonValue(item any) (jsontypes.Normalized, diag.Diagnostics) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return jsontypes.Normalized{}, utils.FrameworkDiagFromError(err)
	}

	return jsontypes.NewNormalizedValue(string(jsonBytes)), nil
}
