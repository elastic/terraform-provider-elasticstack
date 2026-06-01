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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readConnector(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// After terraform import the prior state contains only the resource id.
	// Every other attribute (including the required 'name') is null.
	justImported := model.Name.IsNull()

	oapiClient := client.GetKibanaOapiClient()

	connector, readDiags := kibanaoapi.GetConnector(ctx, oapiClient, resourceID, spaceID)
	if !connectorReadExists(connector, readDiags) {
		return model, false, diags
	}
	diags.Append(readDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	result, found, finishDiags := finishConnectorRead(model, connector, spaceID, resourceID)
	diags.Append(finishDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	// The Kibana API never returns secrets in read responses for security.
	// When importing a connector that is not missing secrets, defaulting
	// secrets to "{}" prevents the perpetual drift reported in
	// https://github.com/elastic/terraform-provider-elasticstack/issues/634
	if justImported && result.Secrets.IsNull() && !result.IsMissingSecrets.ValueBool() {
		result.Secrets = jsontypes.NewNormalizedValue("{}")
	}

	return result, found, diags
}

// connectorReadExists reports whether GetConnector found a connector. A nil
// connector with no read diagnostics means the resource is gone.
func connectorReadExists(connector *models.KibanaActionConnector, readDiags diag.Diagnostics) bool {
	return connector != nil || readDiags.HasError()
}

// finishConnectorRead populates model from a connector returned by the API.
// populateFromAPI errors preserve state (found=true), matching pre-envelope read.
func finishConnectorRead(model tfModel, connector *models.KibanaActionConnector, spaceID, resourceID string) (tfModel, bool, diag.Diagnostics) {
	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}
	var diags diag.Diagnostics
	diags.Append(model.populateFromAPI(connector, compositeID)...)
	return model, true, diags
}
