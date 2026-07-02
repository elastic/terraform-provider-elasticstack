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

// Spike findings summary (fleet-agentless-policy OpenSpec change, Task 3.3,
// 2026-07-01): an empirical probe against a live Kibana 9.4.3 Cloud Hosted
// deployment confirmed that PUT /api/fleet/package_policies/{id} accepts and
// persists the Decision 3 in-place-updatable allowlist (description,
// vars_json, inputs[*].streams[*].vars, global_data_tags,
// additional_datastreams_permissions, var_group_selections, package.title)
// using the "typed" (array) request body -- the "mapped"/simplified body
// used elsewhere in this resource is NOT safe here, since it silently reset
// every input's `enabled` flag on an existing agentless-created policy. The
// one surprising finding: `inputs[*].enabled` was accepted (200) but not
// reliably persisted for the package under test (see overlayInputFromPlan's
// comment for how this resource handles that). The probe also found that
// Kibana's PUT endpoint does NOT actually reject changes to
// name/namespace/package.version/package.name (Decision 3's RequiresReplace
// fields) -- that partitioning is kept anyway as a deliberate Terraform-side
// safety choice, not an API constraint. See
// openspec/changes/archive/2026-07-02-fleet-agentless-policy/design.md,
// Decision 3, for the full empirical investigation and rationale.

package agentlesspolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// mergeVarsInto replaces *dst, an existing `{frozen,type,value}`-shaped typed
// vars map (or nil), with a map containing exactly planVars's keys (a flat
// map[string]any parsed from vars_json or an input/stream's `vars`,
// including a nil or empty map). planVars is the authoritative desired key
// set -- not an incremental overlay -- so any key present in *dst but absent
// from planVars is dropped, matching this file's "send fields explicitly to
// actively clear" design (see buildUpdateBody's comment). For each key that
// does survive, *dst's existing entry (if any) supplies the `type`/`frozen`
// metadata, which the plan's flat vars map does not carry; only `value` is
// taken from planVars. A nil or empty planVars therefore correctly produces
// an empty (non-nil) vars object, so callers no longer need a separate
// "clear" path: passing a nil planVars here does the same thing.
//
// Uses a JSON round trip throughout so the anonymous `*map[string]struct{...}`
// field type of the caller's choosing (see policyshape.TypedVarEntry's doc
// comment) never needs to be spelled out; V is inferred from the caller's
// argument.
func mergeVarsInto[V any](dst **V, planVars map[string]any, diags *diag.Diagnostics) {
	existing := map[string]policyshape.TypedVarEntry{}
	if *dst != nil {
		if b, err := json.Marshal(*dst); err == nil {
			_ = json.Unmarshal(b, &existing)
		}
	}

	result := make(map[string]policyshape.TypedVarEntry, len(planVars))
	for k, v := range planVars {
		e := existing[k] // zero value (no type/frozen) if k is a brand new key
		e.Value = v
		result[k] = e
	}

	b, err := json.Marshal(result)
	if err != nil {
		diags.AddError("Failed to encode vars for the update request", err.Error())
		return
	}

	// *dst must be reset to nil before unmarshaling: json.Unmarshal decodes a
	// JSON object into an *existing* non-nil map by merging keys, never
	// removing ones absent from the new JSON. result's key set is now the
	// complete desired key set, so *dst must be replaced wholesale rather
	// than merged into, or keys dropped from planVars would silently survive.
	*dst = nil
	if err := json.Unmarshal(b, dst); err != nil {
		diags.AddError("Failed to encode vars for the update request", err.Error())
	}
}

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
// and the normal GET+PUT path below runs instead.
//
// kibana_connection (ResourceTimeoutsField) and the Timeouts block are
// intentionally excluded: they are pure provider-side plumbing that is never
// part of the Fleet request body either, so their presence or absence has no
// bearing on whether an API call is needed.
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
		prior.AdditionalDatastreamsPermissions.Equal(plan.AdditionalDatastreamsPermissions) &&
		prior.CreatedAt.Equal(plan.CreatedAt) &&
		prior.UpdatedAt.Equal(plan.UpdatedAt)
}

// updateAgentlessPolicy implements Task 5.3 of the fleet-agentless-policy
// OpenSpec change: calls fleetclient.UpdateAgentlessPolicyViaPackagePolicy
// (PUT /api/fleet/package_policies/{id}, space-aware) with only the
// in-place-updatable allowlist fields from Decision 3 (description,
// vars_json, var_group_selections, inputs, global_data_tags,
// additional_datastreams_permissions, package.title), informed by the spike
// findings in this file's header comment.
//
// The spike found that the "mapped"/simplified request body -- used
// everywhere else in this resource (Create, Read) -- is NOT safe for
// updating `inputs`: it silently reset every input's `enabled` flag, even
// for inputs absent from the request. The "typed" (array) body is required
// instead, built by fetching a fresh typed snapshot
// (fleetclient.GetDefendPackagePolicy, which -- despite its Defend-specific
// name -- is a generic helper already used generically: it is simply
// GetFleetPackagePoliciesPackagepolicyidWithResponse without the
// Format=Simplified query param) and echoing it back with only the
// allowlisted fields overlaid, mirroring the spike's own "mutate a fresh
// GET, PUT it back" methodology.
func updateAgentlessPolicy(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	// spec.md's "Create" requirement ("create_dataset_templates sent only on
	// create" scenario) and the "Operation flags" schema section are explicit
	// that create_dataset_templates, force, and force_delete are
	// create/delete-only knobs that are never read back from the API: a
	// config-only change to one of them "SHALL NOT make any API call". Since
	// none of the three is RequiresReplace, Terraform still invokes this
	// Update callback whenever one changes (it has no way to know the change
	// is inert), so that guarantee has to be enforced here: if req.Prior
	// (the prior state) is identical to plan in every attribute *except*
	// these three, skip the GET+PUT round trip entirely and persist plan
	// as-is.
	if req.Prior != nil && onlyCreateOnlyFlagsChanged(*req.Prior, plan) {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{Model: plan}, diags
	}

	fleetClient := client.GetFleetClient()

	current, getDiags := fleetclient.GetDefendPackagePolicy(ctx, fleetClient, req.WriteID, req.SpaceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}
	if current == nil {
		diags.AddError(
			"Agentless policy not found",
			fmt.Sprintf("Cannot update agentless policy %q: it was not found. It may have been deleted out of band; "+
				"run terraform apply again to detect drift.", req.WriteID),
		)
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	body, bodyDiags := buildUpdateBody(ctx, plan, current)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	var unionBody kbapi.PackagePolicyRequest
	if err := unionBody.FromPackagePolicyRequestTypedInputs(body); err != nil {
		diags.AddError("Failed to build the package policy update request", err.Error())
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	updated, updateDiags := fleetclient.UpdateAgentlessPolicyViaPackagePolicy(ctx, fleetClient, req.SpaceID, req.WriteID, unionBody)
	diags.Append(updateDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	diags.Append(plan.populateFromPackagePolicy(ctx, req.SpaceID, updated)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
	}

	return entitycore.KibanaWriteResult[agentlessPolicyModel]{Model: plan}, diags
}

// buildUpdateBody builds the typed PUT body by echoing current (the fresh
// typed GET) and overlaying only the in-place-updatable allowlist fields
// from plan. name, namespace, policy_id/policy_ids, and package.name/version
// are taken from current (never plan) since they are RequiresReplace and
// must exactly match the existing policy -- the spike found that omitting
// policy_id/policy_ids produces a 400 ("Cannot change agent policies of an
// agentless integration"), even when not actually changing them.
func buildUpdateBody(ctx context.Context, plan agentlessPolicyModel, current *kbapi.PackagePolicy) (kbapi.PackagePolicyRequestTypedInputs, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := current.Name
	body := kbapi.PackagePolicyRequestTypedInputs{
		Name:      &name,
		Namespace: current.Namespace,
		PolicyId:  current.PolicyId,
		PolicyIds: current.PolicyIds,
	}

	if current.Package != nil {
		pkg := kbapi.PackagePolicyRequestPackage{
			Name:    current.Package.Name,
			Version: current.Package.Version,
			Title:   current.Package.Title,
		}

		var pkgModel packageModel
		pkgDiags := plan.Package.As(ctx, &pkgModel, basetypes.ObjectAsOptions{})
		diags.Append(pkgDiags...)
		if !pkgDiags.HasError() && typeutils.IsKnown(pkgModel.Title) {
			title := pkgModel.Title.ValueString()
			pkg.Title = &title
		}
		body.Package = &pkg
	}

	// description, global_data_tags, additional_datastreams_permissions, and
	// var_group_selections must always be sent (never omitted): the spike
	// found that omitting a field from this typed PUT body PRESERVES its
	// existing value rather than clearing it (empirically confirmed against
	// a live Kibana 9.4.3 deployment: omitting `description` left the prior
	// description untouched, while explicitly sending `description: ""`
	// cleared it). Since these attributes are Optional but not Computed in
	// schema.go, Terraform requires state to exactly mirror config, so a
	// config value that was removed must actively clear the API value, not
	// leave it stale.
	description := plan.Description.ValueString()
	body.Description = &description

	tagsRaw := globalDataTagsRawFromModel(ctx, plan.GlobalDataTags, &diags)
	if tagsRaw == nil {
		tagsRaw = []map[string]any{}
	}
	if b, err := json.Marshal(tagsRaw); err != nil {
		diags.AddAttributeError(path.Root("global_data_tags"), "Failed to encode global_data_tags", err.Error())
	} else if err := json.Unmarshal(b, &body.GlobalDataTags); err != nil {
		diags.AddAttributeError(path.Root("global_data_tags"), "Failed to encode global_data_tags for the update request", err.Error())
	}

	var perms []string
	if typeutils.IsKnown(plan.AdditionalDatastreamsPermissions) {
		diags.Append(plan.AdditionalDatastreamsPermissions.ElementsAs(ctx, &perms, false)...)
	}
	if perms == nil {
		perms = []string{}
	}
	body.AdditionalDatastreamsPermissions = &perms

	vgs := map[string]string{}
	if typeutils.IsKnown(plan.VarGroupSelections) {
		diags.Append(plan.VarGroupSelections.ElementsAs(ctx, &vgs, false)...)
	}
	body.VarGroupSelections = &vgs

	// Top-level vars: seed from current's typed ({value,type}) representation
	// (preserving each var's `type` metadata, which the plan's flat
	// vars_json does not carry -- see the varEntry doc comment), then let
	// mergeVarsInto replace it with exactly the plan's key set. vars_json is
	// Optional+Computed (UseStateForUnknown), so it is only genuinely null
	// here if a config explicitly sets `vars_json = null`; mergeVarsInto
	// with a nil planVars handles that the same way as a known-but-empty
	// `vars_json = jsonencode({})` (both clear every existing key).
	if b, err := json.Marshal(current.Vars); err == nil {
		_ = json.Unmarshal(b, &body.Vars)
	}
	var planVars map[string]any
	if typeutils.IsKnown(plan.VarsJSON) {
		sanitized, sd := plan.VarsJSON.SanitizedValue()
		diags.Append(sd...)
		if !sd.HasError() {
			planVars = typeutils.NormalizedTypeToMap[any](jsontypes.NewNormalizedValue(sanitized), path.Root("vars_json"), &diags)
		}
	}
	mergeVarsInto(&body.Vars, planVars, &diags)

	inputs, inputDiags := buildUpdateInputs(ctx, current, plan)
	diags.Append(inputDiags...)
	body.Inputs = &inputs

	return body, diags
}

// buildUpdateInputs fetches current's typed inputs and echoes each one back
// (via a JSON round trip into the request's PackagePolicyRequestTypedInput
// shape, dropping response-only fields like compiled_input/compiled_stream),
// overlaying enabled/condition/vars from the plan's `inputs` map for any
// input the plan configures. Inputs the plan doesn't mention (or the whole
// `inputs` attribute being unknown) are echoed unchanged. See mappedInputKey
// for how a typed input's (PolicyTemplate, Type) is matched against the
// plan's "<policy_template>-<input_type>"-keyed inputs map.
func buildUpdateInputs(ctx context.Context, current *kbapi.PackagePolicy, plan agentlessPolicyModel) ([]kbapi.PackagePolicyRequestTypedInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedInputs, err := current.Inputs.AsPackagePolicyTypedInputs()
	if err != nil {
		diags.AddError("Failed to decode the current policy's inputs", err.Error())
		return nil, diags
	}

	var planInputs map[string]agentlessInputModel
	if typeutils.IsKnown(plan.Inputs.MapValue) {
		planInputs = typeutils.MapTypeAs[agentlessInputModel](ctx, plan.Inputs.MapValue, path.Root("inputs"), &diags)
	}

	result := make([]kbapi.PackagePolicyRequestTypedInput, 0, len(typedInputs))
	for _, in := range typedInputs {
		var reqIn kbapi.PackagePolicyRequestTypedInput
		b, err := json.Marshal(in)
		if err != nil {
			diags.AddError("Failed to encode the current policy's input", err.Error())
			continue
		}
		if err := json.Unmarshal(b, &reqIn); err != nil {
			diags.AddError("Failed to encode the current policy's input for the update request", err.Error())
			continue
		}

		if planIn, ok := planInputs[mappedInputKey(in.PolicyTemplate, in.Type)]; ok {
			overlayInputFromPlan(ctx, &reqIn, planIn, &diags)
		}

		result = append(result, reqIn)
	}

	return result, diags
}

// overlayInputFromPlan mutates reqIn in place, applying the plan's
// enabled/condition/vars for the input itself and, by matching each typed
// stream's DataStream.Dataset against the plan's per-stream map key, for
// each of its streams.
func overlayInputFromPlan(ctx context.Context, reqIn *kbapi.PackagePolicyRequestTypedInput, planIn agentlessInputModel, diags *diag.Diagnostics) {
	// NOTE (Decision 3 spike caveat, see this file's header comment): toggling
	// `enabled` via this PUT endpoint was accepted (200) but NOT reliably
	// persisted for the cloud_security_posture package under test -- it
	// silently reverted to its prior value. We still send the plan's value
	// here regardless (Kibana may honor it for other packages, and a
	// no-op is harmless), and deliberately do NOT add a package-specific
	// workaround: the entitycore envelope always performs a read-after-write
	// refresh following Update (see kibana_resource_envelope.go's
	// runKibanaWrite, which calls the Read callback with this function's
	// caller's returned model before persisting state), so if `enabled`
	// silently failed to persist, the immediately-following Read re-syncs
	// state from the authoritative API response -- state never silently
	// drifts from reality, it just reflects that the change didn't take,
	// which is the correct, honest outcome for Terraform to show on the next
	// plan.
	if typeutils.IsKnown(planIn.Enabled) {
		reqIn.Enabled = planIn.Enabled.ValueBool()
	}

	reqIn.Condition = planIn.Condition.ValueStringPointer()

	var planVars map[string]any
	if typeutils.IsKnown(planIn.Vars) {
		diags.Append(planIn.Vars.Unmarshal(&planVars)...)
	}
	mergeVarsInto(&reqIn.Vars, planVars, diags)

	if reqIn.Streams == nil || !typeutils.IsKnown(planIn.Streams) {
		return
	}

	planStreams := typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, planIn.Streams, path.Root("inputs"), diags)
	for i := range *reqIn.Streams {
		s := &(*reqIn.Streams)[i]
		planStream, ok := planStreams[s.DataStream.Dataset]
		if !ok {
			continue
		}

		if typeutils.IsKnown(planStream.Enabled) {
			s.Enabled = planStream.Enabled.ValueBool()
		}
		s.Condition = planStream.Condition.ValueStringPointer()

		var streamVars map[string]any
		if typeutils.IsKnown(planStream.Vars) {
			diags.Append(planStream.Vars.Unmarshal(&streamVars)...)
		}
		mergeVarsInto(&s.Vars, streamVars, diags)
	}
}
