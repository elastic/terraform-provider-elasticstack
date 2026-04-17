/**
 * Tests for the gating logic used by the PR changelog authoring workflow.
 *
 * The inline scripts in:
 *   .github/workflows-src/pr-changelog-authoring/scripts/validate-pr-changelog.inline.js
 *
 * delegate all parsing and validation to `validateChangelogSectionFull` from
 * pr-changelog-parser.js. These tests exercise those exact code paths and verify
 * the four gating outcomes the workflow depends on:
 *
 *   1. PR has `no-changelog` label  → workflow skips (no validation needed)
 *   2. Valid `## Changelog` present → changelog_valid=true, changelog_present=true
 *   3. Malformed `## Changelog`     → changelog_valid=false (workflow fails)
 *   4. Missing `## Changelog`       → changelog_present=false (agent drafts one)
 */

import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  parseChangelogSectionFull,
  validateChangelogSectionFull,
} = require('./pr-changelog-parser.js');

// ---------------------------------------------------------------------------
// Helpers that mirror what validate-pr-changelog.inline.js does
// ---------------------------------------------------------------------------

/**
 * Simulate the gating decision the inline workflow script makes.
 *
 * @param {string} prBody
 * @returns {{ changelogPresent: boolean, changelogValid: boolean, errors: string[] }}
 */
function simulateGating(prBody) {
  const parsed = parseChangelogSectionFull(prBody);

  if (parsed === null) {
    // No ## Changelog section — agent will draft one
    return { changelogPresent: false, changelogValid: false, errors: [] };
  }

  const validation = validateChangelogSectionFull(parsed);

  if (validation.valid) {
    return { changelogPresent: true, changelogValid: true, errors: [] };
  }

  return { changelogPresent: true, changelogValid: false, errors: validation.errors };
}

// ---------------------------------------------------------------------------
// Gate outcome 1 — no-changelog label
// The label check happens in resolve-pr.inline.js (before validateChangelog),
// so the validation function is never called.  We verify here that the parser
// still returns null when there is no ## Changelog section, confirming the
// label-gated path leaves the body unvalidated.
// ---------------------------------------------------------------------------

test('gating: PR with no-changelog label has no ## Changelog section — parser returns null', () => {
  const prBody = '## Description\n\nThis is an internal refactor.\n';
  const parsed = parseChangelogSectionFull(prBody);
  // The workflow would have already exited at the label check; we confirm the
  // body produces null so nothing would run even if it didn't.
  assert.equal(parsed, null);
});

// ---------------------------------------------------------------------------
// Gate outcome 2 — valid ## Changelog section
// ---------------------------------------------------------------------------

test('gating: valid fix changelog passes — changelog_present=true, changelog_valid=true', () => {
  const prBody = [
    '## Description',
    '',
    'This fixes a regression.',
    '',
    '## Changelog',
    '',
    'Customer impact: fix',
    'Summary: Correct handling of empty API responses',
    '',
  ].join('\n');

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, true);
  assert.deepEqual(result.errors, []);
});

test('gating: valid enhancement changelog passes', () => {
  const prBody = [
    '## Changelog',
    '',
    'Customer impact: enhancement',
    'Summary: Add support for index lifecycle management',
    '',
  ].join('\n');

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, true);
});

test('gating: valid none changelog passes (no summary required)', () => {
  const prBody = '## Changelog\n\nCustomer impact: none\n';

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, true);
});

test('gating: valid breaking changelog with ### Breaking changes passes', () => {
  const prBody = [
    '## Changelog',
    '',
    'Customer impact: breaking',
    'Summary: Remove deprecated attribute from elasticstack_kibana_slo',
    '',
    '### Breaking changes',
    '',
    'The `legacy_mode` attribute has been removed. Update your configs.',
    '',
  ].join('\n');

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, true);
  assert.deepEqual(result.errors, []);
});

// ---------------------------------------------------------------------------
// Gate outcome 3 — malformed ## Changelog section (workflow fails)
// ---------------------------------------------------------------------------

test('gating: malformed changelog (invalid Customer impact value) — changelog_valid=false', () => {
  const prBody = '## Changelog\n\nCustomer impact: patch\nSummary: Some change\n';

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, false);
  assert.ok(result.errors.length > 0, 'should have validation errors');
  assert.ok(
    result.errors.some((e) => e.includes('patch')),
    'error should mention the bad value "patch"'
  );
});

test('gating: malformed changelog (fix without Summary) — changelog_valid=false', () => {
  const prBody = '## Changelog\n\nCustomer impact: fix\n';

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, false);
  assert.ok(
    result.errors.some((e) => e.includes('Summary')),
    'error should mention missing Summary'
  );
});

test('gating: malformed changelog (breaking without ### Breaking changes section) — changelog_valid=false', () => {
  const prBody = '## Changelog\n\nCustomer impact: breaking\nSummary: Remove old API\n';

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, false);
  assert.ok(
    result.errors.some((e) => e.includes('Breaking changes')),
    'error should mention Breaking changes requirement'
  );
});

test('gating: malformed changelog (empty ### Breaking changes heading) — changelog_valid=false', () => {
  const prBody = [
    '## Changelog',
    '',
    'Customer impact: breaking',
    'Summary: A breaking change',
    '',
    '### Breaking changes',
    '',
    '## Other section',
    '',
  ].join('\n');

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, true);
  assert.equal(result.changelogValid, false);
  assert.ok(
    result.errors.some((e) => e.includes('Breaking changes') && e.includes('no content')),
    'error should mention empty Breaking changes'
  );
});

// ---------------------------------------------------------------------------
// Gate outcome 4 — missing ## Changelog section (agent drafts one)
// ---------------------------------------------------------------------------

test('gating: missing ## Changelog — changelog_present=false, triggers agent draft', () => {
  const prBody = '## Description\n\nNo changelog section here.\n\n## Notes\n\nSome notes.\n';

  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, false);
  assert.equal(result.changelogValid, false);
  assert.deepEqual(result.errors, []);
});

test('gating: empty PR body — changelog_present=false, triggers agent draft', () => {
  const result = simulateGating('');
  assert.equal(result.changelogPresent, false);
  assert.equal(result.changelogValid, false);
});

test('gating: null PR body treated as empty — changelog_present=false', () => {
  // The inline script does `const prBody = pr.body ?? ''` before calling the parser,
  // so we simulate the same normalization here.
  const prBody = null ?? '';
  const result = simulateGating(prBody);
  assert.equal(result.changelogPresent, false);
  assert.equal(result.changelogValid, false);
});
