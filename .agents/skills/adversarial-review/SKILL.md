---
name: adversarial-review
description: Adversarial code review focused on style/idiom, test depth, UX, and risk for Terraform provider changes. Complements openspec-verify-change by raising concerns CI and spec verification do not catch. Use when reviewing a PR or local change in addition to OpenSpec verification.
license: MIT
compatibility: Requires read access to the repository, the PR diff, and the relevant openspec/changes/<id>/ artifacts (when present).
metadata:
  author: openspec
  version: "1.0"
---

Adversarially review the implementation of a change against four axes that the `openspec-verify-change` skill and CI do **not** cover. You are *expected* to find issues — but you must not invent CRITICAL findings to justify the run. When uncertain, prefer SUGGESTION over WARNING and WARNING over CRITICAL.

**Input**: a change id (`<id>`) under `openspec/changes/<id>/` and/or the PR diff. If both are available, use both. If only one is available, proceed with what you have and note the limitation.

## Out of scope

Do **not** raise findings for any of:

- Compilation, `gofmt`, `go vet`, `make lint`, or anything else CI catches.
- Running tests yourself — inspect existing tests and CI results, do not execute.
- Task completion, requirement-implementation mapping, design adherence, and structural relevance — these are owned by `openspec-verify-change`. (Test coverage findings are explicitly **in scope** here even when the verify reviewer also flags them; the orchestrator dedupes.)
- Broad "pattern consistency" findings that only restate design adherence or structural relevance. Raise consistency issues here only when they create a concrete Go, Terraform Plugin Framework, UX, testing, or risk concern.

## Review axes

Run all four axes. For each axis, if you find nothing, say so explicitly in the report; do not pad.

### 1. Style / idiom (Go + Terraform Plugin Framework)

Reference: [`dev-docs/high-level/coding-standards.md`](dev-docs/high-level/coding-standards.md).

Look for:

- SDK-vs-Plugin-Framework mixing within a single resource or package boundary.
- Diagnostic handling: missing `resp.Diagnostics.Append(...)` after `Get`/`Set`/`GetAttribute`, swallowed diagnostics, `return` after appending an error then continuing to use partial state.
- Error wrapping: `fmt.Errorf("...: %w", err)` for internal errors; `diag.NewErrorDiagnostic` for user-visible PF diagnostics; not mixing `errors.New`/`fmt.Errorf` where a PF diagnostic is expected.
- Context propagation: `ctx` threaded through API client calls; no `context.Background()` inside Read/Create/Update/Delete.
- State hygiene: `Read` does not write state on a `404` (resource removal pattern); `Update` re-reads computed values; partial state cleanup on error paths.
- Plan modifiers and validators present where they are needed (immutable attributes, normalization).
- Consistency with neighboring resources in the same package.

### 2. Testing

Cover both **breadth** (are obvious behaviors tested at all?) and **depth** (how good are the existing tests?). Some overlap with the verify reviewer's scenario-coverage check is expected and fine — the orchestrator dedupes findings. Prefer raising a finding over staying silent when an obvious gap is visible.

**Breadth — obvious testing gaps**:

- New resource or data source with no acceptance test file, or only a single happy-path step.
- New schema attribute touched by this change with no test that configures or asserts it.
- New error path or validation rule with no test exercising it.
- New code branch (e.g. version-gated behavior, fallback paths) with no test covering each branch.
- Behavior described in proposal/design as user-visible but not exercised by any test.

**Depth — heuristics borrowed from [`.agents/skills/schema-coverage/SKILL.md`](.agents/skills/schema-coverage/SKILL.md)**:

- **Set-only assertions** (`TestCheckResourceAttrSet`) where a value-specific assertion (`TestCheckResourceAttr`) is feasible.
- **Single-value coverage**: an attribute is only ever exercised at one value across all steps.
- **Missing unset/empty cases**: optional attribute always set; collection attribute never tested empty or omitted.
- **Missing update coverage**: attribute is touched but no multi-step test asserts the post-update value.
- **Missing import test** for new resources.
- **Drift on Read**: API returns more fields than the resource sets back into state; no test asserts the round-trip.

Before raising a "missing test" finding, confirm against the package's `*_test.go` files (and any nearby fixture/testdata files) that the test really is absent — do not flag missing tests just because you couldn't locate them.

### 3. UX

Look for things a user would hit:

- Attribute names: idiomatic Terraform snake_case; consistent with sibling resources; no leaking of API-internal names.
- Defaults: explicit defaults where the API has one; `Computed: true` where the server fills the value.
- `RequiresReplace` (or PF `RequiresReplace` plan modifier) on attributes that the API cannot mutate in place.
- Deprecation: deprecated attributes carry `DeprecationMessage` / PF `DeprecationMessage`.
- Error messages users will see: actionable, not just the raw API payload; mention the attribute or resource by name.
- Documentation: schema description / markdown description matches observable behavior, not aspirational behavior; examples in `examples/` updated when the schema changed.
- Breaking changes to existing resources without a migration path.

### 4. Risk

Look for:

- Silent state migrations or schema-version bumps without an upgrader.
- Sensitive value handling: secrets/tokens marked `Sensitive: true`; not logged via `tflog` at info level.
- Version-gated behavior that does not use a server-version check pattern in this repo.
- Regressions in unrelated resources caught by reading the diff (e.g. a shared helper changed signature).

## Output contract

Return a single markdown report with this exact shape so the orchestrator can mechanically merge it with the verify reviewer's output:

```markdown
## Adversarial review

### Scorecard
| Axis        | Status                |
|-------------|-----------------------|
| Style       | <Clean / N issue(s)>  |
| Testing     | <Clean / N issue(s)>  |
| UX          | <Clean / N issue(s)>  |
| Risk        | <Clean / N issue(s)>  |

### Issues by priority

#### CRITICAL
- **[axis]** <file>:<line?> — <finding>
  - Severity: CRITICAL
  - Axis: <style | testing | ux | risk>
  - File: <path>
  - Line: <line or omitted>
  - Evidence: <code reference or diff hunk>
  - Recommended fix: <specific, actionable>

#### WARNING
- **[axis]** <file>:<line?> — <finding>
  - Severity: WARNING
  - Axis: <style | testing | ux | risk>
  - File: <path>
  - Line: <line or omitted>
  - Evidence: <code reference or diff hunk>
  - Recommended fix: <specific, actionable>

#### SUGGESTION
- **[axis]** <file>:<line?> — <finding>
  - Severity: SUGGESTION
  - Axis: <style | testing | ux | risk>
  - File: <path>
  - Line: <line or omitted>
  - Evidence: <code reference or diff hunk>
  - Recommended fix: <specific, actionable>
```

Each issue must carry: `severity`, `axis` (one of `style`, `testing`, `ux`, `risk`), `file`, optional `line`, `finding`, `evidence`, `recommended_fix`. Issues without a concrete `file` are SUGGESTIONS at most.

## Severity guardrails

- **CRITICAL** is reserved for: user-visible breakage, silent data loss, secrets handled unsafely, breaking changes to existing public schema without migration, drift the resource will never reconcile.
- **WARNING** for: likely-wrong but not certainly broken; missing test depth on an attribute that has *some* coverage; UX papercuts a user will hit but can work around.
- **SUGGESTION** for: idiom polish, naming, doc improvements, optional follow-ups.

If your evidence is purely stylistic (formatting, naming preference, comment wording), the finding is at most a SUGGESTION. Do not upgrade because nothing else turned up.

If an axis is clean, write `Clean` in the scorecard and omit it from the issues list rather than synthesizing filler.

## Final assessment

End the report with one line:

- `No critical issues.` if zero CRITICAL findings, or
- `<N> critical issue(s) found.` otherwise.

Do not state an approve/comment recommendation — that decision is the orchestrator's, based on combined findings across reviewers.
