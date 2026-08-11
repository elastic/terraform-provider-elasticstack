---
schema: openspec.code-review.v1
lane: security
change_id: fix-index-mapping-update-unidirectional-check
verdict: approve
---

# Security Review: fix-index-mapping-update-unidirectional-check

## Scope

Reviewed for authn/authz, secrets handling, input validation, injection, and
unsafe deserialization.

Changed files:
- `internal/elasticsearch/index/mappings_value.go` — new `RequiresMappingsUpdate` method
- `internal/elasticsearch/index/mappings_value_test.go` — unit tests
- `internal/elasticsearch/index/index/update.go` — call-site swap
- `internal/elasticsearch/index/index/acc_test.go` — acceptance tests
- `internal/elasticsearch/index/index/testdata/` — test fixtures
- `openspec/changes/fix-index-mapping-update-unidirectional-check/specs/elasticsearch-index/spec.md` — spec delta

## Findings

No CRITICAL or WARNING findings.

### SUGGESTION — Document `decodeMappingPair` call precondition in the function signature comment (low priority)

`internal/elasticsearch/index/mappings_value.go:190`

`decodeMappingPair` calls `json.Unmarshal([]byte(v.ValueString()), ...)` on
both receivers. If called on a null or unknown `MappingsValue`, `ValueString()`
returns `""` (the zero value from the embedded `Normalized`), which causes an
unmarshal error at runtime. The existing comment correctly documents this
precondition ("Must only be called after null/unknown guards") and all current
callers obey it. The suggestion is to make the guard explicit (e.g. a panic or
early-return with a diagnostic) so a future caller cannot silently violate it.
This is a defensiveness suggestion only — there is no reachable bug in the
current change.

## Domain-by-domain assessment

**Authn / authz**: No changes to authentication or authorization paths. The
modified code path (`updateMappings` → `RequiresMappingsUpdate`) sits behind
the Terraform provider framework's own plan/apply lifecycle, which enforces
Elasticsearch credential handling at the `GetElasticsearchClient` call site
above it. No new surface introduced.

**Secrets handling**: No secrets, credentials, or sensitive values are
introduced, logged, or passed through the new code. The mapping JSON that flows
through `RequiresMappingsUpdate` is structural schema metadata, not user data.

**Input validation**: All inputs to `RequiresMappingsUpdate` are
`MappingsValue` structs already validated and normalized by the Terraform plugin
framework and `NewMappingsValue`. The method's null/unknown guards (`typeutils.IsKnown`)
run before any JSON unmarshaling. `json.Unmarshal` into `map[string]any` is
safe: it handles arbitrary valid JSON without reflection-based deserialization
risks. Invalid JSON would surface as a diagnostic error, not a panic.

**Injection**: `indexName` is derived from the state's composite ID (already
persisted by Terraform) and passed to the Elasticsearch client library, not
interpolated into a shell command or SQL query. `planMappings.ValueString()`
is a normalized JSON string passed as a body to the ES REST API — injection is
not a concern for JSON document bodies sent to a typed API client.

**Unsafe deserialization**: Only `encoding/json` (stdlib) is used with
`map[string]any` targets. No `reflect.Value`-based deserialization, no
`unsafe.Pointer`, no `interface{}` coercions beyond standard JSON types.
Recursive helpers (`MappingsSemanticallyEqual`, `propertiesSemanticallyEqual`,
`fieldSemanticallyEqual`) existed before this change; `RequiresMappingsUpdate`
adds one new call through the same stack. No new recursion depth introduced.

**DoS / resource exhaustion**: Not a new risk. The JSON parse-and-traverse
path pre-existed in `StringSemanticEquals`; `RequiresMappingsUpdate` follows
the same pattern. Mapping objects are bounded by Elasticsearch's own mapping
complexity limits, enforced upstream.

## Verdict

**approve** — no security issues found. The change is narrowly scoped,
introduces no new trust boundaries, and all JSON handling is safe.
