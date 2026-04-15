import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { classifyChanges } = require(path.resolve(__dirname, 'classify-changes.js'));

test('empty array → providerChanges === true', () => {
  const result = classifyChanges([]);
  assert.equal(result.providerChanges, 'true');
});

test('all files under openspec/ → providerChanges === false', () => {
  const result = classifyChanges(['openspec/specs/foo.md', 'openspec/specs/bar.md']);
  assert.equal(result.providerChanges, 'false');
});

test('mixed files (some openspec, some not) → providerChanges === true', () => {
  const result = classifyChanges(['openspec/specs/foo.md', 'internal/provider/resource.go']);
  assert.equal(result.providerChanges, 'true');
});

test('only non-openspec files → providerChanges === true', () => {
  const result = classifyChanges(['internal/provider/resource.go', 'go.mod']);
  assert.equal(result.providerChanges, 'true');
});

test('single openspec file → providerChanges === false', () => {
  const result = classifyChanges(['openspec/specs/single.md']);
  assert.equal(result.providerChanges, 'false');
});
