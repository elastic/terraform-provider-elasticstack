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

package cloudconnector

import (
	"context"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, config cloudConnectorsDataSourceModel) (cloudConnectorsDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	client := kbClient.GetFleetClient()

	spaceID := config.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = defaultSpaceID
		config.SpaceID = types.StringValue(defaultSpaceID)
	}

	items, listDiags := fleetclient.ListCloudConnectors(ctx, client, spaceID, listParamsFromModel(config))
	diags.Append(listDiags...)
	if diags.HasError() {
		return config, diags
	}

	mapDiags := mapAPIToDatasourceModel(ctx, &config, spaceID, items)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return config, diags
	}

	return config, diags
}

func listParamsFromModel(config cloudConnectorsDataSourceModel) kbapi.GetFleetCloudConnectorsParams {
	var params kbapi.GetFleetCloudConnectorsParams

	if !config.Kuery.IsNull() && config.Kuery.ValueString() != "" {
		kuery := config.Kuery.ValueString()
		params.Kuery = &kuery
	}

	if !config.Page.IsNull() {
		page := strconv.FormatInt(config.Page.ValueInt64(), 10)
		params.Page = &page
	}

	if !config.PerPage.IsNull() {
		perPage := strconv.FormatInt(config.PerPage.ValueInt64(), 10)
		params.PerPage = &perPage
	}

	return params
}
