import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { writeFileSync, mkdtempSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';

const require = createRequire(import.meta.url);
const { parseTemporaryIdMap, dispatchCodeFactory } = require('./producer-dispatch.js');

function withTempFile(name, content) {
  const dir = mkdtempSync(join(tmpdir(), 'producer-dispatch-test-'));
  const path = join(dir, name);
  writeFileSync(path, content);
  return { path, cleanup: () => rmSync(dir, { recursive: true, force: true }) };
}

test('parseTemporaryIdMap returns entries for valid map', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-elasticstack', number: 42 },
      'issue-2': { repo: 'elastic/terraform-provider-elasticstack', number: 43 },
    })
  );
  const entries = parseTemporaryIdMap(path);
  assert.equal(entries.length, 2);
  assert.deepStrictEqual(entries[0], {
    repo: 'elastic/terraform-provider-elasticstack',
    number: 42,
  });
  assert.deepStrictEqual(entries[1], {
    repo: 'elastic/terraform-provider-elasticstack',
    number: 43,
  });
  cleanup();
});

test('parseTemporaryIdMap returns empty array for empty map', () => {
  const { path, cleanup } = withTempFile('map.json', '{}');
  const entries = parseTemporaryIdMap(path);
  assert.equal(entries.length, 0);
  cleanup();
});

test('parseTemporaryIdMap throws when file is missing', () => {
  assert.throws(() => parseTemporaryIdMap('/nonexistent/path.json'), /not found/);
});

test('parseTemporaryIdMap throws for malformed JSON', () => {
  const { path, cleanup } = withTempFile('map.json', 'not json');
  assert.throws(() => parseTemporaryIdMap(path), /Failed to parse/);
  cleanup();
});

test('parseTemporaryIdMap throws for non-object root (array)', () => {
  const { path, cleanup } = withTempFile('map.json', '[]');
  assert.throws(() => parseTemporaryIdMap(path), /must be a JSON object/);
  cleanup();
});

test('parseTemporaryIdMap throws for non-object root (null)', () => {
  const { path, cleanup } = withTempFile('map.json', 'null');
  assert.throws(() => parseTemporaryIdMap(path), /must be a JSON object/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry missing repo', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': { number: 42 } })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid repo/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with invalid repo format', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': { repo: 'no-slash', number: 42 } })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid repo/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with non-numeric number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-elasticstack', number: 'abc' },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with zero number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-elasticstack', number: 0 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with negative number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-elasticstack', number: -5 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry with decimal number', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({
      'issue-1': { repo: 'elastic/terraform-provider-elasticstack', number: 3.14 },
    })
  );
  assert.throws(() => parseTemporaryIdMap(path), /invalid number/);
  cleanup();
});

test('parseTemporaryIdMap throws for entry that is a primitive', () => {
  const { path, cleanup } = withTempFile(
    'map.json',
    JSON.stringify({ 'issue-1': 42 })
  );
  assert.throws(() => parseTemporaryIdMap(path), /Entry "issue-1" must be an object/);
  cleanup();
});

// ---------------------------------------------------------------------------
// dispatchCodeFactory
// ---------------------------------------------------------------------------

test('dispatchCodeFactory throws when GH_TOKEN and GITHUB_TOKEN are absent', () => {
  const origGh = process.env.GH_TOKEN;
  const origGitHub = process.env.GITHUB_TOKEN;
  delete process.env.GH_TOKEN;
  delete process.env.GITHUB_TOKEN;
  try {
    assert.throws(
      () => dispatchCodeFactory([{ repo: 'elastic/test', number: 1 }], 'test'),
      /GH_TOKEN or GITHUB_TOKEN/
    );
  } finally {
    if (origGh !== undefined) process.env.GH_TOKEN = origGh;
    if (origGitHub !== undefined) process.env.GITHUB_TOKEN = origGitHub;
  }
});

test('dispatchCodeFactory throws for cross-repo dispatch when GITHUB_REPOSITORY is set', () => {
  const origRepo = process.env.GITHUB_REPOSITORY;
  process.env.GITHUB_REPOSITORY = 'elastic/allowed';
  const origToken = process.env.GH_TOKEN;
  process.env.GH_TOKEN = 'test-token';
  try {
    assert.throws(
      () => dispatchCodeFactory([{ repo: 'elastic/different', number: 1 }], 'test'),
      /not the current repository/
    );
  } finally {
    if (origRepo !== undefined) process.env.GITHUB_REPOSITORY = origRepo;
    else delete process.env.GITHUB_REPOSITORY;
    if (origToken !== undefined) process.env.GH_TOKEN = origToken;
    else delete process.env.GH_TOKEN;
  }
});

test('dispatchCodeFactory allows same-repo dispatch when GITHUB_REPOSITORY is set', () => {
  const origRepo = process.env.GITHUB_REPOSITORY;
  process.env.GITHUB_REPOSITORY = 'elastic/terraform-provider-elasticstack';
  const origToken = process.env.GH_TOKEN;
  process.env.GH_TOKEN = 'test-token';
  try {
    // With no spawnSync mock this would actually try to run gh; the test verifies
    // cross-repo guard allows matching repo before shell execution.
    assert.throws(
      () => dispatchCodeFactory([{ repo: 'elastic/terraform-provider-elasticstack', number: 1 }], 'test'),
      /Failed to dispatch/  // gh CLI won't exist in test env
    );
  } finally {
    if (origRepo !== undefined) process.env.GITHUB_REPOSITORY = origRepo;
    else delete process.env.GITHUB_REPOSITORY;
    if (origToken !== undefined) process.env.GH_TOKEN = origToken;
    else delete process.env.GH_TOKEN;
  }
});
