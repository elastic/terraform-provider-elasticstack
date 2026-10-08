## Context

The authoritative implementation-research comment (workflow run
[37594206135](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/37594206135))
is the scope baseline for this proposal; no human direction or contradicting issue/comment content
was present on this run, so its recommendation is adopted as-is.

`status_json` on both `elasticstack_kibana_security_entity_store` and
`elasticstack_kibana_security_entity_store_status` is built straight from the raw Kibana status
response body:

- `readEntityStore` ([`internal/kibana/security_entity_store/read.go:55`](../../../internal/kibana/security_entity_store/read.go))
- `readEntityStoreDataSource` ([`internal/kibana/security_entity_store/data_source_read.go:48`](../../../internal/kibana/security_entity_store/data_source_read.go))

Both do `model.StatusJSON = jsontypes.NewNormalizedValue(string(rawBody))`. `jsontypes.Normalized`'s
semantic-equality comparison round-trips JSON through `interface{}`, which makes object key order
insignificant but leaves JSON array order significant. Kibana's `engines` array can come back in a
different order across two reads of the same logical state (`generic`/`user` swapped), so two
logically-identical responses compare unequal — the direct cause of
`TestAccResourceKibanaSecurityEntityStore_import`'s `ImportStateVerify` failure on `status_json`,
and a latent source of real-user `terraform plan` noise since `status_json`'s only plan modifier is
`UseStateForUnknown()`.

`flattenStatus` already solves this exact instability for `entity_types` via
`sort.Strings(typesList)` ([`internal/kibana/security_entity_store/helpers.go:277`](../../../internal/kibana/security_entity_store/helpers.go)).
`status_json`'s raw-passthrough path has no equivalent.

## Goals / Non-Goals

**Goals:**
- Make `status_json` order-stable across reads of logically-identical Entity Store status data, by
  normalizing `engines` array order before the value is stored.
- Fix the root cause so both the acceptance test flake and the latent real-user plan-noise risk are
  addressed, not just the test symptom.
- Share one normalization helper between the resource and data-source Read paths.

**Non-Goals:**
- No schema, attribute-type, or API changes. `status_json` stays a `jsontypes.Normalized` string
  attribute.
- No change to `entity_types`, `log_extraction`, or the data source's `engines` typed list — these
  are already derived from the typed `entityStoreStatus` struct and already order-stable
  (`entity_types` via `sort.Strings`; `engines` is a list, not a set, and its order already matches
  insertion order of the API response per read — this proposal does not change list ordering
  semantics for the typed `engines` attribute, only for the raw `status_json` passthrough).
- No broader `status_json` schema redesign (e.g. a fully-typed nested attribute) — out of scope per
  the research comment.
- Not investigating why Kibana's API itself returns `engines` in varying order server-side — out of
  scope per the research comment.

## Decisions

- **Approach: normalize `engines` order before building `status_json` (research comment's
  Approach A).** Decode the raw body generically (`map[string]json.RawMessage]`), pull `"engines"`
  as `[]json.RawMessage`, stable-sort (`sort.SliceStable`) by each element's `"type"` field, re-encode
  the sorted array back into the map, and re-marshal the whole map for the final string passed to
  `jsontypes.NewNormalizedValue`. All response fields, including unmodeled and nested fields, must be
  preserved semantically. Re-marshaling may alter formatting, object key order, or other incidental
  byte representation. Exact-byte equality is required only for otherwise-identical response bodies
  with reversed `engines` arrays and for explicit passthrough paths that return the original body.
  Operating on generically-decoded JSON (rather than re-marshaling from the already-decoded
  `entityStoreStatus` struct) avoids dropping any response field not modeled by `entityStoreEngine`.
- **Rejected: ignore `status_json` in `ImportStateVerify` (research comment's Approach B).** This
  is a pure test change that would leave the latent real-user plan-noise risk unaddressed and weaken
  this test's coverage of `status_json`. The issue's own "likely fix" note lists this only as a
  fallback after order normalization; no contradicting signal was found in the issue body or
  comments to justify choosing the weaker fix.
- **Sort key and tie-break.** Elastic's own API documentation for
  `GET /api/security/entity_store/status` does not state that `engines[].type` is unique, but the
  rest of this codebase already treats entity type as a de-facto unique/set key for the same data
  (one engine per installed entity type; `entity_types` is modeled as a Terraform `Set` and
  `flattenStatus` sorts purely by `type`). This proposal follows that same precedent: sort solely by
  `type`, using a *stable* sort so that if two engines ever did share a `type`, their relative order
  from the original response is preserved rather than becoming non-deterministic. This does not
  fully resolve the open question below (whether duplicate types are possible) but ensures the
  normalization degrades safely (stable, not undefined) if they occur.
- **Shared helper location.** The new helper lives in
  `internal/kibana/security_entity_store/helpers.go`, next to `flattenStatus` and
  `getEntityStoreStatus`, and is called from both `read.go` and `data_source_read.go` so the two
  Read paths cannot drift.
- **Error handling.** If the raw body cannot be decoded generically for normalization (should not
  happen, since `getEntityStoreStatus` already successfully `json.Unmarshal`s it into the typed
  `entityStoreStatus` struct beforehand), the helper returns the original, unnormalized body
  unchanged byte-for-byte rather than failing the Read — a cosmetic normalization step should not
  turn into a hard Read failure.

## Open questions

- Does `jsontypes.Normalized.StringSemanticEquals` (v0.2.0) really implement "unordered objects /
  ordered arrays," or something stricter? Doesn't block Approach A's correctness, but affects how
  precisely the fix is described.
- Can two engines ever share the same `type` (e.g. mid-migration), requiring a secondary stable sort
  key?
- Is this reproducible on stable (non-SNAPSHOT) Kibana, or specific to `9.6.0-SNAPSHOT`?
- Are there other consumers of `getEntityStoreStatus`'s raw body besides the two Read paths found
  here?

## Risks / Trade-offs

- [Risk] Sorting only by `type` assumes `type` is an adequate stable key for `engines`. → Mitigation:
  use a stable sort (preserves original relative order for any tie), matching the existing
  `entity_types` precedent; revisit if the "duplicate engine type" open question above is ever
  confirmed to occur in practice.
- [Risk] Generic JSON decode/re-encode of the full response body (rather than only touching
  `engines`) can change numeric formatting or other incidental representation details Kibana emits.
  → Mitigation: preserve the semantic value of every field, including unmodeled and nested fields;
  require exact-byte equality only for otherwise-identical bodies whose `engines` order differs and
  for explicit passthrough paths that return the original body unchanged.
