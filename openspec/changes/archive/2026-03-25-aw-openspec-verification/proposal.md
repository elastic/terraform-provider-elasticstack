## Why

Maintainers want an **opt-in** automation that verifies an **active** OpenSpec change on a pull request (aligned with `.agents/skills/openspec-verify-change/SKILL.md` and normal `openspec` CLI usage), correlates the rest of the PR diff to that proposal, and—when the bot **APPROVE**s—**archives** the change and pushes the result back to the PR branch. Path-based triggers on `openspec/changes/archive/**` are the wrong default: they fire at the wrong lifecycle stage and prevent using `openspec status --change` / `openspec instructions apply --change` against the active change directory.

## What Changes

- Add a **GitHub Agentic Workflow** (markdown + compiled `.lock.yml`) triggered when label **`verify-openspec`** is applied to a pull request (`pull_request` **`labeled`**).
- The workflow SHALL inspect the PR file list and **select at most one** active change id under `openspec/changes/<id>/` (**excluding** `openspec/changes/archive/**`) using strict gating: **noop** if more than one change id has **modified** files, **noop** if any file under active change paths is **added** (new proposal / new change tree), **noop** if there is not exactly one change id with qualifying **updates**—so a review runs only when **exactly one** active proposal tree has been **updated** (modified-only in that sense).
- The agent SHALL run verification using the **standard OpenSpec change id** and skills/CLI (`openspec status --change`, `openspec instructions apply --change`, and the **openspec-verify-change** skill) against `openspec/changes/<id>/`.
- All prior rules for **relevance review** (structural allowlist + `relevant` / `uncertain` / `unassociated`), **PR review body**, **line comments**, and **APPROVE** vs **COMMENT** (no **REQUEST_CHANGES**) remain; the structural allowlist is anchored on the **selected** `openspec/changes/<id>/` and paired canonical specs, not on an archive folder.
- When the submitted review event is **APPROVE**, the workflow SHALL **archive** the targeted change (following repository OpenSpec archive practice—`openspec archive` CLI and/or **openspec-archive-change** skill expectations), commit the outcome, and push to the PR branch via the **`push-to-pull-request-branch`** safe output.
- Document behavior in new capability **`ci-aw-openspec-verification`**.

## Capabilities

### New Capabilities

- `ci-aw-openspec-verification`: Requirements for the label-gated GitHub Agentic Workflow: triggers, change selection / noop rules, verification with active OpenSpec tooling, PR review outputs, post-APPROVE archive and push.

### Modified Capabilities

- _(none)_

## Impact

- **Automation**: New `.github/workflows/*.md` + `.lock.yml`; `safe-outputs` include `push-to-pull-request-branch` and review APIs; `permissions` need **`contents: write`** (or equivalent) for pushes in addition to **`pull-requests: write`**.
- **Secrets / settings**: Org/repo may need to allow Actions to **create/approve** pull requests; pushes from Actions may require a PAT or permissive `GITHUB_TOKEN` per GitHub docs; **`push-to-pull-request-branch`** may require `checkout` / `fetch` configuration (e.g. wildcard fetch) per GitHub Agentic Workflows reference.
- **Human process**: Maintainers apply **`verify-openspec`** when they intend verify-and-archive; PRs that add a new change or touch multiple active proposals are excluded by design (noop).
- **Specs**: Canonical spec delivered via this change’s delta at `openspec/specs/ci-aw-openspec-verification/spec.md` when synced/archived.
