Confirmed the multi-line `GITHUB_OUTPUT` problem. Here is my review.

## Review

### Correct (with evidence)
- **`merge_group` trigger added** — `.github/workflows/provider.yml` `on:` block now lists `merge_group:` (absent in `HEAD~1`, present in `HEAD`). ✅
- **`compute-packages` PR vs non-PR routing** — `git show HEAD:.github/workflows/provider.yml`, the `id: targeted` step: non-PR branch sets `has_packages=true` + `targeted_pkgs=` and `exit 0`; PR branch runs `git fetch origin main --depth=1` then the tool. Matches spec's non-PR (push/workflow_dispatch/merge_group) and PR branches. ✅
- **Tool invocation flags** — `go run ./scripts/targeted-testacc/... --total-shards=2 --shard-index=${{ matrix.shard }}` exactly matches the spec. ✅
- **All expensive steps gated on `has_packages == 'true'`** — Pre-pull fleet image, Start stack, Wait for readiness, Get ES API key, Setup Fleet, Force install synthetics, and TF acceptance tests all carry `if: ... steps.targeted.outputs.has_packages == 'true'`. ✅
- **Test step routes on `targeted_pkgs` emptiness** — `if [[ -n "${{ steps.targeted.outputs.targeted_pkgs }}" ]]; then make targeted-testacc ...; else make testacc ...; fi`. Non-PR path (`make testacc ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`) is byte-identical to the pre-change `HEAD~1` invocation. ✅
- **Teardown always runs** — "Tear down docker compose stack" uses `if: always()`. ✅
- **YAML/syntax** — valid YAML; `matrix.shard` is an integer so unquoted interpolation in `--shard-index=` / `ACCTEST_SHARD_INDEX=` is safe. No structural issues.

### Blocker

**B1 — `targeted_pkgs` is stored as raw newline-separated tool output; spec requires a space-separated list, and the simple `GITHUB_OUTPUT` format cannot represent multi-line values.**
Location: `compute-packages` step (`id: targeted`), PR branch:
```bash
targeted=$(go run ./scripts/targeted-testacc/... --total-shards=2 --shard-index=${{ matrix.shard }})
...
echo "targeted_pkgs=$targeted" >> "$GITHUB_OUTPUT"
```
`scripts/targeted-testacc/main.go:182` emits packages via `fmt.Println(pkg)` — one per line (confirmed). Command substitution preserves internal newlines (verified with `od -c`: the written line is `targeted_pkgs=github.com/foo/a\ngithub.com/foo/b\n`). GitHub Actions' simple `name=value` `$GITHUB_OUTPUT` format only captures the first line; subsequent package lines are malformed (no `=`) and dropped. Net effect: when a shard selects **2+ packages** (the common case for real PRs), `steps.targeted.outputs.targeted_pkgs` silently contains only the **first** package, so only that package's acceptance tests run → false-green CI. The spec explicitly states `targeted_pkgs=<space-separated list>`. Fix: collapse newlines to spaces, e.g. `targeted=$(go run ... | tr '\n' ' ')`, or use the heredoc delimiter syntax for `$GITHUB_OUTPUT`.

**B2 — `TARGETED_PKGS` is passed to `make` unquoted, so a space-separated value word-splits into spurious make targets.**
Location: TF acceptance tests step:
```bash
make targeted-testacc TARGETED_PKGS=${{ steps.targeted.outputs.targeted_pkgs }} ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}
```
With a space-separated `targeted_pkgs` (the spec-required shape, and the shape B1's fix would produce), the shell expands to e.g. `make targeted-testacc TARGETED_PKGS=pkg.a pkg.b ACCTEST_TOTAL_SHARDS=2 ...`. Make parses `TARGETED_PKGS=pkg.a` as a var assignment, then treats `pkg.b` as a **goal** → `No rule to make target 'pkg.b'`, failing the step. Fix: quote it — `TARGETED_PKGS="${{ steps.targeted.outputs.targeted_pkgs }}"`.

B1 and B2 currently partially mask each other (B1 truncates to one package, so B2's splitting rarely triggers), which is precisely why this can produce a false green: the single captured package runs and passes while the other selected packages never run.

### Note (non-blocking)
- **Double tool invocation in CI (wasteful, not incorrect).** The `Makefile` `targeted-testacc` target (line 111) recomputes `TARGETED_PKGS` via `$(eval TARGETED_PKGS := $(shell ... go run ./scripts/targeted-testacc/... ...))`. When CI passes `TARGETED_PKGS=...` on the command line, GNU Make's command-line precedence means the recomputed value is discarded, but `$(shell ...)` is still expanded — so the tool runs a second time (its output thrown away). This doubles tool runtime per shard. Not a spec violation (spec requires passing `TARGETED_PKGS`), but worth a follow-up to short-circuit the Makefile recompute when `TARGETED_PKGS` is already set.
- The `if [[ -n "${{ steps.targeted.outputs.targeted_pkgs }}" ]]` test is correctly quoted; no issue there.

### Residual risks
- If B1/B2 are fixed by space-separating + quoting but the tool ever emits a package path containing shell-special characters, the unquoted-across-`$(...)` capture is still safe (package import paths are benign), so risk is low.
- The Makefile double-invocation (above) remains a latent inefficiency.