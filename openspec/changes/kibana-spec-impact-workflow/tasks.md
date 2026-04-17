## 1. Deterministic impact helper

- [x] 1.1 Create `scripts/kibana-spec-impact/` with commands for baseline selection, entity inventory, diff analysis, and impact reporting.
- [x] 1.2 Derive the canonical Kibana entity inventory from `provider/plugin_framework.go` and `provider/provider.go` instead of a handwritten manifest.
- [x] 1.3 Build reverse-index and diff logic that maps changed `generated/kbapi` methods or types to supported Kibana entities with structured evidence and confidence.

## 2. Repo memory and duplicate suppression

- [x] 2.1 Add a repo-memory seed file under `.github/aw/memory/` for Kibana spec-impact workflow state.
- [x] 2.2 Persist processed baseline and entity-level impact fingerprints so equivalent impacts are not reported twice.
- [x] 2.3 Add helper tests for baseline handling, entity matching, and duplicate-suppression behavior.

## 3. Agentic workflow

- [x] 3.1 Author `.github/workflows-src/kibana-spec-impact/workflow.md.tmpl` using the existing gh-aw workflow conventions.
- [x] 3.2 Add deterministic workflow gating for Kibana spec-impact inputs plus manual execution support, and configure the workflow bootstrap and safe outputs.
- [x] 3.3 Write the agent prompt so it consumes deterministic helper output, creates at most one issue per impacted entity, and suppresses weak or non-actionable matches.
- [x] 3.4 Regenerate checked-in workflow artifacts and ensure the source and generated outputs stay in sync.

## 4. Validation

- [x] 4.1 Add or update tests covering workflow-source generation and any deterministic inline scripts used by the workflow.
- [x] 4.2 Run focused helper and workflow checks, then update the change artifacts if implementation constraints differ from the current proposal or design.
