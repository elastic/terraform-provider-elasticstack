## 1. Parser: end marker support

- [ ] 1.1 In `extractBreakingChanges` in `pr-changelog-parser.js`, add a check inside the `inBreaking` loop: when `fenceType === null` and the line matches `/^\s*<!--\s*\/breaking-changes\s*-->\s*$/`, break out of the loop (before the heading check). The full-line anchor prevents partial matches; `\s*` before `<!--` and after `-->` allows indentation and trailing whitespace; `\s*` around `/breaking-changes` allows internal spacing.
- [ ] 1.2 Add tests to `pr-changelog-parser.test.mjs` for `extractBreakingChanges` end marker behaviour:
  - End marker stops extraction mid-content (content before marker included, content after excluded)
  - End marker with extra internal whitespace (`<!--  /breaking-changes  -->`) is recognised
  - End marker inside a backtick-fenced code block is NOT treated as a stop
  - End marker inside a tilde-fenced code block is NOT treated as a stop
  - End marker before the `### Breaking changes` heading (outside `inBreaking`) is ignored

## 2. Validator: breaking changes require breaking impact

- [ ] 2.1 In `validateChangelogSectionFull` in `pr-changelog-parser.js`, add rule C after the existing two rules: if `parsed.breakingChangesHeadingPresent` is true, `parsed.customerImpact` is non-null, `VALID_CUSTOMER_IMPACTS.has(parsed.customerImpact)` is true, and `parsed.customerImpact !== 'breaking'`, push the error `"### Breaking changes section requires Customer impact: breaking; use <!-- /breaking-changes --> as an end marker."`
- [ ] 2.2 Add tests to `pr-changelog-parser.test.mjs` for rule C:
  - `Customer impact: fix` + `### Breaking changes` present → invalid, correct error message
  - `Customer impact: enhancement` + `### Breaking changes` present → invalid, correct error message
  - `Customer impact: none` + `### Breaking changes` present → invalid, correct error message
  - `Customer impact: patch` (unsupported value) + `### Breaking changes` present → only the "invalid impact" error, NOT the rule-C error
  - `Customer impact: breaking` + `### Breaking changes` present + content → still valid (no regression)

## 3. Verifier comment: document end marker

- [ ] 3.1 In `buildFailureCommentBody` in `pr-changelog-check.js`, add `'<!-- /breaking-changes -->'` as a line in the "Expected format" block immediately after `'<free-form markdown>  (required when Customer impact is "breaking")'`

## 4. PR template: invalid default and breaking example

- [ ] 4.1 In `.github/pull_request_template.md`, replace the pre-filled default `Customer impact: none\nSummary:` with `Customer impact: <none, fix, enhancement, breaking>\nSummary: <single line summary>`
- [ ] 4.2 In `.github/pull_request_template.md`, add a second "Good example" block for the `breaking` case showing `Customer impact: breaking`, `Summary:`, `### Breaking changes`, a single-line description, and `<!-- /breaking-changes -->`
- [ ] 4.3 In `.github/pull_request_template.md`, add a bullet to the format instructions noting that `<!-- /breaking-changes -->` optionally ends the `### Breaking changes` block early (prevents trailing PR content from entering the changelog)

## 5. Verify

- [ ] 5.1 Run `npm test` (or equivalent) in `.github/workflows-src/` and confirm all existing and new tests pass
