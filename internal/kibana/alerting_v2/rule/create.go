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

package rule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alerting_v2/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const ruleNoun = "rule"

func createRule(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	// A generated id is the dangerous case: if the create cannot be completed
	// and no state is saved, a retry mints a new id and creates a second,
	// unmanaged rule. For those, roll the orphan back so a retry stays safe. A
	// supplied id needs no rollback: a retry hits the existence check below and
	// is told to import.
	generatedID := !typeutils.IsKnown(plan.RuleID)
	ruleID := plan.RuleID.ValueString()
	if generatedID {
		ruleID = uuid.NewString()
	}
	importID := clients.CompositeIDValue(req.SpaceID, ruleID).ValueString()

	existing, getDiags := kibanaoapi.GetAlertingV2Rule(ctx, client.GetKibanaOapiClient(), req.SpaceID, ruleID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, diags
	}
	if existing != nil {
		diags.Append(utils.AlreadyExistsDiagnostic(ruleNoun, importID))
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	created, _, putDiags := putRule(ctx, client, req.SpaceID, ruleID, plan)
	diags.Append(putDiags...)
	if diags.HasError() {
		// Roll back on any write error, not only when the server reported the
		// rule created. A transport error reports created=false yet Kibana may
		// have committed the PUT before the connection dropped. The id was
		// confirmed absent above, so an idempotent DELETE (404 is success) is
		// safe whether or not the write committed.
		if generatedID {
			diags.Append(rollbackOrphanedRule(ctx, client, req.SpaceID, ruleID)...)
		}
		return entitycore.KibanaWriteResult[Model]{}, diags
	}
	if !created {
		diags.Append(utils.AlreadyReplacedDiagnostic(ruleNoun, importID))
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	// Own the read-after-write so a read failure on a generated id can be
	// compensated before the envelope discards the unsaved state. This replaces
	// the envelope's own read (SkipReadAfterWrite), so the call count is
	// unchanged.
	state, found, readDiags := readRule(ctx, client, ruleID, req.SpaceID, plan)
	diags.Append(readDiags...)
	if diags.HasError() || !found {
		// A read error or a rule that is unexpectedly absent after a successful
		// create both leave no usable state. Roll the orphan back so a retry of
		// a generated id does not create a duplicate.
		if generatedID {
			diags.Append(rollbackOrphanedRule(ctx, client, req.SpaceID, ruleID)...)
		}
		if !found && !readDiags.HasError() {
			diags.AddError(
				"Alerting v2 rule not found after creation",
				"The rule \""+importID+"\" was created but could not be read back. It may have been deleted concurrently; run a new plan to reconcile.",
			)
		}
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	return entitycore.KibanaWriteResult[Model]{Model: state, SkipReadAfterWrite: true}, diags
}

// rollbackOrphanedRule deletes a rule left behind by a create that could not be
// completed, so a retry does not create a duplicate. A failed rollback is
// reported as its own error because the rule then survives unmanaged and the
// user must act on it.
func rollbackOrphanedRule(ctx context.Context, client *clients.KibanaScopedClient, spaceID, ruleID string) diag.Diagnostics {
	if delDiags := kibanaoapi.DeleteAlertingV2Rule(ctx, client.GetKibanaOapiClient(), spaceID, ruleID); delDiags.HasError() {
		importID := clients.CompositeIDValue(spaceID, ruleID).ValueString()
		var out diag.Diagnostics
		out.AddError(
			"Failed to roll back a partially created Alerting v2 rule",
			"The rule \""+importID+"\" was created but the create could not be completed, and the automatic cleanup also failed, so it now exists unmanaged. "+
				"Delete it manually or import it with `terraform import`, then retry.",
		)
		return out
	}
	return nil
}

// putRule issues the create-or-replace write. created reports whether the server answered 201 (created) rather than 200 (replaced).
func putRule(ctx context.Context, client *clients.KibanaScopedClient, spaceID, ruleID string, plan Model) (created bool, written Model, diags diag.Diagnostics) {
	body, bodyDiags := plan.toAPIPutBody(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return false, plan, diags
	}

	created, putDiags := kibanaoapi.PutAlertingV2Rule(ctx, client.GetKibanaOapiClient(), spaceID, ruleID, body)
	// Prefix the write diagnostics with the composite id, as readRule does, so
	// a failed write still names the rule to import. created is returned even
	// on error: a PUT that reaches the server but returns an empty or malformed
	// body errors here after the rule was created, and the caller needs that
	// signal to roll the orphan back.
	diags.Append(withCompositeID(putDiags, spaceID, ruleID)...)
	if diags.HasError() {
		return created, plan, diags
	}

	plan.ID, plan.SpaceID = clients.CompositeIdentityValues(spaceID, ruleID)
	plan.RuleID = types.StringValue(ruleID)
	return created, plan, diags
}
