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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fleetutils "github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createAgentDownloadSource(ctx context.Context, client *fleet.Client, plan model) (entitycore.KibanaWriteResult[model], diag.Diagnostics) {
	var diags diag.Diagnostics

	// Determine space from plan (first space_ids entry) for CREATE.
	spaceID := fleetutils.SpaceIDFromSet(plan.SpaceIDs)

	body := plan.toAPICreateModel(ctx)

	createResp, createDiags := fleet.CreateAgentDownloadSource(ctx, client, spaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	unwrapped, unwrapDiags := diagutil.UnwrapJSON200(createResp.JSON200, "agent download source")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	item := unwrapped.Item

	// Ensure we keep the operational space information consistent with how Read/Update/Delete will resolve it.
	if plan.SpaceIDs.IsUnknown() {
		plan.SpaceIDs = types.SetNull(types.StringType)
	}

	readState, found, readDiags := readAndHydrateState(ctx, client, item.Id, spaceID, plan.SpaceIDs, plan.KibanaConnection)
	diags.Append(readDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}
	if !found {
		diags.AddError("Unexpected API response", "Created agent download source could not be read back by source_id")
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	return entitycore.KibanaWriteResult[model]{Model: readState, SkipReadAfterWrite: true}, diags
}
