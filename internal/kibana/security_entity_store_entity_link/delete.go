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

package security_entity_store_entity_link

import (
	"context"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteEntityLink(ctx context.Context, client *clients.KibanaScopedClient, _ string, spaceID string, state entityLinkModel) diag.Diagnostics {
	var diags diag.Diagnostics

	entityIDs, d := agentbuilder.SetToStrings(ctx, state.EntityIDs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if len(entityIDs) == 0 {
		return diags
	}

	body := kbapi.PostSecurityEntityStoreResolutionUnlinkJSONRequestBody{
		EntityIds: entityIDs,
	}

	resp, err := client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionUnlinkWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		diags.AddError("Failed to unlink entities", err.Error())
		return diags
	}

	if resp.StatusCode() == http.StatusNotFound {
		return diags
	}
	if resp.StatusCode() != http.StatusOK {
		diags.Append(diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)...)
		return diags
	}

	return diags
}
