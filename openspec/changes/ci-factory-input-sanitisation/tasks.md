## 1. Extend sanitisation in shared library

- [ ] 1.1 Add `stripControlChars` and `stripInvisibleUnicode` helper functions to `.github/workflows-src/lib/sanitize-context.js`
- [ ] 1.2 Add `sanitizeUserContent` composed function that runs all three filters in order
- [ ] 1.3 Export the new functions alongside existing `stripHtmlComments`

## 2. Update tests

- [ ] 2.1 Add test coverage for `stripControlChars` and `stripInvisibleUnicode` in `.github/workflows-src/lib/sanitize-context.test.mjs`
- [ ] 2.2 Add test coverage for `sanitizeUserContent` composed behaviour (sequential filtering, idempotency, non-string inputs)

## 3. Update call sites

- [ ] 3.1 Update `.github/workflows-src/research-factory-issue/scripts/write_context_files.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`
- [ ] 3.2 Update `.github/workflows-src/change-factory-issue/scripts/extract_research_comment.inline.js` and the downstream handling of `prior_research_comment` — add explicit sanitisation for the prior research comment
- [ ] 3.3 Update `.github/workflows-src/change-factory-issue/scripts/sanitize_context.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`
- [ ] 3.4 Update `.github/workflows-src/code-factory-issue/scripts/sanitize_context.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`

## 4. Verify and rebuild locked workflows

- [ ] 4.1 Run the project's workflow generation tooling to regenerate locked workflow YAML from source templates
- [ ] 4.2 Verify the generated workflows reference the updated inline scripts correctly