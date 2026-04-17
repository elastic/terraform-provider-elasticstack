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

const { renderChangelogSection } = require('./changelog-renderer.js');

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
