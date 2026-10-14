## Context

See proposal.md for motivation and `specs/kibana-case-template/spec.md` for requirements. This design follows the recommendation in the issue's implementation-research comment (Approach A below), as refined by a maintainer follow-up comment that resolved the research comment's open questions against a live Kibana 9.5.5 stack; see those comments for the full evaluation of alternatives and the underlying evidence.

Current state and constraints:

- The generated Kibana client exposes `CreateCaseTemplateDefaultSpaceWithResponse`, `GetCaseTemplateDefaultSpaceWithResponse`, `UpdateCaseTemplateDefaultSpaceWithResponse`, and `DeleteCaseTemplateDefaultSpaceWithResponse` (`generated/kbapi/kibana.gen.go:236725-236750` interface, implementations around `186042-186127`). None of these take a `spaceId` parameter — like other Kibana default-space-only endpoints in this repo, named-space routing goes through `kibanautil.SpaceAwarePathRequestEditor(spaceID)` rewriting the request path.
- `CasesTemplateWriteRequest` (`kibana.gen.go:34845-34864`) carries `Definition string`, `Owner CasesOwner`, `Name *string`, `Description *string`, `IsEnabled *bool`, `Tags *CasesTemplateTags` (`CasesTemplateTags = []string`). The same type is the request body for both create and the full-replace update (`CreateCaseTemplateDefaultSpaceJSONRequestBody` / `UpdateCaseTemplateDefaultSpaceJSONRequestBody`).
- `CasesTemplateV2Response` (`kibana.gen.go:34781-34844`) carries `TemplateId`, `TemplateVersion int`, `LatestVersion int`, `IsLatest bool`, `Name`, `Owner`, `DefinitionString string` (raw YAML), `Definition interface{}` (parsed), `Description *string`, `Tags *CasesTemplateTags`, `IsEnabled *bool`, `IsDefault *bool`, `UsageCount *int`, `FieldCount *int`, `Author *string`, `LastUsedAt`/`DeletedAt *time.Time`.
- `GET`'s 200 body is `CasesTemplateV2Response` directly (`GetCaseTemplateDefaultSpaceResponse.JSON200`, `kibana.gen.go:245219-245225`) — no response union, so Read uses the repo's typed raw-handler pattern directly, no union unwrapping needed. `POST`/`PUT`'s 200 body IS a `oneOf` union (`CreateCaseTemplateDefaultSpace200JSONResponseBody` / update equivalent, `kibana.gen.go:81250-81270`), with a working generated `UnmarshalJSON` and `AsCasesTemplateV2Response()`/`AsCasesTemplateDryRunResponse()` accessors (`kibana.gen.go:168473-168591`). Since the provider always sends `dry_run=false` (the `DryRun *bool` param on `CreateCaseTemplateDefaultSpaceParams`/`UpdateCaseTemplateDefaultSpaceParams`, `kibana.gen.go:81245-81265`, is simply omitted), Create/Update call `.AsCasesTemplateV2Response()` on the 200 body — a plain typed accessor, not a raw-body bypass.
- `CasesOwner` is a string enum with constants `CasesOwnerCases = "cases"`, `CasesOwnerObservability = "observability"`, `CasesOwnerSecuritySolution = "securitySolution"` (`kibana.gen.go:482-484`), matching the field-definition resource's `owner`.
- The repo's CRUD-style Kibana resources (`internal/kibana/casefielddefinition`, `internal/kibana/agentbuilderworkflow`) are wired through `entitycore.NewKibanaResource`, with `Read: entitycore.SimpleKibanaRead`, `Delete: entitycore.SimpleKibanaDelete`, and `Create`/`Update` via `entitycore.SimpleKibanaCreate`/`SimpleKibanaUpdate` (generic over `Body any, R any`).
- `customtypes.NormalizedYamlType`/`NormalizedYamlValue` (`internal/utils/customtypes/normalized_yaml_value.go`) provide plan-time YAML syntax validation (`ValidateAttribute`) and a `StringSemanticEquals` hook (canonical-JSON comparison). The sibling `casefielddefinition` resource's `definition` attribute uses this type, and its design.md records that reformat-only edits still produced a non-empty plan in live testing against 9.6 — the type's semantic-equality hook alone did not suppress the diff in practice.
- Version gating follows `internal/kibana/streams/models.go`: a package-level `version.Must(version.NewVersion(...))` plus a `GetVersionRequirements` method satisfying `entitycore.WithVersionRequirements`, enforced automatically by `client.EnforceMinVersion`.

## Deviations from the research comment's recommendation

The issue's implementation-research comment recommended Approach A (an opaque, syntax-validated YAML string via `customtypes.NormalizedYamlType`) and this design adopts that recommendation unchanged. A maintainer follow-up comment, posted after live API probing against a 9.5.5 stack and confirmed with the maintainer, resolved several of the research comment's open questions with specific decisions that differ from the research comment's own proposed defaults. Per the issue-processing contract, these are recorded here as explicit, evidence-backed refinements rather than re-litigated:

- **Attribute name.** The research comment named the YAML attribute `definition`. The follow-up comment specifies **`definition_yaml`** instead, to leave room for a richer typed model later without a breaking rename. This design uses `definition_yaml`.
- **Minimum version.** The research comment's blocking open question proposed gating at 9.5.0 (matching the field-definition resource) pending confirmation. The follow-up comment confirms the actual minimum is **9.5.2**: the merged 9.5 backport of kibana#280144 (commit `c38767533442553879005bf14aca5f10dd576389`, kibana#283476) lands after the `v9.5.0`/`v9.5.1` tags and is included in `v9.5.2`. This design gates at 9.5.2.
- **`latest_version`.** The research comment's schema proposed both `template_version` and `latest_version` as computed attributes. The follow-up comment resolves this: expose **`template_version` only** — the resource always manages the latest version, so a redundant `latest_version` adds no information in the v1 scope (historical-version management is out of scope).
- **`tags` collection type.** The research comment flagged `List[String]` vs. `Set[String]` as an open question pending a round-trip probe. The follow-up comment's live probe (`["z", "a", "z"]` round-tripped unchanged) resolves it: **`List[String]`**, preserving order and duplicates.
- **Null-vs-empty mapping.** The research comment flagged Optional+Computed (matching the field-definition resource's preserve-on-omit pattern) as the default, pending verification. The follow-up comment's probe found the templates API does *not* preserve-on-omit the way field-definitions does: omitting `description`/`tags` on a later `PUT` *clears* them (they become absent from the next `GET`), rather than retaining the prior stored value. This design therefore keeps `description` and `tags` **plain Optional, not Computed** — matching config-to-state directly is sufficient and avoids `UseStateForUnknown` masking a real clear-on-omit.
- **Diff suppression is required behavior, not an accepted limitation.** The sibling field-definition resource treats non-suppressed reformat-only YAML diffs as an accepted v1 limitation (confirmed live: `NormalizedYamlType`'s semantic-equality hook did not suppress the diff in that resource). The follow-up comment requires the opposite outcome here, because every `PUT` against this API creates a new template version server-side — including a `PUT` that resubmits byte-different-but-semantically-identical YAML. An un-suppressed reformat-only diff would therefore silently inflate version history on every plan/apply cycle touching the file, which the maintainer judged unacceptable for this resource specifically. See "`definition_yaml` diff suppression" below for the concrete plan-time mechanism.

None of the above contradicts the research comment's core recommendation (Approach A); they are the blocking/non-blocking open questions that comment explicitly deferred, resolved by the follow-up with live evidence.

## Goals / Non-Goals

**Goals:**
- Implement the resource with the same envelope, package layout, and space-handling conventions as the repo's other CRUD-style Kibana resources so connection handling, timeouts, version gating, and import/identity behave identically.
- Guarantee that a whitespace-, comment-, or mapping-key-order-only edit to `definition_yaml` produces an empty plan and issues no `PUT`, since every `PUT` creates a new template version regardless of whether the YAML changed meaningfully.
- Rely on the templates API's own validation and conflict errors (400 on malformed template schema, 409 on duplicate name per owner) rather than reimplementing those checks client-side.

**Non-Goals:**
- No changes to `entitycore` or to `customtypes.NormalizedYamlType` itself.
- No client-side structural validation of `fields`/`$ref`/`display`/`validation` YAML keys.
- No data source, and no changes to the legacy `/api/cases/configure`-embedded templates.
- No exposure of historical template versions as a Terraform-managed concept.

## Decisions

### Use the `entitycore.KibanaResource` envelope (Approach A)

New package `internal/kibana/casetemplate` embedding `entitycore.KibanaResource[model]`, registered as `elasticstack_kibana_case_template`.

Rationale: matches the repo's CRUD-style Kibana resources and its sibling `casefielddefinition`; version gating via `entitycore.WithVersionRequirements`, space-ID handling via `kbschema.ResourceSpaceIDAttribute()`, and a composite `<space_id>/<template_id>` ID routed through `kibanautil.SpaceAwarePathRequestEditor(spaceID)` all come largely for free. The generic `SimpleKibanaCreate`/`SimpleKibanaUpdate` write helpers are generic over the response type, so unwrapping the create/update response union is a plain `.AsCasesTemplateV2Response()` call, not a bespoke write path.

### Schema mapping

| Attribute | Type | Behavior |
|---|---|---|
| `space_id` | string | Optional, Computed, defaults `"default"`, `RequiresReplace` (`kbschema.ResourceSpaceIDAttribute()`) |
| `owner` | string | Required, one of `cases`\|`securitySolution`\|`observability` (`CasesOwner`), `RequiresReplace` — no update operation changes the owning solution |
| `definition_yaml` | string, `customtypes.NormalizedYamlType` | Required, at most 30000 characters. Case defaults (`name`, `severity`, `category`, `tags`, `assignees`, `connector`, `settings`) and a `fields` array of inline field definitions and/or `$ref` entries into the field library. See "`definition_yaml` diff suppression" for the plan-time behavior required of this attribute beyond the base type. |
| `name` | string | Required, 1-50 chars, unique per `owner` (case-insensitive) — the server also accepts the YAML's own case-default `name:` key as a fallback when the request omits `name`, but this resource always sends an explicit `name` so that fallback is not relied upon. |
| `description` | string | Optional (not Computed), max 1000 chars — omitting it on write clears any previously stored value (see "Deviations" above); explicit `""` is preserved as an empty string and remains distinct from omitted/null. |
| `tags` | `List[String]` | Optional (not Computed), max 200 items, each at most 256 chars — order and duplicates are preserved by the API; omitting it on write clears any previously stored value. |
| `is_enabled` | bool | Optional, Computed, `UseStateForUnknown` — server default `true`; disabled templates are hidden from the case creation flow but remain manageable. |
| `template_id` | string | Computed — server-assigned, stable across versions |
| `template_version` | int64 | Computed — increments on every successful `PUT`, including a `PUT` with an unchanged `definition_yaml` (hence the diff-suppression requirement below) |
| `field_count` | int64 | Computed, nullable — absent on a template with no custom fields in some server responses; map absence to null rather than assuming `0` |
| `usage_count` | int64 | Computed, nullable — absent on a fresh template; map absence to null rather than assuming `0` |
| `is_default` | bool | Computed, nullable — absent on a fresh template; map absence to null rather than assuming `false` |
| `id` | string | Computed, `UseStateForUnknown` — `<space_id>/<template_id>` |

### `definition_yaml` diff suppression

Unlike the sibling field-definition resource (which accepts a non-suppressed reformat-only diff as a documented v1 limitation), this resource must not issue a `PUT` for a YAML edit that is semantically a no-op, because every `PUT` — including one that resubmits unchanged YAML — creates a new template version server-side (confirmed live by the maintainer follow-up comment). `customtypes.NormalizedYamlType`'s `StringSemanticEquals` hook performs a canonical-JSON comparison but its wiring into Create/Update/Read plan reconciliation is not relied upon as sufficient by itself (the sibling resource's live testing showed the type alone did not suppress a reformat-only diff for that resource's attribute). This resource therefore adds an explicit plan modifier on `definition_yaml` that:

- Parses both the prior state value and the planned config value as YAML.
- Compares them for structural/semantic equality (same mechanism as `StringSemanticEquals`'s canonical-JSON comparison, reused directly rather than re-implemented).
- When semantically equal, sets the planned value to the prior state value (`UseStateForUnknown`-style plan-value override), so Terraform's plan diff — and therefore the eventual `PUT` call — is suppressed entirely.
- When not semantically equal, leaves the planned value as the new config value, so a real change still flows through to `PUT` and increments `template_version` as expected.

An acceptance test with `plancheck.ExpectEmptyPlan()` (following `internal/kibana/security_detection_rule/acc_test.go`'s precedent for a pre-apply plan-check assertion) on a pure reformat/reorder/comment edit is required to prove this holds, per the maintainer follow-up comment's explicit instruction. A second test step with `plancheck.ExpectNonEmptyPlan()` on a semantically real change is required alongside it so the suppression logic is proven to be selective, not a blanket no-op.

### Minimum stack version: 9.5.2

Per the maintainer follow-up comment's confirmed evidence: kibana#280144 carries the `v9.5.2` release label, and its merged 9.5 backport (kibana#283476, commit `c38767533442553879005bf14aca5f10dd576389`) lands after the `v9.5.0`/`v9.5.1` tags and is included in `v9.5.2`. The version requirement is expressed as `version.Must(version.NewVersion("9.5.2"))`, following the `streams` package's pattern of using a concrete release tag (adjusted to a `-SNAPSHOT` suffix only if pre-release 9.5.2 build testing during implementation shows it is needed to admit RC/snapshot builds, mirroring the field-definition resource's `9.6.0-SNAPSHOT` precedent).

### Feature-flag prerequisite: `xpack.cases.templates.enabled`

The maintainer follow-up comment confirms the templates routes are additionally gated by the Kibana setting `xpack.cases.templates.enabled`, which defaults to `true` in the inspected 9.5 source but, when set to `false`, removes the routes entirely (404, indistinguishable from a too-old server at the HTTP level). A version check alone cannot guarantee route availability. The schema description documents this prerequisite explicitly, alongside the Technical Preview note, so a 404 against a nominally-supported server points the user at the setting rather than appearing as a provider bug.

### Technical Preview note

The schema description states **Technical Preview**, sourced from the OpenAPI spec's `x-state: Technical preview` on the create/update/delete template operations (GET/List carry no such marker) — the repo treats the API spec as authoritative for API-stability claims, following the field-definition resource's precedent (`internal/kibana/dashboard/schema.go` for the `MarkdownDescription` convention).

### Connector and settings live inside `definition_yaml`

The maintainer follow-up comment's live `dry_run=true` probes confirmed that `connector` and `settings` (including `settings.syncAlerts`) are accepted as keys inside the YAML `definition`, matching the OAS's `ParsedTemplateDefinitionSchema` prose rather than the separate UI docs' description of a UI-only "Configuration tab". The provider therefore does not expose `connector`/`settings` as additional request-level resource attributes; they are left to the user to declare inside `definition_yaml`, consistent with Approach A's opaque-YAML scope. The schema description notes this explicitly, since the UI docs could otherwise mislead a reader into expecting separate attributes.

### Null-vs-empty handling for `description` and `tags`

Resolved by the maintainer follow-up comment's live probes: on minimal create, omitted `description`/`tags` are absent from both the `POST` and subsequent `GET` response bodies. Explicit `description: ""` / `tags: []` are returned as empty values, distinct from absent. After populating both and then omitting them from a later `PUT`, the values are cleared (absent from the next `GET`) rather than preserved. The model therefore maps: API-absent → Terraform null; API-empty-string/empty-list → Terraform empty value; and sends exactly what the config contains on write (including sending nothing when the config is null, which the API then treats as a clear on update). No `UseStateForUnknown` is used for these two attributes, since there is no preserve-on-omit behavior to protect against an inconsistent-result error.

### Components

| Unit | Responsibility |
|---|---|
| `schema.go` | Schema per "Schema mapping" above, including the `definition_yaml` diff-suppression plan modifier. |
| `models.go` | `GetVersionRequirements` (9.5.2 minimum), `GetID`/`GetResourceID`/`GetSpaceID`, mapping between `kbapi.CasesTemplateV2Response` and the model (composite id, null-vs-empty handling for `description`/`tags`/`field_count`/`usage_count`/`is_default`), and `toAPICreateModel`/`toAPIUpdateModel`. |
| `create.go` / `update.go` | `entitycore.SimpleKibanaCreate`/`SimpleKibanaUpdate` wiring, calling `.AsCasesTemplateV2Response()` on the create/update response union. |
| `internal/clients/kibanaoapi/case_template.go` | Thin wrapper (`CreateCaseTemplate`, `GetCaseTemplate`, `UpdateCaseTemplate`, `DeleteCaseTemplate`) calling the generated `...WithResponse` methods with `kibanautil.SpaceAwarePathRequestEditor(spaceID)`, using the repo's typed response-handler helpers for create/update/read and `diagutil.HandleStatusResponse` for delete (HTTP 204 and 404 treated as success), modeled on `internal/clients/kibanaoapi/case_field_definition.go`. |

Import: the import ID is the composite `<space_id>/<template_id>`, passed through to `id` (`resource.ImportStatePassthroughID`); a subsequent read populates the rest.

### Data flow

Plan → `toAPICreateModel`/`toAPIUpdateModel` builds the write body → `SimpleKibanaCreate`/`SimpleKibanaUpdate` calls the wrapper for `req.SpaceID` (and `req.WriteID` for update) → after a successful write, the envelope performs a read and maps its response into the model → `id` is set from `<space_id>/<template_id>`. The mutation response is not used as final state; this follows the resource convention of deriving final state from a read after write. Destroy calls `DeleteCaseTemplate` and treats HTTP 204 and 404 as success (already-deleted); a destroy soft-deletes **all** versions of the template (per OAS prose), and existing cases created from it keep their pinned template reference.

### Failure handling

- Read: HTTP 404 removes the resource from state; other errors become diagnostics.
- Delete: HTTP 404 counts as success; other error statuses surface as destroy-time diagnostics, not a silent success.
- Create/Update: 400 (malformed YAML schema, e.g. invalid `control` value, or other request-shape errors) and 409 (duplicate name within the same owner, case-insensitive) surface unchanged as apply-time error diagnostics — no client-side pre-check beyond `definition_yaml`'s plan-time YAML *syntax* validation (which `customtypes.NormalizedYamlType` already provides).
- Version too old, or the feature flag disabled server-side: the former is caught by the envelope's version requirement before any template API call; the latter surfaces as a 404 from the API (see "Feature-flag prerequisite" above) and is documented, not specially detected.

### Testing approach

- Unit tests for the model/API mapping (`description`/`tags`/`field_count`/`usage_count`/`is_default` present vs. absent, composite-id construction, version requirement) and the `definition_yaml` diff-suppression plan modifier (semantically-equal vs. genuinely-different YAML).
- Acceptance tests in the package (see tasks.md) covering: minimal create, create with all optional attributes, omission vs. explicit-empty for `description`/`tags`, `tags` ordering/duplicates, in-place update of YAML fields, a dedicated reformat-only-edit step asserting `plancheck.ExpectEmptyPlan()` and no new `template_version`, update clearing a previously-set `description`/`tags`, import by composite id, a named (non-default) space, normal destroy, out-of-band deletion followed by refresh/plan, the 400/409 negative cases, an unsupported-version diagnostic on a below-minimum server with no templates API call, and a 9.6+-gated template referencing a field via `$ref`.

## Open questions

Copied verbatim from the issue's implementation-research comment. Most are resolved by the maintainer follow-up comment, as noted inline; the decisions above reflect those resolutions.

- **(Blocking)** Is the "new in Stack 9.5" / kibana#280144 minimum-version claim accurate? Neither the OpenAPI spec nor the retrieved Elastic docs carry a version marker for these paths — only the `x-state: Technical preview` flag on POST/PUT/DELETE was corroborated (GET/List have no marker). Interim policy: gate the whole resource (all CRUD, including read) at the same minimum version the sibling field-definition resource uses (9.5.0), via `entitycore.SingleVersionRequirement`, with an acceptance test that (a) exercises full CRUD at/above the candidate minimum and (b) is skipped via `versionutils.SkipIfUnsupported` below it — do not hard-code 9.5.0 into published documentation until confirmed against the kibana#280144 PR/changelog.
  - *Resolved* by the follow-up comment: the correct minimum is **9.5.2**, not 9.5.0 (see "Deviations" and "Minimum stack version" above).
- Does the YAML `definition` actually accept `connector` and `settings` keys as the OAS prose claims, given the OAS's own example and the separate UI docs only show them configured outside the YAML? Confirm with a live `dry_run=true` create call before finalizing which keys the provider documents as YAML-only vs. request-level.
  - *Resolved* by the follow-up comment: yes, confirmed live (see "Connector and settings live inside `definition_yaml`" above).
- Does formatting-only-edit diff suppression hold for `customtypes.NormalizedYamlType` in practice? This is unverified in this repo for any resource today (the nearest precedent test doesn't actually assert plan emptiness either way) and should be settled by a dedicated acceptance-test step for this resource rather than assumed.
  - *Addressed* by the follow-up comment's explicit requirement (not by the base type alone): this resource adds its own plan-modifier-level suppression on top of `NormalizedYamlType`, proven by a dedicated acceptance test (see "`definition_yaml` diff suppression" above). This is a stricter outcome than the sibling field-definition resource, which accepted non-suppression as a v1 limitation.
- Does the API preserve `tags` ordering and allow duplicate entries, or normalize/dedupe them server-side? Needed to finalize `List[String]` vs. `Set[String]` and to decide Read's reconciliation logic.
  - *Resolved* by the follow-up comment: order and duplicates are preserved; `tags` is `List[String]` (see "Deviations" above).
- Does an omitted `description`/`tags` on write read back as null, empty, or absent from the response JSON? Needed to get Read's null-vs-empty state mapping right and avoid spurious permanent diffs.
  - *Resolved* by the follow-up comment: absent on create; cleared (becoming absent) when omitted on a later update of a previously-populated template (see "Null-vs-empty handling" above).
- Should `template_version`/`latest_version` be exposed as computed attributes at all, given Update always creates a new version and old versions stay retrievable via the `version` query param — or is tracking just the latest version sufficient for v1?
  - *Resolved* by the follow-up comment: expose `template_version` only; omit `latest_version` (see "Deviations" above).

The only item not fully closed by the follow-up comment is a residual implementation-time check: live-verifying the exact 9.5.2 minimum (the follow-up's evidence is from PR/tag metadata, not a live 9.5.2 probe — "Full CRUD was exercised successfully on the local 9.5.5 stack; 9.5.2 itself was not live-tested") and confirming the feature-flag-disabled 404 is indistinguishable from a too-old-server 404 in practice. Both should be recorded in an Implementation findings section once probed during implementation.

## Risks / Trade-offs

- [Gating at 9.5.2 based on tag/PR metadata rather than a live 9.5.2 probe could be off by a patch version] → Accepted for the proposal stage; implementation should live-verify against a 9.5.2 (or 9.5.0/9.5.1, to confirm the routes are genuinely absent there) cluster if available, and adjust the version string before merging if evidence contradicts it.
- [A `definition_yaml`-specific diff-suppression plan modifier adds custom logic beyond the base `NormalizedYamlType`, unlike the sibling resource which accepted non-suppression] → Accepted: the version-inflation consequence of an un-suppressed diff is more costly here (every `PUT` mutates server-side version history) than in the field-definition resource, so the extra complexity is justified; covered by a dedicated unit test plus the `ExpectEmptyPlan`/`ExpectNonEmptyPlan` acceptance pair.
- [`description`/`tags` clear-on-omit (rather than preserve-on-omit) could surprise a user coming from the field-definition resource's preserve-on-omit behavior] → Accepted: documented explicitly in the schema description for both attributes, since the two sibling resources now have genuinely different, evidence-backed server behaviors.
- [A 404 from the feature flag being disabled is indistinguishable from a too-old-server 404] → Accepted: both cases are documented together in the schema description; no provider-side probe is added to disambiguate them.

## Migration Plan

Not applicable: this is a new resource with no impact on existing configurations or state. Rollback is removing the resource from configuration; destroy issues a normal `DELETE`, which soft-deletes all versions of the template (existing cases keep their pinned template reference and are unaffected).
