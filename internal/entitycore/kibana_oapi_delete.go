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

// DeleteWithOapiClient acquires the Kibana OAPI client from client, guards
// against errors, then delegates to fn with ctx, the OAPI client, spaceID,
// and resourceID. This centralises the repeated acquire-guard-delegate pattern
// shared by Kibana delete functions.
func DeleteWithOapiClient(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	spaceID, resourceID string,
	fn func(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics,
) diag.Diagnostics {
	oapiClient, getDiags := client.GetKibanaOapiClient()
	if getDiags.HasError() {
		return getDiags
	}
	return fn(ctx, oapiClient, spaceID, resourceID)
}
