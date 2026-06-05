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

package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readSlo(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapi := client.GetKibanaOapiClient()

	res, fwDiags := kibanaoapi.GetSlo(ctx, oapi, spaceID, resourceID)
	diags.Append(fwDiags...)
	if diags.HasError() {
		return model, false, diags
	}
	if res == nil {
		return model, false, diags
	}

	apiModel := kibanaoapi.SloResponseToModel(spaceID, res)
	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: apiModel.SloID}).String())
	diags.Append(model.populateFromAPI(apiModel)...)
	if diags.HasError() {
		return model, true, diags
	}

	return model, true, diags
}
