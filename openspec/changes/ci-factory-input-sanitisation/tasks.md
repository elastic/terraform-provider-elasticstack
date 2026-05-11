## 1. Extend sanitisation in shared library

- [ ] 1.1 Add `stripControlChars` and `stripInvisibleUnicode` helper functions to `lib/sanitize-context.js`
- [ ] 1.2 Add `sanitizeUserContent` composed function that runs all three filters in order
- [ ] 1.3 Export the new functions alongside existing `stripHtmlComments`

## 2. Update tests

- [ ] 2.1 Add test coverage for `stripControlChars` and `stripInvisibleUnicode` in `sanitize-context.test.mjs`
- [ ] 2.2 Add test coverage for `sanitizeUserContent` composed behaviour (sequential filtering, idempotency, non-string inputs)

## 3. Update call sites

- [ ] 3.1 Update `research-factory/scripts/write_context_files.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`, and add sanitisation for `prior_research_comment`
- [ ] 3.2 Update `change-factory/scripts/sanitize_context.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`
- [ ] 3.3 Update `code-factory/scripts/sanitize_context.inline.js` — change `stripHtmlComments` calls to `sanitizeUserContent`

## 4. Verify and rebuild locked workflows

- [ ] 4.1 Run the project's workflow generation tooling to regenerate locked workflow YAML from source templates
- [ ] 4.2 Verify the generated workflows reference the updated inline scripts correctly