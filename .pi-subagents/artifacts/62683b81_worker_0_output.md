 ## Verification Summary

All OpenSpec planning artifacts for `selective-acceptance-tests` exist and the implementation is structurally complete:

- `scripts/targeted-testacc/` contains the Go tool with all required source files.
- `Makefile` declares `TARGETED_TESTACC_BASE`, `targeted-testacc`, and `targeted-testacc-dry-run`.
- `.github/workflows/provider.yml` adds `merge_group`, a `compute-packages` step, gates expensive steps on `has_packages`, and routes the test step between `targeted-testacc` and `testacc`.

Validation I ran:
- `go test ./scripts/targeted-testacc/...` ✅
- `go build ./...` ✅
- `go vet ./scripts/targeted-testacc/...` ✅
- `npx openspec validate --specs` ✅
- `make targeted-testacc-dry-run` ✅ (selected 43 packages for this diff)
- `make targeted-testacc-dry-run ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1` ✅ (shard 1 got 21 packages because the set is above the 30-package threshold)

### CRITICAL findings

1. **`Makefile:110-117` — `targeted-testacc` does not stop on empty package list.**  
   The recipe prints the “No acceptance test packages selected …” notice and then `exit 0`, but because the next line is a separate shell invocation, `make` continues and invokes `gotestsum --packages=""`, which fails. This violates the spec scenario “No packages selected exits cleanly”.

2. **`openspec/changes/selective-acceptance-tests/tasks.md` — all 32 tasks are still unchecked.**  
   The implementation appears done, but the planning artifact does not reflect completion and must be updated before archiving.

### WARNING findings

- **`scripts/targeted-testacc/entityname.go`** extracts Terraform entity strings from *any* `.go` file, including `*_test.go`. For non-resource packages whose tests contain entity strings, this can trigger false-positive phase-2 selections.
- Manual validation of the `< 30 packages → shard index > 0 emits nothing` path was not performed with a real diff; only the unit tests cover it.
- The `compute-packages` CI step relies on the GitHub Actions default `pipefail` behavior; making it explicit would be safer.

### Archive readiness

**Not ready to archive.** The implementation matches the specs, but the empty-selection Makefile bug is a real failure path and `tasks.md` is not marked complete.