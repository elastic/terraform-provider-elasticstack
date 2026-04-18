## 1. Preconditions and inventory

- [x] 1.1 Confirm on the integration branch that all kbapi migration changes that touch `generated/slo` or `github.com/disaster37/go-kibana-rest/v8` are merged (or explicitly superseded) so this cleanup can compile.
- [x] 1.2 Run `rg 'generated/slo|disaster37/go-kibana-rest'` from the repository root; resolve any hits outside `openspec/changes/**/archive/**` and historical change folders before proceeding.

## 2. Remove generated/slo and its generator

- [x] 2.1 Delete the `generated/slo` directory and remove `generate-slo-client` from the root `Makefile`; retarget `generate-clients` to invoke only `gen` (or the documented kbapi aggregate) with no dependency on a SLO-only generator.
- [x] 2.2 Delete `generated/slo-spec.yml` if nothing else references it after Makefile removal; otherwise document the exception in the change implementation notes.
- [x] 2.3 Search `.github/workflows`, scripts, and `contributing.md` / other automation for `generate-slo-client` or `generated/slo` and update or remove those references.

## 3. Remove go-kibana-rest module wiring

- [x] 3.1 Remove the `require` and `replace` entries for `github.com/disaster37/go-kibana-rest/v8` from root `go.mod`; run `go mod tidy` and commit the resulting `go.sum` updates as part of the implementation PR (not this proposal PR).
- [x] 3.2 Delete `libs/go-kibana-rest` when it exists only as the fork for that replace; confirm no other module or tooling path depends on it.

## 4. Documentation and verification

- [x] 4.1 Update `dev-docs/high-level/generated-clients.md` and `dev-docs/high-level/coding-standards.md` (and any other contributor docs found in 1.2) so they no longer instruct use of `generated/slo` or `libs/go-kibana-rest`.
- [x] 4.2 Run `make build` and `go test ./...` at minimum; fix any compile or test drift from removed packages.
- [x] 4.3 Re-run forbidden-path search from 1.2; ensure only benign matches remain (for example archived OpenSpec text if allowed by policy).
- [x] 4.4 Run `openspec validate` for this change after implementation (or `make check-openspec` per repo workflow) so delta specs and tasks stay coherent.
