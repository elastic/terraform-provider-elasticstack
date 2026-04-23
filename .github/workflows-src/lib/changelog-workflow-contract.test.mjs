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
  assert.match(compiled, /^\s+options:\s*$/m);
  assert.match(compiled, /^\s+- unreleased\s*$/m);
  assert.match(compiled, /^\s+- release\s*$/m);
  assert.match(compiled, /^\s+target_version:\s*$/m);
  assert.match(compiled, /target_version: \$\{\{ github\.event\.inputs\.target_version \|\| '' \}\}/);
  assert.match(compiled, /format\('refs\/heads\/prep-release-\{0\}', github\.event\.inputs\.target_version\)/);
  assert.match(compiled, /- name: Push to generated-changelog branch \(unreleased mode\)/);
  assert.match(compiled, /steps\.resolve_release_context\.outputs\.mode == 'unreleased' &&/);
  assert.doesNotMatch(compiled, /Push to generated-changelog branch \(release mode\)/);
  assert.match(compiled, /id: changelog_engine/);
  assert.match(compiled, /run \.\/scripts\/changelog-engine\/cmd\/changelog-engine/);
  assert.doesNotMatch(compiled, /-previous-tag/);
  assert.doesNotMatch(compiled, /Load changelog engine outputs/);
});
