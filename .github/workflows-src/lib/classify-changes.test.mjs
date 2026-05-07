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

// New skip-path tests

test('CHANGELOG.md only → providerChanges === false', () => {
  const result = classifyChanges(['CHANGELOG.md']);
  assert.equal(result.providerChanges, 'false');
});

test('all files under .agents/ → providerChanges === false', () => {
  const result = classifyChanges(['.agents/skills/foo/SKILL.md', '.agents/skills/bar/SKILL.md']);
  assert.equal(result.providerChanges, 'false');
});

test('.github/ non-workflow files → providerChanges === false', () => {
  const result = classifyChanges(['.github/dependabot.yml', '.github/issue_templates/bug.md']);
  assert.equal(result.providerChanges, 'false');
});

test('.github/workflows/provider.yml → providerChanges === true', () => {
  const result = classifyChanges(['.github/workflows/provider.yml']);
  assert.equal(result.providerChanges, 'true');
});

test('mixed changes (Go file + CHANGELOG) → providerChanges === true', () => {
  const result = classifyChanges(['internal/provider/resource.go', 'CHANGELOG.md']);
  assert.equal(result.providerChanges, 'true');
});

test('mixed .github/workflows/provider.yml + CHANGELOG → providerChanges === true', () => {
  const result = classifyChanges(['.github/workflows/provider.yml', 'CHANGELOG.md']);
  assert.equal(result.providerChanges, 'true');
});
