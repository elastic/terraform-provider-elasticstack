import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { classifyPR } = require(path.resolve(__dirname, 'build-evidence-manifest.js'));

// ---------------------------------------------------------------------------
// classifyPR tests
// ---------------------------------------------------------------------------

test('classifyPR: dependabot PR → internal', () => {
  const pr = { user: { login: 'dependabot[bot]' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /dependabot/);
  assert.equal(result.inclusion_rationale, null);
});

test('classifyPR: github-actions bot PR → internal', () => {
  const pr = { user: { login: 'github-actions[bot]' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /github-actions/);
});

test('classifyPR: openspec-only files → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [
    { filename: 'openspec/specs/foo.md' },
    { filename: 'openspec/specs/bar.md' },
  ];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /openspec/);
});

test('classifyPR: enhancement label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'enhancement' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /enhancement/);
  assert.equal(result.exclusion_rationale, null);
});

test('classifyPR: bug label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'bug' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /bug/);
});

test('classifyPR: breaking-change label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'breaking-change' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: internal label, no provider code → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'chore' }] };
  const files = [{ filename: 'docs/guide.md' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /chore/);
});

test('classifyPR: internal label but touches provider code → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'chore' }] };
  const files = [{ filename: 'internal/provider/resource.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: no labels, touches provider code → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'pkg/utils/helper.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /provider implementation/);
});

test('classifyPR: no labels, no provider code → uncertain', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'docs/some-doc.md' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'uncertain');
  assert.match(result.inclusion_rationale, /uncertain/);
});

test('classifyPR: go.mod file → user-facing (provider path)', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'go.mod' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: new-resource label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'new-resource' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: mixed openspec and provider files → not openspec-only', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [
    { filename: 'openspec/specs/foo.md' },
    { filename: 'internal/provider/resource.go' },
  ];
  const result = classifyPR(pr, files);
  // Should NOT be classified as internal due to openspec-only rule
  // Has provider code → user-facing
  assert.equal(result.classification, 'user-facing');
});
