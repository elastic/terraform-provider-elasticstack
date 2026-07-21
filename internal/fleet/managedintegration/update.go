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

package managedintegration

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// onlyCreateOnlyFlagsChanged reports whether prior and plan are identical in
// every attribute except create_dataset_templates, force, force_delete, and
// skip_topology_check. Those four are create/delete-request-only knobs --
// never part of the Fleet API's read response, and deliberately not
// RequiresReplace (see spec.md's "Schema attributes" requirement and this
// file's updateAgentlessPolicy) -- so a config change confined to them
// carries no information the API needs to see, and per spec.md's "Create"
// requirement changing them "SHALL NOT make any API call".
// skip_topology_check in particular is consulted only by Create's preflight
// check (see create.go) and is never read by updateAgentlessPolicy, so it
// gets the same treatment. Comparing every other field (rather than
// allowlisting "the fields Update actually sends") is deliberately
// conservative: if anything else in the model also drifted -- including
// drift entitycore/Terraform itself wouldn't have produced, such as a
// concurrent out-of-band edit surfaced by a prior Read -- this returns false
// and the normal PUT path below runs instead.
//
// kibana_connection (ResourceTimeoutsField) and the Timeouts block are
// intentionally excluded: they are pure provider-side plumbing that is never
// part of the Fleet request body either, so their presence or absence has no
// bearing on whether an API call is needed.
//
// created_at and updated_at are also intentionally excluded, for a different
// reason than the plumbing fields above: they are purely server-Computed
// (never Optional -- a user's config can never set or influence them), so
// they can never actually be *what changed* between prior and plan; comparing
// them here would only ever be testing an artifact of how the Plugin
// Framework happened to resolve their plan value, not a real signal of user
// intent. That distinction used to matter concretely: created_at/updated_at
// are Computed with no plan modifier that forces them to a known value, so
// the framework marks them Unknown in the plan for every Update regardless of
// whether they are "really" changing -- which made a naive
// prior.CreatedAt.Equal(plan.CreatedAt)-style comparison always false in a
// real Terraform plan (Unknown never equals a known value), permanently
// defeating this short-circuit outside of unit tests that hand-built a plan
// with matching known timestamps. Excluding both fields here fixes that
// directly, without depending on schema.go's plan modifiers lining up a
// particular way. (created_at additionally now carries UseStateForUnknown in
// schema.go, since it never legitimately changes after creation -- but
// updated_at deliberately does NOT, since it genuinely changes on every real
// Update; see schema.go's updated_at comment for why forcing it to look
// unchanged in the plan would be actively wrong.)
func onlyCreateOnlyFlagsChanged(prior, plan agentlessPolicyModel) bool {
	return prior.ID.Equal(plan.ID) &&
		prior.PolicyID.Equal(plan.PolicyID) &&
		prior.Name.Equal(plan.Name) &&
		prior.Description.Equal(plan.Description) &&
		prior.Namespace.Equal(plan.Namespace) &&
		prior.SpaceIDs.Equal(plan.SpaceIDs) &&
		prior.Package.Equal(plan.Package) &&
		prior.PolicyTemplate.Equal(plan.PolicyTemplate) &&
		prior.VarsJSON.Equal(plan.VarsJSON) &&
		prior.VarGroupSelections.Equal(plan.VarGroupSelections) &&
		prior.Inputs.Equal(plan.Inputs) &&
		prior.CloudConnector.Equal(plan.CloudConnector) &&
		prior.GlobalDataTags.Equal(plan.GlobalDataTags) &&
		prior.AdditionalDatastreamsPermissions.Equal(plan.AdditionalDatastreamsPermissions)
}

// updateAgentlessPolicy implements Update for managed integrations. The PUT
// body is a full replace built from the plan (see buildUpdateBody); prior
// state supplies cloud_connector association fields only. The callback returns
// the plan model only; per coding-standards.md the entitycore envelope's
// read-after-write refresh (Read callback) is the sole source of persisted
// state after a real PUT. Mutate responses are not merged into state here.
func updateAgentlessPolicy(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	if req.Prior != nil && onlyCreateOnlyFlagsChanged(*req.Prior, plan) {
		model := preserveKnownComputedFromPrior(ctx, plan, *req.Prior)
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{
			Model:              model,
			SkipReadAfterWrite: true,
		}, diags
	}

	if req.Prior == nil {
		diags.AddError(
			"Managed integration update missing prior state",
			"Internal error: Update requires prior state to preserve cloud_connector on full-replace PUT.",
		)
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	body, bodyDiags := buildUpdateBody(ctx, plan, *req.Prior)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	fleetClient := client.GetFleetClient()

	// TEMPORARY (task 8): PUT already targets managed_integrations; Read on
	// the entitycore read-after-write path still uses the package_policies
	// compat wrapper until task 8 deletes agentless_policy_compat.go.
	updated, updateDiags := fleetclient.UpdateManagedIntegration(ctx, fleetClient, req.SpaceID, req.WriteID, body)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}
	if updated == nil {
		diags.AddError(
			"Managed integration not found",
			fmt.Sprintf("Cannot update managed integration %q: it was not found. It may have been deleted out of band; "+
				"run terraform apply again to detect drift.", req.WriteID),
		)
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	return entitycore.KibanaWriteResult[agentlessPolicyModel]{Model: plan}, diags
}

// buildUpdateBody builds the managed_integrations full-replace PUT body from
// the plan. cloud_connector {enabled, cloud_connector_id} is taken from prior
// state (not plan) so the association is re-sent without write-only fields.
//
// Full-replace semantics: optional request fields use generated `omitempty` JSON
// tags. On Update, a known-null optional attribute is omitted from the body
// (not sent as an empty string or empty collection), which clears the field on
// the API. A known-empty collection (for example `vars_json = jsonencode({})`)
// is sent explicitly when sendExplicitEmptyScalars applies. Unknown top-level
// API-backed optional attributes produce attribute errors (see
// diagnoseUnknownUpdatePlanFields). Known-null optional attributes are omitted
// (omitempty) to clear on full replace.
func buildUpdateBody(ctx context.Context, plan, prior agentlessPolicyModel) (kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody, diag.Diagnostics) {
	return plan.toManagedIntegrationRequestBody(ctx, managedIntegrationRequestOptions{
		omitCreateOnlyFields:     true,
		priorForCloudConnector:   &prior,
		sendExplicitEmptyScalars: true,
	})
}

// preserveKnownComputedFromPrior copies known server-computed values from prior
// state into plan when the Update plan leaves them Unknown (typical for
// updated_at). Used only with SkipReadAfterWrite so direct state write does not
// persist Unknown computed attributes.
func preserveKnownComputedFromPrior(ctx context.Context, plan, prior agentlessPolicyModel) agentlessPolicyModel {
	out := plan
	out.UpdatedAt = entitycore.PreserveStringFromPriorIfUnknown(out.UpdatedAt, prior.UpdatedAt)
	out.CreatedAt = entitycore.PreserveStringFromPriorIfUnknown(out.CreatedAt, prior.CreatedAt)
	out.ID = entitycore.PreserveStringFromPriorIfUnknown(out.ID, prior.ID)
	out.PolicyID = entitycore.PreserveStringFromPriorIfUnknown(out.PolicyID, prior.PolicyID)
	if out.SpaceIDs.IsUnknown() && !prior.SpaceIDs.IsUnknown() {
		out.SpaceIDs = prior.SpaceIDs
	}
	if out.VarsJSON.IsUnknown() && !prior.VarsJSON.IsUnknown() {
		out.VarsJSON = prior.VarsJSON
	}
	if out.Package.IsUnknown() && !prior.Package.IsUnknown() {
		out.Package = prior.Package
		return out
	}
	if !out.Package.IsUnknown() && !prior.Package.IsUnknown() {
		var planPkg, priorPkg packageModel
		if !plan.Package.As(ctx, &planPkg, basetypes.ObjectAsOptions{}).HasError() &&
			!prior.Package.As(ctx, &priorPkg, basetypes.ObjectAsOptions{}).HasError() &&
			planPkg.Title.IsUnknown() && typeutils.IsKnown(priorPkg.Title) {
			planPkg.Title = priorPkg.Title
			if pkgObj, d := types.ObjectValueFrom(ctx, packageAttrTypes(), planPkg); !d.HasError() {
				out.Package = pkgObj
			}
		}
	}
	return out
}
