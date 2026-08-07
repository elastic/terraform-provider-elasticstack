## Why

Kibana is shipping **Alerting v2**, an ES|QL-first alerting engine with its own saved object type (`alerting_rule`), its own HTTP surface under `/api/alerting/v2/`, and its own notification model. It is not a new rule type inside classic alerting: there is no dual-read, no in-place conversion from v1, and the two ID spaces are disjoint. Practitioners who adopt v2 have no way to manage rules as code today.

This change designs a new Terraform resource for v2 rules. It is a **design-only proposal** — the deliverable is this change directory, not provider code. Implementation is tracked separately in [rna-program#673](https://github.com/elastic/rna-program/issues/673) under epic [#655](https://github.com/elastic/rna-program/issues/655).

There is a second reason to do this now, while the API is still experimental. The v1 resource models rule configuration as an opaque `params` JSON string. That defeats plan-time validation, forces the provider to carry a bespoke params validator with a rule-type override table (`kibana-alerting-rule` REQ-018–REQ-021, REQ-043, REQ-044, REQ-051, REQ-052), and is a longstanding practitioner complaint. v2's rule body is a single closed schema with no per-rule-type polymorphism, so the provider can finally express the whole rule as typed Terraform attributes. Getting that shape right — and feeding back the API changes that make it possible — is cheapest before the API is GA.

## What Changes

- Add a new capability `kibana-rule` describing the Terraform resource **`elasticstack_kibana_rule`**, covering: full attribute-by-attribute schema mapped to the v2 rule API, CRUD endpoint selection, composite identity and import, handling of computed and volatile fields, and plan-time validation of the cross-field rules the API enforces.
- Model the ES|QL `query` discriminated union, `schedule`, `state_transition`, `grouping`, and `artifacts` as **typed Terraform attributes**, not a JSON blob.
- Register the resource in `experimentalResources()` behind `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`, as a technical preview with no generated registry documentation. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678) and is designed to need no schema, API, or state change at that point.
- Record a set of **contingent API changes** to request from the Alerting v2 team (see below). One of them — writable `enabled` — materially changes the resource's apply semantics, so it is called out as a decision the design depends on rather than a wish list.

**Not in this change:** any Go code, `generated/kbapi` regeneration, an action-policy resource, or the rename of the v1 resource (the last is argued in Decision 2; delivered separately).

### Decision 1 — a separate resource type, not a discriminator on the existing one

The obvious alternative is an `engine = "v1" | "v2"` (or `alerting_version`) field on `elasticstack_kibana_alerting_rule`. Rejected, for four reasons:

1. **Schema overlap with v1 is close to zero.** v2 has no `rule_type_id`, no `params`, and no `consumer`. v1 has no `kind`, no `query`, no `time_field`, no `recovery_strategy`/`no_data_strategy`/`state_transition`, and no `grouping`. The only fields that survive are `name`, `tags`, and the schedule interval — and in v2 the first two moved inside `metadata` and the third is `schedule.every`.
2. **Actions moved off the rule entirely.** v1's `actions` blocks (with `frequency` and `alerts_filter`) have no v2 counterpart — notifications are no longer configured on the rule. A single resource type would carry a large block that is valid for exactly one value of the discriminator.
3. **The saved object types and ID spaces are disjoint** (`alert` vs `alerting_rule`). `terraform import` takes an ID and a resource type; with a discriminator the provider could not tell which API to query, so import would be ambiguous or would need a probe-both-and-guess step.
4. **`ForceNew` on the discriminator would present destroy-and-recreate as if it were a migration.** No conversion exists. A practitioner who flips `engine` would see a plan that looks like an upgrade and get a deleted rule plus a brand-new one with different history and a different ID.

There is a fifth, structural cost: a discriminated resource forces every attribute to `Optional` so that both variants can validate, which moves all the required/forbidden logic out of the schema and into hand-written `ValidateConfig` code. That is precisely the failure mode this proposal is trying to escape from with `params`.

### Decision 2 — naming

| Object | Proposed Terraform type |
|--------|-------------------------|
| v2 rule | `elasticstack_kibana_rule` |
| v1 rule (rename, future) | `elasticstack_kibana_alerting_rule_classic` |

The tracking issues currently say `elasticstack_kibana_alerting_rule_v2`, marked "name TBD". Three arguments against carrying the version into the type name:

- **`_v2` names age badly.** The version suffix is meaningful only while both engines exist. Once classic alerting is retired, the provider is left with a resource permanently called `_v2` and no `_v1`, and renaming it then is a second breaking change.
- **Dropping the `alerting_` prefix has precedent.** `elasticstack_kibana_maintenance_window` is an alerting-framework object and does not carry the prefix.
- **The product itself is renaming.** Classic alerting is becoming "classic"; `elasticstack_kibana_alerting_rule_classic` tracks that, and leaves the unqualified noun for the thing that is simply "a rule" going forward.

**The main objection is that "rule" is overloaded in this provider** — `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_install_prebuilt_rules`, `elasticstack_elasticsearch_query_ruleset`, plus `elasticstack_kibana_security_enable_rule`. The counter is that all of those are *qualified* nouns: they are security detection rules and query rulesets, and they read as such. None of them claims the bare term. Alerting v2 is the only one whose product name for the object is just "rule", and it is the engine the rest of the stack is converging on. Reviewers who disagree should say so now: this is the cheapest moment to change it.

**Do not reuse `elasticstack_kibana_alerting_rule` for v2.** Renaming v1 vacates the name, but reclaiming it is not safe. A practitioner who upgrades the provider without editing configuration would have Terraform read existing v1 state against the v2 schema. Renaming v1 is a supported *address* move — `terraform-plugin-framework` v1.19.0 (in `go.mod`) provides `resource.ResourceWithMoveState`, and the repo pins Terraform 1.15.8, well past the 1.8 floor for cross-type `moved` blocks — so v1 state can be carried to the new address with a `moved` block and no destroy. Reusing the vacated name is a *schema* substitution and has no such mechanism.

### Decision 3 — release path

Register through `Provider.experimentalResources(...)` in `provider/plugin_framework.go`, alongside `kibanatag.NewResource` and `managedintegration.NewResource`. Practitioners opt in with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`. Because `make docs-generate` runs `tfplugindocs` with that variable set to `false`, no registry documentation page is produced — which is correct while the upstream API is `availability.stability: 'experimental'` and gated off by default.

This mirrors how `elasticstack_kibana_dashboard` shipped and later graduated (`archive/2026-05-14-graduate-kibana-dashboard`): graduation was a one-line registration move plus a spec requirement and generated docs, with no schema or state change. The schema in this proposal is designed so that [#678](https://github.com/elastic/rna-program/issues/678) can be the same one-line move.

## Contingent API changes

Each item below was found by walking the Terraform mapping and hitting something the API does not currently permit. They are ordered by impact on the provider. Item 1 changes the resource's apply semantics; the rest are ergonomics, correctness, or usability of the surrounding surface.

**1. `enabled` is not writable on create or update.** `createRuleDataSchema` has no `enabled` field; `ruleResponseSchema` does. `RulesClient.createRule` hard-codes `enabled: true`, and `upsertRule` copies the existing value forward. The only way to change it is `POST /rules/{id}/_enable` or `_disable`.

For Terraform this means every apply that sets `enabled = false` is **two calls that are not atomic**: create/replace the rule (which starts it running), then disable it. A failure in between leaves a rule enabled that the configuration says should be disabled — and a briefly live rule can fire real notifications.

The provider already carries the scar tissue from this pattern twice: `internal/clients/kibanaoapi/alerting_rule.go` has a `reconcileRuleEnabled` follow-up for v1, and `elasticstack_kibana_security_enable_rule` is an entire resource type that exists only because the detection rule API made the same choice. **Ask: accept `enabled` in the `PUT /rules/{id}` body (and in `PATCH`).** The `_enable`/`_disable` routes can stay for UI use.

**2. Response casing is inconsistent.** `ruleResponseSchema` mixes snake_case configuration fields (`time_field`, `recovery_strategy`, `no_data_strategy`, `state_transition`) with camelCase server fields (`createdBy`, `createdAt`, `updatedBy`, `updatedAt`). The find API compounds it: the request takes `per_page`, the response returns `perPage`. This flows straight into the generated Go types and into every mapping function. **Ask: pick one convention for the v2 surface — snake_case, matching the request bodies.**

**3. Two different fields are called "version".** `metadata.version` is an integer configuration counter incremented on every mutation (surfaced on rule events as `rule.version`). Top-level `version` is the saved object optimistic-concurrency token, a string. Both appear in the same response object. The Terraform schema has to expose both, and no naming makes that unconfusing. **Ask: rename one — `metadata.revision` for the counter, or `occ_token`/`sequence_token` for the OCC string.**

**4. Privilege names embed `alerting-v2`.** The API privileges for rules are `read-alerting-v2-rules` and `write-alerting-v2-rules`. These are user-authored strings in `elasticstack_kibana_security_role` and `elasticstack_elasticsearch_security_role` configurations. If classic alerting is retired and these are renamed, every Terraform-managed role definition that grants them breaks. **Ask: confirm the names are permanent, or settle them before the API is public.**

**5. The `alerting:v2:enabled` gate defaults to off and cannot be set declaratively.** Every v2 route 503s with `ALERTING_DISABLED` until the advanced setting is enabled, and the provider has no resource for Kibana advanced settings. A practitioner cannot express "enable v2, then create these rules" in one configuration; they must click it in the UI first, and Terraform cannot depend on it. **Ask (either):** expose the setting through an API the provider can wrap, or accept that the provider will need an `elasticstack_kibana_advanced_setting` resource before v2 is usable end-to-end as code. Until then the resource must produce a clear diagnostic on 503 rather than a bare HTTP error.

**6. `rule_response.metadata` is inlined rather than shared.** In the OAS produced by [kibana#279519](https://github.com/elastic/kibana/pull/279519), `alerting_v2_new_rule.metadata` is a `$ref` to the named `alerting_v2_rule_metadata` component, but `alerting_v2_rule_response.metadata` is an anonymous inline object — because `ruleResponseSchema` builds it with `metadataSchema.extend({ version })`, which produces a new unnamed schema. The generated Go client therefore gets a shared metadata type on the write path and an anonymous struct on the read path, and the provider has to map the same fields twice. **Ask: give the extended response metadata its own `.meta({ id })`** (for example `alerting_v2_rule_response_metadata`). This is a small addition to the naming work already in that PR.

## OAS status and generated client shape

The OAS work is in flight in [kibana#279519](https://github.com/elastic/kibana/pull/279519), which adds `--include-path /api/alerting/v2/` to `.buildkite/scripts/steps/checks/capture_oas_snapshot.sh` and attaches `.meta({ id })` names to every public v2 schema. Component ids are snake_case and namespaced `alerting_v2_` (alerting v1 already registers bare ids such as `rule_response`, so a bare id would collide), with `new_`/`update_` prefixes on create and update bodies. Discriminated-union variants are named individually so the OAS emits discriminator mappings. The PR is currently draft with CI red on **Check OAS Snapshot**.

**The provider codegen has already been validated against that branch** by Kibana Core, so the generated Go type shapes are known rather than hypothetical. `transform_schema.go` and `oapi-codegen` both exit 0, and:

- Named components become named, reusable Go types. `RuleQuery`, `RuleSchedule`, `RuleMetadata`, `RuleGrouping`, and `RuleArtifact` come through as shared types rather than re-inlined anonymous structs. (One exception, found while checking the bundle for this proposal and not visible in the codegen run: the rule *response* re-inlines its `metadata` rather than referencing the shared component — see contingent change 6 above.)
- The four real unions render as `json.RawMessage` unions with typed accessors. `alerting_v2_rule_query` carries a proper `discriminator` on `format` with mappings to the two named variants.
- The v2 surface is purely additive against the provider baseline — there are no existing `AlertingV2` types.

**Known issue that constrains this design.** zod's `.nullable()` serialises to `anyOf: [T, {nullable: true}]`, which `oapi-codegen` turns into a `json.RawMessage` union wrapper rather than a pointer. `RuleResponse.CreatedBy` therefore arrives as a union instead of `*string`. This is zod-v4 → OAS conversion behaviour, upstream of the naming work and not fixable by it, and it is the biggest ergonomics hit in the generated types.

Counting the pattern across the v2 components and paths in that branch's bundle gives **exactly 37 instances**. Of those, four are on the rule read/write path this resource uses, and their distribution is itself an argument for the endpoint choice in `design.md`:

| Component | Instances | On this resource's path? |
|-----------|-----------|--------------------------|
| `alerting_v2_new_rule` (PUT body) | 1 (`state_transition`) | yes |
| `alerting_v2_rule_response` | 3 (`createdBy`, `updatedBy`, `state_transition`) | yes |
| `alerting_v2_update_rule` (PATCH body) | 8 | no — this design uses PUT |
| other v2 components | 25 | no |

`design.md` proposes handling these in the codegen layer with a generic `transform_schema.go` transformer that collapses the wrapper back to a nullable scalar, following the existing `transformRemoveAnyOfWhenOneOfPresent` and `transformOmitEmptyNullable` precedent, with a per-field unwrap helper in `internal/clients/kibanaoapi` as the fallback.

**Design does not block on the PR merging.** The zod schemas in `@kbn/alerting-v2-schemas` are the source the OAS is generated from, and they are authoritative today; every schema claim in this change is taken from them and cross-checked against the bundle on the PR branch. Only implementation ([rna-program#676](https://github.com/elastic/rna-program/issues/676)) needs the published bundle.

## Capabilities

### New Capabilities

- `kibana-rule`: Terraform resource `elasticstack_kibana_rule` for Kibana Alerting v2 rules — schema, API mapping, CRUD endpoint selection, identity and import, computed and volatile field handling, plan-time cross-field validation, experimental registration, and acceptance-test expectations.

### Modified Capabilities

- _(none)_ — the v1 `kibana-alerting-rule` capability is untouched. Renaming that resource to `elasticstack_kibana_alerting_rule_classic` is argued in Decision 2 but delivered by a separate change.

## Impact

- **Specs**: new delta at `openspec/changes/add-kibana-alerting-v2-rule-resource/specs/kibana-rule/spec.md`, promoted to `openspec/specs/kibana-rule/spec.md` on archive.
- **Implementation (future, [#673](https://github.com/elastic/rna-program/issues/673))**: new package `internal/kibana/rule/`; new wrapper `internal/clients/kibanaoapi/rule_v2.go`; registration in `provider/plugin_framework.go` `experimentalResources()`.
- **Generated client (future, [#676](https://github.com/elastic/rna-program/issues/676))**: bump `github_ref` in `generated/kbapi/Makefile` past the merge of [kibana#279519](https://github.com/elastic/kibana/pull/279519) and regenerate; possibly one new transformer in `transform_schema.go`.
- **Upstream**: six contingent API requests above, to be filed against the Alerting v2 team. Item 1 (writable `enabled`) is the only one that changes this design if granted — see `design.md`.
- **No impact** on existing resources, state, or the default provider surface. The resource is unreachable without `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`.
