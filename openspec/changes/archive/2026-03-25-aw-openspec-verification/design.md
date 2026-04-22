## Context

The repository uses [OpenSpec](https://openspec.dev/) with **active** changes under `openspec/changes/<id>/` and completed work under `openspec/changes/archive/YYYY-MM-DD-<id>/`. [GitHub Agentic Workflows](https://github.github.com/gh-aw/) provide **safe outputs** for PR reviews and for **`push-to-pull-request-branch`**.

## Goals / Non-Goals

**Goals:**

- **Opt-in** via label **`verify-openspec`** so normal PRs are unaffected.
- Deterministic **pre-review** gating from the PR files API: exactly one active change id, **modified-only** updates under that tree (no new files under `openspec/changes/<id>/`).
- After a bot **APPROVE**, **archive** the selected change and **push** commits back to the PR branch so the human can merge a consistent tree.

**Non-Goals:**

- Auto-running on every archive-touching PR.
- Archiving when the review is **COMMENT** or when verification fails gates.
- Using `REQUEST_CHANGES` on the review.

## Decisions

1. **Trigger** — `pull_request` with type **`labeled`**, and the agent or a job condition SHALL ensure the label name is exactly **`verify-openspec`** (avoid running on unrelated labels).

2. **Change selection** — Parse each PR file entry with paths under `openspec/changes/` **excluding** `openspec/changes/archive/**`. Derive `<id>` as the first path segment after `openspec/changes/`. Consider only entries whose GitHub status is **`modified`** for counting “which change was updated.” **Noop** if:
   - any file under `openspec/changes/` (non-archive) has status **`added`** (or otherwise indicates a new change or new artifact file),
   - more than one distinct `<id>` has at least one **`modified`** file,
   - zero distinct `<id>` has **`modified`** files (e.g. label applied but diff unrelated to active changes).

   **Optional strictness:** treat **`removed`**, **`renamed`**, or other non-`modified` statuses under `openspec/changes/<id>/` as **noop** so the workflow only runs for “edit existing proposal” PRs.

3. **Verification** — Use change id `<id>` with **openspec-verify-change** and OpenSpec CLI as in the skill (list/status/instructions as appropriate).

4. **Structural allowlist** — Same idea as before but rooted at **`openspec/changes/<id>/`** plus paired `openspec/specs/<capability>/spec.md` for each delta under `openspec/changes/<id>/specs/`, plus post-archive expectations if the archive step runs in the same run (implementation may stage archive output before push).

5. **Post-APPROVE archive + push** — Only after **`submit_pull_request_review`** with **`APPROVE`**:
   - Run **`openspec archive <id>`** (or equivalent documented repository procedure aligned with **openspec-archive-change**) so the change moves under `openspec/changes/archive/` and canonical specs update per project policy.
   - Commit results and use **`push-to-pull-request-branch`** safe output to update the PR branch. Configure frontmatter per gh-aw docs (`target: triggering`, `title-prefix` / `labels` if policy requires, `fetch: ["*"]` when using wildcard targets).

6. **Permissions** — **`contents: read`** is insufficient for push; compiled workflow SHALL grant **`contents: write`** (or scoped equivalent) for the push safe-output job path.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Label applied on wrong PR | Human discipline; noop rules reduce damage. |
| Archive + push fails mid-flight | Document rollback; workflow may open issue or comment on failure per future hardening. |
| `GITHUB_TOKEN` cannot push to PR branch | Use PAT in `github-token` for safe outputs; see [Triggering CI](https://github.github.io/gh-aw/reference/triggering-ci/) if follow-up CI must run. |
| Protected files in archive path | Use gh-aw **protected-files** policy on `push-to-pull-request-branch` as appropriate. |

## Migration Plan

- Add workflow + compile; document label and permissions for maintainers.
- **Rollback**: remove workflow or disable label in branch protection / automation docs.

## Open Questions

- Whether to auto-remove label **`verify-openspec`** after a run (not required for v1).
- Exact **`openspec archive`** flags vs interactive skill steps when running headless in Actions.
