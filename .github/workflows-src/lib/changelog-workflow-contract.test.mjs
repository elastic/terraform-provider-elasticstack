import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';

const sourcePath = '.github/workflows-src/changelog-generation/workflow.yml.tmpl';
const compiledPath = '.github/workflows/changelog-generation.yml';

function read(path) {
  return readFileSync(path, 'utf8');
}

test('changelog workflow source removes pull_request_target and exposes explicit workflow_dispatch inputs', () => {
  const source = read(sourcePath);

  assert.doesNotMatch(source, /^\s*pull_request_target:/m);
  assert.match(source, /^\s*workflow_dispatch:\s*$/m);
  assert.match(source, /^\s+mode:\s*$/m);
  assert.match(source, /^\s+target_version:\s*$/m);
});

test('compiled changelog workflow matches source expectations', () => {
  const compiled = read(compiledPath);

  assert.doesNotMatch(compiled, /^\s*pull_request_target:/m);
  assert.match(compiled, /^\s*workflow_dispatch:\s*$/m);
  assert.match(compiled, /^\s+mode:\s*$/m);
  assert.match(compiled, /^\s+target_version:\s*$/m);
  assert.match(compiled, /python - <<'PY'/);
  assert.match(compiled, /json\.loads\(output_path\.read_text\(\)\)/);
});
