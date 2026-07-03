 Implemented task 1 (subtasks 1.1–1.8) of the `selective-acceptance-tests` OpenSpec change.

**Files created** under `scripts/targeted-testacc/`:

| File | Main exported APIs / responsibility |
|---|---|
| `main.go` | `run()`: flag parsing, orchestration, stdout plaintext or dry-run output |
| `classifier.go` | `Classifier.Classify()` → `ClassifyResult{ForceAll, Packages, HasCode}`; force-all prefix detection; `testdata/` mapping |
| `depgraph.go` | `BuildImportGraph()`, `BuildReverseDepGraph()`, `WalkReverseDeps()` |
| `entityname.go` | `ExtractEntities()`, `EntityRef`, `EntityRef.FullName()`; regex extraction skipping comment matches |
| `testconsumers.go` | `FindTestConsumers(root, modulePath, entityName)` |
| `acctestpackages.go` | `FindAccTestPackages(root, modulePath)` |
| `selector.go` | `SelectPackages()`, `ApplyShard()` |
| `gitdiff.go` | `ResolveBaseline()`, `GitDiff()` |

**Git commits created:**

- `6cb534de8` feat(targeted-testacc): add Go tool scaffold for selective acceptance tests
- `3553a8bc5` fix(targeted-testacc): ignore comment matches when extracting entity names

**Validation:**

- `go build ./scripts/targeted-testacc/...`: OK
- `go vet ./scripts/targeted-testacc/...`: OK
- `go run ./scripts/targeted-testacc/... --base=HEAD~1 --dry-run`: 0 packages selected for tool-only diff ✅
- `go run ... --base=HEAD~1 --total-shards=2 --shard-index=1 --dry-run`: no packages for this shard ✅
- `go run ... --base=HEAD~6 --dry-run --verbose`: force-all triggered, 130 acc-test packages selected ✅
- `go run ... --base=HEAD~6 --total-shards=2 --shard-index=0`: emits 65 packages ✅
- `go run ... --base=HEAD~6 --total-shards=2 --shard-index=1`: emits 65 packages ✅
- `go run ... --base=HEAD~6 --total-shards=1 --shard-index=1`: emits 0 packages ✅
- `go run ... --base=HEAD --dry-run`: empty diff → 0 packages ✅

**Open risks / notes:**

- Empty diff currently returns zero packages per the explicit task instruction, which diverges from the spec heading “Empty diff defaults to full suite” but aligns with the spec’s “Only docs files changed produces no packages” scenario. This will need to be reconciled if the final validator enforces the literal spec heading.
- No unit tests added yet (task 2); helpers are designed to be stateless/pure so they remain unit-testable when task 2 is picked up.
- Existing unstaged modifications to `openspec/changes/selective-acceptance-tests/*` and `.pi-subagents/` were present before my work; I did not stage or commit them.