## Context

The Copilot setup workflow mirrors acceptance-test-oriented bootstrap (Elastic Stack, `make setup`, Fleet, API key). Repository `package.json` declares `engines.node` as a semver range (and may add Volta or `devEngines.runtime`); those fields are the contract for which Node versions are supported. Copilot sessions that run OpenSpec or `npm ci` need a matching runtime in the setup job; today only Go and Terraform are installed explicitly.

## Goals / Non-Goals

**Goals:**

- Install Node in `copilot-setup-steps` using `actions/setup-node` pinned by commit SHA with **`node-version-file: package.json`**, so the version is resolved from `package.json` per the action’s documented precedence (`volta.node`, then `devEngines.runtime` for node, then `engines.node`). Enable npm caching with `package-lock.json` as the cache dependency path.
- Document the behavior in the `ci-copilot-setup-steps` capability so OpenSpec validation and reviewers treat Node as part of the required toolchain, tied to `package.json` rather than a duplicated literal in the workflow.

**Non-Goals:**

- Changing Elastic, Go, or Terraform versions or step ordering beyond what is needed to place Node sensibly (e.g. before `make setup`, which may invoke OpenSpec).
- Pinning an exact Node patch in YAML; the action resolves a concrete release that satisfies the semver spec read from `package.json`.

## Decisions

1. **`node-version-file: package.json`** — Single source of truth with `engines` (and optional Volta / `devEngines.runtime`). Avoids hardcoding a major in the Copilot workflow when `engines` changes. **Alternative considered:** `node-version: "<major>"` to mirror another workflow — duplicates policy and drifts when `engines` moves.
2. **Omit `node-version` when using `node-version-file`** — The action uses `node-version` over the file if both are set; the workflow must not override the file accidentally.
3. **Enable npm cache with `cache-dependency-path: package-lock.json`** — Speeds repeated Copilot setup runs. **Alternative considered:** no cache — slower, no benefit.
4. **Insert the Node step after checkout and before Go/Terraform / stack steps** — Node must be available before any step that needs `node`/`npm` (e.g. `make setup` → `setup-openspec`). Order: checkout → setup-node → setup-go → setup-terraform matches a sensible toolchain grouping.

## Risks / Trade-offs

- **[Risk] Another workflow (e.g. lint) still hardcodes a major while Copilot uses the file** → **Mitigation:** Prefer converging other jobs on `node-version-file: package.json` in follow-up work; until then, both should satisfy the same `engines` policy.
- **[Risk] Adding `volta.node` or `devEngines.runtime` changes what CI installs without editing the workflow** → **Mitigation:** That is intentional DRY; document precedence in specs so reviewers know the effective source.
- **[Risk] Invalid or missing Node spec in `package.json`** → **Mitigation:** CI fails at setup-node; fix `package.json` or restore `engines.node`.

## Migration Plan

- Land workflow + spec delta together (or spec first per team preference). No runtime migration for existing users; Copilot gains a correct Node path on the next setup run.

## Open Questions

- None for the initial scope.
