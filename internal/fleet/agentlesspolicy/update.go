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

// Spike findings (fleet-agentless-policy OpenSpec change, Task 3.3, 2026-07-01)
// ===============================================================================
//
// Probed live against a cloud-hosted Kibana 9.4.3 deployment. Created a real
// agentless policy via POST /api/fleet/agentless_policies using the
// cloud_security_posture package (v3.4.0, policy_template "cspm", input
// "cloudbeat/cis_aws"), then issued a series of PUT
// /api/fleet/package_policies/{id} requests, inspecting the persisted state
// via a fresh GET after each PUT. The test policy, its hidden agent policy,
// and the cloud_security_posture package installation were all removed
// afterwards; the deployment was left in the state it was found.
//
// Request body shape for PUT: PackagePolicyRequest (the generated kbapi type
// for PutFleetPackagePoliciesPackagepolicyidJSONRequestBody) is a union. The
// "mapped"/simplified variant (KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest,
// same shape as the agentless create body, `inputs` keyed by
// "<policy_template>-<input_type>") is NOT usable for these updates: it does
// not support `global_data_tags`, and -- more importantly -- sending inputs
// via this map/key shape against an existing agentless-created policy did
// not just fail to update inputs, it silently reset every input's `enabled`
// flag to false (see "inputs.enabled" finding below) even for inputs that
// were not present in the request. All findings below use the "typed" array
// variant instead (matching the exact shape returned by GET
// /api/fleet/package_policies/{id}: `inputs` is an array of objects with a
// `type` field, `vars` are `{type, value}` objects), built by mutating a
// fresh GET response and PUTting it back. `policy_ids` (or `policy_id`) must
// be included and match the policy's own ID -- omitting it produces "Cannot
// change agent policies of an agentless integration" (400), even though the
// caller isn't trying to move the policy to a different agent policy.
//
// Candidate in-place-updatable fields (Decision 3 allowlist):
//
//   - description                          -- ACCEPT + PERSIST. Round-tripped
//     cleanly through a follow-up GET.
//   - vars_json (top-level `vars`)          -- ACCEPT + PERSIST. Changing
//     `deployment` from "aws" to "gcp" round-tripped cleanly.
//   - inputs[*].streams[*].vars (a var value on an existing input's stream)
//     -- ACCEPT + PERSIST. Setting `aws.account_type` = "single-account" on
//     the cis_aws input's findings stream round-tripped cleanly.
//   - inputs[*].enabled                     -- ACCEPT, but NOT PERSISTED.
//     Explicitly setting the previously-enabled cis_aws input's `enabled` to
//     `true` (all others `false`, exactly matching the shape returned by a
//     prior GET) returned 200, but a follow-up GET showed EVERY input
//     `enabled: false`, including cis_aws. This was reproduced twice. It may
//     be specific to the cloud_security_posture package (which enforces "only
//     one enabled input is allowed per policy" server-side on create) rather
//     than a general property of agentless PUT; a generic package was not
//     available to cross-check in this environment. Net effect for Task 5:
//     toggling `inputs[*].enabled` via PUT cannot be assumed to take effect
//     for every package; the vars sub-field is confirmed reliable, but
//     enabled-state changes should be verified against the actual response
//     body (not assumed from a 200) and likely warrant an acceptance-test
//     assertion once a second integration is identified (see design.md Open
//     Question 2).
//   - global_data_tags                      -- ACCEPT + PERSIST.
//   - additional_datastreams_permissions    -- ACCEPT + PERSIST.
//   - var_group_selections                  -- ACCEPT + PERSIST, but
//     UNVALIDATED: cloud_security_posture does not define any var groups, and
//     an arbitrary `{"some_group": "some_option"}` was accepted and persisted
//     without any check that "some_group" exists on the package. This means
//     the API will not catch a Terraform-side typo in a var group name; any
//     validation would have to be client-side or rely on a later apply error.
//   - package.title                         -- ACCEPT + PERSIST. Setting a
//     custom title round-tripped cleanly, overriding the registry-populated
//     default.
//
// RequiresReplace fields (Decision 3), probed for contradiction:
//
//   - name              -- ACCEPTED AND PERSISTED via PUT (200, new name
//     round-tripped through a follow-up GET). Kibana does NOT reject a name
//     change on this endpoint.
//   - namespace         -- ACCEPTED AND PERSISTED via PUT (200). One caveat
//     unrelated to immutability: the namespace value is validated against
//     Elasticsearch data-stream naming rules and rejects hyphens (e.g.
//     "spike-alt-namespace" => 400 "Namespace contains invalid characters"),
//     since namespace is a `-`-delimited segment of the backing data stream
//     name. A hyphen-free namespace ("spikealt") was accepted and persisted.
//   - package.version   -- ACCEPTED AND PERSISTED via PUT (200), including
//     setting it to "3.3.0", a version that is NOT installed on the test
//     deployment (only 3.4.0 was installed) and does not appear to exist in
//     the registry for this package. The PUT endpoint does not validate
//     package.version against the registry or installed packages at all; it
//     is stored as an opaque string. This is notably different from Create,
//     which does perform registry validation.
//   - package.name      -- NOT a simple pass-through, but also not rejected
//     merely for being an identity field. Two sub-cases observed:
//     -- an entirely nonexistent package name => 404 "[name] package not
//     installed or found in registry" (registry/installation check).
//     -- an installed-but-non-agentless-capable package ("system", which is
//     installed on the test deployment but does not declare agentless
//     deployment mode support) => 400 "Package \"system\" does not support
//     agentless deployment mode" (agentless-capability check, not an
//     immutability check).
//     No second agentless-capable package was installed on the test
//     deployment to check whether swapping between two agentless-capable
//     packages would fully succeed; based on the above, it is plausible it
//     would, since neither rejection reason was "package.name cannot be
//     changed on an existing policy".
//
// Conclusion / disposition:
//
// The API-level premise behind Decision 3's RequiresReplace list --
// "the package_policies PUT endpoint's behavior for hidden agentless-created
// policies is not fully documented" (design.md, Decision 3 rationale) -- is
// now resolved empirically, and the result CONTRADICTS the assumption that
// these fields are API-enforced immutable: Kibana's PUT endpoint accepts and
// persists changes to name, namespace, and package.version outright, and
// does not reject package.name for identity reasons (only for
// registry/agentless-capability reasons). None of the four RequiresReplace
// candidates tested (name, namespace, package.version, package.name) is
// actually blocked by the API.
//
// Per Task 3.3's own instructions and design.md's framing ("if Kibana
// silently allows changing package.version via PUT, that would mean the
// design's RequiresReplace choice is a Terraform-side policy decision, not
// an API constraint -- still fine, but worth noting"), this document (and
// the corresponding update to design.md Decision 3 / Open Question 1) keeps
// the RequiresReplace partitioning UNCHANGED for Task 4/5 to implement,
// specifically because:
//
//   - Renaming, moving namespace, or bumping package.version in-place via
//     this endpoint bypasses Fleet's normal package-install/upgrade
//     lifecycle (index templates, ingest pipelines, Kibana saved objects
//     for the new version are not provisioned by this call -- Create's
//     registry validation showed that path IS checked there, but PUT does
//     not do the equivalent work). Applying such a change in-place could
//     silently leave the policy referencing a package version that was
//     never actually installed.
//   - Terraform's replacement semantics (destroy + recreate through the
//     validated agentless create path) are the safer default for identity
//     and structural fields even where the raw API is permissive.
//
// This is flagged prominently for orchestrator review: keeping
// RequiresReplace for name/namespace/package.version/package.name is now
// known to be a deliberate Terraform-side safety choice rather than an
// inferred API constraint. See design.md Decision 3 and Open Question 1 for
// the updated rationale.

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

// varEntry mirrors the `{frozen, type, value}` shape used for `vars`
// throughout the "typed" package-policy request/response family
// (KibanaHTTPAPIsUpdatePackagePolicyRequest, PackagePolicyRequestTypedInput,
// PackagePolicyRequestTypedInputStream, PackagePolicyTypedInput, and
// PackagePolicyTypedInputStream). oapi-codegen gives each occurrence its own
// anonymous map[string]struct{Frozen,Type,Value} Go type (structurally
// identical, but not always assignable/convertible to each other -- see e.g.
// PackagePolicyTypedInputStream.Release vs
// PackagePolicyRequestTypedInputStream.Release, which alias the same
// underlying string type via two distinct named types). This local named
// type lets mergeVarsInto operate generically on all of them via a JSON
// marshal/unmarshal round trip instead of hand-spelling each anonymous type.
type varEntry struct {
	Frozen *bool   `json:"frozen,omitempty"`
	Type   *string `json:"type,omitempty"`
	Value  any     `json:"value,omitempty"`
}

// mergeVarsInto overlays planVars (a flat map[string]any parsed from
// vars_json or an input/stream's `vars`) onto *dst, an existing
// `{frozen,type,value}`-shaped typed vars map (or nil), preserving each
// existing entry's `type`/`frozen` metadata and only replacing `value`. New
// keys not already present in *dst are added with only `value` set. Uses a
// JSON round trip throughout so the anonymous `*map[string]struct{...}`
// field type of the caller's choosing (see the type comment above) never
// needs to be spelled out; V is inferred from the caller's argument.
func mergeVarsInto[V any](dst **V, planVars map[string]any, diags *diag.Diagnostics) {
	if len(planVars) == 0 {
		return
	}

	existing := map[string]varEntry{}
	if *dst != nil {
		if b, err := json.Marshal(*dst); err == nil {
			_ = json.Unmarshal(b, &existing)
		}
	}

	for k, v := range planVars {
		e := existing[k]
		e.Value = v
		existing[k] = e
	}

	b, err := json.Marshal(existing)
	if err != nil {
		diags.AddError("Failed to encode vars for the update request", err.Error())
		return
	}
	if err := json.Unmarshal(b, dst); err != nil {
		diags.AddError("Failed to encode vars for the update request", err.Error())
	}
}

// clearVars sets *dst to a non-nil pointer to an empty vars map. Like
// mergeVarsInto, `{}` is used rather than nil because these fields'
// `omitempty` tag only treats a nil *pointer* as empty (see buildUpdateBody's
// comment on why every allowlisted field must be sent explicitly, never
// omitted, to actively clear a value Terraform's config no longer sets).
// `vars` (at the top level, per-input, and per-stream) is Optional but not
// Computed in schema.go, so a config that removes a `vars` block must clear
// the corresponding API value, not silently echo back whatever `vars` the
// fetched typed snapshot already had -- mergeVarsInto alone cannot do this
// because it short-circuits on an empty planVars to avoid clobbering values
// with an accidental no-op call.
func clearVars[V any](dst **V) {
	// json.Unmarshal decodes a JSON object into an *existing* non-nil map by
	// merging keys, never removing ones absent from the new JSON (unlike
	// mergeVarsInto, which is safe because it always seeds `existing` from
	// *dst first, so *dst's key set is always a subset of what gets written
	// back). *dst must be reset to nil here so json.Unmarshal allocates a
	// genuinely fresh, empty map instead of merging `{}` into whatever *dst
	// already pointed to (which would silently leave every existing key in
	// place).
	*dst = nil
	b, err := json.Marshal(map[string]varEntry{})
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, dst)
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
	// vars_json does not carry -- see the varEntry doc comment), then overlay
	// the plan's values. vars_json is Optional+Computed (UseStateForUnknown),
	// so it is only genuinely null here if a config explicitly sets
	// `vars_json = null`; clearVars handles that edge case the same way the
	// per-input/per-stream `vars` overlay does (see overlayInputFromPlan).
	if b, err := json.Marshal(current.Vars); err == nil {
		_ = json.Unmarshal(b, &body.Vars)
	}
	if typeutils.IsKnown(plan.VarsJSON) {
		sanitized, sd := plan.VarsJSON.SanitizedValue()
		diags.Append(sd...)
		if !sd.HasError() {
			planVars := typeutils.NormalizedTypeToMap[any](jsontypes.NewNormalizedValue(sanitized), path.Root("vars_json"), &diags)
			mergeVarsInto(&body.Vars, planVars, &diags)
		}
	} else {
		clearVars(&body.Vars)
	}

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

	if typeutils.IsKnown(planIn.Condition) {
		condition := planIn.Condition.ValueString()
		reqIn.Condition = &condition
	} else {
		reqIn.Condition = nil
	}

	if typeutils.IsKnown(planIn.Vars) {
		var planVars map[string]any
		if err := planIn.Vars.Unmarshal(&planVars); err == nil {
			mergeVarsInto(&reqIn.Vars, planVars, diags)
		}
	} else {
		clearVars(&reqIn.Vars)
	}

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
		if typeutils.IsKnown(planStream.Condition) {
			condition := planStream.Condition.ValueString()
			s.Condition = &condition
		} else {
			s.Condition = nil
		}
		if typeutils.IsKnown(planStream.Vars) {
			var streamVars map[string]any
			if err := planStream.Vars.Unmarshal(&streamVars); err == nil {
				mergeVarsInto(&s.Vars, streamVars, diags)
			}
		} else {
			clearVars(&s.Vars)
		}
	}
}
