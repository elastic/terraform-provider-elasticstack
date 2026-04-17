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

// ### Breaking changes heading appears BEFORE ## Changelog — must be ignored
const BODY_BREAKING_BEFORE_CHANGELOG = `
## Summary

### Breaking changes

This section is outside Changelog and must be ignored.

## Changelog

Customer impact: fix
Summary: Fix handling of stale connections
`.trimStart();

const BODY_ENHANCEMENT_WITH_SUMMARY = `
## Changelog

Customer impact: enhancement
Summary: Add support for index lifecycle management
`.trimStart();

const BODY_BREAKING_NO_BREAKING_SECTION = `
## Changelog

Customer impact: breaking
Summary: Remove deprecated attribute
`.trimStart();

const BODY_SUMMARY_EMPTY_VALUE = `
## Changelog

Customer impact: fix
Summary:
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

// ---------------------------------------------------------------------------
// Test gap #4: ### Breaking changes heading outside ## Changelog is ignored
// ---------------------------------------------------------------------------

test('parseChangelogSection: ### Breaking changes before ## Changelog is not extracted', () => {
  const result = parseChangelogSection(BODY_BREAKING_BEFORE_CHANGELOG);
  assert.ok(result !== null, 'should parse the Changelog section');
  assert.equal(result.customerImpact, 'fix');
  assert.equal(result.breakingChanges, null, '### Breaking changes outside ## Changelog must not be extracted');
});

test('parseChangelogSectionFull: breakingChangesHeadingPresent is false when heading is outside ## Changelog', () => {
  const result = parseChangelogSectionFull(BODY_BREAKING_BEFORE_CHANGELOG);
  assert.ok(result !== null, 'should parse the Changelog section');
  assert.equal(result.breakingChangesHeadingPresent, false);
  assert.equal(result.breakingChanges, null);
});

// ---------------------------------------------------------------------------
// Test gap #5: Customer impact: enhancement
// ---------------------------------------------------------------------------

test('parseChangelogSection: parses Customer impact: enhancement with summary', () => {
  const result = parseChangelogSection(BODY_ENHANCEMENT_WITH_SUMMARY);
  assert.deepEqual(result, {
    customerImpact: 'enhancement',
    summary: 'Add support for index lifecycle management',
    breakingChanges: null,
  });
});

test('validateChangelogSection: valid section with enhancement and summary', () => {
  const parsed = parseChangelogSection(BODY_ENHANCEMENT_WITH_SUMMARY);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

// ---------------------------------------------------------------------------
// Test gap #6: validateChangelogSectionFull positive path for breaking
// ---------------------------------------------------------------------------

test('validateChangelogSectionFull: valid for a complete breaking change entry', () => {
  const parsed = parseChangelogSectionFull(BODY_BREAKING_FULL);
  const result = validateChangelogSectionFull(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

// ---------------------------------------------------------------------------
// Test gap #3 (medium): Customer impact: breaking without ### Breaking changes fails
// ---------------------------------------------------------------------------

test('validateChangelogSectionFull: invalid when Customer impact is breaking but ### Breaking changes is absent', () => {
  const parsed = parseChangelogSectionFull(BODY_BREAKING_NO_BREAKING_SECTION);
  const result = validateChangelogSectionFull(parsed);
  assert.equal(result.valid, false);
  assert.ok(
    result.errors.some((e) => e.includes('breaking') && e.includes('Breaking changes')),
    'error should mention breaking and Breaking changes'
  );
});

// ---------------------------------------------------------------------------
// Test gap #7: Summary: with no trailing text
// ---------------------------------------------------------------------------

test('validateChangelogSection: invalid when Summary has no value', () => {
  const parsed = parseChangelogSection(BODY_SUMMARY_EMPTY_VALUE);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('Summary')), 'error should mention Summary');
});

// ---------------------------------------------------------------------------
// Edge case: ## Changelog is the last section (no trailing ## heading)
// ---------------------------------------------------------------------------

const BODY_CHANGELOG_LAST_SECTION = `
## Description

Fixes a bug.

## Changelog

Customer impact: fix
Summary: Fix handling of nil pointer in cluster client
`.trimStart();

test('parseChangelogSection: parses correctly when ## Changelog is the last section with no trailing heading', () => {
  const result = parseChangelogSection(BODY_CHANGELOG_LAST_SECTION);
  assert.deepEqual(result, {
    customerImpact: 'fix',
    summary: 'Fix handling of nil pointer in cluster client',
    breakingChanges: null,
  });
});

test('validateChangelogSection: valid when ## Changelog is the last section', () => {
  const parsed = parseChangelogSection(BODY_CHANGELOG_LAST_SECTION);
  const result = validateChangelogSection(parsed);
  assert.equal(result.valid, true);
  assert.deepEqual(result.errors, []);
});

// ---------------------------------------------------------------------------
// Edge case: multiple ## sections after ## Changelog — only changelog section parsed
// ---------------------------------------------------------------------------

const BODY_CHANGELOG_WITH_MULTIPLE_FOLLOWING_SECTIONS = `
## Summary

This PR adds a new feature.

## Changelog

Customer impact: enhancement
Summary: Add support for cross-cluster replication

## Review notes

Needs extra attention to the retry logic.

## Test plan

- Run acceptance tests for elasticsearch_index
- Verify cross-cluster replication config
`.trimStart();

test('parseChangelogSection: only parses changelog section content when multiple ## sections follow', () => {
  const result = parseChangelogSection(BODY_CHANGELOG_WITH_MULTIPLE_FOLLOWING_SECTIONS);
  assert.ok(result !== null, 'should parse the Changelog section');
  assert.equal(result.customerImpact, 'enhancement');
  assert.equal(result.summary, 'Add support for cross-cluster replication');
  assert.equal(result.breakingChanges, null);
});

test('parseChangelogSection: does not include content from sections after ## Changelog', () => {
  // Confirm that "Review notes" and "Test plan" content is not present in the
  // parsed section (would only matter if the boundary detection failed)
  const section = parseChangelogSection(BODY_CHANGELOG_WITH_MULTIPLE_FOLLOWING_SECTIONS);
  assert.ok(section !== null);
  // The summary must be the one in Changelog, not text from later sections
  assert.equal(section.summary, 'Add support for cross-cluster replication');
});

// ---------------------------------------------------------------------------
// Tilde-fence handling
// ---------------------------------------------------------------------------

const BODY_TILDE_FENCE = `
## Changelog

Customer impact: fix
Summary: Fix a thing

### Breaking changes

~~~
some tilde-fenced block content
## Not a section header
more content
~~~

## Not a section inside the tilde block
`.trimStart();

test('extractBreakingChanges: handles tilde-fenced block and does not break on ## inside it', () => {
  const result = parseChangelogSection(BODY_TILDE_FENCE);
  assert.ok(result !== null, 'should parse the changelog section');
  assert.equal(result.customerImpact, 'fix');
  // The ## line inside the tilde block must not terminate section extraction
  // We verify via parseChangelogSectionFull that breakingChanges is extracted properly
  const full = parseChangelogSectionFull(BODY_TILDE_FENCE);
  assert.ok(full !== null);
  assert.ok(full.breakingChanges !== null, 'tilde-fenced content must be extracted as breakingChanges');
  assert.ok(full.breakingChanges.includes('tilde-fenced block'), 'breaking changes must include fenced content');
});

test('extractChangelogSection: tilde fence inside Changelog does not cause ## inside it to terminate section', () => {
  const body = [
    '## Changelog',
    'Customer impact: fix',
    'Summary: Fix tilde fence edge case',
    '',
    '~~~',
    '## fake heading inside tilde block',
    '~~~',
    '',
    '## Real next section',
  ].join('\n');

  const result = parseChangelogSection(body);
  assert.ok(result !== null);
  assert.equal(result.customerImpact, 'fix');
  assert.equal(result.summary, 'Fix tilde fence edge case');
});
