## Context

`CHANGELOG.md` has two distinct zones:

```
## [Unreleased]          ← section body (managed by rewriteChangelogSection)
...

## [0.15.0] - 2026-05-14
...

─── link table (bottom of file) ───────────────────────────────
[Unreleased]: .../compare/v0.15.0...HEAD
[0.15.0]: .../compare/v0.14.5...v0.15.0
[0.14.5]: .../compare/v0.14.4...v0.14.5
...
```

`rewriteChangelogSection` handles only the section body. The link table has never been touched by the generator. In release mode the table requires two changes:

1. Update `[Unreleased]:` to point at `vNEW...HEAD`
2. Insert `[NEW]: .../compare/vOLD...vNEW` immediately after it

`previousTag` (needed for the compare URL) is resolved by `resolveChangelogCompareContext` but is never passed to `runChangelogRenderAndWrite`, which is the only place where `CHANGELOG.md` is mutated.

The inline GitHub Actions script (`run-changelog-engine.inline.js`) calls `runChangelogRenderAndWrite` directly and has the same gap — `ctx.previousTag` is in scope but not forwarded.

## Goals / Non-Goals

**Goals:**
- Update the link table automatically in release mode with no manual fixup required.
- Keep the function backward-compatible (unreleased callers pass nothing; new param defaults to `''`).
- Make `rewriteLinkTable` idempotent (re-run safe).

**Non-Goals:**
- Updating the link table in unreleased mode (the `[Unreleased]:` line does not change there).
- Rebuilding the entire link table from scratch or re-ordering existing entries.
- Handling first-ever release with no prior tag (guard makes it a no-op; link table entry construction is not possible without `previousTag`).

## Decisions

### Decision: Separate `rewriteLinkTable` function, same file as `rewriteChangelogSection`

**Chosen:** Add `rewriteLinkTable(content, targetVersion, previousTag)` to `changelog-rewriter.js`, export it, and call it from `runChangelogRenderAndWrite` after `rewriteChangelogSection`.

**Alternatives considered:**

- *Extend `rewriteChangelogSection` to handle the link table too.* Rejected — the link table is structurally separate from section headers/bodies; blending both into one function would widen its responsibility significantly and complicate the already-tested section-rewrite logic.
- *Handle link table in `runChangelogRenderAndWrite` inline.* Rejected — string manipulation logic belongs in `changelog-rewriter.js` where it can be unit-tested in isolation alongside the existing rewriter tests.

### Decision: Extract base URL from the existing `[Unreleased]:` line

The compare URL base (e.g. `https://github.com/elastic/terraform-provider-elasticstack/compare/`) is not passed as a parameter — it is parsed from the existing `[Unreleased]:` line in the content. This avoids introducing a new parameter that callers must supply.

Guard: if no `[Unreleased]:` line is found, `rewriteLinkTable` returns the content unchanged.

### Decision: Idempotency via presence check

Before inserting `[NEW]:`, check whether a line starting with `[targetVersion]:` already exists in the content. If it does, skip insertion. This makes re-runs safe without requiring callers to track whether the table was already updated.

### Decision: `previousTag` defaults to `''`, no-op guard inside the function

```js
function rewriteLinkTable(content, targetVersion, previousTag) {
  if (!targetVersion || !previousTag) return content;
  ...
}
```

This means:
- Unreleased callers (`targetVersion = ''`) get a no-op with no call-site changes needed.
- Release callers with no prior tag (`previousTag = ''`) also get a no-op — the link table cannot be correctly populated without a prior tag anyway.

## Risks / Trade-offs

- **URL parsing fragility** → The regex for extracting the base URL from `[Unreleased]:` is straightforward (`/^\[Unreleased\]:\s*(https?:\/\/.+\/compare\/).*/`). If the existing entry is malformed, the guard (no match → no-op) prevents a corrupt write. Covered by unit tests.
- **Re-run with already-correct table** → Idempotency check prevents duplicate `[NEW]:` entries. The `[Unreleased]:` line is overwritten on each re-run, which is always safe since it always points to the same target.
