## Context

Confirmed against the upstream Kibana OpenAPI spec pinned by `generated/kbapi/Makefile` (`oas_docs/output/kibana.yaml`, schema `kbn-as-code-query`):

```yaml
additionalProperties: false
properties:
  expression: { type: string, description: "A query expression in KQL or Lucene syntax." }
  language:   { type: string, enum: [kql, lucene] }
required: [expression, language]
```

There is no `oneOf` / object branch, and `additionalProperties: false` means Kibana rejects an object-valued `expression`. The generated Go type is `KibanaHTTPAPIsKbnAsCodeQuery{Expression string; Language ...}`. `generated/kbapi/dashboard-paths.json` only `$ref`s this schema and never overrides it, and `transform_schema.go` doesn't touch it either — so this change needs no client regeneration.

Current provider behavior (`internal/kibana/dashboard/models.go`, `schema.go`):

- Write: `text` → `Expression`, `json` → `Expression` (the JSON document stringified). `config_validators.go` / the schema's `ExactlyOneOf` validators enforce exactly one of `text`/`json`.
- Read: the provider cannot know which attribute the practitioner used, so it sniffs the returned string — if it starts with `{` and parses as a JSON object, it populates `json`; otherwise `text`.
- `json` has no API counterpart: Kibana always treats the value as a KQL/Lucene string, not as structured JSON. `text` and `json` are leftovers from an earlier "string | object" union shape that the API no longer has (if it ever did).
- Schema `Version: 1` today (`internal/kibana/dashboard/resource.go`), with a single v0 → v1 `UpgradeState` entry (`migrateV0ToV1` in `state_upgrade.go`, REQ-040) that restructures `options_list_control` / `range_slider_control` panels. That upgrade is orthogonal to this change and must keep working unchanged for state starting at v0.

## Goals / Non-Goals

**Goals**

- Root `query` matches the API shape exactly: `{ language, expression }`, both required when the block is configured.
- Existing state (v0 or v1) upgrades automatically to the new schema version with no data loss: whatever string Kibana was already receiving (from `text` or `json`) ends up in `expression`, unchanged.
- The lossy `{`-sniffing read heuristic is removed entirely — read becomes a direct, unambiguous mapping.

**Non-Goals**

- Preserving `text`/`json` as deprecated aliases (Option B from the issue's exploration comment). The read-back ambiguity this change removes is exactly what an alias approach would keep.
- Any change to non-root `query` blocks (Lens charts, discover-session, typed `vis` panels), which already use `expression` and are structurally independent (different Go models, different schema attributes).
- Any generated-client change; the upstream schema already matches.

## Decisions

1. **Schema (`internal/kibana/dashboard/schema.go`).** Replace the `text` / `json` attributes with a single required `expression` string attribute. Remove the `stringvalidator.ExactlyOneOf` validators on both. Update the `query` block's `MarkdownDescription` to describe `language` + `expression` instead of the union. Bump `Version: 1` → `Version: 2` on the resource schema (`resource.go`).

2. **Model (`internal/kibana/dashboard/models.go`).** `DashboardQueryModel` becomes `{ Language types.String; Expression types.String }`. Drop the `jsontypes.NormalizedValue`/`NormalizedNull` field and, if nothing else in the file still needs it, the `jsontypes` import.

3. **Write mapping.** `dashboardQueryToAPI` (or equivalent) becomes:
   ```go
   query.Language = ...
   query.Expression = m.Query.Expression.ValueString()
   ```
   Delete the `textKnown`/`jsonKnown` branching and the "Invalid dashboard query: exactly one of `query.text` or `query.json`" diagnostic — there is nothing left to validate once there's a single required field (the schema's `Required: true` on `expression` already guarantees it's set whenever `query` is configured).

4. **Read mapping.** Replace the `bytes.TrimSpace` / leading-`{` / `json.Unmarshal` sniffing block with:
   ```go
   q := &models.DashboardQueryModel{
       Language:   types.StringValue(string(data.Data.Query.Language)),
       Expression: types.StringValue(data.Data.Query.Expression),
   }
   ```
   `expression = ""` round-trips as an empty string, not null — this is the exact case reported in issue #5081 and must stay valid and driftless.

5. **State upgrade (`internal/kibana/dashboard/state_upgrade.go`).** Add `migrateV1ToV2`, operating on the raw state map via `stateutil.UnmarshalStateMap` / `MarshalStateMap` (the same pattern `migrateV0ToV1` uses):
   - If `query` is absent or null, leave it untouched.
   - If `query` is a non-null object: set `query["expression"]` from `query["text"]` when `text` is a non-null string; otherwise from `query["json"]` when present. Delete `text` and `json` from the map regardless.
   - The `json` value is copied as-is (it's already the stringified document Kibana was receiving as `Expression`), so there is no behavior change at upgrade time versus what Kibana already stored.

   `UpgradeState` becomes:
   ```go
   func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
       return map[int64]resource.StateUpgrader{
           0: {StateUpgrader: migrateV0ToV2}, // v0 -> v1 -> v2, composed
           1: {StateUpgrader: migrateV1ToV2},
       }
   }
   ```
   `migrateV0ToV2` runs `migrateV0ToV1`'s transform followed by `migrateV1ToV2`'s transform against the same `resp`, since the framework calls the upgrader keyed by the *stored* version and expects it to emit state at the *current* schema version in one call — it does not chain upgraders itself.

6. **Validators.** Remove the now-dead "exactly one of text/json" validator wiring in `schema.go` and any corresponding helper in `config_validators.go` if it has no other caller.

7. **No deprecation period.** This is a one-shot breaking rename, not a deprecate-then-remove cycle, because: (a) the resource is Kibana Technical Preview, (b) keeping `text`/`json` as aliases would keep the exact read-back ambiguity (which attribute to populate on read) that this change exists to remove, and (c) the state upgrader makes the migration automatic — practitioners only need to edit `.tf` source, not touch state.

## Tests

**Unit**

- `migrateV1ToV2`: text-only, json-only, empty-string `text`, null `query`, absent `query`, already-v2-shaped input (idempotent — no `text`/`json` keys present), non-object `query` value (left alone / error path consistent with `migrateV0ToV1`'s existing handling of malformed input).
- `migrateV0ToV2` composition: a v0 state containing both a flat `options_list_control` and a root `query.text` ends up with `by_field` *and* `expression` populated correctly in one upgrade call.
- `dashboardQueryToAPI` (or renamed equivalent): `expression` copies straight through, including `expression = ""`.
- Read-side mapping: `Expression` populates directly from the API string; no sniffing logic remains to test.
- Update existing tests that reference `text`/`json`: `models_dashboard_root_test.go` (remove the "neither/both set" cases — no longer reachable), `models_optional_root_blocks_test.go`, `create_test.go`, `pinned_panels_mapping_test.go`, and schema validation tests (`text` → `expression`; add a "missing expression" case since it's now `Required`).

**Acceptance**

- Rename `text = "..."` to `expression = "..."` across the ~20 `testdata/**/main.tf` fixtures that set a root `query` (mechanical).
- Delete `TestAccResourceDashboardRootQueryJSON` and its testdata (no `json` branch left to test).
- Rewrite `TestAccResourceDashboardQueryTransition` as an in-place `expression` update (e.g., one KQL expression to a different expression, and/or a language change), since there is no longer a text ↔ json transition to exercise.
- **New upgrade test (required).** Two-step test using the pattern in `internal/kibana/alertingrule/acc_test.go`: step 1 creates a dashboard with `query = { language = "kql", text = "..." }` using `ExternalProviders` pinned to the last released version (schema v1); step 2 switches to `ProtoV6ProviderFactories: acctest.Providers` with `query = { language = "kql", expression = "..." }` and asserts `query.expression` is populated with the prior `text` value, `query.text`/`query.json` are absent, and the plan is empty.

## Migration Plan

- Schema version 1 → 2. The `UpgradeState` map above makes the migration automatic on the next `terraform apply` or `terraform plan -refresh-only`; no manual `terraform state` commands are required.
- Practitioners on `query.text` or `query.json` MUST update `.tf` source to `query.expression` — this part is not automatable, since Terraform re-reads configuration from source, not from the upgraded state.
- Documented in the CHANGELOG under "Breaking changes" (see `proposal.md`), with a before/after config example and a note that `json` values carry over unchanged because Kibana was already treating them as a plain string.

## Risks / Trade-offs

- **Breaking change for any existing `query.text` / `query.json` user.** Accepted: the resource is Technical Preview, the rename eliminates a real read-back ambiguity, and state migration is automatic — only source edits are required of practitioners.
- **`json` users lose the "structured query" framing**, even though it was never real — Kibana always stored it as a plain string. The CHANGELOG entry makes this explicit so practitioners aren't surprised that their JSON document is now just the `expression` string.
- **v0 → v2 composition risk.** Composing two upgrader transforms in one call is more complex than a single transform; mitigated by a dedicated composition unit test (see Tests) in addition to each transform's own unit tests.

## Open questions

None. This proposal and design are derived directly from the issue reporter's own exploration and recommendation (issue #5081 comments); no implementation-research comment is present on the issue, and no open questions were raised in that exploration that aren't already resolved by the decisions above.
