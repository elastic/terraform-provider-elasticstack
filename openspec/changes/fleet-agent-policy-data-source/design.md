## Context

Fleet exposes agent policies through `GET /api/fleet/agent_policies/{agentPolicyId}` (by ID) and
`GET /api/fleet/agent_policies` (list, with optional `Kuery`/pagination). The Terraform provider
already ships `elasticstack_fleet_agent_policy` as a full CRUD resource
(`internal/fleet/agentpolicy`), but has no read-only way to resolve an existing policy's `policy_id`
from a consuming stack that doesn't own it. The only two Fleet data sources with a comparable
single-match lookup shape are `elasticstack_fleet_enrollment_tokens`
(`internal/fleet/enrollmenttokens`) and `elasticstack_fleet_output`
(`internal/fleet/outputds`); both are built on the shared `entitycore.KibanaDataSource[Model]`
envelope (`internal/entitycore/data_source_envelope.go`).

This document was informed by an automated implementation-research comment on issue #5324
(workflow run `38970672707`) and by maintainer decisions posted on the issue afterward. The
maintainer's decisions are the final say on every point where they speak; they close all four open
questions the research comment left outstanding, and on one of them (`kql-field-mapping`) they
change the recommended lookup implementation. See **Deviation from the research recommendation**
below.

## Goals / Non-Goals

**Goals:**

- A read-only `elasticstack_fleet_agent_policy` data source that resolves exactly one agent policy
  by `policy_id` or by `name` (exactly one of the two, both `Optional+Computed` so the unused
  selector is populated as an output).
- An optional `space_id`, consistent with the other Fleet data sources.
- A `name` lookup that is correct (never silently returns a wrong or over-broad match) and does not
  depend on an unverified assumption about KQL escaping behavior.
- An error that fails closed on zero or multiple matches for `name`, naming the conflicting
  `policy_id`s when there is more than one.
- Exported attributes mirroring the resource's read-only surface, mapped directly from the raw API
  response (no drift-avoidance null-preservation — a data source has no configured value to
  preserve nulls against).

**Non-Goals:**

- A filterable, list-typed data source returning multiple agent policies. The issue asks for a
  scalar lookup with an error on ambiguity, not a list; a list data source, if ever needed, is a
  separate, thin-result type.
- Any change to the existing resource's schema, CRUD behavior, or `populateFromAPI` mapping
  function.
- Lookup by any key other than `policy_id` or `name`.
- Resolving `terraform_remote_state` cross-stack coupling in general.

## Decisions

### Decision 1: New, additive `internal/fleet/agentpolicyds` package

Build the data source as its own package, following the exact structural precedent of the three
existing Fleet data sources (`enrollmenttokens`, `outputds`, `integrationds`): built on
`entitycore.NewKibanaDataSource[Model]`, with `schema.go`, `models.go`, `read.go`, and
`data_source.go`.

**Why:** Keeps the change additive-only (no edits to the existing resource package's tested mapping
logic), matches every other Fleet data source's shape, and avoids inheriting the resource's
drift-avoidance null-preservation and description-normalization behavior
(`internal/fleet/agentpolicy/models.go:91-237`), which would be actively wrong for a data source
whose config is always null for computed fields. An alternative — exporting and parameterizing the
resource's `populateFromAPI` for reuse by both resource and data source — was considered and
rejected: it would touch and re-test the existing, already-covered resource mapping function purely
to serve a new data source, for a larger and riskier diff than an additive-only package.

### Decision 2: Selector validation via per-attribute `ExactlyOneOf`

`policy_id` and `name` are each `Optional+Computed` with
`stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName(<other>))`. This exact
per-attribute, relative-path pattern already has in-repo precedent on two data sources
(`internal/elasticsearch/ingest/processor_set_pf_data_source.go:77-94`,
`internal/elasticsearch/ingest/processor_script_pf_data_source.go:76-93`).

**Why `Optional+Computed` and not plain `Optional`:** both attributes double as outputs — whichever
one wasn't used as the selector is populated from the API response. `ExactlyOneOf` in
terraform-plugin-framework-validators counts configuration values only, so a value filled in by a
previous read does not trip the validator on the next plan; only the attribute actually set in
config is evaluated. The data source's `GetResourceID()` mirrors
`enrollmenttokens`' pattern exactly (`internal/fleet/enrollmenttokens/models.go:40-45`): return
`PolicyID` when set, else the literal placeholder `"_"`. The Read function filters on
`config.PolicyID`/`config.Name` directly, not the envelope-resolved id, for the same reason
`enrollmenttokens/read.go:31-33` gives for its own policy filter: `"_"` is itself a valid
`policy_id`.

### Decision 3: `id` is set explicitly by the read function, not by the envelope

The shared envelope (`internal/entitycore/data_source_envelope.go:295-345`) only plumbs the model
through; it has no knowledge of a model's `id` field. The read/mapper callback must explicitly set
`ID = types.StringValue(api.Id)`, mirroring `enrollmenttokens/read.go:60-69`'s explicit assignment.
Both the by-ID and by-name lookup paths must populate `id` this way, and both test cases must assert
it.

### Decision 4: By-ID lookup reuses the existing client wrapper; by-name lookup is a new, Kuery-filtered helper

- `policy_id` set → reuse `fleet.GetAgentPolicy(ctx, client, policyID, spaceID)`
  (`internal/clients/fleet/agent_policy.go:31-38`), which already wraps
  `GetFleetAgentPoliciesAgentpolicyidWithResponse` with `kibanautil.SpaceAwarePathRequestEditor`;
  404 is already mapped to the envelope's `found=false` → the standard "not found" diagnostic via
  `kibanaoapi.HandleGetTypedResponse`.
- `name` set → a new client helper, `fleet.GetAgentPolicyByName`, in
  `internal/clients/fleet/agent_policy.go`, calling `GetFleetAgentPoliciesWithResponse` with
  `Kuery: fmt.Sprintf("ingest-agent-policies.name:\"%s\"", escapedName)` for the target space (mirroring
  the existing `Kuery` usage precedent in `internal/clients/fleet/enrollment_tokens.go:48`, which
  passes an unquoted `policy_id:` term; this lookup additionally quotes and escapes the value because
  policy names, unlike policy IDs, may contain KQL-significant characters). See **Deviation from the
  research recommendation** below for why this replaces the research's unfiltered-pagination design.
  The helper then applies a client-side exact compare (`item.Name == name`) against every item the
  filtered query returns, so a bad escape or an unexpectedly broad server-side match fails closed
  instead of silently returning a wrong policy. If the filtered response spans more than one page
  (`Page`/`PerPage`/`Total` at `generated/kbapi/kibana.gen.go:250260-250265`), the helper paginates
  through the full filtered result set before deciding: it never reports a unique match without
  reaching the server-reported `Total` for the filtered query, and it errors out (rather than
  guessing) if a page comes back short of its expected size, if `Total` changes between requests, or
  if the same `Id` is seen twice. Because the lookup is now scoped to a server-side filter instead of
  the full unfiltered policy list, the per-request page size can stay small without the O(total
  policies) cost the research's fallback design would have had.
  - Zero exact matches after a complete, anomaly-free pagination of the filtered result → data
    source error: `no agent policy found matching name %q in space %q`.
  - More than one exact match → data source error listing the conflicting `policy_id`s (maintainer
    decision; see below), e.g.: `multiple agent policies match name %q in space %q: policy_id "a",
    policy_id "b"`.
- Escaping: construct the quoted Kuery value by escaping backslashes and double quotes in the raw
  `name` (`\` → `\\`, `"` → `\"`) exactly once, then interpolate it into the quoted literal with
  `fmt.Sprintf("ingest-agent-policies.name:\"%s\"", escapedName)`, per the documented KQL escaping
  rules (https://www.elastic.co/docs/reference/query-languages/kql). The client-side
  exact compare is the correctness backstop regardless of whether this escaping is complete for every
  edge case.

### Decision 5: Data-source-specific attribute mapping, not resource mapping reuse

Attribute mapping is a new function in the `agentpolicyds` package, not a call into the resource's
unexported `agentPolicyModel.populateFromAPI`. The resource's mapping intentionally suppresses a
non-empty API value back to null when the *current model value* is null (`preserveNullStr`,
`internal/fleet/agentpolicy/models.go:106-111`) — correct for the resource's drift avoidance, wrong
for a data source whose config is always null for computed fields (it would hide real, non-empty
output/host IDs). The data source instead always reports the raw API value directly:
`types.StringPointerValue(api.DataOutputId)`, `types.StringPointerValue(api.FleetServerHostId)`,
etc., with no null-preservation. `description` is reported as-is from the API (`*string`; nil stays
null, a returned empty string stays empty) — maintainer-confirmed: "An omitted description is stored
as an empty string. Report the API value as returned." Monitoring booleans are computed
unconditionally: `api.MonitoringEnabled != nil && slices.Contains(*api.MonitoringEnabled,
kbapi.KibanaHTTPAPIsAgentPolicyResponseMonitoringEnabledLogs)` (and the `metrics` equivalent), so a
policy with monitoring turned off after being on is correctly reported `false`, never a stale `true`.

### Decision 6: No per-attribute version gate for the core lookup

The OpenAPI spec and public docs state no minimum-version floor for either the by-ID or by-name GET
endpoint, so the data source's core lookup needs no version gate of its own (unlike the resource's
`is_protected` 8.10.0 gate, which is about *setting* that field on write, not reading it). All other
exported attributes are `Computed`-only in this data source, and the resource's per-attribute
version gates (`internal/fleet/agentpolicy/version_requirements.go:32-143`) only fire on a
*configured write* (the `addMinimum` guard checks whether the attribute was set in config/plan), so
reading an older/absent field on a GET response never triggers them. `space_id`-scoped behavior is
the one area with a documented version dependency on the read side: Fleet space awareness is
default-on only for new 9.1.0+ deployments and requires an explicit one-time migration on upgraded
deployments.

### Decision 7: Schema and attribute set

Every attribute besides `policy_id`/`name`/`space_id`/`id` is `Computed`-only. The full attribute
list, types, and API field mapping are recorded in the delta spec
(`openspec/changes/fleet-agent-policy-data-source/specs/fleet-agent-policy/spec.md`). `skip_destroy`
and `sys_monitoring` are excluded — both are write-time-only resource behaviors (destroy-skip flag,
create-time `RequiresReplace` setting) with no meaningful read-only value to export.

### Decision 8: Registration

Register in `provider/plugin_framework.go` next to the other Fleet data sources
(`enrollmenttokens.NewDataSource`, `outputds.NewDataSource`, `integrationds.NewDataSource`). Add
`examples/data-sources/elasticstack_fleet_agent_policy/data-source.tf` and
`docs/data-sources/fleet_agent_policy.md`.

## Deviation from the research recommendation

The implementation-research comment's recommended design (Approach A) made the by-name lookup's
*primary* path an unfiltered pagination of the entire agent-policy list for the target space,
followed by a client-side exact-name compare — explicitly as "the conservative choice" because the
research could not confirm whether `ingest-agent-policies.name` is a `keyword` (exact) or analyzed
(tokenized) saved-object field, and therefore could not confirm a `Kuery` pre-filter would behave
correctly.

The maintainer's follow-up comment on the issue closes that open question with a direct,
version-pinned verification: on Kibana 9.5.5, `name` is a `keyword` field on both saved-object types
(`ingest-agent-policies`, `fleet-agent-policies`) from 8.6 through 9.5.5; the list API accepts the
legacy-prefixed, quoted term `ingest-agent-policies.name:"<escaped>"` as an exact match (hyphenated
names are not tokenized), while `fleet-agent-policies.name:...` is rejected at request validation
(Kibana rewrites the legacy prefix onto the active saved-object type). This directly contradicts the
research's "conservative" full-list-pagination-as-primary-design choice, which was explicitly
conditioned on that mapping being *unconfirmed*. Per the maintainer's explicit instruction ("Do not
page the unfiltered policy list as the primary design. That full-list scan was the conservative
choice while this mapping was unconfirmed."), this design adopts the server-side `Kuery` pre-filter
as the primary by-name lookup path (Decision 4), keeping the client-side exact-match compare as a
fail-closed backstop rather than as the sole correctness mechanism.

The maintainer's comment also closes the research's other three open questions, none of which
required a design change beyond what's reflected above:

- **`sort-field-support`**: confirmed closed — `id` is not an accepted `sortField` (`GET
  /api/fleet/agent_policies?sortField=id` returns 400 `Unknown sort field id` on Kibana 9.5.5);
  `name` and `updated_at` are accepted. Since the primary by-name path is now a `Kuery`-filtered
  query rather than a full-list scan, pagination sort order is lower-stakes than in the research's
  design; if a future change ever needs to page the filtered (or, hypothetically, unfiltered) list,
  it SHALL sort by `name`, never by `id`.
- **`canonical-spec-file`**: confirmed closed — `openspec/specs/fleet-agent-policy/spec.md` is the
  canonical spec; the data source requirements are added there as a delta to that capability. The
  other candidate, `openspec/specs/elasticstack-fleet-agent-policy/spec.md`, is a leftover bugfix
  spec and is left untouched.
- **`ambiguous-match-error-detail`**: confirmed closed — the multi-match error enumerates the
  conflicting `policy_id`s (Decision 4). A duplicate `name` at create time already returns a Fleet
  409 naming the existing policy's id (`Agent Policy '<id>' already exists with name '<name>'`), so
  surfacing IDs in this data source's ambiguous-match error is consistent with information Fleet
  already exposes for the same `name` field elsewhere.

## Open questions

The items below are the research comment's original open questions, preserved for traceability. All
four are resolved by the maintainer's decisions described above; none remain blocking.

- Is `id` an accepted value for `GetFleetAgentPoliciesParams.SortField` on this endpoint, across every
  Kibana version this provider supports? The generated client only proves the `SortField`/`SortOrder`
  parameters exist (kibana.gen.go:84274-84307), not which field names the server actually accepts;
  needs a real-endpoint probe before implementation. Even once confirmed, a stable sort only reduces
  (does not eliminate) the chance of a concurrent create/delete reshuffling paginated results, so the
  helper's anomaly checks (duplicate-`Id`, short-page, changing-`Total`) remain load-bearing regardless
  of the answer. (non-blocking: the client-side exact-match-after-full-pagination design is correct
  with or without a working sort key, which only affects anomaly-detection quality, not correctness of
  the happy path) — **Resolved**: maintainer confirmed `id` is rejected (400 `Unknown sort field id`
  on Kibana 9.5.5); `name`/`updated_at` are accepted; sort by `name` if pagination of the filtered
  list is ever needed.
- Is the `ingest-agent-policies.name` saved-object field mapped as keyword (exact) or analyzed
  (tokenized/substring) text in the versions of Kibana this provider supports, and does a correctly
  backslash-escaped `Kuery` value (per the documented KQL escaping rules) behave consistently across
  them? Confirming this would let the name lookup add a server-side `Kuery` pre-filter as a performance
  optimization on top of the client-side exact match this research recommends as the correctness
  baseline. (non-blocking for correctness, since the recommended approach doesn't depend on the
  answer; blocking only for the optimization) — **Resolved**: maintainer confirmed `name` is a
  `keyword` field (not analyzed) on both saved-object types, 8.6 through 9.5.5; the legacy-prefixed,
  quoted `ingest-agent-policies.name:"<escaped>"` Kuery term is an exact match. This promotes the
  `Kuery` pre-filter from "future optimization" to the primary by-name lookup path (see Deviation
  section above).
- Which of `openspec/specs/fleet-agent-policy/spec.md` and
  `openspec/specs/elasticstack-fleet-agent-policy/spec.md` is the canonical spec file for this entity,
  and should the data source's requirements be added there or in a new spec? (non-blocking for design,
  blocking for the OpenSpec authoring step) — **Resolved**: `openspec/specs/fleet-agent-policy/spec.md`
  is canonical; `elasticstack-fleet-agent-policy/spec.md` is a leftover bugfix spec and is left alone.
- Should the two-matches-found error enumerate the conflicting `policy_id`s (more actionable) or just
  report the count, given that returning extra API response data in an error message is a minor
  information-exposure consideration for multi-tenant Kibana deployments? (non-blocking) —
  **Resolved**: enumerate the conflicting `policy_id`s.
