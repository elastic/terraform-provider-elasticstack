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

package security_entity_store_resolution_group

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readResolutionGroup(ctx context.Context, client *clients.KibanaScopedClient, config resolutionGroupModel) (resolutionGroupModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	spaceID := config.SpaceID.ValueString()
	if spaceID == "" {
		spaceID = "default"
	}

	resp, err := client.GetKibanaOapiClient().API.GetSecurityEntityStoreResolutionGroupWithResponse(
		ctx,
		&kbapi.GetSecurityEntityStoreResolutionGroupParams{EntityId: config.EntityID.ValueString()},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		diags.AddError("Failed to read resolution group", err.Error())
		return config, diags
	}

	if resp.StatusCode() == http.StatusNotFound {
		diags.AddError(
			"Resolution group not found",
			"The specified entity does not exist or has no resolution group.",
		)
		return config, diags
	}
	if resp.StatusCode() != http.StatusOK {
		diags.Append(diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)...)
		return config, diags
	}

	result := config
	diags.Append(result.populateFromAPI(spaceID, resp.Body)...)
	return result, diags
}
