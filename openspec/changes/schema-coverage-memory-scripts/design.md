## Context

The `schema-coverage-rotation` workflow currently embeds memory behavior directly in the agent instructions: the prompt describes the JSON layout, how to initialize the repo-memory file, how to rebuild the entity inventory from docs, how to select the next entities by age, and when to write timestamps back. That is brittle because repo memory is materialized only after built-in action hooks complete, so the prompt cannot move this work into pre-activation the way issue-slot gating can.

This change keeps the logic in the agent path but moves it out of prose and into repository-local scripts that run against the live repo-memory workspace after hooks have finished.

## Goals / Non-Goals

**Goals:**
- Encapsulate schema-coverage memory bootstrap, entity reconciliation, oldest-first selection, and timestamp persistence in a Go command under `scripts/schema-coverage-rotation`.
- Let the agent invoke those scripts after repo-memory hooks complete instead of reconstructing memory behavior from prompt prose.
- Build the canonical entity inventory from the provider's registered resources and data sources, not from documentation filenames.
- Preserve the existing selection semantics: `null` timestamps first, then oldest timestamps across resources and data sources while retaining entity type.
- Remove detailed memory JSON format instructions from the workflow prompt.

**Non-Goals:**
- Moving repo-memory-dependent logic into pre-activation.
- Changing the schema-coverage issue-slot gate or open-issue cap.
- Changing the schema-coverage analysis rubric, issue body requirements, or safe-output behavior.

## Decisions

Implement the helper as a Go command under `scripts/schema-coverage-rotation`.
The workflow should direct the agent to run a checked-in Go command after repo-memory hooks finish. Keeping the logic in Go makes it easier to import the provider package, reuse repository types, and test the selection and persistence logic with the normal Go toolchain.

Alternative considered: implement the helper in shell or JavaScript.
Rejected because the canonical entity list now needs to derive from registered provider entities, which is most directly accessible from Go code in this repository.

Alternative considered: keep the memory logic in prompt prose.
Rejected because it duplicates deterministic behavior in natural language, increases prompt size, and exposes the internal memory schema as part of the agent contract.

Use a command-oriented interface rather than documenting the JSON schema to the agent.
The scripts should expose a small set of operations, such as preparing the canonical inventory and selecting entities, then recording analysis completion. The prompt can name those commands and expected outputs without describing the structure of the memory JSON itself.

Alternative considered: continue documenting the memory format and let the agent edit JSON directly.
Rejected because direct JSON editing is more error-prone and makes future memory-shape changes harder.

Derive the canonical entity list from provider registrations.
The Go command should import the `provider` package and inspect the registered entities exposed by both `provider/plugin_framework.go` and `provider/provider.go`. For Plugin Framework entities, it should instantiate the provider and resolve registered resource and data source type names from the provider's registrations. For Plugin SDK entities, it should read the registered `ResourcesMap` and `DataSourcesMap`. The command should union those registrations by entity type and name, preserve existing timestamps for still-registered entities, add newly discovered entities with `null`, and remove entities that are no longer registered by either provider implementation.

Alternative considered: continue scanning `docs/resources/*.md` and `docs/data-sources/*.md`.
Rejected because documentation can lag the implementation, while provider registrations are the source of truth for entities that the provider actually serves.

Alternative considered: maintain the tracked entity inventory only in the seed file.
Rejected because the canonical entity set must follow the current provider registrations on each run.

Return machine-readable selection results.
The selection command should emit structured output that includes entity names and types so the agent can iterate deterministically without reparsing memory internals.

Alternative considered: emit only plain-text names.
Rejected because the agent also needs entity type and should not infer it from naming conventions after selection.

Persist timestamps through a dedicated post-analysis command.
After each analyzed entity, the agent should call a script command that records the current UTC timestamp regardless of whether an issue was created. That preserves rotation fairness without requiring the agent to hand-edit the memory file.

Alternative considered: batch all timestamp updates at the end of the run.
Rejected because a per-entity update is safer if the run stops partway through analysis.

## Risks / Trade-offs

- Hiding the memory schema behind scripts can make debugging less obvious -> Ensure the commands produce clear machine-readable output and helpful failure messages.
- Repo-memory paths are runtime-specific -> Accept the live memory path as an argument so the scripts do not hardcode environment-specific locations.
- Partial writes could corrupt the memory file -> Use replace-on-success file writes so updates are atomic from the workflow's point of view.
- Provider-registration discovery spans both Plugin Framework and Plugin SDK implementations -> Normalize both sources into the same `resource` / `data source` inventory model, de-duplicate by fully qualified Terraform type name, and prune entries absent from both sources.
- Importing the provider package may accidentally include experimental entities -> Use the default provider constructors and environment so the command reflects the repository's normal registered entity set unless the workflow explicitly opts in to experimental registrations later.

## Migration Plan

1. Add a Go command under `scripts/schema-coverage-rotation` for memory preparation, entity selection, and timestamp persistence.
2. Implement canonical entity discovery by importing the provider registrations from the `provider` package and normalizing Plugin Framework and Plugin SDK entities into the same memory model.
3. Update `.github/workflows/schema-coverage-rotation.md` so the agent instructions tell the agent which Go command(s) to run after repo-memory hooks complete.
4. Remove the detailed memory JSON schema and hand-authored selection algorithm from the prompt.
5. Recompile `.github/workflows/schema-coverage-rotation.lock.yml` and validate the OpenSpec and workflow artifacts.
