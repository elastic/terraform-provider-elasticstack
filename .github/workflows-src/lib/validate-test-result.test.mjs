import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { validateTestResult } = require(path.resolve(__dirname, 'validate-test-result.js'));

// Note: the preflight-skip path (needs.preflight.outputs.should_run=false) is not
// tested here because the workflow skips the test-validation job entirely in that
// case. Only post-preflight states are reachable by this function.

test('providerChanges === "" → passed false, reason mentions "did not produce a valid output"', () => {
  const result = validateTestResult({ providerChanges: '', testResult: 'success' });
  assert.equal(result.passed, false);
  assert.match(result.reason, /did not produce a valid output/);
});

test('providerChanges === false, testResult === skipped → passed true', () => {
  const result = validateTestResult({ providerChanges: 'false', testResult: 'skipped' });
  assert.equal(result.passed, true);
});

test('providerChanges === false, testResult === success → passed true', () => {
  const result = validateTestResult({ providerChanges: 'false', testResult: 'success' });
  assert.equal(result.passed, true);
});

test('providerChanges === false, testResult === failure → passed false', () => {
  const result = validateTestResult({ providerChanges: 'false', testResult: 'failure' });
  assert.equal(result.passed, false);
});

test('providerChanges === false, testResult === cancelled → passed false', () => {
  const result = validateTestResult({ providerChanges: 'false', testResult: 'cancelled' });
  assert.equal(result.passed, false);
});

test('providerChanges === true, testResult === success → passed true', () => {
  const result = validateTestResult({ providerChanges: 'true', testResult: 'success' });
  assert.equal(result.passed, true);
});

test('providerChanges === true, testResult === failure → passed false', () => {
  const result = validateTestResult({ providerChanges: 'true', testResult: 'failure' });
  assert.equal(result.passed, false);
});

test('providerChanges === true, testResult === skipped → passed false', () => {
  const result = validateTestResult({ providerChanges: 'true', testResult: 'skipped' });
  assert.equal(result.passed, false);
});

test('providerChanges === true, testResult === cancelled → passed false', () => {
  const result = validateTestResult({ providerChanges: 'true', testResult: 'cancelled' });
  assert.equal(result.passed, false);
});
