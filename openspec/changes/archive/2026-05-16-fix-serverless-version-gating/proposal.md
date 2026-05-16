## Why

Two server-version enforcement functions use a raw `serverVersion.LessThan(...)` comparison with no
serverless short-circuit:

| Function | File | Attribute | Min version |
|---|---|---|---|
| `datastreamoptions.EnforceMinServerVersion` | `datastreamoptions/version_gating.go` | `data_stream_options` | ES 9.1.0 |
| `validateIgnoreMissingComponentTemplatesVersion` | `template/version_gating.go` | `ignore_missing_component_templates` | ES 8.7.0 |

Serverless Elasticsearch clusters report a version that may be lower than these thresholds, so both
functions incorrectly reject configuration that is valid on serverless. `client.EnforceMinVersion`
already handles serverless correctly (it short-circuits with `return true, nil` when the cluster
flavour is `"serverless"`), but neither function routes through it.

The `entitycore` resource envelope already calls `enforceVersionRequirements(ctx, client, &planModel)`
before every write and during Read — but only when the model implements the `WithVersionRequirements`
interface. Neither `componenttemplate.Data` nor `template.Model` currently implements this interface.

## What Changes

- **`datastreamoptions/version_gating.go`** — replace `EnforceMinServerVersion` with a
  `GetVersionRequirements` helper function and a package-level `MinSupportedVersion` constant (9.1.0).
  The helper inspects the template object and returns a `[]entitycore.VersionRequirement` rather than
  performing the version comparison itself.

- **`componenttemplate/models.go`** — add `GetVersionRequirements()` on `Data`, delegating to the
  `datastreamoptions.GetVersionRequirements` helper. The envelope now handles enforcement
  transparently, before the write callback and during Read.

- **`componenttemplate/create.go`** — remove the manual `serverVersion` fetch and
  `EnforceMinServerVersion` call from `writeComponentTemplate`. Enforcement is handled by the envelope.

- **`template/models.go`** — add `GetVersionRequirements()` on `Model`, returning two
  `VersionRequirement` entries: one for `data_stream_options` (via the shared helper) and one for
  `ignore_missing_component_templates` (when non-empty, minimum ES 8.7.0).

- **`template/create.go` and `template/update.go`** — remove the `serverVersion` fetch entirely;
  replace both validator calls with a loop over `plan.GetVersionRequirements()` +
  `client.EnforceMinVersion`. Because index template Create/Update are overridden on the concrete
  type, enforcement cannot be fully transparent and instead uses an explicit loop.

- **`template/version_gating.go`** — delete; no callers remain after the above changes.

- **`datastreamoptions/version_gating.go`** — `EnforceMinServerVersion` is replaced; file content
  becomes the new `GetVersionRequirements` + `MinSupportedVersion`.

- **Unit tests** — two new dedicated test files:
  - `componenttemplate/version_requirements_test.go` (`TestData_GetVersionRequirements`)
  - `template/version_requirements_test.go` (`TestModel_GetVersionRequirements`)
  The existing unit test in `template/expand_flatten_test.go` (lines 211–227) that calls
  `datastreamoptions.EnforceMinServerVersion` directly is refactored to test
  `Model.GetVersionRequirements()` instead.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- **`elasticsearch-index-component-template`**: REQ-027 updated to specify that version gating for
  `template.data_stream_options` is implemented via the `entitycore.WithVersionRequirements` interface
  (`componenttemplate.Data.GetVersionRequirements()`), which routes enforcement through
  `client.EnforceMinVersion` and correctly handles Serverless clusters.

- **`elasticsearch-index-template`**: REQ-012 and REQ-033 updated to specify that both version gates
  (`ignore_missing_component_templates` ≥ 8.7.0 and `data_stream_options` ≥ 9.1.0) are implemented
  via `template.Model.GetVersionRequirements()`, routing enforcement through
  `client.EnforceMinVersion` and correctly handling Serverless clusters. Both Create and Update
  replace the explicit `serverVersion` fetch and validator calls with a loop over
  `GetVersionRequirements()`.

## Impact

- `internal/elasticsearch/index/datastreamoptions/version_gating.go` — replace function, add constant
- `internal/elasticsearch/index/componenttemplate/create.go` — remove manual version check
- `internal/elasticsearch/index/componenttemplate/models.go` — add `GetVersionRequirements` method
- `internal/elasticsearch/index/template/version_gating.go` — delete
- `internal/elasticsearch/index/template/create.go` — replace `serverVersion` block with requirements loop
- `internal/elasticsearch/index/template/update.go` — replace `serverVersion` block with requirements loop
- `internal/elasticsearch/index/template/models.go` — add `GetVersionRequirements` method
- `internal/elasticsearch/index/componenttemplate/version_requirements_test.go` — new test file
- `internal/elasticsearch/index/template/version_requirements_test.go` — new test file
- `internal/elasticsearch/index/template/expand_flatten_test.go` (lines 211–227) — refactor to use `Model.GetVersionRequirements()`
- `openspec/specs/elasticsearch-index-component-template/spec.md` — update REQ-027
- `openspec/specs/elasticsearch-index-template/spec.md` — update REQ-012 and REQ-033
