# Specs

This folder contains **feature specs** intended for autonomous implementation.

## Spec index

Keep this table updated whenever you create or update a spec file.

| Spec | Status | Updated | Summary |
| --- | --- | --- | --- |
| `dashboard-spec-bump-25-2.md` | done | 2026-02-25 | Split root query into `query.text`/`query.json`; checks + dashboard acc suite green |

## Rules

- **Extremely concise**: only feature context + task list.
- **Self-contained**: include every detail an autonomous agent needs (constraints, exact outputs, file locations, CLI flags, data streams/index names, acceptance checks).
- **Task-driven**: the spec *is* the plan; implementation should follow it.
- **Continuously updated**: **after each task is completed, the agent must update the spec** (mark task done, record any decisions/changes, adjust remaining tasks).
- **Created collaboratively**: when drafting a new spec, the agent should briefly interview you, draft the spec, and iterate with your answers until it is implementation-ready.
- **Loop-based execution**: the agent runs in a loop across fresh sessions; each session should complete **exactly one task** then exit.
- **Required final tasks**: every spec MUST include:
  - A task to run all tests, checks, and formatting commands—fix any issues found
  - A final task to create a commit with all changes

## Spec format (template)

Create one file per feature: `spec/<feature>.md`

Conventions:
- Keep tasks **small** and checkable.
- **Acceptance** is "it depends": include the exact verification needed for the task.
- When relevant, explicitly list impacted output formats: `otel`, `elastic`, `fieldsense`.
- Default "done" is **builds + tests pass**, but each spec should state its own done criteria.

## Status workflow

- **draft**: actively writing/refining the spec with you
- **ready**: complete + unambiguous; ready for an agent to implement
- **in-progress**: an agent has started implementation
- **done**: completion criteria met (as defined in the spec)

## Agent loop protocol (how to execute a spec)

When a session is started with a spec as the prompt:

- **Pick work**: select the **first unchecked** task in `## Tasks`.
- **Do one task only**: implement *only* that task (including its acceptance checks).
- **Update the spec**:
  - Mark the task complete (`[x]`).
  - Update `## Additional Context` with any discoveries/gotchas that help remaining tasks.
  - Adjust remaining tasks if reality differs (split/merge/reword as needed).
  - Update `## Status`:
    - `in-progress` when the first implementation task begins
    - `done` only when the spec's "Definition of done" is met
  - Update the **Spec index** row (status/date/summary).
- **Exit**: stop after updating the spec so the next fresh session can continue.

```md
# <Feature name>

## Status
draft | ready | in-progress | done

## Context
- **Problem**: <what's wrong / missing; why it matters>
- **Worktree**: <required: worktree name for isolated execution>
- **Scope**: <what is in / out>
- **Constraints**: <perf, compatibility, deps, no-breaking-changes, etc.>
- **Repo touchpoints**: <files/dirs likely involved, commands, datasets impacted>
- **Formats impacted**: <otel|elastic|fieldsense|none>
- **Definition of done**: <e.g., builds + tests pass; plus any feature-specific checks>

## Tasks
- [ ] 1) <task> (owner: agent)
  - **Change**: <precise behavior/code change>
  - **Files**: <exact paths>
  - **Acceptance**: <how to verify; exact commands/output>
  - **Spec update**: mark done + update remaining tasks/context

- [ ] 2) ...

- [ ] N-1) Run all checks and fix issues (owner: agent)
  - **Change**: Run all tests, linting, type-checking, and formatting; fix any failures
  - **Files**: any files with issues
  - **Acceptance**: All repo checks pass (e.g., `yarn test && yarn lint && yarn typecheck`)
  - **Spec update**: mark done

- [ ] N) Create commit (owner: agent)
  - **Change**: Stage all changes and create a descriptive commit
  - **Files**: none (git operation)
  - **Acceptance**: `git status` shows clean working tree; `git log -1` shows the new commit
  - **Spec update**: mark done

## Additional Context
<append-only notes that help complete remaining tasks (discoveries, links, constraints, decisions made implicitly, gotchas)>
```
