## 1. Update workflow bootstrap

- [x] 1.1 Update `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` so `network.allowed` includes `go`.
- [x] 1.2 Extend the Go environment handoff step so it exports `GOPATH` and `GOMODCACHE` alongside `GOROOT`.
- [x] 1.3 Regenerate `.github/workflows/openspec-verify-label.md` and the compiled `.lock.yml` from the workflow source.

## 2. Align requirements and verify

- [x] 2.1 Sync the `ci-aw-openspec-verification` canonical spec with the approved requirement changes from this delta spec.
- [x] 2.2 Validate the updated change and specs with `make check-openspec` or equivalent `npx openspec validate --all`.
- [ ] 2.3 Run or inspect a representative verify-label workflow execution **on a branch that includes this change’s workflow revision** to confirm AWF no longer fails solely due to blocked Go module access in the review phase. (Not satisfied from CI history alone until such a run exists.)
