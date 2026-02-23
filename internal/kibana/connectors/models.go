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

package connectors

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
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

func (model tfModel) GetID() (*clients.CompositeID, diag.Diagnostics) {
	compID, sdkDiags := clients.CompositeIDFromStr(model.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	return compID, nil
}

func (model tfModel) toAPIModel() (models.KibanaActionConnector, diag.Diagnostics) {
	apiModel := models.KibanaActionConnector{
		ConnectorID:     model.ConnectorID.ValueString(),
		SpaceID:         model.SpaceID.ValueString(),
		Name:            model.Name.ValueString(),
		ConnectorTypeID: model.ConnectorTypeID.ValueString(),
	}

	if typeutils.IsKnown(model.Config) {
		sanitizedConfig, diags := model.Config.SanitizedValue()
		if diags.HasError() {
			return models.KibanaActionConnector{}, diags
		}
		apiModel.ConfigJSON = sanitizedConfig
	}

	if typeutils.IsKnown(model.Secrets) {
		apiModel.SecretsJSON = model.Secrets.ValueString()
	}

	return apiModel, nil
}

func (model *tfModel) populateFromAPI(apiModel *models.KibanaActionConnector, compositeID *clients.CompositeID) diag.Diagnostics {
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
