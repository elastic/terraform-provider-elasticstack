## Context

See proposal.md for motivation and `specs/kibana-case-field-definition/spec.md` for requirements. This design follows the recommendation in the issue's implementation-research comment (Approach A below); see that comment for the full evaluation of alternatives.

Current state and constraints:

- The generated Kibana client exposes `CreateCaseFieldDefinitionDefaultSpaceWithResponse`, `GetCaseFieldDefinitionDefaultSpaceWithResponse`, `UpdateCaseFieldDefinitionDefaultSpaceWithResponse`, and `DeleteCaseFieldDefinitionDefaultSpaceWithResponse` (`generated/kbapi/kibana.gen.go:235068-235079,267236-267271`). None of these take a `spaceId` parameter — like other Kibana default-space-only endpoints in this repo, named-space routing goes through `kibanautil.SpaceAwarePathRequestEditor(spaceID)` rewriting the request path, which needs confirming against a running cluster (see Open questions).
- `CasesFieldDefinitionWriteRequest` (create and update body before PUT-only exceptions) and `CasesFieldDefinitionPutRequest` (`generated/kbapi/kibana.gen.go:33919-33985`) both carry `Definition string`, `Owner string`, `Name *string`, `Description *string`, `IsGlobal *bool`. `CasesFieldDefinitionResponse` (`generated/kbapi/kibana.gen.go:33941-33964`) carries the same fields plus `FieldDefinitionId string` and `DisplayOrder *int`. The 200 response bodies for create/update (`CreateCaseFieldDefinitionDefaultSpace200JSONResponseBody` / update equivalent) are response unions (`union json.RawMessage`), but `CasesFieldDefinitionResponse`'s own JSON tags decode directly from the raw response body, so the provider does not need the generated union's `.AsCasesFieldDefinitionResponse()` accessor: `kibanaoapi.HandleMutateRawResponse[kbapi.CasesFieldDefinitionResponse]` (`internal/clients/kibanaoapi/responses.go:54-65`) unmarshals the raw body directly, following the same raw-handler pattern used elsewhere in the package.
- `CreateCaseFieldDefinitionDefaultSpaceParams` and `UpdateCaseFieldDefinitionDefaultSpaceParams` each carry only `DryRun *bool` (`generated/kbapi/kibana.gen.go:80629-80649`).
- The repo's CRUD-style Kibana resources (`internal/kibana/connectors/resource.go:50-68`, `internal/kibana/agentbuilderworkflow`) are wired through `entitycore.NewKibanaResource`, with `Read: entitycore.SimpleKibanaRead`, `Delete: entitycore.SimpleKibanaDelete`, and `Create`/`Update` via `entitycore.SimpleKibanaCreate`/`SimpleKibanaUpdate` (`internal/entitycore/kibana_write_helpers.go:89-93,120-124`, generic over `Body any, R any`). The one non-CRUD counter-example, `internal/kibana/import_saved_objects`, is an action-style resource with deliberate no-op `Read`/`Delete` and is not a relevant precedent here.
- `customtypes.NormalizedYamlType`/`NormalizedYamlValue` (`internal/utils/customtypes/normalized_yaml_value.go:64-116`) already provide YAML-aware `StringSemanticEquals`, used today by `agentbuilderworkflow`'s `configuration_yaml` (`internal/kibana/agentbuilderworkflow/schema.go:51-55`).
- Version gating follows `internal/kibana/streams/models.go:183-197`: a package-level `version.Must(version.NewVersion(...))` plus a `GetVersionRequirements` method satisfying `entitycore.WithVersionRequirements`, enforced automatically by `client.EnforceMinVersion` (`internal/entitycore/version_requirements.go:143-156`).

## Goals / Non-Goals

**Goals:**
- Implement the resource with the same envelope, package layout, and space-handling conventions as the repo's other CRUD-style Kibana resources so connection handling, timeouts, version gating, and import/identity behave identically.
- Rely on the field library API's own identity-mismatch errors (400 for `name` vs. YAML mismatch, 409 `field_identity_immutable` for YAML `type` changes) rather than reimplementing those checks client-side.

**Non-Goals:**
- No changes to `entitycore`.
- No `dry_run` pre-flight validation exposed on the resource (see Decisions).
- No data source, and no case-templates resource (out of scope; see proposal.md).
- No custom plan modifier parsing `definition`'s YAML to pre-empt identity-mismatch errors.

## Decisions

### Use the `entitycore.KibanaResource` envelope (Approach A)

New package `internal/kibana/casefielddefinition` embedding `entitycore.KibanaResource[model]`, registered as `elasticstack_kibana_case_field_definition`.

Rationale: matches the repo's CRUD-style Kibana resources; version gating via `entitycore.WithVersionRequirements`, space-ID handling via `kbschema.ResourceSpaceIDAttribute()`, and a composite `<space_id>/<field_definition_id>` ID routed through `kibanautil.SpaceAwarePathRequestEditor(spaceID)` all come largely for free. The generic `SimpleKibanaCreate`/`SimpleKibanaUpdate` write helpers are generic over the response type, so no bespoke write path or hand-rolled CRUD (the manual `import_saved_objects`-style alternative) is needed purely to unwrap the response union — see the issue's research comment for the full alternative-approaches analysis.

### Schema mapping

| Attribute | Type | Behavior |
|---|---|---|
| `space_id` | string | Optional, Computed, defaults `"default"`, `RequiresReplace` (`kbschema.ResourceSpaceIDAttribute()`) |
| `owner` | string | Required, `RequiresReplace` — no update operation changes the owning solution |
| `name` | string | Optional, Computed, `RequiresReplace`, `UseStateForUnknown` — the server extracts it from the YAML and returns it on the response when omitted from config |
| `definition` | string, `customtypes.NormalizedYamlType` | Required. Its `StringSemanticEquals` hook is designed to suppress diffs for semantically-identical YAML (whitespace/key-order only); treat a non-empty plan on reformat-only edits as an acceptable v1 limitation if the acceptance test (task group 2) shows the suppression does not hold in practice, rather than engineering around it immediately. |
| `description` | string | Optional, Computed (not Optional-only) — the provider does not yet know whether Kibana's `PUT` preserves or clears a previously-stored value when `description` is omitted (see Open questions); Computed plus a read-after-write refresh avoids a perpetual diff regardless of which behavior the live-cluster probe confirms. Revisit if the probe shows Optional-only is more correct. |
| `is_global` | bool | Optional, Computed — no documented server default on create; Computed avoids assuming either an absent-on-omit or a defaulted response (see Open questions). |
| `field_definition_id` | string | Computed — server-assigned UUID |
| `display_order` | int64 | Computed — server-assigned, `>= 0` |
| `id` | string | Computed, `UseStateForUnknown` — `<space_id>/<field_definition_id>` |

### `dry_run` is not a resource attribute in v1

Per the OpenAPI spec (`oas.yaml:22125-22135,22331-22342`), `dry_run=true` runs server-side authorization, full body validation, and name-uniqueness/identity checks the provider cannot replicate client-side — it is not a local-plan equivalent. Omitting it is a scope/request-cost tradeoff: these checks remain apply-time-only failures (surfaced as 400/409 from `Create`/`Update`) in this version. Revisit if real-world apply failures on those checks turn out to be common enough to justify a `ValidateConfig`-time pre-flight call.

### Technical Preview note

The schema description states **Technical Preview**, sourced from the OpenAPI spec's `x-state: Technical preview` on all five field-definition operations (`oas.yaml:22113,22203,22263,22319`, and the PUT op) — the repo treats the API spec as authoritative for API-stability claims. The Elastic UI docs badge the *feature* itself as GA in Stack 9.5+ (UI/doc availability, not the public API's own stability marker); both facts are stated in the schema description since they could otherwise look contradictory to a reader.

### Identity-mismatch errors are not special-cased

Because `name` is `RequiresReplace`, a user can only reach the in-place `Update` path (not a replace) by editing `definition`'s internal YAML while leaving the top-level `name` attribute untouched. On `PUT`, the provider sends the *unchanged* `name` from state alongside the edited YAML:

- If the edited YAML's `name` key no longer matches the (unchanged) top-level `name` being sent, the server returns **400**.
- If the YAML's `type` key changes (independent of `name`), the server returns **409** with `attributes.code = "field_identity_immutable"`.

Both surface through standard error-diagnostic reporting with no special-case logic. The schema description documents this split so users understand why an in-place `definition`-only edit to `name` vs. `type` fails with different, specific errors.

### Components

| Unit | Responsibility |
|---|---|
| `schema.go` | Schema per "Schema mapping" above. |
| `models.go` | `GetVersionRequirements` (9.5.0 minimum), `GetID`/`GetResourceID`/`GetSpaceID`, mapping between `kbapi.CasesFieldDefinitionResponse` and the model (composite id, `description`/`is_global` nullability), and `toAPICreateModel`/`toAPIUpdateModel`. |
| `create.go` / `update.go` | `entitycore.SimpleKibanaCreate`/`SimpleKibanaUpdate` wiring, following the `agentbuilderworkflow` shape. |
| `internal/clients/kibanaoapi/case_field_definition.go` | Thin wrapper (`CreateCaseFieldDefinition`, `GetCaseFieldDefinition`, `UpdateCaseFieldDefinition`, `DeleteCaseFieldDefinition`) calling the generated `...WithResponse` methods with `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, using `HandleMutateRawResponse[kbapi.CasesFieldDefinitionResponse]` / `HandleGetRawResponse[kbapi.CasesFieldDefinitionResponse]` for create/update/read and `diagutil.HandleStatusResponse` for delete (404 treated as success), modeled on `internal/clients/kibanaoapi/agentbuilder_skill.go`. |

Import: the import ID is the composite `<space_id>/<field_definition_id>`, passed through to `id` (`resource.ImportStatePassthroughID`); a subsequent read populates the rest.

### Data flow

Plan → `toAPICreateModel`/`toAPIUpdateModel` builds the write body → `SimpleKibanaCreate`/`SimpleKibanaUpdate` calls the wrapper for `req.SpaceID` (and `req.WriteID` for update) → the response is mapped directly into the model (no separate read-after-write call is required, since `CasesFieldDefinitionResponse` is already the full authoritative shape) → `id` is set from `<space_id>/<field_definition_id>`. Destroy calls `DeleteCaseFieldDefinition` and treats HTTP 404 as success (already-deleted).

### Failure handling

- Read: HTTP 404 removes the resource from state (`HandleGetRawResponse` returns `(nil, nil)`, treated as not-found by `entitycore.SimpleKibanaRead`); other errors become diagnostics.
- Delete: HTTP 404 counts as success; a 409 while the field is referenced by an active template/custom field surfaces as a destroy-time diagnostic, not a silent success.
- Create/Update: 400 (name/YAML `name` mismatch) and 409 (`field_identity_immutable`, or demotion of a linked global field) surface unchanged as apply-time error diagnostics — no client-side pre-check (see "Identity-mismatch errors are not special-cased").
- Version too old: the envelope's version requirement fails the operation before any field-definition API call, following the resource-level (nil `AttributePath`) pattern in `internal/entitycore/kibana_resource_envelope_test.go:2720-2799`.

### Testing approach

- Unit tests for the model/API mapping (`description`/`is_global` present vs. absent, composite-id construction, version requirement) and the YAML custom-type validation path, following existing `agentbuilderworkflow` unit test patterns.
- Acceptance tests in the package (see tasks.md) covering: minimal create, create with all optional fields, omission of `description`/`is_global`, update of `description` (set then remove, probing the actual clear-vs-preserve behavior), update of `is_global`, in-place update of non-identity YAML fields, reformat-only YAML update (semantic-equality no-op), replacement on a consistent `name` change, the 400 name-mismatch negative case, the 409 `field_identity_immutable` negative case, import by composite id, a named (non-default) space, normal destroy, external-deletion-then-refresh (kept as a distinct step from normal destroy), a 409 destroy-blocked negative case, and the resource-level version-gate diagnostic. Tests should resolve the open questions below as part of the acceptance-test work, and the schema/write mapping should be adjusted afterward if the probed behavior does not match the Computed-field assumption above.

## Open questions

- **Server-side omission semantics for `description` and `is_global` are unverified.** Does omitting `description` on a `PUT` clear a previously-stored value, or preserve it (partial-update semantics)? Does omitting `is_global` on `POST` leave it absent in the response, or does the server apply an undocumented default? Neither the OpenAPI spec nor the generated client settles this — confirm against a running 9.5+ Kibana before finalizing the `description`/`is_global` read/update mapping (acceptance steps 3-4 in the acceptance outline are designed to surface the answer).
- Can the 409 "cannot demote a global field linked to an active custom field" conflict (toggling `is_global` from `true` to `false`) be translated into a clearer provider error message, or should it be left as the raw API error for the first version?
- Named (non-default) spaces: the generated client only exposes default-space paths (no `spaceId` parameter on the generated methods, `generated/kbapi/kibana.gen.go:203111` et al.); confirm `kibanautil.SpaceAwarePathRequestEditor` correctly rewrites these specific paths against a running cluster — not yet verified live.
- Is a companion "case template" resource/data source (`/api/cases/templates`) planned that will reference this field by `name`/`field_definition_id`? Out of scope here, but affects whether `name`/`field_definition_id` need more prominent documentation as stable references.

## Risks / Trade-offs

- [A mistyped `owner` or a `name`/`type` identity conflict only surfaces as an apply-time 400/409, not a plan-time validation error] → Accepted: documented in the schema description (see "Identity-mismatch errors are not special-cased"); no client-side YAML parsing is added to pre-empt it (see proposal non-goals).
- [`description`/`is_global` Computed-field choice may not match the server's actual partial-update semantics once verified live] → Tracked as a blocking open question; the acceptance-test task group is scoped to resolve it and adjust the schema/mapping before this change is considered complete.
- [The envelope enforces the version requirement on every lifecycle operation, including read and delete] → Accepted: a state entry against a pre-9.5 server is not expected in practice, since the resource could never have been created there.
- [Reformat-only YAML edits may not produce an empty plan in practice, despite `NormalizedYamlType`'s semantic-equality hook] → Accepted as a documented v1 limitation if the acceptance test shows this, rather than a release blocker (see Schema mapping note on `definition`).

## Migration Plan

Not applicable: this is a new resource with no impact on existing configurations or state. Rollback is removing the resource from configuration; destroy issues a normal `DELETE` against the field definition (no reset-to-empty semantics are needed, unlike a singleton-settings resource).
