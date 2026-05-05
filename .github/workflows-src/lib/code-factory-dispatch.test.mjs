import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { validateDispatchInputs, normalizeIssueEventContext } = require('./code-factory-dispatch.js');

test('validateDispatchInputs accepts matching repository and valid issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '42',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /issue #42/);
  assert.equal(result.issue_number, 42);
});

test('validateDispatchInputs rejects cross-repository dispatch', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '42',
    dispatchIssueRepo: 'someone/else',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /Cross-repository dispatch is not supported/);
  assert.equal(result.issue_number, undefined);
});

test('validateDispatchInputs rejects non-numeric issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: 'abc',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects zero issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '0',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects negative issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '-1',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects empty issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects decimal issue number', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '3.14',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
    currentRepository: 'elastic/terraform-provider-elasticstack',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not a valid positive integer/);
});

test('validateDispatchInputs rejects issue number with leading zeros mismatch', () => {
  const result = validateDispatchInputs({
    dispatchIssueNumber: '007',
    dispatchIssueRepo: 'elastic/terraform-provider-elasticstack',
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
