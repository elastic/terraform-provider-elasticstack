import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { writeFileSync, mkdtempSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';

const require = createRequire(import.meta.url);
const { parseTemporaryIdMap } = require('./producer-dispatch.js');

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
