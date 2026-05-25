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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createConnector(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics

	diags.Append(enforceUserSuppliedConnectorIDVersion(ctx, client, planModel)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	modelForAPI := planModel
	if typeutils.IsKnown(req.Config.SecretsWo) {
		modelForAPI.SecretsWo = req.Config.SecretsWo
	}

	apiModel, apiDiags := modelForAPI.toAPIModel()
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	oapiClient, getDiags := client.GetKibanaOapiClient()
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	connectorID, createDiags := kibanaoapi.CreateConnector(ctx, oapiClient, apiModel)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	compositeID := clients.CompositeID{
		ClusterID:  req.SpaceID,
		ResourceID: connectorID,
	}
	planModel.ID = types.StringValue(compositeID.String())
	planModel.ConnectorID = types.StringValue(connectorID)

	return entitycore.KibanaWriteResult[tfModel]{Model: planModel}, diags
}
