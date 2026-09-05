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
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/getapikey"
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

// GetAPIKey reads the API key identified by id. restrictToOwned mirrors the
// resource's `restrict_to_owned` attribute: when true, the request is scoped
// to keys owned by the currently authenticated user, so a key that exists
// but is owned by a different user comes back empty, which this function
// treats the same as "not found" - i.e. the resource is treated as
// non-existent rather than erroring. When false (the default), the lookup is
// not scoped by owner at all.
func GetAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, restrictToOwned bool) (*types.ApiKey, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, diags := CallOrNotFound(func() (*getapikey.Response, error) {
		req := typedClient.Security.GetApiKey().Id(id)
		if restrictToOwned {
			req = req.Owner(true)
		}
		return req.Do(ctx)
	}, "Unable to get an apikey")
	if diags.HasError() || res == nil {
		return nil, diags
	}

	if len(res.ApiKeys) == 0 {
		// Not found, or (when restrictToOwned is true) found but not owned
		// by the current authenticated user. Both cases are treated as
		// "does not exist" so the resource disappears from state instead of
		// erroring.
		return nil, diags
	}

	return SingleOrNotFoundDiag(res.ApiKeys, id, "apikey")
}

// DeleteAPIKey invalidates the API key identified by id.
//
// Elasticsearch only authorizes an id-scoped invalidate request under the
// narrower `manage_own_api_key` cluster privilege when `owner: true` is set
// on the request (in which case it only takes effect if the key is actually
// owned by the calling user); invalidating by id with `owner: false` (or
// omitted) requires the broader `manage_api_key` privilege but works
// regardless of who owns the key.
//
// To make the common "delete my own key" case work with just
// `manage_own_api_key`, without requiring `manage_api_key` for every caller,
// DeleteAPIKey first attempts the invalidate with `owner: true`. If that
// doesn't invalidate the key (for example because it's owned by a different
// user, or the caller lacks `manage_own_api_key`), and restrictToOwned is
// false, it retries with `owner: false`. An error is only reported if that
// second attempt also fails. When restrictToOwned is true (the resource's
// `restrict_to_owned` attribute), no fallback is attempted: the key is only
// ever invalidated if it is owned by the calling user, so a key owned by
// someone else is never touched.
func DeleteAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, restrictToOwned bool) fwdiag.Diagnostics {
	invalidated, err := invalidateAPIKey(ctx, apiClient, id, true)
	if err != nil && IsNotFoundElasticsearchError(err) {
		return nil
	}
	if err == nil && invalidated {
		return nil
	}

	if restrictToOwned {
		if err != nil {
			return DeleteWithNotFoundAsSuccess(err, "Unable to delete an apikey")
		}
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"Unable to delete an apikey",
			fmt.Sprintf(
				`Elasticsearch did not report "%s" as invalidated when scoped to the current authenticated user (owner=true). `+
					`It may be owned by a different user; set "restrict_to_owned" to false to allow deleting keys owned by other users (requires the "manage_api_key" cluster privilege).`,
				id,
			),
		)}
	}

	// Fall back to an unscoped invalidate request. This requires the
	// broader `manage_api_key` privilege but succeeds regardless of who
	// owns the key.
	invalidated, err = invalidateAPIKey(ctx, apiClient, id, false)
	if err != nil {
		return DeleteWithNotFoundAsSuccess(err, "Unable to delete an apikey")
	}

	if !invalidated {
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(
			"Unable to delete an apikey",
			fmt.Sprintf(
				`Elasticsearch did not report "%s" as invalidated in the invalidate API key response (invalidated_api_keys/previously_invalidated_api_keys).`,
				id,
			),
		)}
	}

	return nil
}

// invalidateAPIKey issues a single Invalidate API Key request for id with the
// given owner flag, returning whether id was reported as invalidated.
func invalidateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, owner bool) (bool, error) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Security.InvalidateApiKey().Request(&invalidateapikey.Request{
		Ids:   []string{id},
		Owner: &owner,
	}).Do(ctx)
	if err != nil {
		return false, err
	}

	return apiKeyIDInvalidated(res, id), nil
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
