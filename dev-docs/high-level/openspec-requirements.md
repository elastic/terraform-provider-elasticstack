# OpenSpec requirements

This repository uses [OpenSpec](https://openspec.dev/) for **living functional requirements**: specs live under [`openspec/specs/`](../../openspec/specs/) as the canonical source of truth.

## Layout

- **`openspec/specs/<capability>/spec.md`** — One capability per directory (e.g. `elasticsearch-security-role`, `ci-build-lint-test`). Each file includes:
  - **`## Purpose`** — Scope in plain language.
  - **`## Schema`** (optional) — HCL/YAML sketch for Terraform resources or workflows; appendix-style, not a substitute for code.
  - **`## Requirements`** — `### Requirement: …` blocks using **SHALL** / **MUST** (RFC 2119).
  - **`#### Scenario: …`** — Given / When / Then checks reviewers and agents can trace.

- **`openspec/changes/`** — Proposed deltas (proposal, design, tasks, delta specs). Use OpenSpec’s change workflow when a feature spans multiple specs or needs review before implementation.

- **`openspec/config.yaml`** — Project OpenSpec configuration.

## Authoring a new Terraform entity spec

1. Pick a stable capability id: `elasticsearch-<area>-<resource>` or `kibana-…` as appropriate.
2. Create `openspec/specs/<capability>/spec.md` with Purpose, Schema (if useful), and Requirements with scenarios.
3. Run **`make check-openspec`** or **`openspec validate --specs`** after **`make setup`** (installs the CLI via `npm ci`).
4. Use the **existing-entity-requirements** or **new-entity-requirements** agent skill for structure and completeness.

## CI

The **lint** job validates specs structurally with `openspec validate --specs` (telemetry disabled in CI). That checks format and normative keywords; it does **not** prove the Go implementation matches every requirement — use the **requirements-verification** skill and code review for that.

## References

- OpenSpec docs: [GitHub — OpenSpec](https://github.com/Fission-AI/OpenSpec)
- Contributor setup (Node version): [contributing.md](./contributing.md)
