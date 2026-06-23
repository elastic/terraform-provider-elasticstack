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

// Package agentbuildercommon provides shared read and delete helpers used by
// the four AgentBuilder entity packages (agent, skill, tool, workflow) to
// eliminate boilerplate CRUD duplication.
package agentbuildercommon

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ReadEntity fetches an AgentBuilder entity by ID and applies it to a prior
// model. getFn is called with the oapi client, spaceID, and resourceID; it
// returns (nil, nil-diags) when the entity is not found (404). populateFn
// receives a pointer to a copy of prior and the fetched API data, and must
// populate it; its diagnostics are appended. The returned bool is false when
// the entity was not found or when an error diagnostic was produced.
func ReadEntity[M any, A any](
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID, spaceID string,
	prior M,
	getFn func(context.Context, *kibanaoapi.Client, string, string) (*A, diag.Diagnostics),
	populateFn func(*M, *A) diag.Diagnostics,
) (M, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	entity, d := getFn(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if entity == nil {
		return prior, false, diags
	}

	diags.Append(populateFn(&prior, entity)...)
	return prior, true, diags
}

// DeleteEntity deletes an AgentBuilder entity by calling deleteFn with the
// oapi client, spaceID, and resourceID.
func DeleteEntity(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID, spaceID string,
	deleteFn func(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics,
) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()
	return deleteFn(ctx, oapiClient, spaceID, resourceID)
}
