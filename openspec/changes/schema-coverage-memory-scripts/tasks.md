## 1. Add schema-coverage memory helper commands

- [ ] 1.1 Add a Go command under `scripts/schema-coverage-rotation` that bootstraps the working memory file and reconciles the canonical entity inventory from the provider's registered resources and data sources, including pruning entities that are no longer registered.
- [ ] 1.2 Implement provider-backed entity discovery by importing the `provider` package, reading Plugin Framework registrations from `provider/plugin_framework.go`, reading Plugin SDK registrations from `provider/provider.go`, and normalizing both into one de-duplicated inventory.
- [ ] 1.3 Add command support for selecting the next entities by oldest timestamp while preserving entity type, and for recording an analyzed entity's current UTC timestamp after each analysis run.
- [ ] 1.4 Add focused Go tests for inventory building, oldest-first selection, and timestamp persistence behavior.

## 2. Update the workflow prompt to use scripts

- [ ] 2.1 Update `.github/workflows/schema-coverage-rotation.md` so the agent is instructed to run the Go helper command(s) after repo-memory hooks complete instead of following inline memory-format and selection prose.
- [ ] 2.2 Ensure the prompt-to-script contract preserves the current bootstrap, selection, and post-analysis timestamp-update semantics without exposing the JSON memory structure in detail.

## 3. Rebuild and verify workflow artifacts

- [ ] 3.1 Recompile `.github/workflows/schema-coverage-rotation.lock.yml` from the markdown workflow source.
- [ ] 3.2 Run the relevant OpenSpec and workflow validation checks for memory bootstrap, entity selection, prompt integration, and timestamp persistence behavior.
