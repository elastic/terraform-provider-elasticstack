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
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	return readAndHydrateState(ctx, fleetClient, resourceID, spaceID, prior.SpaceIDs, prior.KibanaConnection, prior.Auth, prior.Ssl)
}

// secretIDs holds the ids of the managed secrets Fleet currently stores for a
// download source. A nil field means Fleet reports no secret id for it, for
// example because secret storage is unavailable and the value was merged into
// the plain field instead.
type secretIDs struct {
	sslKey       *string
	authPassword *string
	authAPIKey   *string
}

func (ids secretIDs) forPath(attrPath string) *string {
	switch attrPath {
	case sslKeyWOPath:
		return ids.sslKey
	case authPasswordWOPath:
		return ids.authPassword
	case authAPIKeyWOPath:
		return ids.authAPIKey
	default:
		return nil
	}
}

// fetchSecretIDs reads the source by id and extracts the secret ids Fleet
// currently holds, so an update can reference an unchanged secret rather than
// resend its literal. A source that no longer exists yields no ids.
func fetchSecretIDs(ctx context.Context, client *fleet.Client, sourceID, spaceID string) (secretIDs, diag.Diagnostics) {
	var ids secretIDs

	getResp, diags := fleet.GetAgentDownloadSource(ctx, client, sourceID, spaceID)
	if diags.HasError() || getResp == nil {
		return ids, diags
	}
	body, unwrapDiags := diagutil.UnwrapJSON200(getResp.JSON200, "agent download source")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return ids, diags
	}

	secrets := body.Item.Secrets
	if secrets == nil {
		return ids, diags
	}
	if secrets.Ssl != nil && secrets.Ssl.Key != nil {
		if v, err := secrets.Ssl.Key.AsGetFleetAgentDownloadSourcesSourceid200JSONResponseBodyItemSecretsSslKey0(); err == nil {
			ids.sslKey = nonEmptyIDPointer(v.Id)
		}
	}
	if secrets.Auth != nil {
		if secrets.Auth.Password != nil {
			if v, err := secrets.Auth.Password.AsGetFleetAgentDownloadSourcesSourceid200JSONResponseBodyItemSecretsAuthPassword0(); err == nil {
				ids.authPassword = nonEmptyIDPointer(v.Id)
			}
		}
		if secrets.Auth.ApiKey != nil {
			if v, err := secrets.Auth.ApiKey.AsGetFleetAgentDownloadSourcesSourceid200JSONResponseBodyItemSecretsAuthApiKey0(); err == nil {
				ids.authAPIKey = nonEmptyIDPointer(v.Id)
			}
		}
	}
	return ids, diags
}

func nonEmptyIDPointer(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}

func readAndHydrateState(
	ctx context.Context,
	client *fleet.Client,
	sourceID string,
	spaceID string,
	preservedSpaceIDs types.Set,
	preservedKibanaConnection types.List,
	fallbackAuth *authModel,
	fallbackSsl types.Object,
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
	var apiSsl agentDownloadSourceSSLAPIModel
	if item.Ssl != nil {
		apiSsl = agentDownloadSourceSSLAPIModel{
			Certificate:            item.Ssl.Certificate,
			CertificateAuthorities: item.Ssl.CertificateAuthorities,
			Key:                    item.Ssl.Key,
		}
	}
	sslObj, sslDiags := hydrateSsl(ctx, apiSsl, fallbackSsl)
	diags.Append(sslDiags...)
	if diags.HasError() {
		return state, false, diags
	}
	state.Ssl = sslObj

	auth, authDiags := hydrateAuth(ctx, item.Auth, fallbackAuth)
	diags.Append(authDiags...)
	state.Auth = auth
	return state, true, diags
}

func hydrateAuth(ctx context.Context, apiAuth *downloadSourceAuth, fallback *authModel) (*authModel, diag.Diagnostics) {
	if apiAuth == nil && fallback == nil {
		return nil, nil
	}

	var api downloadSourceAuth
	if apiAuth != nil {
		api = *apiAuth
	}
	var prior authModel
	if fallback != nil {
		prior = *fallback
	}

	headers, diags := authHeadersFromAPI(ctx, api.Headers)
	// Write-only values never reach state, and the API echoes them in the
	// plain fields when secret storage is unavailable, so a plain credential
	// is only hydrated when the prior plan or state configured it.
	hasPrior := fallback != nil
	return &authModel{
		Username: types.StringPointerValue(api.Username),
		Password: hydratePlainSecret(api.Password, prior.Password, hasPrior),
		APIKey:   hydratePlainSecret(api.ApiKey, prior.APIKey, hasPrior),
		Headers:  headers,
	}, diags
}

// hydratePlainSecret hydrates a plain secret attribute. With no prior value
// (import) the API value is used. With a prior object that did not configure
// the attribute, the API value is ignored because it may be an echoed
// write-only literal. Otherwise a redacted API value keeps the prior value.
func hydratePlainSecret(apiValue *string, prior types.String, hasPrior bool) types.String {
	if hasPrior && !typeutils.IsKnown(prior) {
		return types.StringNull()
	}
	return typeutils.PreserveRedactedSecret(apiValue, prior)
}

func authHeadersFromAPI(ctx context.Context, apiHeaders *[]downloadSourceAuthHeader) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	headers := typeutils.SliceToListType(ctx, typeutils.Deref(apiHeaders), authHeadersElementType(ctx), path.Root("auth").AtName("headers"), &diags,
		func(header downloadSourceAuthHeader, _ typeutils.ListMeta) authHeaderModel {
			return authHeaderModel{Key: types.StringValue(header.Key), Value: types.StringValue(header.Value)}
		})
	return headers, diags
}

func authHeadersElementType(ctx context.Context) attr.Type {
	return typeutils.ElementTypeOf(getSchema(ctx).Attributes["auth"].(schema.SingleNestedAttribute).Attributes["headers"].GetType())
}
