## Context

The repository already tracks Kibana OpenAPI updates by bumping `generated/kbapi/Makefile` through Renovate and regenerating `generated/kbapi/kibana.gen.go`. What is missing is a repeatable mechanism that turns those upstream API changes into actionable provider follow-up.

This change introduces a hybrid agentic workflow. Deterministic repository-local code handles the parts that need to be testable and stable: trigger gating, baseline selection, generated client diffing, entity inventory, reverse indexing, impact matching, and duplicate suppression. The agent only receives structured impact evidence plus limited repository context and is responsible for turning that evidence into maintainable issue summaries.

The design needs to fit existing repository conventions:

- workflow sources are authored under `.github/workflows-src/` and compiled into checked-in workflow artifacts;
- repo memory is used to persist workflow state across runs;
- provider entity discovery should come from the registered Plugin Framework and SDK providers rather than from handwritten lists;
- Kibana coverage is strongest for entities that use `generated/kbapi` through `internal/clients/kibanaoapi`, while legacy `go-kibana-rest` and `generated/slo` consumers remain follow-up scope.

## Goals / Non-Goals

**Goals:**

- Detect Kibana spec or generated client changes from a stable stored baseline rather than from only the immediately previous commit.
- Determine impacted Kibana Terraform entities using deterministic matching with inspectable evidence.
- Open one issue per impacted entity with a concise summary of likely new fields, widened options, or new capabilities.
- Avoid repeatedly creating equivalent issues for the same baseline-to-target transition.
- Reuse existing repository patterns for gh-aw workflows, repo memory, and script testing.

**Non-Goals:**

- Fully model every Kibana-facing client in the first version.
- Guarantee semantic understanding of all upstream behavioral changes when no tracked symbol changes.
- Produce implementation plans inside the created issue; a later agent workflow will handle that.
- Replace Renovate or change how `generated/kbapi` is regenerated.

## Decisions

### Use a hybrid workflow instead of a fully agent-driven classifier

The workflow will separate evidence collection from summarization. Deterministic helpers under `scripts/kibana-spec-impact/` will compute the baseline, diff tracked artifacts, build the entity reverse index, and emit a JSON impact report. The gh-aw workflow will then instruct the agent to read that report and create issues only for supported high-confidence matches.

Alternative considered: let the agent inspect raw diffs and infer impacted entities directly. This would be faster to implement but harder to test, harder to debug, and more likely to drift in issue quality over time.

### Use provider registrations as the canonical entity inventory

Entity discovery will be derived from `provider/plugin_framework.go` and `provider/provider.go`, following the same pattern already used by the schema-coverage rotation helper. This prevents the workflow from relying on manual allowlists that can go stale when resources or data sources are added or removed.

Alternative considered: maintain a handwritten manifest of Kibana entities. This was rejected because it duplicates provider truth and adds maintenance burden.

### Persist processed state in repo memory

The workflow will maintain repo memory that stores at least:

- the last processed baseline revision,
- fingerprints for previously reported entity impacts,
- optional metadata needed to suppress duplicates or resume after partial runs.

This keeps duplicate suppression deterministic and aligned with existing repository automation patterns.

Alternative considered: rely only on open-issue title matching or issue search. This was rejected because issue text is an unstable dedupe key and cannot reliably represent baseline-to-target equivalence.

### Limit V1 deterministic matching to `generated/kbapi` consumers

The first implementation will only claim high-confidence impact for entities that can be matched through `generated/kbapi` and `internal/clients/kibanaoapi` symbol usage. The workflow may surface weaker evidence for `transform_schema.go` changes, but those should remain agent-reviewed and conservative in V1.

Alternative considered: include `generated/slo` and `go-kibana-rest` from day one. This was rejected for V1 because those surfaces use different client-generation and call patterns, increasing complexity before the base approach is proven.

### Define a strict evidence contract between helper and agent

The helper output will include baseline and target revisions, changed symbols grouped by kind, impacted entities, matched implementation paths, and a confidence level per entity. The agent prompt will instruct the agent not to create issues without deterministic evidence unless the workflow explicitly opts into lower-confidence handling in a later version.

Alternative considered: pass arbitrary helper logs to the agent. This was rejected because it makes the agent prompt brittle and weakens testability.

## Risks / Trade-offs

- [Shared generated types create broad match sets] → Restrict V1 issue creation to high-confidence entity-local matches and let the agent suppress weak generic matches.
- [Transform-only spec changes may not map cleanly to symbols] → Include transform file changes in the evidence report and document them as lower-confidence cases rather than silently ignoring them.
- [Repo memory can suppress needed follow-up if fingerprints are too coarse] → Use baseline-to-target revision data plus changed-symbol fingerprints rather than title-only or entity-only dedupe.
- [Push-only triggering can miss nonstandard update paths] → Include `workflow_dispatch` and a scheduled safety-net run.
- [Deterministic matching misses legacy Kibana clients] → Explicitly scope V1 to `generated/kbapi` consumers and leave extension points for `generated/slo` and `go-kibana-rest`.

## Migration Plan

1. Add the new OpenSpec capability and implementation tasks.
2. Implement the helper command and its unit tests.
3. Add repo-memory seed state for the new workflow.
4. Author the gh-aw workflow source and compile the checked-in workflow artifacts.
5. Validate the helper logic and workflow generation locally.
6. Roll out with conservative issue caps and high-confidence issue creation only.

Rollback is straightforward: disable or revert the workflow source and generated artifacts, and remove or ignore the corresponding repo-memory branch or seed file.

## Resolved decisions (implementation)

- **Initial baseline when repo memory is empty**: resolve to `git rev-parse <target>~1` so the first run compares the parent commit to the current target; after each successful workflow completion, `last_analyzed_target_sha` advances to the analyzed `target_sha`.
- **Schedule frequency**: weekly on Monday (`cron: weekly on monday`) with push-based triggers on `main` for kbapi/kibanaoapi paths; maintainers can use `workflow_dispatch`.
- **Transform-only changes (V1)**: surface paths under `transform_schema_hints` in the report and keep the gate active for agent review, but **do not** open issues from transform hints alone; issues are only for `high_confidence_impacts` from kbapi/kibanaoapi matching.
- **Issue cap vs dedupe**: the workflow caps new issues per run; repo memory records dedupe fingerprints **only** for entities that actually received an issue (`--issued` list). The analysis baseline always advances after a successful run so stale baselines do not block future diffs.
