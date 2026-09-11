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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// writeAgentDownloadSource dispatches Create and Update through the envelope's
// shared write callback: req.Prior is nil for Create and non-nil for Update.
func writeAgentDownloadSource(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[model],
) (entitycore.KibanaWriteResult[model], diag.Diagnostics) {
	fleetClient := client.GetFleetClient()

	if req.Prior == nil {
		return createAgentDownloadSource(ctx, fleetClient, req.Plan)
	}
	return updateAgentDownloadSource(ctx, fleetClient, req.Plan, *req.Prior)
}

// finalizeWrite reads back the just-written source by ID and hydrates it into
// the final write result, or returns a diagnostic naming verb (e.g. "Created",
// "Updated") if the read-back doesn't find it.
func finalizeWrite(ctx context.Context, client *fleet.Client, sourceID, spaceID string, plan model, verb string) (entitycore.KibanaWriteResult[model], diag.Diagnostics) {
	var diags diag.Diagnostics

	readState, found, readDiags := readAndHydrateState(ctx, client, sourceID, spaceID, plan.SpaceIDs, plan.KibanaConnection)
	diags.Append(readDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}
	if !found {
		diags.AddError("Unexpected API response", verb+" agent download source could not be read back by source_id")
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	return entitycore.KibanaWriteResult[model]{Model: readState, SkipReadAfterWrite: true}, diags
}
