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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// onlyCreateOnlyFlagsChanged is true when prior and plan match on every
// attribute except create_dataset_templates, force, force_delete, and
// skip_topology_check. Those knobs are create/delete-only (not in the Fleet
// read body); changing them alone must not trigger PUT (see
// openspec/specs/fleet-managed-integration/spec.md). All other model fields
// are compared so unrelated drift still runs the normal update path.
//
// Not compared: kibana_connection and Timeouts (provider-only plumbing);
// created_at and updated_at (Computed-only — users cannot change them, and
// updated_at is often Unknown in the plan so a naive Equal would always fail).
func onlyCreateOnlyFlagsChanged(prior, plan managedIntegrationModel) bool {
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

// updateManagedIntegration implements Update for managed integrations. The PUT
// body is a full replace built from the plan (see buildUpdateBody); prior
// state supplies cloud_connector association fields only. The callback returns
// the plan model only; per coding-standards.md the entitycore envelope's
// read-after-write refresh (Read callback) is the sole source of persisted
// state after a real PUT. Mutate responses are not merged into state here.
func updateManagedIntegration(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[managedIntegrationModel],
) (entitycore.KibanaWriteResult[managedIntegrationModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	if req.Prior != nil && onlyCreateOnlyFlagsChanged(*req.Prior, plan) {
		model := preserveKnownComputedFromPrior(ctx, plan, *req.Prior)
		return entitycore.KibanaWriteResult[managedIntegrationModel]{
			Model:              model,
			SkipReadAfterWrite: true,
		}, diags
	}

	if req.Prior == nil {
		diags.AddError(
			"Managed integration update missing prior state",
			"Internal error: Update requires prior state to preserve cloud_connector on full-replace PUT.",
		)
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}

	body, bodyDiags := buildUpdateBody(ctx, plan, *req.Prior)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}

	fleetClient := client.GetFleetClient()

	_, updateDiags := fleetclient.UpdateManagedIntegration(ctx, fleetClient, req.SpaceID, req.WriteID, body)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[managedIntegrationModel]{}, diags
	}

	return entitycore.KibanaWriteResult[managedIntegrationModel]{Model: plan}, diags
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
func buildUpdateBody(ctx context.Context, plan, prior managedIntegrationModel) (kbapi.PutFleetManagedIntegrationsPolicyidJSONRequestBody, diag.Diagnostics) {
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
func preserveKnownComputedFromPrior(ctx context.Context, plan, prior managedIntegrationModel) managedIntegrationModel {
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
