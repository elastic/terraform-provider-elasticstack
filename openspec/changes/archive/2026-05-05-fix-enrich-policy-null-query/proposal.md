## Why

When an enrich policy is created without a `query` attribute, subsequent applies
cause unwanted resource replacement. The Elasticsearch API can return a non-nil
`*types.Query` pointer (pointing to a zero-value struct) even when no query was
configured. Marshaling this value produces JSON `null` bytes, which `GetEnrichPolicy`
stores as the Go string `"null"`. On the next read, Terraform sees `query = "null"`
(a string) in state versus `query = null` (Terraform null) in the configuration,
and because `query` has `RequiresReplace`, it schedules resource recreation on every
subsequent apply.

This is reported in [issue #1084](https://github.com/elastic/terraform-provider-elasticstack/issues/1084).

## What Changes

- **Client-layer fix**: In `GetEnrichPolicy` (`internal/clients/elasticsearch/enrich.go`),
  check whether `json.Marshal(policy.Query)` produces `null` bytes and skip the assignment
  if so, leaving `queryStr` empty and allowing `populateFromPolicy` to store TF null.
- **Model-layer cleanup**: The defensive `"null"` check in `populateFromPolicy`
  (`internal/elasticsearch/enrich/models.go`) becomes a belt-and-suspenders guard;
  it can remain for defence-in-depth or be removed — decision is left to the implementer.
- **Idempotency test**: Add a second apply step to `TestAccResourceEnrichPolicyQueryOmitted`
  (`internal/elasticsearch/enrich/acc_test.go`) that re-applies the same configuration
  and asserts no replacement is planned.
- **Test helper fix**: Strengthen `checkEnrichPolicyQueryNull` so it rejects the string
  value `"null"` as a valid null (i.e., no longer accepts the bug symptom as passing).
- **Spec update**: Update REQ-013 in `openspec/specs/elasticsearch-enrich-policy/spec.md`
  to explicitly require that a marshaled-null API response (JSON `null` bytes from
  `policy.Query`) SHALL be treated identically to an absent `query` field.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `elasticsearch-enrich-policy`: REQ-013 (Query mapping — read path) updated to cover
  the case where the API response contains a `query` field that marshals to JSON `null`.

## Impact

- `internal/clients/elasticsearch/enrich.go` — skip `queryStr` when marshal produces `null`
- `internal/elasticsearch/enrich/models.go` — optional cleanup of now-redundant `"null"` guard
- `internal/elasticsearch/enrich/acc_test.go` — new idempotency step + hardened helper
- `openspec/specs/elasticsearch-enrich-policy/spec.md` — spec update REQ-013
