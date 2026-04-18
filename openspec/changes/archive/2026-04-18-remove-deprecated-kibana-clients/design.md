## Context

Today the repository still ships a second Kibana OpenAPI client under `generated/slo` (produced by `make generate-slo-client` from `generated/slo-spec.yml`) and patches `github.com/disaster37/go-kibana-rest/v8` through a local `replace` into `libs/go-kibana-rest`. Multiple OpenSpec changes are migrating individual resources to `generated/kbapi` and `internal/clients/kibanaoapi`. This change is **migration plan item 9**: once those consumers are gone, delete the legacy trees and module wiring so only the consolidated kbapi stack remains.

## Goals / Non-Goals

**Goals:**

- Eliminate `generated/slo` from the repository (sources, generator inputs used only for that tree, and Makefile/CI hooks).
- Eliminate `github.com/disaster37/go-kibana-rest/v8` from `go.mod` and eliminate the `libs/go-kibana-rest` fork directory when it exists solely for the replace directive.
- Prove absence of deprecated imports and update contributor-facing docs that still mention the old paths.

**Non-Goals:**

- Migrating any remaining resource logic to kbapi (that belongs in the earlier per-entity migration changes).
- Changing Terraform schema or runtime behavior of resources except as required to compile after dependency removal (there should be no such need if preconditions are met).
- Replacing `go-kibana-rest` with a different third-party Kibana client; the single supported path is `generated/kbapi`.

## Decisions

1. **Strict sequencing** — Implementation runs only after the kbapi migrations that still reference `generated/slo` or `go-kibana-rest` are merged on the target branch. **Rationale:** Avoids a broken intermediate state. **Alternative:** delete packages first — rejected as it would not compile.

2. **Delete generator inputs with the package** — Remove `generated/slo-spec.yml` if the Makefile is the only consumer, so contributors are not tempted to revive the old generator. **Rationale:** Single source of truth for SLO shapes lives in the kbapi OpenAPI pipeline. **Alternative:** keep the YAML archived — rejected unless legal/product needs require retaining it (not assumed here).

3. **`generate-clients` becomes kbapi-only** — After removal, `generate-clients` SHALL depend only on `gen` (or equivalent documented aggregate that does not reference `generated/slo`). **Rationale:** Preserves a familiar “regenerate everything” entrypoint without resurrecting the SLO split.

4. **Module cleanup** — Remove both `require` and `replace` for `go-kibana-rest`, delete `libs/go-kibana-rest`, then `go mod tidy`. **Rationale:** A dangling replace or empty directory confuses Renovate and new contributors.

## Risks / Trade-offs

- **[Risk] Preconditions not actually met** — A hidden import in tests or tools could reappear on rebase. **Mitigation:** Mandatory `rg` / `go list`/compile gate in tasks; run full `go test ./...` and `make build`.

- **[Risk] CI or release scripts invoke `generate-slo-client`** — Removing the target breaks jobs. **Mitigation:** Grep workflows and `Makefile` callers in tasks; update or delete those steps.

- **[Risk] External forks depend on `libs/go-kibana-rest` in this repo** — Unlikely for the provider module, but if discovered, document fork removal in proposal Impact. **Mitigation:** Confirm no submodule or cross-module reference before deleting the directory.

## Migration Plan

1. Merge all prerequisite kbapi migration changes; confirm default branch has zero imports of deprecated paths.
2. Apply this change: delete trees, edit Makefile and `go.mod`, refresh docs, verify build and tests.
3. Rollback: revert the single cleanup commit (or branch) if a missed import is found post-merge; no runtime feature flag is involved.

## Open Questions

- Whether `generated/slo-spec.yml` must be retained for compliance or historical comparison; default assumption is **delete** if unused outside the removed Makefile target.
