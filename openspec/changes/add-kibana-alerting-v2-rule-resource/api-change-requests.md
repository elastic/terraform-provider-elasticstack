# Alerting v2 — API change requests from the Terraform provider

> **Status.** The corresponding Terraform resource (`elasticstack_kibana_rule`) ships as an **experimental** technical-preview resource: registered via `experimentalResources()`, gated by `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, no generated registry docs. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678).

Companion to `proposal.md` in this change. Each section is written to stand alone so it can be pasted into a GitHub issue against the Alerting v2 team without the reader needing the Terraform context.

Every claim below is taken from the zod schemas in `x-pack/platform/packages/shared/response-ops/alerting-v2-schemas/src/`, the routes and client in `x-pack/platform/plugins/shared/alerting_v2/server/`, or the OAS bundle produced on the branch of [kibana#279519](https://github.com/elastic/kibana/pull/279519).

**Tracking:** [rna-program#673](https://github.com/elastic/rna-program/issues/673) (resource), epic [#655](https://github.com/elastic/rna-program/issues/655), graduation [#678](https://github.com/elastic/rna-program/issues/678).

## What already works

Worth stating up front, because it makes the list below short: **`PUT /api/alerting/v2/rules/{id}` with a client-supplied id is exactly what Terraform needs.** It gives one write path for create and update, idempotent convergence when an apply is interrupted, and full-replace semantics that match a declarative tool. Most Kibana APIs the provider wraps do not have this, and the resource design depends on it. If none of the requests below are granted, the resource still works — item 1 costs atomicity, and the rest cost ergonomics or a future breaking change.

## Summary

| # | Request | Category | Cost of deferring past GA |
|---|---------|----------|---------------------------|
| 1 | Accept `enabled` in the `PUT`/`PATCH` rule body | Correctness | Apply stays non-atomic; rules briefly run when configured off |
| 2 | Settle casing on the v2 surface | Ergonomics | Permanent mapping-code cost; not schema-forcing |
| 3 | Rename one of the two fields called `version` | Breaking later | Renames Terraform attributes — breaking provider change |
| 4 | zod `.nullable()` → `anyOf` in generated OAS | Upstream | 37 awkward Go types; provider has a workaround |

Numbering matches `proposal.md` items 1–3; item 4 here is the zod/`anyOf` note (provider workaround; no Alerting v2 team action).

---

## 1. Accept `enabled` in the rule create/replace body

**Category:** correctness. This is the only functional defect in the list.

**Current behaviour.** `createRuleDataSchema` has no `enabled` field, so neither `POST /rules` nor `PUT /rules/{id}` accepts one. `ruleResponseSchema` does return it. In `lib/rules_client/rules_client.ts`:

- `createRule` hard-codes `enabled: true` (with the comment "A freshly created rule is always enabled").
- `upsertRule` carries `existingAttrs.enabled` forward on replace.

The only way to change it is `POST /rules/{id}/_enable` or `POST /rules/{id}/_disable`.

**Why this is a problem for Terraform.** `enabled` is an ordinary attribute in a practitioner's configuration, so every apply that sets `enabled = false` becomes two calls that are not atomic: create or replace the rule — which starts it running — and then disable it. A failure between the two leaves a rule enabled that the configuration says should be off, and Terraform reports an error without a way to finish the job.

This is worse in v2 than it would be in classic alerting: a rule that runs for a few seconds can fire real notifications before the disable call lands.

**Prior art in the provider, twice.** `internal/clients/kibanaoapi/alerting_rule.go` carries a `reconcileRuleEnabled` follow-up for classic alerting. `elasticstack_kibana_security_enable_rule` is an entire resource type that exists only because the detection rule API made this same choice. We would rather not add a third.

**Ask.** Accept `enabled` in the `PUT /rules/{id}` body, and in `PATCH /rules/{id}`. Keep `_enable` and `_disable` for UI use.

**Provider impact if granted.** Small and contained: send the field, delete the follow-up call. It changes no Terraform schema and does not block graduation either way.

## 2. Settle casing on the v2 surface

**Category:** ergonomics. Not schema-forcing.

**Current behaviour.** `ruleResponseSchema` mixes conventions: snake_case configuration fields (`time_field`, `recovery_strategy`, `no_data_strategy`, `state_transition`) alongside camelCase server fields (`createdBy`, `createdAt`, `updatedBy`, `updatedAt`). The find API compounds it — the request takes `per_page` and the response returns `perPage`.

**Why this is a problem for Terraform.** It flows into the generated Go types and into every mapping function the provider writes, permanently. It is a readability and maintenance cost rather than a defect.

**To be clear, this does not force a Terraform schema change.** The provider exposes `created_by` and `created_at` regardless of what the API calls them, so a later rename would be absorbed in the mapping layer. That is why this ranks below item 3 despite looking similar.

**Ask.** Pick one convention for the v2 surface — snake_case, matching the request bodies — while the API is still experimental.

## 3. Rename one of the two fields called `version`

**Category:** breaking if deferred. Decide before GA.

**Current behaviour.** `ruleResponseSchema` returns two unrelated fields with the same name:

- `metadata.version` — an integer configuration revision counter, incremented on every mutation (including `_enable` and `_disable`), surfaced on rule events as `rule.version`.
- `version` — the saved object optimistic-concurrency token, a string, used by `PATCH` and internally by `PUT`.

**Why this is a problem for Terraform.** Both are computed attributes the provider must expose, so both appear in the resource schema and in generated documentation. There is no naming that makes `version` and `metadata.version` unconfusing to a practitioner reading a plan diff, and they behave differently enough that the distinction matters.

**Ask.** Rename one. `metadata.revision` for the counter reads well, or `occ_token` / `sequence_token` for the concurrency string.

**Cost of deferring.** The Terraform attribute names follow the API. Renaming after the resource graduates is a breaking provider change requiring a state migration for every managed rule. This is the only item on this list that would force a Terraform schema change, which is why it should land before [#678](https://github.com/elastic/rna-program/issues/678).

## 4. zod `.nullable()` produces an `anyOf` wrapper in the generated OAS

**Category:** upstream conversion behaviour. Provider has a workaround. Flagged because it is broader than alerting.

**Current behaviour.** zod v4's `.nullable()` serialises to `anyOf: [T, {nullable: true}]`. `oapi-codegen` reads any `anyOf` as a union and emits a `json.RawMessage` wrapper with typed accessors rather than a pointer, so `RuleResponse.CreatedBy` arrives as a union instead of `*string`.

Counting the pattern across the v2 components and paths in the bundle on the [kibana#279519](https://github.com/elastic/kibana/pull/279519) branch gives **exactly 37 occurrences**. On the rule read/write path this resource uses:

| Component | Count |
|-----------|-------|
| `alerting_v2_new_rule` (PUT body) | 1 |
| `alerting_v2_rule_response` | 3 |
| `alerting_v2_update_rule` (PATCH body) | 8 |
| other v2 components | 25 |

**Why we are flagging it rather than just fixing it.** This is zod v4 → OAS conversion behaviour, upstream of the naming work in that PR and not fixable by it. It will affect every Kibana route that migrates to zod v4, not only alerting v2, so a fix in the conversion layer would be worth considerably more than our workaround.

**Provider-side plan meanwhile.** Add a transformer to `generated/kbapi/transform_schema.go` that collapses `anyOf: [T, {nullable: true}]` into `T` with `nullable: true`, so `oapi-codegen` emits a pointer. That file already has two global transformers with the same shape, and one generic rule fixes all 37 at once.

**Ask.** No action required from the Alerting v2 team. Raised for whoever owns the zod → OAS conversion, and recorded here so the count and the affected components are written down.
