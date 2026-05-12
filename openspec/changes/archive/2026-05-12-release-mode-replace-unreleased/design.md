## Context

The shared changelog engine (`.github/workflows-src/lib/changelog-engine-factory.js`) calls `rewriteChangelogSection` (`.github/workflows-src/lib/changelog-rewriter.js`) to mutate `CHANGELOG.md`. The rewriter has three branches today:

1. **Target heading already exists** → replace the targeted block (`targetStart..sectionEnd`) with the new section.
2. **Target heading missing AND release mode AND `## [Unreleased]` exists** → insert the new section *after* the Unreleased block (the buggy path).
3. **Target heading missing AND no Unreleased** → prepend the new section.

Release mode in practice always hits branch 2 the first time it runs against a `prep-release-X.Y.Z` branch, because the Unreleased section is the canonical buffer of pending work. The current behavior preserves both, producing duplicated content (see PR #2857).

The unreleased-mode generation re-prepends an `## [Unreleased]` heading on the next scheduled run when one is absent (already covered by `rewriter.test.mjs:46-51`), so removing the heading in release mode does not create a permanent gap — it just defers Unreleased recreation to the next scheduled changelog-generation run after the release lands.

## Goals / Non-Goals

**Goals:**

- Release mode produces a single, correct `## [X.Y.Z] - <date>` section, with no Unreleased block remaining in the file, regardless of whether the run is the first attempt or a re-run.
- Behavior is convergent: running the workflow N times against the same `prep-release-*` branch yields the same `CHANGELOG.md`.
- Unreleased-mode behavior is unchanged.
- Existing rewriter contract for "target heading already present" is preserved (idempotent replacement of the version block).

**Non-Goals:**

- Refactoring the rewriter's split-by-lines approach or adopting a proper markdown AST.
- Leaving an empty `## [Unreleased]` placeholder above the new release section (Option B from exploration). Deferred — adds complexity without clear value, and scheduled unreleased generation re-creates the heading naturally.
- Changing how the release section body is rendered (`renderChangelogSection`), how PRs are collected, or how the prep-release branch / PR is managed.
- Touching the `ci-release-pr-preparation` workflow.

## Decisions

### Decision 1: Replace, don't insert

Release mode always removes any `## [Unreleased]` block when emitting the new versioned section. Equivalent to: in release mode, the "target slot" for replacement is `## [Unreleased]` when the target version heading is absent, and continues to be the existing `## [X.Y.Z]` block when present.

**Alternative considered — Option B (replace + leave empty Unreleased stub):** Keeps a visible `## [Unreleased]` placeholder above the release section. Pros: makes the "Unreleased is now empty" state explicit during release-PR review. Cons: adds a render concern (what does an empty Unreleased look like? `## [Unreleased]\n` only? With a placeholder bullet?), and scheduled unreleased generation already re-creates the heading on the next run. Rejected as low value vs. complexity.

**Alternative considered — Leave current behavior, document workaround:** Have maintainers strip the Unreleased section manually as part of the release-PR checklist. Rejected: the workflow exists to *avoid* manual cleanup, and the bug is mechanical, not judgement-driven.

### Decision 2: Strip Unreleased on every release-mode run, even when target heading exists

The current "target heading already exists" branch only replaces the versioned block and leaves `## [Unreleased]` alone. After this change, release-mode runs also strip any Unreleased block. This makes the rewriter convergent: re-running the workflow against an already-prepped release branch (e.g., after pushing a follow-up commit to `prep-release-X.Y.Z`) cannot resurrect a stale Unreleased section.

**Alternative considered — Strip only on first run:** Leaves the door open for a re-run to disagree with the first run if someone manually re-added an Unreleased heading. Rejected: not worth defending an unusual edit pattern; convergent behavior is easier to reason about.

### Decision 3: Implementation shape

In `rewriteChangelogSection`, after computing `targetStart` (versioned heading) and `unreleasedStart`:

- In release mode, build the set of line ranges to remove:
  - `[unreleasedStart, findSectionEnd(unreleasedStart))` if `unreleasedStart !== -1`.
  - `[targetStart, findSectionEnd(targetStart))` if `targetStart !== -1`.
- Replace the *earliest* range with `newSectionContent` and drop the other range(s).
- If neither exists, prepend (current branch 3, unchanged).

Keeping it line-based avoids introducing a markdown parser dependency.

### Decision 4: Tests assert new behavior, not legacy

The three existing tests that pin the buggy "both headings present" outcome are inverted to assert that only `## [X.Y.Z]` is present after release-mode rewrite. New tests cover:

- Re-run case: file already contains both `## [Unreleased]` and `## [X.Y.Z]` — both are collapsed into one `## [X.Y.Z]`.
- No-Unreleased case (already covered by prepend path) — keep as-is.

## Risks / Trade-offs

- **Risk**: A maintainer hand-edits `prep-release-X.Y.Z` to inject curated Unreleased content not driven by PR bodies. → **Mitigation**: This is already incompatible with the existing "full-section regeneration from authoritative range" spec; the rewriter regenerates the section from PR records every run. No regression in flexibility.
- **Risk**: Downstream tooling parses `CHANGELOG.md` and expects `## [Unreleased]` to persist alongside released versions. → **Mitigation**: No such tooling exists in this repo; release-notes generation reads the rendered `## [X.Y.Z]` block, and the unreleased PR workflow re-creates the heading on its next run.
- **Trade-off**: Between release-cut and the next scheduled unreleased generation, `CHANGELOG.md` on `main` has no `## [Unreleased]` heading. Acceptable: the section would be empty anyway immediately after a release.

## Migration Plan

None required. The change ships with the next merged commit to `main`. The first release after this lands produces a clean release section; no historical changelog rewriting needed. Rollback is a single revert of the rewriter + tests + regenerated workflow YAML.

## Open Questions

None.
