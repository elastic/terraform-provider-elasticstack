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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readAgentDownloadSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	prior model,
) (model, bool, diag.Diagnostics) {
	fleetClient := client.GetFleetClient()

	return readAndHydrateState(ctx, fleetClient, resourceID, spaceID, prior.SpaceIDs, prior.KibanaConnection, prior.Auth)
}

func readAndHydrateState(
	ctx context.Context,
	client *fleet.Client,
	sourceID string,
	spaceID string,
	preservedSpaceIDs types.Set,
	preservedKibanaConnection types.List,
	fallbackAuth *authModel,
) (model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var state model

	getResp, readDiags := fleet.GetAgentDownloadSource(ctx, client, sourceID, spaceID)
	diags.Append(readDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	if getResp == nil {
		// Ensure we also exercise the list endpoint before deciding the resource is gone.
		listResp, listDiags := fleet.ListAgentDownloadSources(ctx, client, spaceID)
		diags.Append(listDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		if listResp != nil && listResp.JSON200 != nil {
			for _, listItem := range listResp.JSON200.Items {
				if listItem.Id == sourceID {
					diags.AddError("Unexpected API response", "Read by source_id returned not found but list endpoint still includes this source.")
					return state, false, diags
				}
			}
		}
		return state, false, diags
	}

	body, unwrapDiags := diagutil.UnwrapJSON200(getResp.JSON200, "agent download source")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	item := body.Item
	state.ID = types.StringValue(item.Id)
	state.SourceID = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	state.Host = types.StringValue(item.Host)
	state.Default = types.BoolPointerValue(item.IsDefault)
	state.ProxyID = typeutils.NonEmptyStringishPointerValue(item.ProxyId)
	state.SpaceIDs = preservedSpaceIDs
	state.KibanaConnection = preservedKibanaConnection
	state.Auth = hydrateAuth(ctx, item.Auth, fallbackAuth)
	return state, true, diags
}

func hydrateAuth(ctx context.Context, apiAuth *downloadSourceAuth, fallback *authModel) *authModel {
	if apiAuth == nil && fallback == nil {
		return nil
	}

	var apiUsername, apiPassword, apiKey *string
	if apiAuth != nil {
		apiUsername = apiAuth.Username
		apiPassword = apiAuth.Password
		apiKey = apiAuth.ApiKey
	}

	var fallbackPassword, fallbackAPIKey types.String
	if fallback != nil {
		fallbackPassword = fallback.Password
		fallbackAPIKey = fallback.APIKey
	}

	return &authModel{
		Username: types.StringPointerValue(apiUsername),
		Password: preserveRedactedSecret(apiPassword, fallbackPassword),
		APIKey:   preserveRedactedSecret(apiKey, fallbackAPIKey),
		Headers:  types.ListNull(authHeadersElementType(ctx)),
	}
}

func preserveRedactedSecret(apiValue *string, fallback types.String) types.String {
	switch {
	case apiValue != nil:
		return types.StringPointerValue(apiValue)
	case typeutils.IsKnown(fallback):
		return fallback
	default:
		return types.StringNull()
	}
}

func authHeadersElementType(ctx context.Context) attr.Type {
	authAttr := getSchema(ctx).Attributes["auth"].(schema.SingleNestedAttribute)
	return authAttr.Attributes["headers"].GetType().(attr.TypeWithElementType).ElementType()
}
