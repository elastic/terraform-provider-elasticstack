# Design

## Context

See `proposal.md` for motivation and `specs/kibana-advanced-settings/spec.md` for requirements. This
document covers how the provider will deliver them.

- **How Kibana exposes advanced settings.** Settings are stored per space in the `config` saved
  object and served by Kibana's UI settings routes (`src/core/packages/ui-settings/server-internal/src/routes/`:
  `get.ts`, `set.ts`, `set_many.ts`, `delete.ts`). `GET /api/kibana/settings` returns only
  user-set values, shaped `{ "settings": { "<key>": { "userValue": <any>, "isOverridden"?: bool } } }`.
  `POST /api/kibana/settings` takes `{ "changes": { "<key>": <any or null> } }` and returns the same
  shape; a `null` value removes the user value, restoring the Kibana default. Registered settings
  are validated on write (HTTP 400 on a bad value or on a setting pinned by Kibana server
  configuration); unregistered keys and `null` skip validation. Writes require the
  `manage_advanced_settings` privilege (a live 403 names it). All routes are space-aware
  (`/s/{space}/api/kibana/settings`).
- **Verified live (Kibana 9.5.6, stateful).** A `null` change removes the user value and the key
  disappears from `GET`; a user value equal to its Kibana default is retained; a setting pinned by
  Kibana server configuration is rejected with HTTP 400 on write (including a `null` reset) and
  appears in `GET` with `isOverridden: true`; a write for an unregistered key succeeds, and so does
  any value for a setting that has no schema; Kibana's schema for a registered boolean or number
  setting also accepts the strings `"true"`/`"false"` and numeric strings (storing them as strings),
  so only values such as `"maybe"` are rejected; Kibana returns integers beyond the JavaScript safe
  range rounded. The spaces API answers 200 for a space the credentials can read and 403, not 404,
  for a space they have no privilege on, whether or not it exists (credentials holding only
  `advancedSettings: all` in one space can read that space and no other).
- **Internal routes.** The routes declare no public access level in any Kibana release checked, so
  they are internal: Kibana 9.x (where `server.restrictInternalApis` defaults to true) serves them
  only when the request carries the `x-elastic-internal-origin: Kibana` header, they have no public
  contract, and Kibana's own OpenAPI document does not describe them usefully. Elastic could change
  or remove them.
- **Spaces.** Kibana does not report a missing or deleted space on these routes: reads and writes
  against a space that never existed succeed (a write creates an orphaned `config` saved object),
  and a deleted space keeps serving the settings previously written to it. Whether a space exists
  must therefore come from the spaces API.
- **Generated client gap.** These routes are absent from the Kibana OpenAPI export, both at the
  commit pinned in `generated/kbapi/Makefile` and on Kibana `main`, so `generated/kbapi` has no
  advanced-settings operations.
- **Provider precedent.** `internal/kibana/defaultdataview` is a space-scoped singleton built on
  `entitycore.NewKibanaResource` (id equals `space_id`, no import) and is the structural template.
  The envelope supplies the `kibana_connection` block, `timeouts`, scoped-client resolution,
  read-after-write, and the prior state on update. `generated/kbapi/dashboard-paths.json` is the
  precedent for a paths-only overlay for routes the upstream export does not describe.
- **Adjacent resource.** `elasticstack_kibana_default_data_view` writes the `defaultIndex` setting.

## Goals / Non-Goals

**Goals:**

- Reuse the Kibana resource envelope and the `defaultdataview` shape instead of introducing new
  resource infrastructure.
- Perform every operation with one Kibana settings request, so a set-and-reset change is a single call
  (any atomic application is a Kibana property; the design's correctness does not rely on it, see D7).
- Keep the logic that decides what to send and what to record in pure functions that are unit
  testable without Kibana.
- Make the generated-client change purely additive.
- Leave the resource's identity and schema free of anything that would block adding global
  advanced settings later.

**Non-Goals:**

- No scope abstraction, interface, or parameter for global settings in this change. Extensibility
  comes from not reserving conflicting names and from the overlay and wrapper being easy to extend.
- No registry of setting names or types, in the provider or in the schema.
- No use of Kibana's per-key `DELETE` or per-key `POST` routes, and no use of the global settings
  routes.
- No cleanup of unrelated stale text in `generated/kbapi/README.md` (for example its mention of a
  `transformFilterPaths` allow list that does not exist). Only a short section for the new overlay
  is added.

## Decisions

### D1. Reach Kibana through a paths-only overlay in `generated/kbapi`

Add `generated/kbapi/advanced-settings-paths.json`, embedded and merged by a new
`injectAdvancedSettingsAPIPaths` transform registered next to `injectDashboardAPIPaths` (which
replaces whole path entries and runs first in the transformer list). The overlay defines only `GET`
and `POST` on `/api/kibana/settings`, with inline schemas and no new component schemas, as the
`generated/kbapi/README.md` requires for overlays. Response entries are modeled as
`{ userValue: <any>, isOverridden?: bool }` and the request body as `{ changes: { <key>: <any, null allowed> } }`.

The paths are left unprefixed and are not added to the `spaceIdPaths` list; the space is applied at
call time with `kibanautil.SpaceAwarePathRequestEditor`, as dashboards do. This follows the
repository standard of using the generated client for new Kibana API interactions, keeps the
overlay removable once upstream describes these routes, and keeps the change to the roughly 13 MB
generated file purely additive (verified: regenerating at the pinned commit adds 285 lines and
removes none). The overlay's operation IDs are `get-kibana-settings` and `post-kibana-settings`.

Because the routes are internal (see Context), every call from the wrapper carries the
`x-elastic-internal-origin: Kibana` header through a request editor; the header is harmless where it
is not enforced. The overlay and wrapper are the only places that know about these routes, which
confines any later move to a public route, or any removal by Elastic, to those two units.

The wrapper does not consume the generated typed response. It decodes the raw body itself so that
each value keeps its exact JSON text (the overlay's `userValue` is untyped and would lose exact
numbers), and it reports a 404 as a nil result for reads and as a `notFound` flag for writes. The
overlay's response schemas therefore document the shape without being consumed.

### D2. One bulk request per operation; `null` is the reset

Create, update and destroy each send a single bulk `POST` whose `changes` carry values for keys to
set and `null` for keys to reset (a `null` change was confirmed live to remove the user value). The
per-key `DELETE` route is not used: it would add a second request path and give up the single-call
property.

### D3. Resource shape follows `defaultdataview`

- Package `internal/kibana/advancedsettings`, type name `advanced_settings`
  (`elasticstack_kibana_advanced_settings`), registered in `provider/plugin_framework.go`.
- Built with `entitycore.NewKibanaResource`. The model embeds the envelope's timeouts, Kibana
  connection and space ID field helpers; `GetID()` returns the `id` and `GetResourceID()` returns
  the space ID, so identity resolution falls back to the space ID (it is not a composite ID).
- No import, no schema version, no state upgrader.
- `timeouts` and `kibana_connection` are injected by the envelope and are not declared in the
  schema file.

Schema contract (requirements are in the spec):

```hcl
resource "elasticstack_kibana_advanced_settings" "example" {
  id       = <computed, string>           # equals space_id; UseStateForUnknown
  space_id = <optional, computed, string> # default "default"; RequiresReplace

  string_settings = <optional, map of string>          # setting key => plain string value
  json_settings   = <optional, map of normalized JSON> # setting key => JSON value (boolean, number, string, array, object); never null

  # injected by the envelope: kibana_connection block, timeouts attribute
}
```

### D4. Two typed maps for values

Kibana setting values are JSON of any type and the API gives no way to learn a setting's type, so
the provider cannot infer it from a single map of strings. `string_settings` carries the common
string case without escaping; `json_settings` carries every other type as JSON text. The element
type of `json_settings` is `jsontypes.Normalized` so that formatting and key-order differences do
not produce diffs (the module is already a dependency). Terraform converts booleans and numbers to
strings when assigned to a `map(string)`, which is why the attributes are named for their content
and why a rejected `string_settings` write carries a hint (D7).

The library's semantic equality for `jsontypes.Normalized` ignores formatting and key order but
compares number text exactly (`1` and `1.0` differ), and Kibana returns `1.50` as `1.5`. The
provider therefore adds its own value comparison (`json_equality.go`): numbers are compared exactly
by value (`1.50` equals `1.5`, `1e2` equals `100`, while integers that really differ, such as
Kibana's rounded values beyond 2^53, still differ), and when Kibana's value is equal the recorded
text is kept instead of Kibana's. The same comparison decides which keys an update sends. The
framework applies semantic equality to the results of create, read and update, not when comparing a
changed configuration to state, so reformatting a value in the configuration can plan an in-place
update that sends nothing to Kibana. Everything else about typing is in the spec.

### D5. Component structure

| Unit | Location | Responsibility |
|---|---|---|
| Paths overlay | `generated/kbapi/advanced-settings-paths.json`, a new `injectAdvancedSettingsAPIPaths` transform, regenerated `kibana.gen.go` | Describe `GET` and `POST /api/kibana/settings`; produce the typed client. |
| Client wrapper | `internal/clients/kibanaoapi/advanced_settings.go` | `GetAdvancedSettings(ctx, client, spaceID)` returns a `map[string]json.RawMessage` of user-set values keyed by setting name (decoded from the raw body so each value keeps its exact JSON text), and nil on 404, so a non-nil empty map means "no user-set settings"; the `isOverridden` flag is ignored. `ApplyAdvancedSettings(ctx, client, spaceID, changes)` sends one bulk `POST` where a nil value means reset, keeps exact number text, reports a 404 through a `notFound` result, and turns Kibana errors into diagnostics carrying Kibana's message. Both calls send the internal-origin header. The only code that sees HTTP or generated types. |
| Space check | `internal/kibana/advancedsettings`, using the existing `GetSpace` wrapper in `internal/clients/kibanaoapi/spaces.go` | For a space other than `default`, reports whether the space exists before read, create, update and destroy. Kibana's settings routes cannot report this themselves (see Context). |
| Resource wiring | `internal/kibana/advancedsettings/resource.go` | `NewKibanaResource` with name `advanced_settings`, plus `NewResource`. |
| Schema | `schema.go` | `id` via `entitycore.IDAttribute()`, `space_id` via `kbschema.ResourceSpaceIDAttributeRequiresReplaceOnly()`, `string_settings` as a map of strings, `json_settings` as a map of `jsontypes.Normalized`. |
| Model | `models.go` | Embeds the envelope field helpers; `GetResourceID()` returns the space ID. |
| Validation | `validate.go` | Per-map validators for key length 1-1024 and no null entries; a `ValidateConfig` on the resource wrapper for the cross-map rule (a key in only one map) and for rejecting the JSON literal `null`. |
| Pure conversions | `convert.go` | `buildChanges(plan, prior)`: new or changed declared keys plus nil for keys managed before and no longer declared. `refreshState(model, userProvided)`: rebuilds both maps from Kibana for managed keys only. The main unit-test seam. |
| Lifecycle | `create.go`, `update.go`, `read.go`, `delete.go` | Create and update share one write function; delete resets all managed keys, or sends nothing when there are none. |
| Value comparison | `json_equality.go` | `jsonValuesEqual` compares decoded JSON recursively with numbers compared exactly by value; `refreshState` keeps the recorded text when Kibana's value is equal, and `buildChanges` omits keys whose value is unchanged. |
| Error guidance | `error_guidance.go`, used by the write path and Read | `withJSONSettingsHint` appends the hint that non-string values belong in `json_settings` to each wrapper-reported error diagnostic of a write whose sent changes included `string_settings` entries (`sendsStringSetting`). `missingSettingsRouteDiagnostic` builds the error naming the route for a settings-route 404 in read, create, update and destroy, and never carries the hint. |
| Acceptance helper | `internal/acctest/kibana_advanced_settings.go` | `AdvancedSettingsKibanaClient` (pre-check, stateful and minimum-version gate, client) and `NewAdvancedSettingsTestSpace` (random space removed in cleanup, tolerating a missing space), shared by the wrapper and resource acceptance tests. |
| Docs and descriptions | `descriptions.go` with `descriptions/*.md`, `examples/resources/elasticstack_kibana_advanced_settings/resource.tf`, generated `docs/resources/kibana_advanced_settings.md`, a short section in `generated/kbapi/README.md` | Embedded long descriptions, a self-contained example, generated docs. |

The envelope calls the lifecycle functions, which call the pure conversions and then the wrapper.

### D6. Data flow

State lives in Kibana (the truth) and in Terraform state (managed keys with their last-read
values). The managed key set is always the keys of the two maps; there is no private state.

| Operation | Flow |
|---|---|
| Create | First the space check: a space other than `default` that does not exist fails the create before anything is written. Plan maps become `changes`: `string_settings` values as JSON strings, `json_settings` values parsed from their JSON text with numbers kept exact. One bulk `POST`, skipped if there are no keys. The write sets `id` and `space_id`. The envelope's read-after-write then runs Read against the planned keys and that result becomes state. |
| Update | Same space check first. Diff plan against the prior state the envelope supplies. Send only keys that are new or changed (semantic JSON equality), plus `null` for every prior key no longer declared in either map. A key moving between maps counts as changed. One bulk `POST`, skipped if there is nothing to send. Read-after-write follows. |
| Read | Space check first: a missing space removes the resource from state without reading settings. Then `GET` the space's user-set settings (the "user-provided" map of key to JSON value that `refreshState` receives); a 404 from the route itself is an error and is never read as "no settings", so managed keys are not dropped by mistake. For each managed key: absent in Kibana means dropped from state. A `string_settings` key records the string, or the JSON text if Kibana now holds a non-string. A `json_settings` key records the JSON text, keeping the recorded text when Kibana's value is equal under the comparison in D4. Unmanaged keys are ignored. A map that is null in the model stays null (most configurations set only one of the two maps); a map that is non-null is rebuilt and may become empty. This keeps an omitted map from turning into an empty one and causing an inconsistent-result error after apply. |
| Delete | Space check first: a missing space is success with no settings request. Otherwise `null` for every managed key in one bulk `POST`; no managed keys means no call. |

Read dropping a key that was reset in Kibana means the prior state lacks it, so the next update
treats it as new and re-applies it; drift correction needs no special code. Writers outside
Terraform are last-write-wins per key.

### D7. Error handling

| Failure | Detected by | Behavior | State effect |
|---|---|---|---|
| Invalid key length, null entry, JSON syntax error, JSON `null`, key in both maps | Attribute validators and `ValidateConfig`, at plan time | Error naming the key; no Kibana call | None |
| Value unknown at plan time, invalid once known | The write path re-validates and parses every JSON value before sending | Error identifying the key, before any settings request (a spaces lookup for a non-default space may already have run) | None |
| Kibana 400 (invalid value for a registered setting, or a setting pinned by Kibana server configuration) | Existing response helpers extracting Kibana's `{statusCode, error, message}` body | Error diagnostic with Kibana's message; if the changes actually sent included entries from `string_settings` (not merely declared in configuration, since updates send only changed keys), the json_settings hint is appended to each error's detail | Failed create: resource not in state. Failed update: prior state kept. |
| A 200 response with a malformed or non-JSON body (for example a proxy login page) on read | The wrapper's body decoding | Error diagnostic; nothing is read as "no settings" | Same as any other failure |
| Kibana 403, 401, other 4xx/5xx, transport errors | Same helpers | Error with Kibana's message, plus the same json_settings hint when the changes actually sent included `string_settings` entries | Same as above |
| Space does not exist (never a 404 from the settings routes) | The space check using the spaces API, for spaces other than `default` | Read removes the resource from state; destroy succeeds; create and update fail with an error saying the space must exist, and write nothing | Read: removed. Create: not in state. Update: prior state kept. |
| The space check fails for a reason other than the space being missing (403, 5xx, transport error) | The space check | Surfaced as an error. It is never treated as "missing", so the resource is never silently dropped and a destroy is never skipped. | Same as any other failure |
| 404 from the settings route itself (for example a deployment that lacks the route) | Wrapper reports not found (a nil result on read, `notFound` on writes) | Surfaced as an error naming the missing route (`GET` or `POST /api/kibana/settings`) for read, create, update and destroy; on read the managed keys are kept, never read as "no settings"; the error carries no json_settings hint | Same as any other failure |

State is only written after a successful call and a read-back, so correctness does not depend on
Kibana applying a bulk request atomically: if part of a rejected request were applied, the next
refresh reconciles from Kibana. No new retries are added and the envelope's default timeouts apply.
Setting values are not logged beyond what the existing transport logging already does.

### D8. Testing approach

| Layer | Runs against | Covers |
|---|---|---|
| Unit: conversions | Nothing external | `buildChanges` and `refreshState` as table-driven tests: typing, removed keys becoming `null`, moved keys, dropped keys, non-string values in a string map, semantic JSON equality (exact-number comparison in `jsonValuesEqual`). |
| Unit: validators and schema | Nothing external | Key length, null entries, key in both maps, invalid JSON, JSON `null`, unknown-at-plan values; an early test that a map of normalized-JSON elements applies semantic equality (new to this repository). |
| Unit: client wrapper | `httptest` server, using the existing `newTestClient` pattern | Space-aware path (`/s/ops/...` versus unprefixed default), the internal-origin header, request body with `null` resets and exact numbers, response parsing with mixed value types, 404 handling, Kibana error-message extraction, including 400 for a setting pinned by Kibana server configuration, 403, 500 and unstructured error bodies. |
| Unit: space check | A stubbed spaces API | `default` is never checked; a missing space is reported; other spaces-API failures surface as errors. |
| Unit: overlay transform | Existing `transform_schema_test.go` pattern | The new paths are injected and operation IDs are as expected. |
| Acceptance | A live stack (`TF_ACC=1`) with the provider's `acctest` helpers | Create, update, key-removal reset, destroy; both value maps and all JSON types; unknown or custom keys; empty config; space replacement; out-of-band changes and out-of-band space deletion via `PreConfig` plus refresh; unmanaged keys surviving apply and destroy; an invalid-value error; the `kibana_connection` override. |
| Example plan test | Existing `TestAccExamples_planOnly` | The new example is self-contained. |
| Docs check | `make docs-generate` and the lint job | Generated docs are current. |

Accommodations:

- Advanced settings are shared stack state, so every acceptance test creates its own randomly named
  space through the `kibanaoapi` space wrappers (the shared acctest helper removes it again in
  cleanup; this also lets a test seed or change settings before or between steps) and prefers
  unregistered `custom:` keys; real settings such as `dateFormat:tz` appear only inside per-test
  spaces. The exceptions are tests that deliberately target a space that does not exist.
- There is no resource ID for the standard destroy check, so acceptance tests use a small custom
  check that reads the space's settings through the new wrapper and asserts managed keys are gone
  and unmanaged keys remain. The same wrapper performs the out-of-band edits.
- The repository's Kibana 9.4+ test configuration pins two settings through `uiSettings.overrides`,
  so the "pinned by Kibana server configuration" behavior is also exercised by an acceptance test,
  covered by the shared 9.5.0 floor below. A missing-privilege user needs a restricted account, so that
  scenario is covered by the wrapper's `httptest` unit tests (whose messages were copied from the
  live stack); the tasks record the layer chosen for each spec scenario.
- Acceptance tests are developed and run on the latest published Elastic Stack, and a single shared
  helper gates them to stateful deployments whose server version is 9.5.0 or later (the gate reads
  the acceptance stack's server version; 9.5.0 is the earliest version the behavior and Kibana's
  message texts were verified on), because CI runs acceptance tests across a
  stack-version matrix that includes older stacks; behavior on older versions and on serverless is
  unverified.
- Deleting a test space leaves its `config` saved object behind in Kibana, so acceptance tests that
  use random spaces accumulate harmless orphaned objects on a long-lived stack.
- A real `terraform destroy` refreshes first, so a space deleted out of band normally removes the
  resource at refresh. The acceptance harness closes each test with `terraform destroy -refresh=false`
  and a `PlanOnly` step does not persist refreshed state, so acceptance tests that delete a space (or
  reset a key) in a `PlanOnly` step's `PreConfig` reach the real destroy path without a refresh, with a
  guard asserting that the destroy ran. The destroy path is also covered by unit tests against the
  fake Kibana server (the space check itself against a stubbed spaces getter), and the
  update-after-deletion-without-refresh case only by a unit test.
- Every spec scenario maps to at least one test in whichever layer can run it.

### D9. Security and permissions

Writes require Kibana's `manage_advanced_settings` privilege; a live 403 response names it, while
settings reads are allowed for a user with no roles (both verified live). The spaces lookup
needs read access to the target space: verified live, credentials holding only
`advancedSettings: all` in one space get 200 for that space and 403 (not 404) for any other space,
existing or not, and the space check surfaces that 403 as an error. The routes are internal and
unsupported as a public API (see Context), which the resource documentation states. Values are recorded in state in
plaintext and are not marked sensitive, because readable plans are the purpose of the resource and
advanced settings are not secrets; the documentation tells users not to put secrets in them.

### D10. Extensibility for global settings

The resource's `id` is its `space_id` and its schema reserves no names that a later global
attribute pair or separate resource would need. Adding `/api/kibana/global_settings` later is an
additional overlay entry plus one wrapper function, with no change to this resource's identity,
schema or state.

### D11. Overlapping managers

Two resources that declare the same key in one space overwrite each other, and
`elasticstack_kibana_default_data_view` already manages `defaultIndex`. The provider does not try to
detect this; the resource documentation warns about it.

### D12. Performance and rollout

One bulk `POST` per write operation and one `GET` per read, plus a spaces-API `GET` per operation for
spaces other than `default` (create and update make two, since the read-after-write's Read checks
again). The change is purely additive (a new
resource and new overlay entries), so there is no migration and rollback is not using the resource.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| The settings routes are internal with no public contract, and Kibana 9.x only serves them with an internal-origin header, so Elastic could change or remove them | The overlay and wrapper are the only units that know the routes, so a change is confined to them. The documentation states the status. Raise with the Kibana team that these routes should become public. |
| A deleted space leaves its settings behind, so a recreated space with the same ID silently gets its old settings back, and a write to a never-created space creates an orphaned object | The space check prevents writes to missing spaces and drops the resource when a space disappears. The documentation notes that settings can outlive a deleted space. |
| Kibana returns integers beyond the JavaScript safe range rounded, so such a value in `json_settings` shows as drift | Documented limitation; the provider keeps exact number text on the way out. |
| Minimum Kibana version and serverless support are unconfirmed: behavior is verified only on current stateful Kibana (9.5.6), though the route source exists in every release checked | Acceptance tests are gated to stateful deployments on server version 9.5.0 and later through one shared helper, so older stacks in CI's version matrix skip them instead of failing on unverified behavior; the documentation states what was verified and what was not. |
| A map of normalized-JSON elements has no precedent in this repository | Early unit test. Fallback: a plain `map(string)` with custom normalization. |
| The overlay shadows upstream if Kibana documents these routes later | A README note (added); drop the overlay once upstream covers it. |
| String `"true"` and boolean `true` look identical in drift detection | Documented limitation. |
| Reformatting a number or JSON value in the configuration plans an in-place update that sends nothing, because the framework applies semantic equality to results, not to configuration changes | Harmless: the update sends no request and read-after-write records the planned text; the documentation says so. |
| Kibana's error text may not name the offending key | Surface it verbatim, with the `json_settings` hint for type mismatches. |
| Regeneration needs network access to download the pinned Kibana spec | Run regeneration where network is available; the overlay is the only input that changes. |
