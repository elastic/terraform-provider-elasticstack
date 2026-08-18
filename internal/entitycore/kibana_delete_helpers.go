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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SimpleKibanaDelete returns a [KibanaDeleteFunc] that resolves the scoped
// Kibana OAPI client and delegates directly to apiDelete. Use for resources
// whose delete callback needs nothing beyond client, spaceID, and resourceID,
// for example:
//
//	Delete: entitycore.SimpleKibanaDelete[agentModel](kibanaoapi.DeleteAgent),
func SimpleKibanaDelete[T KibanaResourceModel](
	apiDelete func(ctx context.Context, client *kibanaoapi.Client, spaceID, resourceID string) diag.Diagnostics,
) KibanaDeleteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, _ T) diag.Diagnostics {
		return apiDelete(ctx, client.GetKibanaOapiClient(), spaceID, resourceID)
	}
}

// SimpleKibanaDeleteWithParams is [SimpleKibanaDelete] for Kibana APIs whose
// delete endpoint takes a per-resource params struct (for example
// *kbapi.DeleteListParams) instead of a bare resourceID string. buildParams
// constructs that struct from resourceID and the prior model:
//
//	Delete: entitycore.SimpleKibanaDeleteWithParams[Model, kbapi.DeleteListParams](buildDeleteListParams, kibanaoapi.DeleteList),
func SimpleKibanaDeleteWithParams[T KibanaResourceModel, P any](
	buildParams func(resourceID string, prior T) *P,
	apiDelete func(ctx context.Context, client *kibanaoapi.Client, spaceID string, params *P) diag.Diagnostics,
) KibanaDeleteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior T) diag.Diagnostics {
		return apiDelete(ctx, client.GetKibanaOapiClient(), spaceID, buildParams(resourceID, prior))
	}
}
