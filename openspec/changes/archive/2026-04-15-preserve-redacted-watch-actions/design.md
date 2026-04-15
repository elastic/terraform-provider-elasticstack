## Context

`elasticstack_elasticsearch_watch` currently treats `actions` like any other normalized JSON blob: create/update unmarshals the configured JSON into the Put Watch request, and read marshals the Get Watch response back into state. That breaks for action secrets because Watcher may return nested placeholders such as `::es_redacted::` instead of the original password value.

The immediate failure happens after a successful create or update. The resource performs a read-after-write, stores the redacted placeholder in `actions`, and later reuses that state-derived JSON when another update touches an unrelated field. Elasticsearch rejects the placeholder as an invalid password value, so ordinary watch updates fail.

## Goals / Non-Goals

**Goals:**

- Preserve last-known action secret values in Terraform state when Watcher read responses redact those values.
- Keep refresh authoritative for non-secret action fields so real drift outside redacted leaves still appears in state.
- Scope the fix to the watch resource without introducing a provider-wide JSON comparison type.
- Add focused regression coverage for the redacted-actions update path.

**Non-Goals:**

- Introduce a new generic custom Terraform type for redaction-aware JSON equality.
- Change the `actions` schema shape or split secret values into separate Terraform attributes.
- Recover original secret values during import or any first read where Terraform has no prior concrete `actions` value to preserve.
- Broaden the change to `trigger`, `input`, `condition`, `metadata`, or other resources unless they show the same concrete failure mode.

## Decisions

### 1. Preserve redacted leaves during read instead of changing `actions` type semantics

The implementation will keep `actions` as a normalized JSON string and adjust watch read/state mapping so redacted leaves from the API are replaced with the corresponding prior value from plan/state before final state is written.

Why:

- The bug is caused by read-time state replacement, so fixing read-time behavior addresses the root cause directly.
- Preserving the prior concrete secret keeps future password changes visible to Terraform, because state continues to hold the last applied value instead of a wildcard placeholder.
- The provider already uses resource-specific preservation patterns when APIs omit or redact sensitive values, so this approach matches existing practice better than introducing new type semantics.

Alternatives considered:

- Add a custom `actions` type with redaction-aware `StringSemanticEquals`: rejected as the primary fix because treating `::es_redacted::` as semantically equal to any configured string can hide legitimate secret rotations.
- Preserve the entire prior `actions` document whenever any redaction appears: rejected because it would hide unrelated drift in non-secret action fields that the API still returns accurately.

### 2. Merge only matching paths where the API returns a redaction sentinel

The read helper will decode both the API `actions` document and the prior Terraform `actions` value, walk them recursively, and replace only string leaves equal to the Watcher redaction sentinel with the prior value at the same path when one exists.

Why:

- Path-level replacement keeps the API authoritative everywhere except at the exact redacted leaf.
- Restricting preservation to matching paths avoids copying stale values into unrelated branches when the action structure changes.
- A recursive merge helper is straightforward to unit test without involving Terraform framework internals.

Alternatives considered:

- Compare the full JSON blobs semantically and keep state unchanged when they are "close enough": rejected because the resource still needs a concrete merged document to write back into state.
- Special-case only known password field names: rejected because the issue is defined by API redaction behavior, not by a stable field-name allowlist.

### 3. Keep imports and first reads explicit about their limitation

When Terraform has no prior concrete `actions` value, the resource will store the API response as returned, even if it contains redacted placeholders.

Why:

- There is no trustworthy prior secret value to restore on import or the first refresh after state loss.
- Making this limitation explicit keeps the design honest and avoids inventing hidden storage just for this fix.

Alternatives considered:

- Introduce private-state secret storage for actions: rejected as out of scope for this bug fix because it adds a broader state-management pattern and migration burden.

## Risks / Trade-offs

- **[Risk] Out-of-band cluster changes to redacted action secrets will not appear in Terraform state once a prior concrete value is being preserved** -> Mitigation: limit preservation to redacted leaves only and document that imported/first-read resources without prior values cannot recover the original secret.
- **[Risk] Merge logic could accidentally preserve stale values when action structure changes significantly** -> Mitigation: only reuse a prior value when the API path is redacted and a corresponding prior path exists; otherwise keep the API value.
- **[Risk] Acceptance coverage may be harder if Watcher redaction behavior varies by stack version or action shape** -> Mitigation: add deterministic unit coverage for the merge helper and keep the acceptance case narrowly focused on a known redacted auth path.

## Migration Plan

1. Add a watch-local helper that merges API `actions` JSON with prior Terraform `actions` JSON, preserving only redacted string leaves.
2. Update watch read/state mapping to use that merged document before writing `actions` into state.
3. Add unit coverage for nested redaction-preservation behavior and extend watch acceptance coverage for unrelated updates after redacted action secrets are present.
4. Validate the OpenSpec change and use the new tests as the implementation guardrail.

## Open Questions

- None.
