package api_key

import (
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

type tfModel struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	KeyID                   types.String         `tfsdk:"key_id"`
	Name                    types.String         `tfsdk:"name"`
	RoleDescriptors         jsontypes.Normalized `tfsdk:"role_descriptors"`
	Expiration              types.String         `tfsdk:"expiration"`
	ExpirationTimestamp     types.Int64          `tfsdk:"expiration_timestamp"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
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

	if utils.IsKnown(model.RoleDescriptors) {
		diags := model.RoleDescriptors.Unmarshal(&apiModel.RolesDescriptors)
		if diags.HasError() {
			return models.ApiKey{}, diags
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
