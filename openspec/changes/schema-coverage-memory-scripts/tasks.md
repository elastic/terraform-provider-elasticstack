## 1. Add schema-coverage memory helper commands

- [x] 1.1 Add a Go command under `scripts/schema-coverage-rotation` that accepts the live working memory file path, bootstraps that file when needed, reconciles the canonical entity inventory from the provider's registered resources and data sources, and writes updated memory atomically.
- [x] 1.2 Implement provider-backed entity discovery by importing the `provider` package, reading Plugin Framework registrations from `provider/plugin_framework.go`, reading Plugin SDK registrations from `provider/provider.go`, and normalizing both into one de-duplicated inventory.
- [x] 1.3 Add command support for selecting the next entities by oldest timestamp with stable JSON output and deterministic tie-breaking while preserving entity type, and for recording an analyzed entity's current UTC timestamp after each analysis run.
- [x] 1.4 Add focused Go tests for inventory building, oldest-first selection, and timestamp persistence behavior.

## 2. Update the workflow prompt to use scripts

- [x] 2.1 Update the authored workflow prompt source under `.github/workflows-src/schema-coverage-rotation/` so the agent is instructed to run the Go helper command(s) after repo-memory hooks complete instead of following inline memory-format and selection prose.
- [x] 2.2 Ensure the prompt-to-script contract preserves the current bootstrap, selection, and post-analysis timestamp-update semantics without exposing the JSON memory structure in detail.

## 3. Rebuild and verify workflow artifacts

- [ ] 3.1 Recompile `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml` from the authored workflow source under `.github/workflows-src/schema-coverage-rotation/`.
- [ ] 3.2 Run the relevant OpenSpec and workflow validation checks for memory bootstrap, entity selection, prompt integration, and timestamp persistence behavior.
