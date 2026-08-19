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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeleteByIDParams returns a [KibanaDeleteFunc] for resources whose delete
// request is a params struct built from the resourceID (and, optionally,
// fields read off the model, such as a namespace type) rather than the bare
// resourceID accepted by [SimpleKibanaDelete]. Use for example:
//
//	Delete: entitycore.DeleteByIDParams[exceptionListModel](
//	    func(id string, m exceptionListModel) *kbapi.DeleteExceptionListParams {
//	        params := &kbapi.DeleteExceptionListParams{Id: &id}
//	        ...
//	        return params
//	    },
//	    kibanaoapi.DeleteExceptionList,
//	),
func DeleteByIDParams[T KibanaResourceModel, P any](
	newParams func(id string, model T) *P,
	apiDelete func(ctx context.Context, client *kibanaoapi.Client, spaceID string, params *P) diag.Diagnostics,
) KibanaDeleteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, model T) diag.Diagnostics {
		return apiDelete(ctx, client.GetKibanaOapiClient(), spaceID, newParams(resourceID, model))
	}
}

// ReadByIDParams returns a [KibanaReadFunc] for resources whose read request
// is a params struct built from the resourceID rather than the bare
// resourceID accepted by [SimpleKibanaRead]. Resources that also need a retry
// with an agnostic namespace type should use
// [ReadByIDParamsWithAgnosticNamespaceRetry] instead.
func ReadByIDParams[T KibanaResourceModel, P any, R any](
	newParams func(id string) *P,
	apiGet func(ctx context.Context, client *kibanaoapi.Client, spaceID string, params *P) (*R, diag.Diagnostics),
	populate func(model *T, ctx context.Context, spaceID string, data *R) diag.Diagnostics,
) KibanaReadFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior T) (T, bool, diag.Diagnostics) {
		var diags diag.Diagnostics

		data, d := apiGet(ctx, client.GetKibanaOapiClient(), spaceID, newParams(resourceID))
		diags.Append(d...)
		if diags.HasError() {
			return prior, false, diags
		}

		if data == nil {
			return prior, false, diags
		}

		diags.Append(populate(&prior, ctx, spaceID, data)...)
		return prior, true, diags
	}
}

// ReadByIDParamsWithAgnosticNamespaceRetry returns a [KibanaReadFunc] for the
// security exception list/list-item resources: it builds read params from
// the resourceID plus an optional namespace-type field, and — when the
// namespace type was not known (for example during import) and the initial
// read reports not-found — retries once with namespace_type=agnostic before
// giving up. namespaceTypeOf reads the namespace type off the prior model;
// newParams builds the initial params; setAgnosticNamespace mutates params in
// place to request the agnostic namespace for the retry.
func ReadByIDParamsWithAgnosticNamespaceRetry[T KibanaResourceModel, P any, R any](
	namespaceTypeOf func(model T) types.String,
	newParams func(id string, namespaceType types.String) *P,
	setAgnosticNamespace func(params *P),
	apiGet func(ctx context.Context, client *kibanaoapi.Client, spaceID string, params *P) (*R, diag.Diagnostics),
	populate func(model *T, ctx context.Context, spaceID string, data *R) diag.Diagnostics,
) KibanaReadFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior T) (T, bool, diag.Diagnostics) {
		var diags diag.Diagnostics

		oapiClient := client.GetKibanaOapiClient()
		namespaceType := namespaceTypeOf(prior)
		params := newParams(resourceID, namespaceType)

		data, d := apiGet(ctx, oapiClient, spaceID, params)
		diags.Append(d...)
		if diags.HasError() {
			return prior, false, diags
		}

		// If namespace_type was not known (e.g., during import) and the item was
		// not found, retry once with namespace_type=agnostic.
		if data == nil && !typeutils.IsKnown(namespaceType) {
			setAgnosticNamespace(params)
			data, d = apiGet(ctx, oapiClient, spaceID, params)
			diags.Append(d...)
			if diags.HasError() {
				return prior, false, diags
			}
		}

		if data == nil {
			return prior, false, diags
		}

		diags.Append(populate(&prior, ctx, spaceID, data)...)
		return prior, true, diags
	}
}
