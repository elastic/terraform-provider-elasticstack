> **Status.** `elasticstack_kibana_rule` ships as an **experimental** technical-preview resource: registered via `experimentalResources()`, gated by `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, no generated registry docs. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678).

## Why

Kibana is shipping **Alerting v2**, an ES|QL-first alerting engine with its own saved object type (`alerting_rule`), its own HTTP surface under `/api/alerting/v2/`, and its own notification model. It is not a new rule type inside classic alerting: there is no dual-read, no in-place conversion from v1, and the two ID spaces are disjoint. Practitioners who adopt v2 have no way to manage rules as code today.

This change designs a new Terraform resource for v2 rules. It is a **design-only proposal** — the deliverable is this change directory, not provider code. Implementation is tracked separately in [rna-program#673](https://github.com/elastic/rna-program/issues/673) under epic [#655](https://github.com/elastic/rna-program/issues/655).

There is a second reason to do this now, while the API is still experimental. The v1 resource models rule configuration as an opaque `params` JSON string. That defeats plan-time validation, forces the provider to carry a bespoke params validator with a rule-type override table (`kibana-alerting-rule` [REQ-018–REQ-021](../../specs/kibana-alerting-rule/spec.md#requirement-plan-time-params-validation-req-018req-021), [REQ-043](../../specs/kibana-alerting-rule/spec.md#requirement-xpackuptimealertsmonitorstatus-primary-params-struct-req-043), [REQ-044](../../specs/kibana-alerting-rule/spec.md#requirement-xpackuptimealertsmonitorstatus-legacy-fallback-filters-completeness-req-044), [REQ-051](../../specs/kibana-alerting-rule/spec.md#requirement-discriminator-validation-coverage-guard-req-051), [REQ-052](../../specs/kibana-alerting-rule/spec.md#requirement-params-validation-override-table-req-052)), and is a longstanding practitioner complaint. v2's rule body is a single closed schema with no per-rule-type polymorphism, so the provider can finally express the whole rule as typed Terraform attributes. Getting that shape right — and feeding back the API changes that make it possible — is cheapest before the API is GA.

## What Changes

- Add a new capability `kibana-rule` describing the Terraform resource **`elasticstack_kibana_rule`**, covering: full attribute-by-attribute schema mapped to the v2 rule API, CRUD endpoint selection, composite identity and import, handling of computed and volatile fields, and plan-time cross-field validation for the constraints the API enforces.
- Register the resource in `experimentalResources()` behind `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`, as a technical preview with no generated registry documentation. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678) and is designed to need no schema, API, or state change at that point.
- Record a set of **contingent API changes** to request from the Alerting v2 team (see below). One of them — writable `enabled` — materially changes the resource's apply semantics, so it is called out as a decision the design depends on rather than a wish list.

**Not in this change:** any Go code, `generated/kbapi` regeneration, an action-policy resource.

### Decision 1 — a separate resource type, not a discriminator on the existing one

The obvious alternative is an `engine = "v1" | "v2"` (or `alerting_version`) field on `elasticstack_kibana_alerting_rule`. That alternative is a poor fit, for four reasons:

1. **Schema overlap with v1 is close to zero.** v2 has no `rule_type_id`, no `params`, and no `consumer`. v1 has no `kind`, no `query`, no `time_field`, no `recovery_strategy`/`no_data_strategy`/`state_transition`, and no `grouping`. The only fields that survive are `name`, `tags`, and the schedule interval — and in v2 the first two moved inside `metadata` and the third is `schedule.every`.
2. **Actions moved off the rule entirely.** v1's `actions` blocks (with `frequency` and `alerts_filter`) have no v2 counterpart — notifications are no longer configured on the rule. A single resource type would carry a large block that is valid for exactly one value of the discriminator.
3. **The saved object types and ID spaces are disjoint** (`alert` vs `alerting_rule`). `terraform import` takes an ID and a resource type; with a discriminator the provider could not tell which API to query, so import would be ambiguous or would need a probe-both-and-guess step.
4. **`ForceNew` on the discriminator would present destroy-and-recreate as if it were a migration.** No conversion exists. A practitioner who flips `engine` would see a plan that looks like an upgrade and get a deleted rule plus a brand-new one with different history and a different ID.

There is a fifth, structural cost: a discriminated resource forces every attribute to `Optional` so that both variants can validate, which moves all the required/forbidden logic out of the schema and into hand-written `ValidateConfig` code. That is precisely the failure mode this proposal is trying to escape from with `params`.

### Decision 2 — naming

| Object | Terraform type |
|--------|----------------|
| v2 rule (this change) | `elasticstack_kibana_rule` |
| v1 / classic rule | `elasticstack_kibana_alerting_rule` (unchanged; rename is not viable — see below) |

Three arguments against carrying the version into the new type name (`…_v2`):

- **`_v2` names age badly.** The version suffix is meaningful only while both engines exist. Once classic alerting is retired, the provider is left with a resource permanently called `_v2` and no `_v1`, and renaming it then would be a second breaking change.
- **Dropping the `alerting_` prefix has precedent.** `elasticstack_kibana_maintenance_window` is an alerting-framework object and does not carry the prefix. More importantly, not every v2 rule produces an alert: `kind = "signal"` is a first-class rule kind that emits signals rather than alerts, so baking `alerting` into the type name is inaccurate for part of the resource's surface.
- **The product's own name for the v2 object is "rule".** That leaves the unqualified noun for the thing that is simply "a rule" going forward, while classic keeps the historical `alerting_rule` type name it already ships under.

**Classic rename is not viable.** Renaming `elasticstack_kibana_alerting_rule` would be a breaking change for customers. Classic therefore **keeps** that type name. Product docs may still say "classic"; the Terraform type name does not follow.

**Two main objections to the v2 name.** First, **"rule" is overloaded in this provider** — `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_install_prebuilt_rules`, `elasticstack_elasticsearch_query_ruleset`, plus `elasticstack_kibana_security_enable_rule`. Those are *qualified* nouns (security detection rules, query rulesets) and none claims the bare term. Alerting v2 is the platform-level alerting implementation in Kibana going forward, so it is the component that should take the bare name even though "rule" is overloaded elsewhere. Second, **"rule" alone may not carry enough product meaning** for practitioners who think in terms of an "alerting" product and expect that word in the type name. The counter is that the product's own name for the object is just "rule", and that `kind = "signal"` makes a permanent `alerting_` qualifier inaccurate. Reviewers who disagree should say so now: this is the cheapest moment to change it.

**Do not reuse `elasticstack_kibana_alerting_rule` for v2.** That name stays with classic. Putting v2 behind it would be a *schema* substitution: a practitioner who upgrades the provider without editing configuration would have Terraform read existing v1 state against the v2 schema. There is no safe mechanism for that, and this change does not need the classic name — it introduces a new type alongside it.

### Decision 3 — release path

Register through `Provider.experimentalResources(...)` in `provider/plugin_framework.go`, alongside `kibanatag.NewResource` and `managedintegration.NewResource`. Practitioners opt in with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`. Because `make docs-generate` runs `tfplugindocs` with that variable set to `false`, no registry documentation page is produced — which is correct while the upstream API is `availability.stability: 'experimental'`.

This mirrors how `elasticstack_kibana_dashboard` shipped and later graduated (`archive/2026-05-14-graduate-kibana-dashboard`): graduation was a one-line registration move plus a spec requirement and generated docs, with no schema or state change. The schema in this proposal is designed so that [#678](https://github.com/elastic/rna-program/issues/678) can be the same one-line move.

## Contingent API changes

Each item below was found by walking the Terraform mapping and hitting something the API does not currently permit. They are ordered by impact on the provider. Item 1 changes the resource's apply semantics; the rest are ergonomics, correctness, or usability of the surrounding surface.

**1. `enabled` is not writable on create or update.** `createRuleDataSchema` has no `enabled` field; `ruleResponseSchema` does. `RulesClient.createRule` hard-codes `enabled: true`, and `upsertRule` copies the existing value forward. The only way to change it is `POST /rules/{id}/_enable` or `_disable`.

For Terraform this means every apply that sets `enabled = false` is **two calls that are not atomic**: create/replace the rule (which starts it running), then disable it. A failure in between leaves a rule enabled that the configuration says should be disabled — and a briefly live rule can fire real notifications.

The provider already provides workarounds for this pattern twice: `internal/clients/kibanaoapi/alerting_rule.go` has a `reconcileRuleEnabled` follow-up for v1, and `elasticstack_kibana_security_enable_rule` is an entire resource type that exists only because the detection rule API made the same choice. **Ask: accept `enabled` in the `PUT /rules/{id}` body (and in `PATCH`).** The `_enable`/`_disable` routes can stay for UI use.

**2. Response casing is inconsistent.** `ruleResponseSchema` mixes snake_case configuration fields (`time_field`, `recovery_strategy`, `no_data_strategy`, `state_transition`) with camelCase server fields (`createdBy`, `createdAt`, `updatedBy`, `updatedAt`). The find API compounds it: the request takes `per_page`, the response returns `perPage`. This flows straight into the generated Go types and into every mapping function. **Ask: pick one convention for the v2 surface — snake_case, matching the request bodies.**

**3. Two different fields are called "version".** `metadata.version` is an integer configuration counter incremented on every mutation (surfaced on rule events as `rule.version`). Top-level `version` is the saved object optimistic-concurrency token, a string. Both appear in the same response object. The Terraform schema has to expose both, and no naming makes that unconfusing. **Ask: rename one — `metadata.revision` for the counter, or `occ_token`/`sequence_token` for the OCC string.**

## Generated client

v2 OAS lands via [kibana#279519](https://github.com/elastic/kibana/pull/279519); implementation ([#676](https://github.com/elastic/rna-program/issues/676)) regenerates `kbapi` from that bundle. Schema claims in this change are taken from the zod sources and the regenerated OAS on that PR.

Component ids are `Kibana_HTTP_APIs_alerting_*`. Local codegen can run against a copy of the PR's `oas_docs/output/kibana.yaml` before merge. Committing `github_ref` in `generated/kbapi/Makefile` waits until the PR is on `elastic/kibana`, because that Makefile fetches from `raw.githubusercontent.com/elastic/kibana/$(github_ref)/...`.

## Capabilities

### New Capabilities

- `kibana-rule`: Terraform resource `elasticstack_kibana_rule` for Kibana Alerting v2 rules — schema, API mapping, CRUD endpoint selection, identity and import, computed and volatile field handling, plan-time cross-field validation, experimental registration, and acceptance-test expectations.

### Modified Capabilities

- _(none)_ — the v1 `kibana-alerting-rule` capability is untouched. Its Terraform type name stays `elasticstack_kibana_alerting_rule`; Decision 2 records why a rename is not viable.

## Impact

- **Specs**: new delta at `openspec/changes/add-kibana-alerting-v2-rule-resource/specs/kibana-rule/spec.md`, promoted to `openspec/specs/kibana-rule/spec.md` on archive.
- **Implementation (future, [#673](https://github.com/elastic/rna-program/issues/673))**: new package `internal/kibana/rule/`; new wrapper `internal/clients/kibanaoapi/rule_v2.go`; registration in `provider/plugin_framework.go` `experimentalResources()`.
- **Generated client (future, [#676](https://github.com/elastic/rna-program/issues/676))**: regenerate `kbapi` from the [kibana#279519](https://github.com/elastic/kibana/pull/279519) OAS; bump `github_ref` in `generated/kbapi/Makefile` once that PR is on `elastic/kibana`.
- **Upstream**: three contingent API requests above, to be filed against the Alerting v2 team. Item 1 (writable `enabled`) is the only one that changes this design if granted — see `design.md`.
- **No impact** on existing resources, state, or the default provider surface. The resource is unreachable without `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`.
