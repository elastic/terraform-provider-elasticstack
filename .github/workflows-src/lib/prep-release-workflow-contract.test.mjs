import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';

const workflowPath = '.github/workflows/prep-release.yml';

function read(path) {
  return readFileSync(path, 'utf8');
}

test('prep-release runs release-mode changelog generation before PR management and without unused engine output file', () => {
  const workflow = read(workflowPath);

  const engineStep = workflow.indexOf('- name: Run shared changelog engine in release mode');
  const prCheckStep = workflow.indexOf('- name: Check if release PR already exists');
  const createPrStep = workflow.indexOf('- name: Create release PR');

  assert.notEqual(engineStep, -1, 'workflow should invoke the shared changelog engine');
  assert.notEqual(prCheckStep, -1, 'workflow should check for an existing release PR');
  assert.notEqual(createPrStep, -1, 'workflow should create a release PR when needed');
  assert.ok(engineStep < prCheckStep, 'release-mode changelog generation should occur before PR reuse checks');
  assert.ok(engineStep < createPrStep, 'release-mode changelog generation should occur before PR creation');
  assert.doesNotMatch(workflow, /-output\s+"\$RUNNER_TEMP\/changelog-engine-outputs"/, 'prep-release should not pass an unused engine output file');
});

test('prep-release keeps deterministic single-commit intent and stable reuse contract', () => {
  const workflow = read(workflowPath);

  assert.match(workflow, /BRANCH="prep-release-\$\{TARGET_VERSION\}"/, 'workflow should derive a deterministic prep-release branch name');
  assert.match(workflow, /git checkout -B "\$\{BRANCH\}" "origin\/\$\{BRANCH\}"/, 'workflow should reset the local branch from the remote branch on rerun');
  assert.match(workflow, /git commit -m "chore\(release\): prepare \$\{TARGET_VERSION\} release"/, 'workflow should create the single deterministic release-preparation commit');
  assert.match(workflow, /gh pr list --head "\$\{BRANCH\}" --state open --json url --jq '\.\[0\]\.url'/, 'workflow should reuse an existing PR for the branch');
  assert.match(workflow, /--label no-changelog/, 'workflow should apply the no-changelog label on PR creation');
  assert.match(workflow, /gh pr edit "\$\{\{ steps\.pr-check\.outputs\.PR_URL \}\}" --add-label no-changelog/, 'workflow should reapply the no-changelog label when reusing a PR');
});
