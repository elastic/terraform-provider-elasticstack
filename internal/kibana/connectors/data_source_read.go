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
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readConnectorDataSource(ctx context.Context, client *clients.KibanaScopedClient, model connectorDataSourceModel) (connectorDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	spaceID := ""
	if !model.SpaceID.IsNull() {
		spaceID = model.SpaceID.ValueString()
	}
	if spaceID == "" {
		spaceID = "default"
		model.SpaceID = types.StringValue("default")
	}

	connectorName := model.Name.ValueString()

	connectorType := ""
	if !model.ConnectorTypeID.IsNull() {
		connectorType = model.ConnectorTypeID.ValueString()
	}

	foundConnectors, searchDiags := kibanaoapi.SearchConnectors(ctx, oapiClient, connectorName, spaceID, connectorType)
	diags.Append(searchDiags...)
	if diags.HasError() {
		return model, diags
	}

	if len(foundConnectors) == 0 {
		diags.AddError(
			"error while creating elasticstack_kibana_action_connector datasource",
			fmt.Sprintf("connector with name [%s/%s] and type [%s] not found", spaceID, connectorName, connectorType),
		)
		return model, diags
	}

	if len(foundConnectors) > 1 {
		diags.AddError(
			"error while creating elasticstack_kibana_action_connector datasource",
			fmt.Sprintf("multiple connectors found with name [%s/%s] and type [%s]", spaceID, connectorName, connectorType),
		)
		return model, diags
	}

	connector := foundConnectors[0]
	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: connector.ConnectorID}
	model.ID = types.StringValue(compositeID.String())
	model.ConnectorID = types.StringValue(connector.ConnectorID)
	model.SpaceID = types.StringValue(connector.SpaceID)
	model.Name = types.StringValue(connector.Name)
	model.ConnectorTypeID = types.StringValue(connector.ConnectorTypeID)
	if connector.ConfigJSON != "" {
		model.Config = jsontypes.NewNormalizedValue(connector.ConfigJSON)
	} else {
		model.Config = jsontypes.NewNormalizedNull()
	}
	model.IsDeprecated = types.BoolValue(connector.IsDeprecated)
	model.IsMissingSecrets = types.BoolValue(connector.IsMissingSecrets)
	model.IsPreconfigured = types.BoolValue(connector.IsPreconfigured)

	return model, diags
}
