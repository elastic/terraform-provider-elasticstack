## 1. Add maintenance targets

- [x] 1.1 Add a Makefile target that checks whether `.github/workflows/openspec-verify-label.md` `runtimes.go.version` matches the Go version declared in `go.mod`.
- [x] 1.2 Add a Makefile target that syncs `.github/workflows/openspec-verify-label.md` `runtimes.go.version` from `go.mod` and recompiles `.github/workflows/openspec-verify-label.lock.yml`.
- [x] 1.3 Update `make check-lint` so it runs the drift-check target.
- [x] 1.4 Leave `make renovate-post-upgrade` unchanged.

## 2. Update workflow source

- [x] 2.1 Add `runtimes.go.version` to `.github/workflows/openspec-verify-label.md` using the repository's current Go version.
- [x] 2.2 Preserve the explicit `actions/setup-go` step in `.github/workflows/openspec-verify-label.md` so the runner environment continues to read `go.mod` before repository setup commands run.
- [x] 2.3 Update workflow instructions or comments to distinguish runner Go setup through `actions/setup-go` from agent Go runtime configuration through `runtimes.go.version`.

## 3. Regenerate and verify compiled workflow

- [x] 3.1 Recompile `.github/workflows/openspec-verify-label.lock.yml` from the updated markdown workflow source.
- [x] 3.2 Confirm the compiled workflow requests the intended agent Go runtime while the source workflow still preserves runner-side `actions/setup-go`, review, archive, and cleanup behavior.
- [x] 3.3 Verify the explicit Go version in workflow frontmatter matches the version declared in `go.mod`.
