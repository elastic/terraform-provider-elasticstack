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
)

// updateSecretResolver returns the resolver for an update. When a configured
// write-only value has a stored hash it reads the source immediately before the
// PUT to learn the secret ids Fleet holds, so an unchanged value can be
// referenced by id instead of resent. Without a stored hash every value is a
// literal and the lookup is skipped.
func updateSecretResolver(
	ctx context.Context,
	client *fleet.Client,
	plan model,
	sourceID, spaceID string,
	storedHashes map[string][]byte,
) (secretResolver, diag.Diagnostics) {
	values, diags := plan.writeOnlyValues(ctx)
	if diags.HasError() {
		return nil, diags
	}

	var lookup bool
	for attrPath := range values {
		if len(storedHashes[attrPath]) > 0 {
			lookup = true
			break
		}
	}
	if !lookup {
		return literalSecrets, diags
	}

	ids, idDiags := fetchSecretIDs(ctx, client, sourceID, spaceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return nil, diags
	}
	return func(attrPath, value string) secretValue {
		return secretWriteBranch(value, storedHashes[attrPath], ids.forPath(attrPath))
	}, diags
}

// updateAgentDownloadSource sends the update. storedHashes holds the private
// state hash of each write-only value from its last successful apply, keyed by
// attribute path; plan carries the write-only values read from configuration.
func updateAgentDownloadSource(
	ctx context.Context,
	client *fleet.Client,
	plan model,
	prior model,
	storedHashes map[string][]byte,
) (entitycore.KibanaWriteResult[model], diag.Diagnostics) {
	var diags diag.Diagnostics

	sourceID := plan.SourceID.ValueString()

	// Read the existing spaces from prior state to determine where to update. When
	// space_ids changes (e.g. prepending a new space), we must query the space where
	// the resource currently exists (prior state), not where it will exist (plan).
	// Otherwise the API call fails with 404.
	spaceID := fleetutils.SpaceIDFromSet(prior.SpaceIDs)

	resolve, resolveDiags := updateSecretResolver(ctx, client, plan, sourceID, spaceID, storedHashes)
	diags.Append(resolveDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	body, bodyDiags := plan.toAPIUpdateModel(ctx, prior, resolve)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	updateResp, updateDiags := fleet.UpdateAgentDownloadSource(ctx, client, sourceID, spaceID, body)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	unwrapped, unwrapDiags := diagutil.UnwrapJSON200(updateResp.JSON200, "agent download source")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[model]{}, diags
	}

	item := unwrapped.Item
	written, writeDiags := finalizeWrite(ctx, client, item.Id, spaceID, plan, "Updated")
	diags.Append(writeDiags...)
	return written, diags
}
