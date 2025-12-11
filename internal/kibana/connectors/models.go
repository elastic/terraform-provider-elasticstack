package connectors

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID               types.String         `tfsdk:"id"`
	KibanaConnection types.List           `tfsdk:"kibana_connection"`
	ConnectorID      types.String         `tfsdk:"connector_id"`
	SpaceID          types.String         `tfsdk:"space_id"`
	Name             types.String         `tfsdk:"name"`
	ConnectorTypeID  types.String         `tfsdk:"connector_type_id"`
	Config           ConfigValue          `tfsdk:"config"`
	Secrets          jsontypes.Normalized `tfsdk:"secrets"`
	IsDeprecated     types.Bool           `tfsdk:"is_deprecated"`
	IsMissingSecrets types.Bool           `tfsdk:"is_missing_secrets"`
	IsPreconfigured  types.Bool           `tfsdk:"is_preconfigured"`
}

func (model tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compId, nil
}

func (model tfModel) toAPIModel() (models.KibanaActionConnector, diag.Diagnostics) {
	apiModel := models.KibanaActionConnector{
		ConnectorID:     model.ConnectorID.ValueString(),
		SpaceID:         model.SpaceID.ValueString(),
		Name:            model.Name.ValueString(),
		ConnectorTypeID: model.ConnectorTypeID.ValueString(),
	}

	if utils.IsKnown(model.Config) {
		sanitizedConfig, diags := model.Config.SanitizedValue()
		if diags.HasError() {
			return models.KibanaActionConnector{}, diags
		}
		apiModel.ConfigJSON = sanitizedConfig
	}

	if utils.IsKnown(model.Secrets) {
		apiModel.SecretsJSON = model.Secrets.ValueString()
	}

	return apiModel, nil
}

func (model *tfModel) populateFromAPI(apiModel *models.KibanaActionConnector, compositeID *clients.CompositeId) diag.Diagnostics {
	model.ID = types.StringValue(compositeID.String())
	model.ConnectorID = types.StringValue(apiModel.ConnectorID)
	model.SpaceID = types.StringValue(apiModel.SpaceID)
	model.Name = types.StringValue(apiModel.Name)
	model.ConnectorTypeID = types.StringValue(apiModel.ConnectorTypeID)
	model.IsDeprecated = types.BoolValue(apiModel.IsDeprecated)
	model.IsMissingSecrets = types.BoolValue(apiModel.IsMissingSecrets)
	model.IsPreconfigured = types.BoolValue(apiModel.IsPreconfigured)

	if apiModel.ConfigJSON != "" {
		var diags diag.Diagnostics
		model.Config, diags = NewConfigValueWithConnectorID(apiModel.ConfigJSON, apiModel.ConnectorTypeID)
		if diags.HasError() {
			return diags
		}
	}

	return nil
}
