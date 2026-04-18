## Why

The provider still carries two deprecated Kibana HTTP stacks (`github.com/disaster37/go-kibana-rest/v8` via the vendored `libs/go-kibana-rest` fork and the standalone OpenAPI package under `generated/slo`). Those duplicates complicate auth, transport, regeneration, and dependency hygiene. They should be removed only as a **final cleanup** once every in-repo consumer has moved to `generated/kbapi` and `internal/clients/kibanaoapi`, so this change is explicitly sequenced **after** the outstanding kbapi migration changes land.

## What Changes

- Delete the `generated/slo` Go package and remove **all** imports of `github.com/elastic/terraform-provider-elasticstack/generated/slo` from provider and test code.
- Remove Makefile targets, Docker/OpenAPI generator wiring, and any repository inputs used **only** to produce `generated/slo` (including `generate-slo-client`, `generated/slo-spec.yml` if nothing else references it, and `generate-clients` chaining into that path).
- Remove `github.com/disaster37/go-kibana-rest/v8` from the root module: drop the `require` entry, delete the `replace github.com/disaster37/go-kibana-rest/v8 => ./libs/go-kibana-rest` directive, remove the `libs/go-kibana-rest` tree when it exists only to satisfy that replace, and run `go mod tidy` so the module graph is clean.
- Search the tree (including workflows, scripts, and contributor docs that describe deprecated paths) and confirm **no** remaining references to the deprecated import paths or generators; align high-level dev docs that still list deprecated clients.
- **BREAKING** for contributors who relied on `make generate-slo-client` or direct use of `libs/go-kibana-rest`: those workflows are removed; Kibana client work uses `generated/kbapi` / `make gen` (and related documented targets) only.

## Capabilities

### New Capabilities

- `provider-go-module-kibana-clients`: Normative requirements that the root Go module and source tree exclude deprecated Kibana client packages (`github.com/disaster37/go-kibana-rest/v8`, `github.com/elastic/terraform-provider-elasticstack/generated/slo`) and the vendored fork directory once migrations are complete.

### Modified Capabilities

- `makefile-workflows`: Remove standalone SLO OpenAPI generation (`generate-slo-client`) and redefine `generate-clients` so it no longer depends on or produces `generated/slo`; update the delta to remove the old requirement and add a consolidated codegen requirement.

## Impact

- **Preconditions:** This change **MUST NOT** start until prior migration changes have merged so that no package imports `generated/slo` or `github.com/disaster37/go-kibana-rest/v8`. Concretely, that includes (at minimum) the OpenSpec-tracked kbapi migrations such as `migrate-kibana-slo-to-kbapi`, `migrate-kibana-spaces-to-kbapi`, `migrate-kibana-security-role-to-kbapi`, `migrate-kibana-synthetics-monitor-to-kbapi`, `migrate-kibana-synthetics-private-location-to-kbapi`, and `migrate-kibana-import-saved-objects-to-kbapi`, plus any other in-flight changes still touching those import paths. If a straggling reference remains, complete or supersede that work before applying this cleanup.
- **Build / codegen:** Root `Makefile`, possible CI references to `generate-slo-client`, `go.mod` / `go.sum`, directory `libs/go-kibana-rest`, directory `generated/slo`, and `generated/slo-spec.yml` when unused elsewhere.
- **Documentation:** `dev-docs/high-level/generated-clients.md`, `dev-docs/high-level/coding-standards.md`, and any other docs or comments that mention the deprecated clients or Makefile targets.
- **Verification:** `go test ./...`, `make build`, and repository-wide search (for example `rg`) for forbidden import paths; `openspec validate` after spec deltas are written.
