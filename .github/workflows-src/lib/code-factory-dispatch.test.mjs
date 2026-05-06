import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { validateDispatchInputs, normalizeIssueEventContext } = require('./code-factory-dispatch.js');

test('validateDispatchInputs accepts valid issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '42',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /issue #42/);
  assert.equal(result.issue_number, 42);
});

test('validateDispatchInputs rejects non-numeric issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: 'abc',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects zero issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '0',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects negative issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '-1',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects empty issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects decimal issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '3.14',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects issue number with leading zeros mismatch', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '007',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('normalizeIssueEventContext extracts issue fields from issues event', () => {
  const result = normalizeIssueEventContext({
    eventName: 'issues',
    payload: {
      issue: { number: 42, title: 'Hello', body: 'World' },
    },
  });

  assert.equal(result.issue_number, 42);
  assert.equal(result.issue_title, 'Hello');
  assert.equal(result.issue_body, 'World');
});

test('normalizeIssueEventContext returns null number for non-issues event', () => {
  const result = normalizeIssueEventContext({
    eventName: 'workflow_dispatch',
    payload: {},
  });

  assert.equal(result.issue_number, null);
  assert.equal(result.issue_title, '');
  assert.equal(result.issue_body, '');
});

test('normalizeIssueEventContext handles missing issue fields', () => {
  const result = normalizeIssueEventContext({
    eventName: 'issues',
    payload: {
      issue: {},
    },
  });

  assert.equal(result.issue_number, null);
  assert.equal(result.issue_title, '');
  assert.equal(result.issue_body, '');
});
