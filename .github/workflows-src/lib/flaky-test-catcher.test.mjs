import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const require = createRequire(import.meta.url);
const { classifyRuns, computeGate, filterIssues } = require('./flaky-test-catcher.js');

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/flaky-test-catcher.md');
const lockPath = path.resolve(__dirname, '../../workflows/flaky-test-catcher.lock.yml');

// ---------------------------------------------------------------------------
// classifyRuns
// ---------------------------------------------------------------------------

test('classifyRuns returns empty failedRunIds and zero count for empty input', () => {
  const result = classifyRuns([]);
  assert.deepEqual(result.failedRunIds, []);
  assert.equal(result.totalRunCount, 0);
});

test('classifyRuns counts only countable runs and identifies failed ones', () => {
  const runs = [
    { id: 1, conclusion: 'success' },
    { id: 2, conclusion: 'failure' },
    { id: 3, conclusion: 'failure' },
    { id: 4, conclusion: 'cancelled' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, [2, 3]);
  // cancelled is not countable, so totalRunCount is 3
  assert.equal(result.totalRunCount, 3);
});

test('classifyRuns returns no failed IDs when all runs succeed', () => {
  const runs = [
    { id: 10, conclusion: 'success' },
    { id: 11, conclusion: 'success' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, []);
  assert.equal(result.totalRunCount, 2);
});

test('classifyRuns returns all IDs when all runs fail', () => {
  const runs = [
    { id: 20, conclusion: 'failure' },
    { id: 21, conclusion: 'failure' },
    { id: 22, conclusion: 'failure' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, [20, 21, 22]);
  assert.equal(result.totalRunCount, 3);
});

test('classifyRuns excludes null-conclusion (in-progress) runs from totalRunCount', () => {
  const runs = [
    { id: 30, conclusion: null },
    { id: 31, conclusion: 'failure' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, [31]);
  assert.equal(result.totalRunCount, 1);
});

test('classifyRuns excludes skipped runs from totalRunCount', () => {
  const runs = [
    { id: 40, conclusion: 'skipped' },
    { id: 41, conclusion: 'success' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, []);
  assert.equal(result.totalRunCount, 1);
});

test('classifyRuns excludes cancelled runs from totalRunCount', () => {
  const runs = [
    { id: 50, conclusion: 'cancelled' },
    { id: 51, conclusion: 'cancelled' },
    { id: 52, conclusion: 'failure' },
    { id: 53, conclusion: 'success' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, [52]);
  assert.equal(result.totalRunCount, 2);
});

test('classifyRuns treats timed_out runs as non-failures but counts them', () => {
  const runs = [
    { id: 60, conclusion: 'timed_out' },
    { id: 61, conclusion: 'failure' },
  ];
  const result = classifyRuns(runs);
  assert.deepEqual(result.failedRunIds, [61]);
  assert.equal(result.totalRunCount, 2); // timed_out is countable but not a failure
});

// ---------------------------------------------------------------------------
// computeGate
// ---------------------------------------------------------------------------

const slotsAvailable = {
  open_issues: 1,
  issue_slots_available: 2,
  gate_reason: '2 slot(s) available: 1 open flaky-test issue(s), cap is 3.',
};

const slotsFull = {
  open_issues: 3,
  issue_slots_available: 0,
  gate_reason: 'Issue cap reached: 3 open flaky-test issue(s), cap is 3. Agent job will be skipped.',
};

const slotsEmpty = {
  open_issues: 0,
  issue_slots_available: 3,
  gate_reason: '3 slot(s) available: 0 open flaky-test issue(s), cap is 3.',
};

test('computeGate returns has_ci_failures=false when no failures', () => {
  const result = computeGate([], slotsAvailable);
  assert.equal(result.has_ci_failures, 'false');
  assert.ok(result.gate_reason.includes('No CI failures'));
  assert.ok(result.gate_reason.includes('skipped'));
});

test('computeGate returns has_ci_failures=true when failures exist and slots available', () => {
  const result = computeGate([100, 101], slotsAvailable);
  assert.equal(result.has_ci_failures, 'true');
  assert.ok(result.gate_reason.includes('2 failed run'));
});

test('computeGate returns has_ci_failures=true but notes cap when slots are 0', () => {
  const result = computeGate([200], slotsFull);
  assert.equal(result.has_ci_failures, 'true');
  assert.ok(result.gate_reason.includes('1 failed run'));
  assert.ok(result.gate_reason.includes('Issue cap reached'));
});

test('computeGate no-failures branch ignores issue slots', () => {
  // Even with slots full, no failures means skip
  const result = computeGate([], slotsFull);
  assert.equal(result.has_ci_failures, 'false');
  assert.ok(result.gate_reason.includes('No CI failures'));
});

test('computeGate with many failures and all slots available', () => {
  const ids = [1, 2, 3, 4, 5];
  const result = computeGate(ids, slotsEmpty);
  assert.equal(result.has_ci_failures, 'true');
  assert.ok(result.gate_reason.includes('5 failed run'));
});

test('computeGate gate_reason forwards full issueSlots gate_reason text', () => {
  const result = computeGate([99], slotsAvailable);
  assert.equal(result.has_ci_failures, 'true');
  assert.ok(result.gate_reason.includes(slotsAvailable.gate_reason));
});

test('computeGate with failures and exactly 1 slot available', () => {
  const slots = {
    open_issues: 2,
    issue_slots_available: 1,
    gate_reason: '1 slot(s) available: 2 open flaky-test issue(s), cap is 3.',
  };
  const result = computeGate([42], slots);
  assert.equal(result.has_ci_failures, 'true');
  assert.ok(result.gate_reason.includes('1 failed run'));
  assert.ok(result.gate_reason.includes('1 slot'));
});

// ---------------------------------------------------------------------------
// filterIssues
// ---------------------------------------------------------------------------

test('filterIssues returns empty array for empty input', () => {
  assert.deepEqual(filterIssues([]), []);
});

test('filterIssues keeps all items without pull_request field', () => {
  const items = [{ id: 1 }, { id: 2 }, { id: 3 }];
  assert.equal(filterIssues(items).length, 3);
});

test('filterIssues removes all items with pull_request set', () => {
  const items = [
    { id: 1, pull_request: { url: 'https://...' } },
    { id: 2, pull_request: {} },
  ];
  assert.deepEqual(filterIssues(items), []);
});

test('filterIssues keeps only real issues in a mixed list', () => {
  const items = [
    { id: 1 },
    { id: 2, pull_request: { url: 'https://...' } },
    { id: 3 },
    { id: 4, pull_request: {} },
  ];
  const result = filterIssues(items);
  assert.equal(result.length, 2);
  assert.deepEqual(result.map(i => i.id), [1, 3]);
});

test('workflow includes dispatch instruction and compiled lock contains dispatch_code_factory job', () => {
  const source = readFileSync(workflowPath, 'utf8');
  const lock = readFileSync(lockPath, 'utf8');
  assert.match(source, /dispatch_code_factory/);
  assert.match(source, /Dispatch/);
  assert.match(lock, /dispatch_code_factory/);
  assert.match(lock, /"dispatch-code-factory":\{"description":"Dispatch code-factory for each created issue"\}/);
  assert.match(lock, /"dispatch_code_factory"/);
  assert.match(lock, /"labels":\["flaky-test"\]/);
});
