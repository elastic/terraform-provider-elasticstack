import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  extractBreakingChanges,
  parseChangelogSection,
  parseChangelogSectionFull,
  validateChangelogSection,
  validateChangelogSectionFull,
} = require('./pr-changelog-parser.js');

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const BODY_FIX_WITH_SUMMARY = `
## Description

Fixes a bug in the provider.

## Changelog

Customer impact: fix
Summary: Correct handling of empty API responses

## Notes

Nothing else.
`.trimStart();

const BODY_NONE_NO_SUMMARY = `
## Changelog

Customer impact: none
`.trimStart();

const BODY_BREAKING_FULL = `
## Changelog

Customer impact: breaking
Summary: Remove deprecated attribute from elasticstack_kibana_slo

### Breaking changes

The \`legacy_mode\` attribute has been removed.

- Attribute \`legacy_mode\` is no longer accepted.
- Existing configs referencing \`legacy_mode\` must be updated.

\`\`\`hcl
# Before
resource "elasticstack_kibana_slo" "example" {
  legacy_mode = true
}

# After — remove the attribute entirely
resource "elasticstack_kibana_slo" "example" {
}
\`\`\`

## Other section

Not part of breaking changes.
`.trimStart();

const BODY_NO_CHANGELOG = `
## Description

No changelog section here.

## Notes

Some notes.
`.trimStart();

const BODY_INVALID_IMPACT = `
## Changelog

Customer impact: patch
Summary: Some change
`.trimStart();

const BODY_FIX_MISSING_SUMMARY = `
## Changelog

Customer impact: fix
`.trimStart();

const BODY_BREAKING_CHANGES_EMPTY = `
## Changelog

Customer impact: breaking
Summary: A breaking change

### Breaking changes

## Other section
`.trimStart();

const BODY_BREAKING_CHANGES_FENCED_ONLY = `
## Changelog

Customer impact: breaking
Summary: Breaking change with only code block

### Breaking changes

\`\`\`json
{
  "removed": true
}
\`\`\`
`.trimStart();

// ---------------------------------------------------------------------------
// parseChangelogSection
// ---------------------------------------------------------------------------

test('parseChangelogSection: returns null when no ## Changelog section is present', () => {
  const result = parseChangelogSection(BODY_NO_CHANGELOG);
  assert.equal(result, null);
});

test('parseChangelogSection: returns null for empty string', () => {
  assert.equal(parseChangelogSection(''), null);
});

test('parseChangelogSection: parses Customer impact and Summary for a fix', () => {
  const result = parseChangelogSection(BODY_FIX_WITH_SUMMARY);
  assert.deepEqual(result, {
    customerImpact: 'fix',
    summary: 'Correct handling of empty API responses',
    breakingChanges: null,
  });
});

test('parseChangelogSection: parses Customer impact: none with no summary', () => {
  const result = parseChangelogSection(BODY_NONE_NO_SUMMARY);
  assert.deepEqual(result, {
    customerImpact: 'none',
    summary: null,
    breakingChanges: null,
  });
});

test('parseChangelogSection: parses breaking impact with Summary and extracts breakingChanges', () => {
  const result = parseChangelogSection(BODY_BREAKING_FULL);
  assert.equal(result.customerImpact, 'breaking');
  assert.equal(result.summary, 'Remove deprecated attribute from elasticstack_kibana_slo');
  assert.ok(result.breakingChanges !== null, 'breakingChanges should not be null');
  assert.ok(result.breakingChanges.includes('legacy_mode'), 'breakingChanges should include legacy_mode');
  assert.ok(result.breakingChanges.includes('```hcl'), 'breakingChanges should include fenced code block');
});

test('parseChangelogSection: parses invalid Customer impact value without error', () => {
  const result = parseChangelogSection(BODY_INVALID_IMPACT);
  assert.equal(result.customerImpact, 'patch');
});

// ---------------------------------------------------------------------------
// validateChangelogSection
// ---------------------------------------------------------------------------

test('validateChangelogSection: valid section with fix and summary', () => {
  const parsed = parseChangelogSection(BODY_FIX_WITH_SUMMARY);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

test('validateChangelogSection: valid section with Customer impact: none (no summary required)', () => {
  const parsed = parseChangelogSection(BODY_NONE_NO_SUMMARY);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

test('validateChangelogSection: valid section with breaking impact, summary, and breaking changes', () => {
  const parsed = parseChangelogSection(BODY_BREAKING_FULL);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

test('validateChangelogSection: invalid when Customer impact has unsupported value', () => {
  const parsed = parseChangelogSection(BODY_INVALID_IMPACT);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('patch')), 'error should mention the bad value');
});

test('validateChangelogSection: invalid when Summary is missing and Customer impact is fix', () => {
  const parsed = parseChangelogSection(BODY_FIX_MISSING_SUMMARY);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('Summary')), 'error should mention Summary');
});

test('validateChangelogSection: returns error when parsed is null', () => {
  const result = validateChangelogSection(null);
  assert.equal(result.valid, false);
  assert.ok(result.errors.length > 0);
});

// ---------------------------------------------------------------------------
// validateChangelogSectionFull — empty breaking changes
// ---------------------------------------------------------------------------

test('validateChangelogSectionFull: invalid when ### Breaking changes is present but empty', () => {
  const parsed = parseChangelogSectionFull(BODY_BREAKING_CHANGES_EMPTY);
  const result = validateChangelogSectionFull(parsed);
  assert.equal(result.valid, false);
  assert.ok(
    result.errors.some((e) => e.includes('Breaking changes') && e.includes('no content')),
    'error should mention empty Breaking changes section'
  );
});

// ---------------------------------------------------------------------------
// extractBreakingChanges
// ---------------------------------------------------------------------------

test('extractBreakingChanges: returns null when section is not present', () => {
  assert.equal(extractBreakingChanges(BODY_FIX_WITH_SUMMARY), null);
});

test('extractBreakingChanges: returns null for empty string', () => {
  assert.equal(extractBreakingChanges(''), null);
});

test('extractBreakingChanges: extracts content including lists and fenced code blocks', () => {
  const content = extractBreakingChanges(BODY_BREAKING_FULL);
  assert.ok(content !== null, 'should extract content');
  assert.ok(content.includes('legacy_mode'), 'should include prose with attribute name');
  assert.ok(content.includes('- Attribute'), 'should include list item');
  assert.ok(content.includes('```hcl'), 'should include fenced code block');
  // Should not include the next ## section
  assert.ok(!content.includes('## Other section'), 'should not include next ## section');
});

test('extractBreakingChanges: returns null when heading is present but content is empty', () => {
  assert.equal(extractBreakingChanges(BODY_BREAKING_CHANGES_EMPTY), null);
});

test('extractBreakingChanges: handles fenced code block as the only content', () => {
  const content = extractBreakingChanges(BODY_BREAKING_CHANGES_FENCED_ONLY);
  assert.ok(content !== null, 'should extract fenced code block content');
  assert.ok(content.includes('```json'), 'should include fenced code block');
  assert.ok(content.includes('"removed": true'), 'should include code content');
});
