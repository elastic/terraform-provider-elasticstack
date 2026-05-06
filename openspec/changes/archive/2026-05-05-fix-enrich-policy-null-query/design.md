## Context

The enrich policy resource and data source are fully implemented on Terraform Plugin
Framework (PF). The API client is in `internal/clients/elasticsearch/enrich.go`; the
Terraform model and state-mapping are in `internal/elasticsearch/enrich/models.go`.

The `query` field is typed as `jsontypes.Normalized` (an optional, normalized JSON string).
`RequiresReplace` is set on it, so any state/plan mismatch causes resource recreation.

### Root cause

`GetEnrichPolicy` does:

```go
var queryStr string
if policy.Query != nil {
    queryBytes, err := json.Marshal(policy.Query)
    ...
    queryStr = string(queryBytes)
}
```

When Elasticsearch returns `"query": null` (an explicit JSON null in the Get response),
the go-elasticsearch typed client deserializes this into a non-nil `*types.Query` pointing
to a zero-value struct. Because `policy.Query != nil` passes, `json.Marshal` is called.
For a zero-value `types.Query{}` the marshaler returns the bytes `null`, making
`queryStr = "null"` — a non-empty string that is not valid JSON representing a query.

`populateFromPolicy` has a workaround `policy.Query != "null"` that converts this back to
TF null, but it is fragile and located at the wrong layer. The real fix belongs where the
data crosses the API boundary.

### Why fix at the client layer

Moving the null-byte check to `GetEnrichPolicy` keeps `populateFromPolicy` clean. The
model layer's job is to map typed Go values to Terraform types, not to compensate for
JSON serialization artefacts produced by the go-elasticsearch typed API.

## Goals / Non-Goals

**Goals:**
- Ensure that a policy created without `query` produces `query = null` in state on every
  read, regardless of what the Elasticsearch API returns for the `query` field.
- Add an acceptance test that verifies two consecutive applies of the same no-query
  configuration produce no planned changes (idempotency).
- Harden the `checkEnrichPolicyQueryNull` test helper to detect the regression.

**Non-Goals:**
- Changing the behavior when `query` IS explicitly set (no impact on that path).
- Handling the case where Elasticsearch returns a `match_all` default query for policies
  where the user explicitly omitted `query` and wants Terraform to track it. If
  Elasticsearch returns a real non-null query body, that body SHOULD be reflected in
  state — the user can suppress drift by explicitly setting `query = jsonencode({match_all = {}})`.

## Decisions

### Fix in GetEnrichPolicy, not populateFromPolicy

**Decision**: Check `string(queryBytes) != "null"` inside `GetEnrichPolicy` before
assigning `queryStr`. Keep (but do not rely on) the `populateFromPolicy` guard as a
last-resort safety net.

**Rationale**: `GetEnrichPolicy` is the canonical boundary between the API response and
the Go model. Fixing it there prevents the bad value from propagating at all. The model
layer guard becomes defensive-only and can be removed in a follow-up cleanup.

**Alternative considered**: Fix only in `populateFromPolicy`. Rejected: leaves the
`models.EnrichPolicy` struct able to hold the string `"null"` as a meaningful value,
which is misleading.

### Keep the populateFromPolicy guard

**Decision**: Leave the `policy.Query != "null"` check in `populateFromPolicy` unchanged.

**Rationale**: Belt-and-suspenders. If another code path ever populates `models.EnrichPolicy`
with `"null"`, the model layer will still produce the correct TF null without a new bug.

**Alternative considered**: Remove the guard to simplify the code. Acceptable but not
required; deferring to avoid scope creep.

### Idempotency test as a second step in existing test

**Decision**: Extend `TestAccResourceEnrichPolicyQueryOmitted` with a second
`resource.TestStep` using the same config but `PlanOnly: true` and asserting no diff.

**Rationale**: The existing test only covers the create step. A second plan-only step
exercises the Read → Plan cycle and will fail if `query = "null"` reappears in state.

### Harden checkEnrichPolicyQueryNull

**Decision**: Change the helper to also fail when `value == "null"` (currently it
accepts this as "null enough"). Return an error if the attribute exists and its value is
the non-empty string `"null"`.

**Rationale**: The helper was written permissively to pass tests that hit the bug. Making
it strict turns it into a regression guard.

## Risks / Trade-offs

- The fix is minimal and targeted; it does not change the Terraform schema or any
  provider-visible behavior beyond fixing the bug. Risk of regression is low.
- If Elasticsearch introduces an API change that returns non-null structured query objects
  for policies created without a query, the fix will correctly surface those queries in
  state, which may cause existing no-query policies to show a diff. This is correct
  behavior and can be addressed separately if it occurs.
