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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// updateAgentlessPolicy is a stub for Task 3 of the fleet-agentless-policy
// OpenSpec change. Full implementation -- calling
// fleetclient.UpdateAgentlessPolicyViaPackagePolicy
// (PUT /api/fleet/package_policies/{id}) with the in-place-updatable
// allowlist from Decision 3, informed by the spike findings documented
// above -- lands in Task 5.
func updateAgentlessPolicy(
	_ context.Context,
	_ *clients.KibanaScopedClient,
	_ entitycore.KibanaWriteRequest[agentlessPolicyModel],
) (entitycore.KibanaWriteResult[agentlessPolicyModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	diags.AddError(
		"Not yet implemented",
		"The elasticstack_fleet_agentless_policy resource's Update operation is not yet implemented "+
			"(see openspec/changes/fleet-agentless-policy, Task 5).",
	)
	return entitycore.KibanaWriteResult[agentlessPolicyModel]{}, diags
}
