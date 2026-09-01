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
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateConnector(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	return entitycore.SimpleKibanaUpdate[tfModel, models.KibanaActionConnector, string](
		func(plan tfModel, _ context.Context, writeID string) (models.KibanaActionConnector, diag.Diagnostics) {
			modelForAPI := plan
			if typeutils.IsKnown(req.Config.SecretsWo) {
				modelForAPI.SecretsWo = req.Config.SecretsWo
			}

			apiModel, diags := modelForAPI.toAPIModel()
			if diags.HasError() {
				return apiModel, diags
			}

			apiModel.ConnectorID = writeID
			return apiModel, diags
		},
		func(ctx context.Context, oapiClient *kibanaoapi.Client, _, _ string, apiModel models.KibanaActionConnector) (*string, diag.Diagnostics) {
			connectorID, diags := kibanaoapi.UpdateConnector(ctx, oapiClient, apiModel)
			if diags.HasError() {
				return nil, diags
			}
			return &connectorID, diags
		},
		func(plan *tfModel, _ context.Context, spaceID string, connectorID *string) diag.Diagnostics {
			compositeID := clients.CompositeID{
				ClusterID:  spaceID,
				ResourceID: *connectorID,
			}
			plan.ID = types.StringValue(compositeID.String())
			plan.ConnectorID = types.StringValue(*connectorID)
			return nil
		},
	)(ctx, client, req)
}
