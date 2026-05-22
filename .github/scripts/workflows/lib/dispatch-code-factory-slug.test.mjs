import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const workflowsDir = path.resolve(__dirname, '../../../workflows');

/** Mirrors the shared fragment bash slug pipeline. */
function slugify(name) {
  return name.toLowerCase().split(' ').join('-');
}

function workflowName(workflowFile) {
  const source = readFileSync(path.join(workflowsDir, workflowFile), 'utf8');
  const match = source.match(/^name:\s*(.+)$/m);
  assert.ok(match, `expected name: in ${workflowFile}`);
  return match[1].trim();
}

const consumers = [
  { file: 'flaky-test-catcher.md', name: 'Flaky Test Catcher', slug: 'flaky-test-catcher' },
  { file: 'semantic-function-refactor.md', name: 'Semantic Function Refactor', slug: 'semantic-function-refactor' },
  { file: 'schema-coverage-rotation.md', name: 'Schema Coverage Rotation', slug: 'schema-coverage-rotation' },
  { file: 'duplicate-code-detector.md', name: 'Duplicate Code Detector', slug: 'duplicate-code-detector' },
];

test('slugify maps the four consumer workflow display names to expected slugs', () => {
  for (const { name, slug } of consumers) {
    assert.equal(slugify(name), slug);
  }
});

test('each consumer workflow name frontmatter slugifies to the expected SOURCE_WORKFLOW slug', () => {
  for (const { file, name, slug } of consumers) {
    assert.equal(workflowName(file), name);
    assert.equal(slugify(workflowName(file)), slug);
  }
});
