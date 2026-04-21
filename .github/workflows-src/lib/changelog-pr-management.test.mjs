import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

const {
  buildUnreleasedPRBody,
  buildReleasePRBody,
  manageUnreleasedPR,
  refreshReleasePR,
} = require('./changelog-pr-management.js');

// ---------------------------------------------------------------------------
// buildUnreleasedPRBody
// ---------------------------------------------------------------------------

test('buildUnreleasedPRBody: body contains Generated: date', () => {
  const body = buildUnreleasedPRBody({ compareRange: 'v1.0.0...HEAD' });
  assert.match(body, /\*\*Generated:\*\* \d{4}-\d{2}-\d{2}/);
});

test('buildUnreleasedPRBody: body contains Compare range', () => {
  const body = buildUnreleasedPRBody({ compareRange: 'v1.0.0...HEAD' });
  assert.ok(body.includes('**Compare range:** `v1.0.0...HEAD`'), 'Compare range line not found');
});

test('buildUnreleasedPRBody: body contains Do not make manual edits notice', () => {
  const body = buildUnreleasedPRBody({ compareRange: 'v1.0.0...HEAD' });
  assert.ok(body.includes('Do not make manual edits to the `generated-changelog` branch.'), 'Warning notice not found');
});

// ---------------------------------------------------------------------------
// buildReleasePRBody
// ---------------------------------------------------------------------------

test('buildReleasePRBody: body contains Generated: date', () => {
  const body = buildReleasePRBody({ targetVersion: '2.0.0', compareRange: 'v1.0.0...v2.0.0' });
  assert.match(body, /\*\*Generated:\*\* \d{4}-\d{2}-\d{2}/);
});

test('buildReleasePRBody: body contains Version:', () => {
  const body = buildReleasePRBody({ targetVersion: '2.0.0', compareRange: 'v1.0.0...v2.0.0' });
  assert.ok(body.includes('**Version:** `2.0.0`'), 'Version line not found');
});

test('buildReleasePRBody: body contains Compare range:', () => {
  const body = buildReleasePRBody({ targetVersion: '2.0.0', compareRange: 'v1.0.0...v2.0.0' });
  assert.ok(body.includes('**Compare range:** `v1.0.0...v2.0.0`'), 'Compare range line not found');
});

// ---------------------------------------------------------------------------
// manageUnreleasedPR: existing PR → updated
// ---------------------------------------------------------------------------

test('manageUnreleasedPR: existing open PR → updates body, returns action=updated', async () => {
  const updateCalls = [];
  const createCalls = [];

  const github = {
    rest: {
      pulls: {
        list: async () => ({
          data: [{ number: 42, html_url: 'https://github.com/org/repo/pull/42' }],
        }),
        update: async (args) => {
          updateCalls.push(args);
          return { data: {} };
        },
        create: async (args) => {
          createCalls.push(args);
          return { data: { number: 99, html_url: 'https://github.com/org/repo/pull/99' } };
        },
      },
    },
  };

  const result = await manageUnreleasedPR({
    github,
    owner: 'org',
    repo: 'repo',
    compareRange: 'v1.0.0...HEAD',
  });

  assert.equal(result.prAction, 'updated');
  assert.equal(result.prNumber, 42);
  assert.equal(result.prUrl, 'https://github.com/org/repo/pull/42');

  assert.equal(updateCalls.length, 1, 'pulls.update should be called once');
  assert.equal(updateCalls[0].pull_number, 42);
  assert.ok(updateCalls[0].body.includes('Do not make manual edits'), 'body should include warning text');

  assert.equal(createCalls.length, 0, 'pulls.create should not be called');
});

// ---------------------------------------------------------------------------
// manageUnreleasedPR: no existing PR → created
// ---------------------------------------------------------------------------

test('manageUnreleasedPR: no existing PR → creates new PR, returns action=created with head=generated-changelog', async () => {
  const createCalls = [];
  const updateCalls = [];

  const github = {
    rest: {
      pulls: {
        list: async () => ({ data: [] }),
        update: async (args) => {
          updateCalls.push(args);
          return { data: {} };
        },
        create: async (args) => {
          createCalls.push(args);
          return { data: { number: 7, html_url: 'https://github.com/org/repo/pull/7' } };
        },
      },
    },
  };

  const result = await manageUnreleasedPR({
    github,
    owner: 'org',
    repo: 'repo',
    compareRange: 'v0.9.0...HEAD',
  });

  assert.equal(result.prAction, 'created');
  assert.equal(result.prNumber, 7);
  assert.equal(result.prUrl, 'https://github.com/org/repo/pull/7');

  assert.equal(createCalls.length, 1, 'pulls.create should be called once');
  assert.equal(createCalls[0].head, 'generated-changelog');
  assert.equal(createCalls[0].base, 'main');

  assert.equal(updateCalls.length, 0, 'pulls.update should not be called');
});

// ---------------------------------------------------------------------------
// refreshReleasePR: prNumber present → pulls.update called
// ---------------------------------------------------------------------------

test('refreshReleasePR: prNumber present → calls pulls.update with correct body', async () => {
  const updateCalls = [];
  const warnings = [];

  const github = {
    rest: {
      pulls: {
        update: async (args) => {
          updateCalls.push(args);
          return { data: {} };
        },
      },
    },
  };

  const core = {
    info: () => {},
    warning: (msg) => warnings.push(msg),
  };

  await refreshReleasePR({
    github,
    core,
    owner: 'org',
    repo: 'repo',
    prNumber: 55,
    compareRange: 'v1.0.0...v2.0.0',
    targetVersion: '2.0.0',
  });

  assert.equal(updateCalls.length, 1, 'pulls.update should be called once');
  assert.equal(updateCalls[0].pull_number, 55);
  assert.ok(updateCalls[0].body.includes('**Version:** `2.0.0`'), 'body should include version');
  assert.ok(updateCalls[0].body.includes('**Compare range:** `v1.0.0...v2.0.0`'), 'body should include compare range');
  assert.match(updateCalls[0].body, /\*\*Generated:\*\* \d{4}-\d{2}-\d{2}/);

  assert.equal(warnings.length, 0, 'no warnings should be emitted');
});

// ---------------------------------------------------------------------------
// refreshReleasePR: prNumber absent → core.warning called, no API call
// ---------------------------------------------------------------------------

test('refreshReleasePR: prNumber absent → emits warning, does not call pulls.update', async () => {
  const updateCalls = [];
  const warnings = [];

  const github = {
    rest: {
      pulls: {
        update: async (args) => {
          updateCalls.push(args);
          return { data: {} };
        },
      },
    },
  };

  const core = {
    info: () => {},
    warning: (msg) => warnings.push(msg),
  };

  await refreshReleasePR({
    github,
    core,
    owner: 'org',
    repo: 'repo',
    prNumber: null,
    compareRange: 'v1.0.0...v2.0.0',
    targetVersion: '2.0.0',
  });

  assert.equal(updateCalls.length, 0, 'pulls.update should not be called');
  assert.equal(warnings.length, 1, 'one warning should be emitted');
  assert.ok(warnings[0].includes('skipping PR metadata refresh'), 'warning message should mention skipping');
});
