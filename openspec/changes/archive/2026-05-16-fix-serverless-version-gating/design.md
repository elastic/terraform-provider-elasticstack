## Context

Two Elasticsearch resource types enforce server version minimums using a raw
`serverVersion.LessThan(...)` comparison with no serverless short-circuit:

- `datastreamoptions.EnforceMinServerVersion` (called from `componenttemplate/create.go` and
  `template/create.go` / `template/update.go`) — gates `data_stream_options` at ES 9.1.0.
- `validateIgnoreMissingComponentTemplatesVersion` (called from `template/create.go` /
  `template/update.go`) — gates `ignore_missing_component_templates` at ES 8.7.0.

Serverless clusters report a version that may be lower than either threshold. `client.EnforceMinVersion`
already handles serverless correctly (`if flavor == "serverless" { return true, nil }`), but neither
function routes through it.

The `entitycore` resource envelope exposes a `WithVersionRequirements` interface
(`internal/entitycore/version_requirements.go`). When a decoded model satisfies this interface,
`enforceVersionRequirements` is called automatically:
- During Read (line 251 of `resource_envelope.go`)
- Before the write callback in `runWrite` (line 358 of `resource_envelope.go`)

For component templates the write callback (`writeComponentTemplate`) is not overridden beyond what
the envelope provides — once the model implements `WithVersionRequirements`, the envelope handles
enforcement for free. For index templates, Create/Update are fully overridden on the concrete
`Resource` type, so the explicit `serverVersion` fetch and validator calls must be replaced with an
explicit loop over `plan.GetVersionRequirements()` + `client.EnforceMinVersion`.

Reference implementations of `WithVersionRequirements`:
- `internal/kibana/maintenance_window/models.go:80`
- `internal/kibana/agentbuilderagent/models.go:67`

## Goals / Non-Goals

**Goals:**

- Fix the serverless blind-spot in both version-gating functions by routing enforcement through
  `client.EnforceMinVersion` (which short-circuits for serverless).
- Implement `entitycore.WithVersionRequirements` on `componenttemplate.Data` (transparent envelope
  enforcement) and `template.Model` (explicit loop in Create/Update).
- Move the `data_stream_options` version definition (`MinSupportedVersion` constant +
  `GetVersionRequirements` helper) into the `datastreamoptions` package, making it the single
  authority for that logic.
- Delete both `version_gating.go` files once no callers remain.
- Remove `serverVersion` entirely from index template Create/Update.
- Add dedicated unit tests for `GetVersionRequirements` on each model.
- Cover Read-time enforcement (the envelope enforces requirements on Read; this is confirmed desirable
  — an old-cluster stateful resource with `data_stream_options` in state will error on `terraform refresh`).

**Non-Goals:**

- `internal/elasticsearch/index/template_sdk_shared.go:validateDataStreamOptionsVersion` — same
  serverless blind-spot but marked "Used by tests" and not on the production code path; excluded.
- Changing the numeric version thresholds (9.1.0 / 8.7.0) themselves.
- Acceptance-test skip helpers — they already use `versionutils.SkipIfUnsupported` with `FlavorAny`.
- Whether `index.MinSupportedDataStreamOptionsVersion` in `template_constants.go` becomes a thin
  re-export of `datastreamoptions.MinSupportedVersion` — acceptable either way; acceptance tests are
  unaffected.

## Decisions

### Decision 1: `datastreamoptions` package owns the version constant and shared helper

**Rationale:** The package is already the authoritative home for `data_stream_options` logic. Adding
`MinSupportedVersion` and `GetVersionRequirements` there gives both `componenttemplate` and `template`
packages a single, importable source for the version check rather than duplicating the constant or
importing it from `index` (which has no other reason to know about `datastreamoptions`).

**Alternative considered:** Keep `MinSupportedDataStreamOptionsVersion` in `template_constants.go`
(the `index` package) and add a new wrapper elsewhere. Rejected as it scatters the version definition
across packages.

### Decision 2: `componenttemplate.Data.GetVersionRequirements` delegates entirely to the shared helper

**Rationale:** The template object layout is identical between component templates and index
templates. No condition logic is needed in `Data` itself. Delegating directly to
`datastreamoptions.GetVersionRequirements(d.Template)` keeps the method minimal and reduces the
chance of drift.

### Decision 3: `template.Model.GetVersionRequirements` returns both requirements

**Rationale:** Both `data_stream_options` (≥ 9.1.0) and `ignore_missing_component_templates`
(≥ 8.7.0) apply to the index template resource. Expressing them in one method provides a single
call-site replacement for both `serverVersion` validator calls in Create/Update.

### Decision 4: Index template Create/Update use an explicit requirements loop, not the envelope

**Rationale:** Index template Create/Update are overridden on the concrete `Resource` type; the
envelope's `enforceVersionRequirements` is invoked for Read only. The explicit loop mirrors the
pattern in other overridden resources and keeps enforcement visible at the call site.

### Decision 5: Dedicated test files, not folded into expand_flatten_test.go

**Rationale:** The issue author explicitly requested dedicated test files. This also keeps
`expand_flatten_test.go` focused on expand/flatten coverage and makes the new unit tests easier to
find.

## Risks / Trade-offs

- **[Risk]** New Read-time enforcement: `terraform refresh` on a stateful cluster below ES 8.7.0 (or
  9.1.0) with the relevant attributes in state will now error rather than silently succeed.
  **Mitigation:** Confirmed desirable by the issue author; consistent with existing Kibana envelope
  semantics.

- **[Risk]** `datastreamoptions.EnforceMinServerVersion` is deleted; any caller that was not identified
  will fail to compile.
  **Mitigation:** Both callers are explicitly listed in the references. `grep` confirms the function is
  only called from `componenttemplate/create.go` and `template/expand_flatten_test.go` (and the latter
  is refactored).

- **[Risk]** The existing unit test at `template/expand_flatten_test.go:211–227` calls
  `datastreamoptions.EnforceMinServerVersion` directly and will break when that function is deleted.
  **Mitigation:** The test is explicitly in-scope and is refactored to call
  `Model.GetVersionRequirements()` instead.

## Open Questions

_(All previously open questions have been resolved by issue author comments. No blocking questions remain.)_
