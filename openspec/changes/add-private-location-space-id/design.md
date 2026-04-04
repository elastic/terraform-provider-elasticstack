## Context

Private location CRUD in `libs/go-kibana-rest/kbapi` builds URLs with `basePath("", privateLocationsSuffix)`, which resolves to the default space (`spaceBasesPath` treats `""` and `"default"` as the non-prefixed path). Synthetics monitors already pass a `space` string into `basePath(space, monitorsSuffix)`, producing `/s/<space_id>/api/synthetics/...` when the space is not default. The Terraform resource `elasticstack_kibana_synthetics_private_location` does not expose a space attribute, so non-default spaces are unreachable.

## Goals / Non-Goals

**Goals:**

- Add optional `space_id` on the private location resource, aligned with naming and semantics used by `elasticstack_kibana_synthetics_monitor` (`space_id` optional; empty string means default space).
- Thread the chosen space through `KibanaSynthetics.PrivateLocation` Create, Get, and Delete so all requests hit the correct space-scoped API paths.
- Document replacement behavior when `space_id` changes and keep import behavior coherent (import id remains the Kibana private location id; practitioners set `space_id` to match the location’s space).

**Non-Goals:**

- Changing composite import id format to embed space (unless existing patterns in the repo require it—prefer keeping import id as today plus explicit `space_id` in configuration).
- In-place update support (still unsupported).
- OpenAPI/generated `kbapi` client migration for this resource.

## Decisions

1. **Extend legacy `kbapi` private location functions with a `space string` parameter** (same order and semantics as monitor APIs: empty and `"default"` map to the default path). **Rationale:** Centralizes URL rules in `spaceBasesPath` and matches established monitor behavior. **Alternatives considered:** Per-request REST path override in the resource only (duplicates logic and risks drift).

2. **`space_id` uses `RequiresReplace` and is stored in state** after create/read. **Rationale:** Changing space implies a different API namespace; same as other identity-related attributes on this resource. **Alternatives considered:** Suppressing replace (would leave state inconsistent with API).

3. **Default when unset:** Omit attribute or set to `""` → default space, consistent with monitor resource patterns in tests and docs.

## Risks / Trade-offs

- **[Risk] Signature change in `kbapi` breaks callers** → **Mitigation:** Update all call sites in-repo (provider + tests under `libs/go-kibana-rest`); run `make build` and targeted tests.

- **[Risk] Acceptance tests may not always create a secondary Kibana space** → **Mitigation:** Follow existing monitor acceptance patterns for space-scoped tests; if stack fixtures lack a second space, document manual verification or skip conditions consistent with `testing.md`.

- **[Risk] Import without `space_id` assumes default space** → **Mitigation:** Document that imported locations in non-default spaces require `space_id` in configuration before refresh; 404 on wrong space surfaces as remove-from-state per existing read behavior.

## Migration Plan

- **Deploy:** Provider upgrade is backward compatible when `space_id` is omitted (default space behavior unchanged).
- **Rollback:** Revert provider version; state may contain `space_id` attributes—older provider versions may ignore unknown attributes depending on Terraform version and schema; practitioners can remove the attribute from config if needed.

## Open Questions

- None blocking implementation; confirm Elastic version support for space-scoped Synthetics private location APIs matches the existing minimum Kibana version gate for this resource (`8.12.0` in acceptance tests).
