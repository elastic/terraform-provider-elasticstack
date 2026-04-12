import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { computeIssueSlots, ISSUE_CAP, SCHEMA_COVERAGE_LABEL } = require('./schema-coverage-slots.js');

// ---------------------------------------------------------------------------
// computeIssueSlots — below-cap cases
// ---------------------------------------------------------------------------

test('computeIssueSlots returns correct slots when no issues are open', () => {
  const result = computeIssueSlots(0);
  assert.equal(result.open_schema_coverage_issues, 0);
  assert.equal(result.issue_slots_available, 3);
  assert.ok(result.gate_reason.length > 0, 'expected a non-empty gate_reason');
  assert.ok(!result.gate_reason.includes('skipped'), 'expected gate_reason to not mention skipped when slots available');
});

test('computeIssueSlots returns correct slots when one issue is open', () => {
  const result = computeIssueSlots(1);
  assert.equal(result.open_schema_coverage_issues, 1);
  assert.equal(result.issue_slots_available, 2);
});

test('computeIssueSlots returns one slot when two issues are open', () => {
  const result = computeIssueSlots(2);
  assert.equal(result.open_schema_coverage_issues, 2);
  assert.equal(result.issue_slots_available, 1);
});

// ---------------------------------------------------------------------------
// computeIssueSlots — at-cap cases
// ---------------------------------------------------------------------------

test('computeIssueSlots returns zero slots when at the cap', () => {
  const result = computeIssueSlots(3);
  assert.equal(result.open_schema_coverage_issues, 3);
  assert.equal(result.issue_slots_available, 0);
  assert.ok(result.gate_reason.includes('skipped'), 'expected gate_reason to mention agent job skipped');
});

test('computeIssueSlots returns zero slots when above the cap', () => {
  const result = computeIssueSlots(5);
  assert.equal(result.open_schema_coverage_issues, 5);
  assert.equal(result.issue_slots_available, 0);
});

test('computeIssueSlots never returns negative slots', () => {
  const result = computeIssueSlots(100);
  assert.equal(result.issue_slots_available, 0);
});

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

test('ISSUE_CAP is 3', () => {
  assert.equal(ISSUE_CAP, 3);
});

test('SCHEMA_COVERAGE_LABEL is schema-coverage', () => {
  assert.equal(SCHEMA_COVERAGE_LABEL, 'schema-coverage');
});

// ---------------------------------------------------------------------------
// gate_reason content
// ---------------------------------------------------------------------------

test('computeIssueSlots gate_reason includes open count and cap for below-cap case', () => {
  const result = computeIssueSlots(1);
  assert.ok(result.gate_reason.includes('1'), 'expected open count in gate_reason');
  assert.ok(result.gate_reason.includes('3'), 'expected cap in gate_reason');
});

test('computeIssueSlots gate_reason includes open count and cap for at-cap case', () => {
  const result = computeIssueSlots(3);
  assert.ok(result.gate_reason.includes('3'), 'expected count/cap in gate_reason');
});
