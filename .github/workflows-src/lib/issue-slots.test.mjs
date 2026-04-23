import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { computeIssueSlots } = require('./issue-slots.js');

test('computeIssueSlots returns slots and static open-issues output', () => {
  const result = computeIssueSlots({
    label: 'schema-coverage',
    issueCap: '3',
    openIssueCount: 1,
  });

  assert.equal(result.open_issues, 1);
  assert.equal(result.issue_slots_available, 2);
  assert.ok(result.gate_reason.includes('schema-coverage'));
  assert.ok(result.gate_reason.includes('3'));
});

test('computeIssueSlots returns zero slots when at the cap', () => {
  const result = computeIssueSlots({
    label: 'duplicate-code',
    issueCap: 3,
    openIssueCount: 3,
  });

  assert.equal(result.open_issues, 3);
  assert.equal(result.issue_slots_available, 0);
  assert.ok(result.gate_reason.includes('skipped'));
});

test('computeIssueSlots never returns negative slots', () => {
  const result = computeIssueSlots({
    label: 'duplicate-code',
    issueCap: 3,
    openIssueCount: 100,
  });

  assert.equal(result.issue_slots_available, 0);
});

test('computeIssueSlots rejects invalid label', () => {
  assert.throws(
    () => computeIssueSlots({ label: '   ', issueCap: 3, openIssueCount: 0 }),
    /non-empty string/,
  );
});

test('computeIssueSlots rejects invalid cap', () => {
  assert.throws(
    () => computeIssueSlots({ label: 'schema-coverage', issueCap: 'abc', openIssueCount: 0 }),
    /non-negative integer/,
  );
});
