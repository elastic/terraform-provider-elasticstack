## Context

Private location CRUD in `libs/go-kibana-rest/kbapi` builds URLs with `basePath("", privateLocationsSuffix)`, which resolves to the default space (`spaceBasesPath` treats `""` and `"default"` as the non-prefixed path). Synthetics monitors already pass a `space` string into `basePath(space, monitorsSuffix)`, producing `/s/<space_id>/api/synthetics/...` when the space is not default. The Terraform resource `elasticstack_kibana_synthetics_private_location` does not expose a space attribute, so non-default spaces are unreachable.

## Goals / Non-Goals

**Goals:**

- Add optional `space_id` on the private location resource, aligned with naming and semantics used by `elasticstack_kibana_synthetics_monitor` (`space_id` optional; empty string means default space).
- Thread the chosen space through `KibanaSynthetics.PrivateLocation` Create, Get, and Delete so all requests hit the correct space-scoped API paths.
- Document replacement behavior when `space_id` changes and keep import behavior aligned with other space-scoped Kibana resources: practitioners MAY import with a composite id `<space_id>/<private_location_id>` so the space segment is unambiguous; read/delete resolve the effective space from configured `space_id` when present, and otherwise from the composite id’s space segment when the stored id is composite (for example immediately after import, before `space_id` is present in state).

**Non-Goals:**

- Replacing the established provider composite import convention (`<space_id>/<resource_id>` for non-default spaces) with import ids that omit the space and rely on configuration alone.
- In-place update support (still unsupported).
- OpenAPI/generated `kbapi` client migration for this resource.

## Decisions

1. **Extend legacy `kbapi` private location functions with a `space string` parameter** (same order and semantics as monitor APIs: empty and `"default"` map to the default path). **Rationale:** Centralizes URL rules in `spaceBasesPath` and matches established monitor behavior. **Alternatives considered:** Per-request REST path override in the resource only (duplicates logic and risks drift).

2. **`space_id` uses `RequiresReplace` and is stored in state** after create/read. **Rationale:** Changing space implies a different API namespace; same as other identity-related attributes on this resource. **Alternatives considered:** Suppressing replace (would leave state inconsistent with API).

3. **Default when unset:** Omit attribute or set to `""` → default space, consistent with monitor resource patterns in tests and docs.

4. **Import uses the provider’s composite id form for non-default spaces:** `terraform import ... <space_id>/<private_location_id>`. **`effectiveSpaceID`** (in the resource implementation) chooses the Kibana space for API calls from configured `space_id` when it is set; if `space_id` is null, unknown, or empty and the stored resource id parses as composite, the space segment from that id is used (for example right after import before refresh materializes `space_id`). **Rationale:** Matches other space-scoped resources in this provider and avoids relying on config-only import semantics that differ from documented import syntax. **Alternatives considered:** Import identifier only the bare private location id with mandatory `space_id` in configuration—rejected in favor of consistency with existing composite import patterns.

## Risks / Trade-offs

- **[Risk] Signature change in `kbapi` breaks callers** → **Mitigation:** Update all call sites in-repo (provider + tests under `libs/go-kibana-rest`); run `make build` and targeted tests.

- **[Risk] Acceptance tests may not always create a secondary Kibana space** → **Mitigation:** Follow existing monitor acceptance patterns for space-scoped tests; if stack fixtures lack a second space, document manual verification or skip conditions consistent with `testing.md`.

- **[Risk] Wrong space on import** → **Mitigation:** Document composite import `<space_id>/<private_location_id>` for non-default spaces; the implementation derives the API space from configured `space_id` when set, and otherwise from the composite id’s space segment when `space_id` is empty or not yet in state. If the practitioner imports with only the bare private location id and omits `space_id`, behavior matches default-space assumptions; 404 on wrong space surfaces as remove-from-state per existing read behavior.

## Migration Plan

- **Deploy:** Provider upgrade is backward compatible when `space_id` is omitted (default space behavior unchanged).
- **Rollback:** Revert provider version; state may contain `space_id` attributes—older provider versions may ignore unknown attributes depending on Terraform version and schema; practitioners can remove the attribute from config if needed.

## Open Questions

- None blocking implementation; confirm Elastic version support for space-scoped Synthetics private location APIs matches the existing minimum Kibana version gate for this resource (`8.12.0` in acceptance tests).
