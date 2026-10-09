import assert from 'node:assert/strict';
import test, { afterEach } from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const removeStaleOutcomeLabels = require('../research-factory/remove-stale-outcome-labels.js');

const originalInput = process.env.INPUT_ISSUE_NUMBER;
afterEach(() => {
  if (originalInput === undefined) {
    delete process.env.INPUT_ISSUE_NUMBER;
  } else {
    process.env.INPUT_ISSUE_NUMBER = originalInput;
  }
});

function setup({ inputIssueNumber, payloadIssueNumber, absent = [], failing = [] } = {}) {
  if (inputIssueNumber === undefined) {
    delete process.env.INPUT_ISSUE_NUMBER;
  } else {
    process.env.INPUT_ISSUE_NUMBER = inputIssueNumber;
  }
  const removed = [];
  const outputs = {};
  const github = {
    rest: {
      issues: {
        removeLabel: async ({ issue_number, name }) => {
          removed.push({ issue_number, name });
          if (failing.includes(name)) throw Object.assign(new Error('boom'), { status: 500 });
          if (absent.includes(name)) throw Object.assign(new Error('nf'), { status: 404 });
        },
      },
    },
  };
  const core = { setOutput: (k, v) => { outputs[k] = v; }, info() {}, warning() {} };
  const context = {
    repo: { owner: 'o', repo: 'r' },
    payload: payloadIssueNumber ? { issue: { number: payloadIssueNumber } } : {},
  };
  return { github, core, context, removed, outputs };
}

test('removes both outcome labels for an issue event', async () => {
  const env = setup({ payloadIssueNumber: 7 });
  await removeStaleOutcomeLabels(env);
  assert.deepEqual(
    env.removed.map((r) => r.name).sort(),
    ['ready-for-change-factory', 'research-needs-human'],
  );
  assert.ok(env.removed.every((r) => r.issue_number === 7));
});

test('uses the dispatch input issue number without an event payload', async () => {
  const env = setup({ inputIssueNumber: '9' });
  await removeStaleOutcomeLabels(env);
  assert.equal(env.removed.length, 2);
  assert.ok(env.removed.every((r) => r.issue_number === 9));
});

test('tolerates absent labels and reports a reason output', async () => {
  const env = setup({ payloadIssueNumber: 7, absent: ['ready-for-change-factory', 'research-needs-human'] });
  await removeStaleOutcomeLabels(env);
  assert.equal(env.outputs.stale_outcome_labels_removed, 'true');
  assert.match(env.outputs.stale_outcome_labels_removed_reason, /ready-for-change-factory/);
});

test('a failing removal is reported without throwing', async () => {
  const env = setup({ payloadIssueNumber: 7, failing: ['research-needs-human'] });
  await removeStaleOutcomeLabels(env);
  assert.equal(env.outputs.stale_outcome_labels_removed, 'false');
  assert.match(env.outputs.stale_outcome_labels_removed_reason, /Failed to remove/);
});

test('never touches the needs-human classifier label', async () => {
  const env = setup({ payloadIssueNumber: 7 });
  await removeStaleOutcomeLabels(env);
  assert.ok(!env.removed.some((r) => r.name === 'needs-human'));
});

test('skips with a reason when no issue number is available', async () => {
  const env = setup();
  await removeStaleOutcomeLabels(env);
  assert.equal(env.removed.length, 0);
  assert.equal(env.outputs.stale_outcome_labels_removed, 'false');
});

test('the input issue number takes precedence over the event payload', async () => {
  const env = setup({ inputIssueNumber: '9', payloadIssueNumber: 7 });
  await removeStaleOutcomeLabels(env);
  assert.ok(env.removed.every((r) => r.issue_number === 9));
});

test('a non-numeric input falls back to the event payload', async () => {
  const env = setup({ inputIssueNumber: 'abc', payloadIssueNumber: 7 });
  await removeStaleOutcomeLabels(env);
  assert.ok(env.removed.length === 2 && env.removed.every((r) => r.issue_number === 7));
});

test('mixed 404 and 500 results report failure and mention both outcomes', async () => {
  const env = setup({ payloadIssueNumber: 7, absent: ['ready-for-change-factory'], failing: ['research-needs-human'] });
  await removeStaleOutcomeLabels(env);
  assert.equal(env.outputs.stale_outcome_labels_removed, 'false');
  assert.match(env.outputs.stale_outcome_labels_removed_reason, /was not present/);
  assert.match(env.outputs.stale_outcome_labels_removed_reason, /Failed to remove/);
});
