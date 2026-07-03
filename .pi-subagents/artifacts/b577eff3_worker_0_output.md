 Implemented the required `scripts/targeted-testacc/` fixes.

Changed files:
- `scripts/targeted-testacc/main.go`
- `scripts/targeted-testacc/acctestpackages.go`
- `scripts/targeted-testacc/entityname.go`
- `scripts/targeted-testacc/depgraph.go`

Commits created:
- `8ce930660 fix(scripts/targeted-testacc): empty diff defaults to full suite and phase-2 consumers filtered to acc packages`
- `6a753930f fix(scripts/targeted-testacc): correct TestAcc regex, skip unparseable Go files, avoid mutating caller slices`

Validation:
- `go vet ./scripts/targeted-testacc/...` passed
- `go build ./scripts/targeted-testacc/...` passed
- `go run ./scripts/targeted-testacc/... --base=HEAD --dry-run` selected all 129 acceptance test packages
- `go run ./scripts/targeted-testacc/... --base=HEAD~1 --dry-run` selected 0 packages (only tool source files changed)
- `go test ./scripts/targeted-testacc/...` had no test files

Open risks/questions:
- The optional verbose warning for unknown component suffixes was not implemented.
- Pre-existing unstaged changes remain in `openspec/changes/selective-acceptance-tests/*`, `scripts/targeted-testacc/gitdiff.go`, and `scripts/targeted-testacc/selector.go`; they were left untouched.