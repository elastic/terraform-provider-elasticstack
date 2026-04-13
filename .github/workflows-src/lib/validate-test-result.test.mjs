import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { validateTestResult } = require(path.resolve(__dirname, 'validate-test-result.js'));

test('preflightShouldRun === false → passed true (regardless of other inputs)', () => {
  const result = validateTestResult({ preflightShouldRun: 'false', providerChanges: '', testResult: '' });
  assert.equal(result.passed, true);
});

test('preflightShouldRun === false → passed true even with providerChanges and testResult set', () => {
  const result = validateTestResult({ preflightShouldRun: 'false', providerChanges: 'true', testResult: 'failure' });
  assert.equal(result.passed, true);
});

test('preflightShouldRun === true, providerChanges === "" → passed false, reason mentions "did not produce a valid output"', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: '', testResult: 'success' });
  assert.equal(result.passed, false);
  assert.match(result.reason, /did not produce a valid output/);
});

test('preflightShouldRun === true, providerChanges === false, testResult === skipped → passed true', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'false', testResult: 'skipped' });
  assert.equal(result.passed, true);
});

test('preflightShouldRun === true, providerChanges === false, testResult === success → passed true', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'false', testResult: 'success' });
  assert.equal(result.passed, true);
});

test('preflightShouldRun === true, providerChanges === false, testResult === failure → passed false', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'false', testResult: 'failure' });
  assert.equal(result.passed, false);
});

test('preflightShouldRun === true, providerChanges === false, testResult === cancelled → passed false', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'false', testResult: 'cancelled' });
  assert.equal(result.passed, false);
});

test('preflightShouldRun === true, providerChanges === true, testResult === success → passed true', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'true', testResult: 'success' });
  assert.equal(result.passed, true);
});

test('preflightShouldRun === true, providerChanges === true, testResult === failure → passed false', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'true', testResult: 'failure' });
  assert.equal(result.passed, false);
});

test('preflightShouldRun === true, providerChanges === true, testResult === skipped → passed false', () => {
  const result = validateTestResult({ preflightShouldRun: 'true', providerChanges: 'true', testResult: 'skipped' });
  assert.equal(result.passed, false);
});
