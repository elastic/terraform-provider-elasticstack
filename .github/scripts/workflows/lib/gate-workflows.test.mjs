import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { gateWorkflows } = require(path.resolve(__dirname, 'gate-workflows.js'));

test('test-passed (classify=true, test=success) → passed', () => {
  const result = gateWorkflows({
    classifyResult: 'true',
    testResult: 'success',
  });
  assert.equal(result.passed, true);
  assert.match(result.reason, /Workflow tests succeeded/);
});

test('test-skipped-legitimately (classify=false, test=skipped) → passed', () => {
  const result = gateWorkflows({
    classifyResult: 'false',
    testResult: 'skipped',
  });
  assert.equal(result.passed, true);
  assert.match(result.reason, /legitimately skipped/);
});

test('unexpected-skip (classify=true, test=skipped) → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'true',
    testResult: 'skipped',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Unexpected skip/);
});

test('test-failed (classify=true, test=failure) → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'true',
    testResult: 'failure',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failure/);
});

// Edge cases

test('test-cancelled (classify=true, test=cancelled) → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'true',
    testResult: 'cancelled',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /cancelled/);
});

test('invalid classifyResult → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'maybe',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Invalid classify result/);
});

test('invalid testResult → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'true',
    testResult: 'unknown',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Invalid test result/);
});

test('classify=false, test=success → passed', () => {
  const result = gateWorkflows({
    classifyResult: 'false',
    testResult: 'success',
  });
  assert.equal(result.passed, true);
  assert.match(result.reason, /Workflow tests succeeded/);
});

test('classify=false, test=failure → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'false',
    testResult: 'failure',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failure/);
});

test('classify=false, test=cancelled → failed', () => {
  const result = gateWorkflows({
    classifyResult: 'false',
    testResult: 'cancelled',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /cancelled/);
});
