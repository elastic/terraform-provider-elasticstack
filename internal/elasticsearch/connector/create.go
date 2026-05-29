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

package connector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func createConnector(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[ContentConnectorData],
) (entitycore.WriteResult[ContentConnectorData], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan

	connectorID, createDiags := esclient.CreateConnector(ctx, client, req.WriteID, data.toCreateConnectorBody())
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[ContentConnectorData]{Model: data}, diags
	}

	data.ConnectorID = fwtypes.StringValue(connectorID)

	diags.Append(applyConnectorFanOut(ctx, client, connectorID, data, req.Config, nil, req.Private, false)...)

	return entitycore.WriteResult[ContentConnectorData]{Model: data}, diags
}
