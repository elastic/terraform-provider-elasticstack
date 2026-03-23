# OpenSpec requirements

This repository uses [OpenSpec](https://openspec.dev/) for **living functional requirements**: specs live under [`openspec/specs/`](../../openspec/specs/) as the canonical source of truth.

## Layout

- **`openspec/specs/<capability>/spec.md`** — One capability per directory (e.g. `elasticsearch-security-role`, `ci-build-lint-test`). Each file includes:
  - **`## Purpose`** — Scope in plain language.
  - **`## Schema`** (optional) — HCL/YAML sketch for Terraform resources or workflows; appendix-style, not a substitute for code.
  - **`## Requirements`** — `### Requirement: …` blocks using **SHALL** / **MUST** (RFC 2119).
  - **`#### Scenario: …`** — Given / When / Then checks reviewers and agents can trace.

- **`openspec/changes/<name>/`** — OpenSpec **change** directories: **proposal**, **design**, **tasks**, and **delta specs** (the proposed deltas until they land in `openspec/specs/`). **Use this workflow for all changes** that add or update requirements—features, fixes, new Terraform resources or data sources, CI behavior, documentation of behavior, and so on. Create a change with **`openspec new change`**, implement from it, then **sync** delta specs into `openspec/specs/` or **archive** the change when done (**openspec-propose**, **openspec-sync-specs**, **openspec-archive-change** skills).

- **`openspec/config.yaml`** — Project OpenSpec configuration.

## Authoring requirements (changes)

**Default path** — Same as above: **`openspec new change "<name>"`**, then fill **proposal**, **design**, **tasks**, and **delta specs** (see **openspec-propose**). After **`make setup`**, validate with **`openspec validate --all`** (canonical specs plus active changes under `openspec/changes/`); **`make check-openspec`** runs **`openspec validate --specs`** and is what CI uses once deltas are synced into `openspec/specs/`.

### New Terraform resource or data source

1. Pick a stable capability id: `elasticsearch-<area>-<resource>` or `kibana-…` as appropriate (this names the delta spec directory under the change).
2. Use the **new-entity-requirements** skill for research (API clients, Elastic docs, user interview) and what belongs in each artifact.
3. After implementation, **sync** or **archive** (**openspec-sync-specs**, **openspec-archive-change**).

### Documenting an existing entity from code

Use the **existing-entity-requirements** skill for structure and completeness when capturing behavior in delta specs (still via a change).

### Editing `openspec/specs/` directly

Avoid for new work; prefer a change and sync. Direct edits are only for **tiny** follow-ups (e.g. typo, link) or when the canonical tree already reflects merged work and you are not starting a new change.

## CI

The **lint** job validates specs structurally with `openspec validate --specs` (telemetry disabled in CI). That checks format and normative keywords; it does **not** prove the Go implementation matches every requirement — use the **requirements-verification** skill and code review for that.

## References

- OpenSpec docs: [GitHub — OpenSpec](https://github.com/Fission-AI/OpenSpec)
- Contributor setup (Node version): [contributing.md](./contributing.md)
