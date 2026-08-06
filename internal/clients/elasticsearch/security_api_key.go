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

package elasticsearch

import (
	"context"
	"fmt"
	"slices"

	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/invalidateapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updateapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updatecrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func CreateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createapikey.Request) (*createapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.CreateApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updateapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.UpdateApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update apikey", err.Error())
		return diags
	}

	return diags
}

// GetAPIKey reads the API key identified by id. owner mirrors the resource's
// `owner` attribute: when true, the request is scoped to keys owned by the
// currently authenticated user (matching the `owner` semantics used by
// DeleteAPIKey), so a key that exists but is owned by a different user comes
// back empty, which this function treats the same as "not found" - i.e. the
// resource is treated as non-existent rather than erroring.
func GetAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, owner bool) (*types.ApiKey, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	req := typedClient.Security.GetApiKey().Id(id)
	if owner {
		req = req.Owner(true)
	}

	res, err := req.Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Unable to get an apikey", err.Error())
		return nil, diags
	}

	if len(res.ApiKeys) == 0 {
		// Not found, or (when owner is true) found but not owned by the
		// current authenticated user. Both cases are treated as "does not
		// exist" so the resource disappears from state instead of erroring.
		return nil, diags
	}

	if len(res.ApiKeys) != 1 {
		diags.AddError(
			"Unable to find an apikey in the cluster",
			fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
		)
		return nil, diags
	}

	apiKey := res.ApiKeys[0]
	return &apiKey, diags
}

// DeleteAPIKey invalidates the API key identified by id. owner controls the
// `owner` flag sent on the Invalidate API Key request: Elasticsearch only
// authorizes an id-scoped invalidate request under the `manage_own_api_key`
// cluster privilege when `owner` is `true` (in which case it is understood to
// target only keys owned by the calling user); invalidating by id with
// `owner: false` (or omitted) requires the broader `manage_api_key` privilege.
func DeleteAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, owner bool) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.InvalidateApiKey().Request(&invalidateapikey.Request{
		Ids:   []string{id},
		Owner: &owner,
	}).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete an apikey", err.Error())
		return diags
	}

	if !apiKeyIDInvalidated(res, id) {
		diags.AddError(
			"Unable to delete an apikey",
			fmt.Sprintf(
				`Elasticsearch did not report "%s" as invalidated in the invalidate API key response (invalidated_api_keys/previously_invalidated_api_keys). `+
					`It may be owned by a different user; if so, set "owner" to false and ensure the connection is granted "manage_api_key" to delete keys owned by other users.`,
				id,
			),
		)
	}

	return diags
}

// apiKeyIDInvalidated reports whether id appears in either the
// invalidated_api_keys or previously_invalidated_api_keys lists of an
// Invalidate API Key response. Elasticsearch can return a 200 response with
// error_count == 0 while silently dropping ids that don't match the request's
// filters (e.g. an `owner` mismatch), so the response body must be checked
// rather than relying solely on the absence of a transport error.
func apiKeyIDInvalidated(res *invalidateapikey.Response, id string) bool {
	if res == nil {
		return false
	}
	if slices.Contains(res.InvalidatedApiKeys, id) {
		return true
	}
	return slices.Contains(res.PreviouslyInvalidatedApiKeys, id)
}

func CreateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createcrossclusterapikey.Request) (*createcrossclusterapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.CreateCrossClusterApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create cross cluster apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updatecrossclusterapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient := apiClient.GetESClient()

	_, err := typedClient.Security.UpdateCrossClusterApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update cross cluster apikey", err.Error())
		return diags
	}

	return diags
}
