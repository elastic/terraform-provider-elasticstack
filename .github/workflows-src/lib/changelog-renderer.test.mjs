import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

// The renderer uses `//include: ./pr-changelog-parser.js` (an inline
// compilation directive). When required directly the parser symbols are not
// automatically in scope, so we inject them into global before requiring the
// renderer module.
const parser = require('./pr-changelog-parser.js');
Object.assign(global, parser);

const { renderChangelogSection, normalizeBulletPrefix } = require('./changelog-renderer.js');

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makePR(overrides = {}) {
  return {
    number: 1,
    url: 'https://github.com/org/repo/pull/1',
    labels: [],
    body: null,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

test('single PR with Customer impact: fix and Summary renders a correct bullet', () => {
  const pr = makePR({
    number: 42,
    url: 'https://github.com/org/repo/pull/42',
    body: '## Changelog\nCustomer impact: fix\nSummary: fixed a nasty bug\n',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.errors.length, 0);
  assert.equal(result.included.length, 1);
  assert.equal(result.excluded.length, 0);
  assert.ok(result.sectionBody.includes('### Changes'));
  assert.ok(result.sectionBody.includes('- fixed a nasty bug ([#42](https://github.com/org/repo/pull/42))'));
});

test('PR with Customer impact: none is excluded from bullets', () => {
  const pr = makePR({
    number: 7,
    url: 'https://github.com/org/repo/pull/7',
    body: '## Changelog\nCustomer impact: none\n',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.included.length, 0);
  assert.equal(result.excluded.length, 1);
  assert.equal(result.excluded[0].reason, 'Customer impact: none');
  assert.equal(result.sectionBody, '');
});

test('PR with no-changelog label is excluded entirely', () => {
  const pr = makePR({
    number: 3,
    url: 'https://github.com/org/repo/pull/3',
    labels: ['no-changelog'],
    body: null,
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.included.length, 0);
  assert.equal(result.excluded.length, 1);
  assert.equal(result.excluded[0].reason, 'no-changelog label');
});

test('PR with ## Changelog but no Customer impact: line fails assembly', () => {
  const pr = makePR({
    number: 5,
    url: 'https://github.com/org/repo/pull/5',
    body: '## Changelog\nSummary: something happened\n',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, false);
  assert.equal(result.errors.length, 1);
  assert.ok(
    result.errors[0].reason.includes('missing the required Customer impact field'),
    `Expected missing Customer impact error, got: ${result.errors[0].reason}`,
  );
});

test('PR with Customer impact: breaking, Summary, and ### Breaking changes renders correctly', () => {
  const body = [
    '## Changelog',
    'Customer impact: breaking',
    'Summary: removed the old API endpoint',
    '',
    '### Breaking changes',
    'The `/v1/legacy` endpoint has been removed. Migrate to `/v2/endpoint`.',
    '',
  ].join('\n');

  const pr = makePR({
    number: 99,
    url: 'https://github.com/org/repo/pull/99',
    body,
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true, `Expected success but got errors: ${JSON.stringify(result.errors)}`);
  assert.equal(result.included.length, 1);
  assert.ok(result.sectionBody.includes('### Breaking changes'));
  assert.ok(result.sectionBody.includes('### Changes'));
  assert.ok(result.sectionBody.includes('- removed the old API endpoint ([#99](https://github.com/org/repo/pull/99))'));
  assert.ok(result.sectionBody.includes('The `/v1/legacy` endpoint has been removed'));
});

test('PR with an invalid Customer impact value fails assembly', () => {
  const pr = makePR({
    number: 11,
    url: 'https://github.com/org/repo/pull/11',
    body: '## Changelog\nCustomer impact: refactor\nSummary: cleaned up internals\n',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, false);
  assert.equal(result.errors.length, 1);
  assert.ok(
    result.errors[0].reason.includes('failed validation'),
    `Expected validation failure error, got: ${result.errors[0].reason}`,
  );
});

test('PR entirely missing ## Changelog and without no-changelog label fails assembly', () => {
  const pr = makePR({
    number: 20,
    url: 'https://github.com/org/repo/pull/20',
    body: '## Description\nSome description but no changelog section.',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, false);
  assert.equal(result.errors.length, 1);
  assert.ok(
    result.errors[0].reason.includes('no parseable ## Changelog section'),
    `Expected missing changelog section error, got: ${result.errors[0].reason}`,
  );
});

// ---------------------------------------------------------------------------
// Mixed batch: valid fix, no-changelog label, Customer impact: none,
//              breaking change — combined output is correct
// ---------------------------------------------------------------------------

test('mixed batch renders correct combined output for fix, no-changelog, none, and breaking PRs', () => {
  const fixPR = makePR({
    number: 101,
    url: 'https://github.com/org/repo/pull/101',
    body: '## Changelog\nCustomer impact: fix\nSummary: Fix the widget factory\n',
  });

  const noChangelogPR = makePR({
    number: 102,
    url: 'https://github.com/org/repo/pull/102',
    labels: ['no-changelog'],
    body: null,
  });

  const nonePR = makePR({
    number: 103,
    url: 'https://github.com/org/repo/pull/103',
    body: '## Changelog\nCustomer impact: none\n',
  });

  const breakingBody = [
    '## Changelog',
    'Customer impact: breaking',
    'Summary: Remove the old authentication endpoint',
    '',
    '### Breaking changes',
    '',
    'The `/v1/auth` endpoint has been removed. Use `/v2/auth` instead.',
    '',
  ].join('\n');

  const breakingPR = makePR({
    number: 104,
    url: 'https://github.com/org/repo/pull/104',
    body: breakingBody,
  });

  const result = renderChangelogSection([fixPR, noChangelogPR, nonePR, breakingPR]);

  // Overall result must succeed
  assert.equal(result.success, true, `Expected success but got errors: ${JSON.stringify(result.errors)}`);
  assert.deepEqual(result.errors, []);

  // Included: only fix (#101) and breaking (#104)
  assert.equal(result.included.length, 2, 'only fix and breaking PRs should be included');
  const includedNumbers = result.included.map((p) => p.prNumber);
  assert.ok(includedNumbers.includes(101), 'fix PR must be included');
  assert.ok(includedNumbers.includes(104), 'breaking PR must be included');

  // Excluded: no-changelog (#102) and none (#103)
  assert.equal(result.excluded.length, 2, 'no-changelog and none PRs should be excluded');
  const excludedMap = Object.fromEntries(result.excluded.map((p) => [p.prNumber, p.reason]));
  assert.equal(excludedMap[102], 'no-changelog label');
  assert.equal(excludedMap[103], 'Customer impact: none');

  // Section body contains both ### Breaking changes and ### Changes subsections
  assert.ok(result.sectionBody.includes('### Breaking changes'), 'sectionBody must have ### Breaking changes');
  assert.ok(result.sectionBody.includes('### Changes'), 'sectionBody must have ### Changes');

  // Breaking change prose from #104 is present
  assert.ok(
    result.sectionBody.includes('/v1/auth'),
    'sectionBody must contain the breaking-change prose from #104',
  );

  // Change bullet for fix PR (#101) is present
  assert.ok(
    result.sectionBody.includes('- Fix the widget factory ([#101](https://github.com/org/repo/pull/101))'),
    'sectionBody must contain the fix bullet for #101',
  );

  // Change bullet for breaking PR (#104) is present
  assert.ok(
    result.sectionBody.includes('- Remove the old authentication endpoint ([#104](https://github.com/org/repo/pull/104))'),
    'sectionBody must contain the breaking change bullet for #104',
  );

  // PRs excluded from bullets must not appear as change bullets
  assert.ok(!result.sectionBody.includes('#102'), 'no-changelog PR #102 must not appear in sectionBody');
  assert.ok(!result.sectionBody.includes('#103'), 'none PR #103 must not appear in sectionBody');
});

// ---------------------------------------------------------------------------
// Customer impact: none with ### Breaking changes — excluded entry includes breakingChanges
// ---------------------------------------------------------------------------

test('Customer impact: breaking PR with Breaking changes block is included and breakingChanges rendered', () => {
  const body = [
    '## Changelog',
    'Customer impact: breaking',
    'Summary: An internal refactor removes an undocumented field',
    '',
    '### Breaking changes',
    'This internal refactor technically removes an undocumented field.',
    '',
  ].join('\n');

  const pr = makePR({
    number: 55,
    url: 'https://github.com/org/repo/pull/55',
    body,
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.included.length, 1, 'PR should be in included');
  assert.equal(result.excluded.length, 0, 'PR should not be in excluded');
  assert.equal(result.included[0].summary, 'An internal refactor removes an undocumented field');
  assert.ok(
    result.included[0].breakingChanges !== undefined && result.included[0].breakingChanges !== null,
    'included entry must carry breakingChanges when present',
  );
  assert.ok(
    result.included[0].breakingChanges.includes('undocumented field'),
    'breakingChanges must contain the prose from the PR',
  );
  // Breaking changes from this PR are still rendered in the section body
  assert.ok(result.sectionBody.includes('### Breaking changes'), 'sectionBody must still have ### Breaking changes');
  assert.ok(result.sectionBody.includes('undocumented field'), 'sectionBody must include the breaking change prose');
});

test('Customer impact: none PR without Breaking changes block has no breakingChanges in excluded entry', () => {
  const pr = makePR({
    number: 7,
    url: 'https://github.com/org/repo/pull/7',
    body: '## Changelog\nCustomer impact: none\n',
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.excluded.length, 1);
  assert.equal(result.excluded[0].reason, 'Customer impact: none');
  assert.equal(result.excluded[0].breakingChanges, undefined, 'breakingChanges should not be set when absent');
});

// ---------------------------------------------------------------------------
// normalizeBulletPrefix — no-space edge case
// ---------------------------------------------------------------------------

// Regression: Customer impact: none + ### Breaking changes still works at release time
// because release-time rendering skips the breaking-impact match check.
test('Customer impact: none PR with Breaking changes block is excluded and breakingChanges preserved', () => {
  const body = [
    '## Changelog',
    'Customer impact: none',
    '',
    '### Breaking changes',
    'This is a legacy internal change with breaking implications.',
  ].join('\n');

  const pr = makePR({
    number: 56,
    url: 'https://github.com/org/repo/pull/56',
    body,
  });

  const result = renderChangelogSection([pr]);

  assert.equal(result.success, true);
  assert.equal(result.excluded.length, 1, 'PR should be in excluded');
  assert.equal(result.excluded[0].reason, 'Customer impact: none');
  assert.ok(
    result.excluded[0].breakingChanges &&
    result.excluded[0].breakingChanges.includes('legacy internal change'),
    'excluded entry must carry breakingChanges when present',
  );
  // Breaking changes from none PRs are still rendered in the section body
  assert.ok(result.sectionBody.includes('### Breaking changes'), 'sectionBody must still have ### Breaking changes');
  assert.ok(
    result.sectionBody.includes('legacy internal change'),
    'sectionBody must include the breaking change prose',
  );
});

test('normalizeBulletPrefix: normalizes bullet with no space after dash', () => {
  assert.equal(normalizeBulletPrefix('-fix bug'), '- fix bug');
});

test('normalizeBulletPrefix: normalizes standard bullet with space', () => {
  assert.equal(normalizeBulletPrefix('- fix bug'), '- fix bug');
});

test('normalizeBulletPrefix: normalizes asterisk bullet', () => {
  assert.equal(normalizeBulletPrefix('* fix bug'), '- fix bug');
});

test('normalizeBulletPrefix: normalizes plus bullet with no space', () => {
  assert.equal(normalizeBulletPrefix('+fix bug'), '- fix bug');
});
