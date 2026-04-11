## Why

The `schema-coverage-rotation` workflow currently describes entity discovery, selection order, and memory JSON updates directly in the agent prompt. That logic depends on repo memory that is only available after built-in action hooks, so it should live in workspace scripts the agent can run against the real memory files instead of encoding the memory contract in prompt prose.

## What Changes

- Add a Go helper under `scripts/schema-coverage-rotation` that loads the schema-coverage memory file, bootstraps it from the repository seed when needed, and maintains the canonical entity inventory by reading the provider's registered resources and data sources directly while removing entities that are no longer registered.
- Add scripted selection logic that chooses the next entities to analyze by oldest timestamp across resources and data sources while preserving entity type.
- Add scripted memory-update logic so analyzed entities have their timestamps persisted after each run without requiring the prompt to describe the JSON structure in detail.
- Replace the current detailed memory-format and entity-selection prose in the workflow prompt with concise instructions telling the agent which script commands to run after repo-memory hooks complete.

## Capabilities

### New Capabilities
- `ci-schema-coverage-rotation-memory`: script-driven schema-coverage entity discovery, selection, and memory persistence for the rotation workflow

### Modified Capabilities
<!-- None. -->

## Impact

- `.github/workflows/schema-coverage-rotation.md`
- `.github/workflows/schema-coverage-rotation.lock.yml`
- `scripts/schema-coverage-rotation/`
- `provider/plugin_framework.go` and `provider/provider.go`
- New Go-based schema-coverage helper commands invoked by the agent after repo-memory initialization
- Memory bootstrap and persistence flow rooted at `.github/aw/memory/schema-coverage.json` and the repo-memory working copy
