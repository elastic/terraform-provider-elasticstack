> **Status.** `elasticstack_kibana_rule` ships as an **experimental** technical-preview resource: registered via `experimentalResources()`, gated by `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, no generated registry docs. Graduation is [rna-program#678](https://github.com/elastic/rna-program/issues/678).

## Context

See `proposal.md` — Why, for motivation, and the naming and release-path decisions. This document covers how the resource maps onto the API and why each mapping choice was made. Requirements are in `specs/kibana-rule/spec.md`; this file does not restate them.

**Source of truth.** Alerting v2 is not yet in Kibana's published OAS bundle on `main`. Field shapes come from the zod schemas in `x-pack/platform/packages/shared/response-ops/alerting-v2-schemas/src/` (principally `rule_data_schema.ts`), cross-checked against the OAS on [kibana#279519](https://github.com/elastic/kibana/pull/279519); artifact `data` follows [kibana#281751](https://github.com/elastic/kibana/pull/281751).

**Provider mechanics assumed.** The resource follows the current Plugin Framework pattern: a package under `internal/kibana/` wiring `entitycore.NewKibanaResource[Model]`, with the envelope owning `Metadata`, `Configure`, `Schema` (injecting `kibana_connection` and `timeouts`), and the CRUD orchestration, and the concrete resource supplying a schema factory, a model, four lifecycle callbacks, and `ImportState`. `internal/kibana/maintenance_window/` is the closest recent example. API calls go through a wrapper in `internal/clients/kibanaoapi/` that returns `diag.Diagnostics` rather than `error`.

## Goals / Non-Goals

**Goals**

- A mapping in which every rule field is a typed Terraform attribute (with `artifacts[].data` as the sole JSON object string), and every statically checkable API constraint is a schema validator.
- One write path for create and update, so a re-apply after a partial failure converges.
- Volatile server fields that never produce a diff on an unchanged configuration.
- A design that survives graduation ([#678](https://github.com/elastic/rna-program/issues/678)) without a schema, API, or state change.

**Non-Goals**

- Modelling action policies (a separate future resource, not part of this change).

## Decisions

### D1. `PUT /rules/{id}` for both create and update; never `POST` or `PATCH`

The API offers three ways to write a rule:

- `POST /rules` — create, server picks the id.
- `PUT /rules/{id}` — create or replace the rule at a known id.
- `PATCH /rules/{id}` — change only the fields you send.

This design uses `PUT` for both create and update.

**Why not `PATCH`.** Terraform configs describe the whole object, not a list of edits. If a practitioner deletes a block from their `.tf` file, the next apply should remove it from the rule. With `PUT`, that happens for free — you send the config as the new body and anything missing is gone. With `PATCH`, the provider has to notice each removed field and send an explicit `null` for it, and the API treats "field not sent" and "field sent as null" as different things on nearly every attribute. Getting that translation wrong is how you end up with perpetual diffs and "omit this key so Kibana doesn't wipe the value" special cases — both of which the classic v1 resource already carries. `PUT` does not have that class of bug.

**Why not `POST` for create.** If an apply is interrupted after the rule is created but before Terraform records the id, a retry with `POST` fails with "already exists". `PUT` to the same id just replaces the rule and converges. That only works when we know the id up front, which is D2.

**The cost.** `PUT` does not take a concurrency token, so two writers racing will overwrite each other. That is last-write-wins, which is how the rest of this provider already behaves and what Terraform assumes: the configuration owns the object. Getting conflict detection would mean switching to `PATCH`, storing the token in state, and handling 409s — including from the provider's own enable/disable calls, which bump the token. Not worth it.

Full replace also means a writable field we forget to send gets reset. D8 covers how the write path protects fields the Terraform schema does not model.

Alternatives considered: `POST` to create and `PATCH` to update (two code paths, the partial-update problems above, and no safe retry after a partial create); `POST` to create and `PUT` to update (still no safe retry, and update still needs a known id anyway).

### D2. Always address the rule by a known id; generate one when the practitioner does not

`PUT` needs an id before the first write. When `rule_id` is not configured, the provider generates one at apply time and persists it. This keeps a single write path and preserves the property that a practitioner who *does* pin `rule_id` gets genuine idempotency.

The alternative — `POST` when `rule_id` is unset, `PUT` when it is set — was rejected because it makes the resource's failure and recovery behaviour depend on whether an optional attribute happens to be set, which is hard to document and harder to test.

`rule_id` and `space_id` are `RequiresReplace`, matching the classic resource and the fact that the rule saved object is `namespaceType: 'multiple-isolated'` — a rule cannot move between spaces.

### D3. The query union becomes two mutually exclusive blocks, not one block with a `format` field

`querySchema` is `z.discriminatedUnion('format', [composed, standalone])`. HCL has no sum type, so there are two ways to land it:

1. A single `query` block with `format` plus every field from both variants, all `Optional`, and hand-written validation deciding which combination is legal.
2. Two top-level blocks, `composed_query` and `standalone_query`, exactly one of which may be set, with each field carrying its real `Required`/`Optional` marking.

Option 1 is rejected for the same reason `proposal.md` rejects an `engine` discriminator on the classic resource: a discriminator inside one schema forces every attribute to `Optional`, which moves the required-field logic out of the schema and into imperative code, where it does not appear in generated documentation, does not show up in `terraform providers schema`, and has to be reimplemented for every new variant.

Option 2 keeps the constraints in the schema. `composed_query.base` and `composed_query.breach.segment` are genuinely `Required`; so is `standalone_query.breach.query`. Exactly-one-of is a single object validator at the top level. `format` is not exposed — the provider derives it, since it is fully determined by which block is present, and a user-set `format` that disagreed with the configured fields would just be a new error case to validate.

**Composed — API**

```json
{
  "query": {
    "format": "composed",
    "base": "FROM metrics | WHERE host.name IS NOT NULL",
    "breach": {
      "segment": "| WHERE cpu > 90"
    },
    "recovery": {
      "segment": "| WHERE cpu < 70"
    }
  }
}
```

**Composed — Terraform**

```hcl
composed_query {
  base = "FROM metrics | WHERE host.name IS NOT NULL"

  breach {
    segment = "| WHERE cpu > 90"
  }

  recovery {
    segment = "| WHERE cpu < 70"
  }
}
```

**Standalone — API**

```json
{
  "query": {
    "format": "standalone",
    "breach": {
      "query": "FROM metrics | WHERE cpu > 90"
    },
    "recovery": {
      "query": "FROM metrics | WHERE cpu < 70"
    },
    "no_data": {
      "query": "FROM metrics | STATS count()"
    }
  }
}
```

**Standalone — Terraform**

```hcl
standalone_query {
  breach {
    query = "FROM metrics | WHERE cpu > 90"
  }

  recovery {
    query = "FROM metrics | WHERE cpu < 70"
  }

  no_data {
    query = "FROM metrics | STATS count()"
  }
}
```

The `breach` / `recovery` / `no_data` wrapper objects are preserved as nested single blocks rather than flattened to `breach_segment` and friends. Each currently holds exactly one field, so flattening would read better today, but the wrappers exist precisely so the API can add fields to them; flattening now means a schema break later.

Cross-field validation that genuinely spans attributes — the seven refinements in spec REQ-007 — cannot be expressed in the schema and go in `ValidateConfig`, guarded on known values.

### D4. Volatile field handling

Four fields change on every successful write: `metadata.version` and `version` (both incremented or reissued by `writeRuleAttrs`), `updatedAt`, and `updatedBy`. `metadata.version` also bumps on `_enable` and `_disable`, which the provider itself calls.

They are `Computed` **without** `UseStateForUnknown`. That is the deliberate opposite of the usual provider habit. `UseStateForUnknown` says "if nothing else changed, this value will not change", which is true for `created_at` but false for these four: if anything else in the resource changes, they will change too. Marking them plain `Computed` gives the correct behaviour in both directions — Terraform keeps the prior value when the plan is otherwise empty, and shows known-after-apply when it is not.

`created_at` and `created_by` do get `UseStateForUnknown`: `upsertRule` explicitly carries `createdBy` and `createdAt` forward from the existing object on replace.

The `version` OCC token is exposed as a read-only attribute rather than hidden. It is useful for out-of-band tooling and for debugging conflicts, and hiding a field that appears in the API response tends to generate questions. It is never sent back; `PUT` does not accept it.

`proposal.md` contingent change 3 asks for one of the two "version" fields to be renamed. If that happens the Terraform attribute names change with it, which is a schema change — this is the strongest reason to land that request before graduation.

### D5. `enabled` needs a second call, and that is worth fixing upstream

`createRuleDataSchema` has no `enabled`. `RulesClient.createRule` hard-codes `enabled: true`; `upsertRule` copies `existingAttrs.enabled` forward. So:

- create with `enabled = true` → one call;
- create with `enabled = false` → `PUT`, then `POST .../_disable`;
- update with no change to `enabled` → one call, because `PUT` preserves it;
- update that toggles `enabled` → `PUT`, then `_enable` or `_disable`.

The reconciliation mirrors `reconcileRuleEnabled` in `internal/clients/kibanaoapi/alerting_rule.go`, and runs before the authoritative read so state reflects the final server state.

The window between the two calls is the problem, and it is worse here than in v1. A create with `enabled = false` briefly produces a running rule that can fire notifications before the disable lands. That is the substance of contingent change 1.

If `enabled` becomes writable in the `PUT` body, the change is confined to the client wrapper: send the field, delete the follow-up. Spec REQ-011 is written so its scenarios hold either way, and `enabled` stays `Optional + Computed` with a `true` default in both worlds — so this is not a schema change and does not block graduation.

### D6. Space handling by request editor, not by path rewriting

The v2 rule saved object is `namespaceType: 'multiple-isolated'`, so rules are space-scoped and the composite `<space_id>/<rule_id>` identity applies.

`transform_schema.go` has a `spaceIdPaths` allowlist that rewrites selected paths into `/s/{spaceId}` variants, but the classic alerting resource does not use it — it passes `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to the generated client instead, rewriting the URL per request. This design does the same. It keeps the v2 paths out of the transform allowlist, so a `kbapi` regeneration that pulls in new v2 endpoints does not need a matching transform change, and it matches the resource this one is most likely to be read alongside.

### D7. Resource-level minimum version gate, value to be determined

The provider supports Elastic Stack 8.0+ and gates features with `entitycore.VersionRequirement` (see `maintenance_window/models.go` for the resource-level form). This resource will need one, since v2 does not exist on older stacks — a request against a stack without the plugin gets a 404 from the HTTP layer, which the resource would misread as "rule deleted" and try to recreate, looping.

The value is not yet knowable; see Open Questions. Until it is, the read path must distinguish "the rule is gone" from "this endpoint does not exist here". The safest available signal is that the engine-disabled 503 (spec REQ-013) and a 404 from an unregistered route are both non-`ALERTING_DISABLED` responses on a stack that has never had v2, so a version gate is the real fix rather than response sniffing.

### D8. Read before write, so full replace cannot clear fields the schema does not model

Under D1 every write sends the whole object, so a writable field missing from the request body is reset to its server-side default. Today no such field exists — the schema in `specs/kibana-rule/spec.md` covers all of `createRuleDataSchema`. The design has to hold when that stops being true.

Without that protection, every apply that touches a managed rule would send only the fields Terraform knows about. If Kibana later adds a writable field that the schema does not yet model — or if someone set that field in the UI — the next apply clears it. Terraform's plan is empty, because from its point of view nothing changed; the practitioner sees the loss only when the rule behaves differently in production. Silent data loss is the failure mode.

That miss is also easy for maintainers: `generated/kbapi/Makefile` pins a Kibana commit, Renovate bumps it and regenerates `kibana.gen.go` in the same PR (the most recent bump was roughly 2,400 lines), and a new optional field on `createRuleDataBaseSchema` lands somewhere in that diff with no obvious link to "every managed rule now gets this wiped on apply."

So the write path reads first: `GET` the rule, populate the request struct from the response, overwrite the fields Terraform owns, `PUT` the merged body. Fields the generated client knows about but the schema does not model survive untouched.

The overwrite must be unconditional for every modeled field, nulls included. Skipping nulls would turn full replace back into a partial update and break the semantics D1 exists to provide — removing a block from configuration has to clear the value server side (spec REQ-012).

Seeding the request from the response means copying between the response and request types. Both sides are plain pointers, so that copy is mechanical.

This is the usual pattern for full-replace APIs ([AzureRM best practices](https://hashicorp.github.io/terraform-provider-azurerm/topics/best-practices/); this repo already does it in `internal/fleet/agentpolicy/update.go`). Cost: one extra `GET` per update; create is unaffected. The wider read-to-write window is more last-write-wins exposure, which D1 already accepts.

It only preserves fields the pinned generated client knows about — a field on a newer Kibana that is not yet in `kibana.gen.go` still cannot round-trip. Keep Renovate cadence tight; that gap is not a reason to switch write strategies.

Read-before-write alone would quietly preserve unknown fields forever and never force anyone to model them. Pair it with a fail-closed unit test that compares the `PUT` body's JSON tags against a checked-in allowlist (modeled + explicitly waived). A new upstream field then fails the Renovate PR by name until someone models or waives it.

## Risks / Trade-offs

- **[Risk] The schema is designed against an experimental API that can still change.** The rule body has a snapshot test upstream (`rule_data_schema.test.ts`) that fails when a top-level field is added, so additions are visible in review — but they will still land as provider schema changes. → **Mitigation:** ship behind `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`, where schema changes are acceptable, and treat graduation ([#678](https://github.com/elastic/rna-program/issues/678)) as the freeze point. Land contingent change 3 (the two `version` fields) before then — it is the only contingent ask that would force a Terraform schema rename. Contingent change 1 (`enabled`) changes apply semantics but not the schema.

- **[Risk] `PUT` is last-write-wins.** A rule modified in the Kibana UI between Terraform's refresh and its write is silently overwritten. → **Mitigation:** this is standard Terraform behaviour and standard for this provider; document it in the resource description rather than reaching for `PATCH` and inheriting its problems (D1).

- **[Risk] A field added upstream that the schema does not model would be silently cleared by full-replace `PUT`.** This is data loss with no plan diff, and it arrives through a Renovate regeneration PR too large to review field by field. → **Mitigation:** read before write so unmodeled fields round-trip, plus a fail-closed field-accounting test that turns the regeneration PR red (D8).

- **[Risk] The non-atomic `enabled` reconciliation can leave a rule enabled when configuration says otherwise,** so a briefly live rule can fire notifications. → **Mitigation:** report the partial state explicitly (spec REQ-011) rather than failing silently, and pursue contingent change 1.

- **[Trade-off] Two top-level query blocks diverge from the API's single `query` object,** so the mapping is not one-to-one and the spec needs a mapping table. → Accepted: the alternative pushes required-field validation out of the schema (D3), which is the specific defect this change exists to avoid.

- **[Trade-off] Preserving the `breach` / `recovery` / `no_data` wrapper blocks costs a nesting level** for objects that hold one field each today. → Accepted for forward compatibility; flattening later would be a breaking schema change, unflattening would not have been needed.

## Migration Plan

Not applicable in the usual sense: this is a new resource with no prior state, no state upgrader, and no default-surface exposure. Practitioners opt in via `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, and removing the resource from a configuration destroys the rule as normal.

Two adjacent migrations are worth recording because they are frequently confused with this one:

- **Renaming the classic resource is not viable.** Renaming `elasticstack_kibana_alerting_rule` would be a breaking change for customers. Classic keeps its type name. See `proposal.md` Decision 2.
- **v1 → v2 rule migration** is a server-side concern and has no Terraform path. See Open Questions.

## Open Questions

1. **Minimum Kibana version for the gate in D7.** The version in which `/api/alerting/v2/rules` first ships as a public route is not yet determined; the feature is unreleased. Implementation must establish it and add a resource-level `VersionRequirement` with a diagnostic naming the version. This does not change the schema, the approach, or the task breakdown — only the constant — which is why it is deferred rather than resolved. Serverless short-circuits the provider's version checks, so serverless behaviour needs a separate confirmation.

2. **Interaction with server-side v1 → v2 rule migration.** If Elastic ships a server-side migration and a rule managed by `elasticstack_kibana_alerting_rule` is migrated, its saved object type changes from `alert` to `alerting_rule`. The v1 `GET` then 404s, the classic resource removes it from state (`kibana-alerting-rule` REQ-001), and the next apply tries to recreate it — producing either a duplicate rule or a confusing failure, with no signal to the practitioner that a migration happened. This is not something this change can fix from the v2 side, and the migration's shape is not yet known. It needs a decision with the Alerting v2 team about what a migrated rule looks like to the v1 API, and it may need a follow-up change to the classic resource (for example, detecting the type change and emitting a diagnostic that points at the v2 resource instead of silently planning a recreate). Recorded here so it is not discovered during a migration.
