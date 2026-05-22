import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { gateProvider } = require(path.resolve(__dirname, 'gate-provider.js'));

test('all-pass (classify=true, all success) → passed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'success',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, true);
  assert.match(result.reason, /all jobs succeeded/);
});

test('all-skipped-legitimately (classify=false, all skipped) → passed', () => {
  const result = gateProvider({
    classifyResult: 'false',
    buildResult: 'skipped',
    lintResult: 'skipped',
    golangciLintResult: 'skipped',
    testResult: 'skipped',
  });
  assert.equal(result.passed, true);
  assert.match(result.reason, /legitimately skipped/);
});

test('unexpected-skip (classify=true, build=skipped, others=success) → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'skipped',
    lintResult: 'success',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Unexpected skip/);
});

test('any-failure (classify=true, build=failure) → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'failure',
    lintResult: 'success',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
});

test('cancelled (classify=true, test=cancelled) → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'success',
    golangciLintResult: 'success',
    testResult: 'cancelled',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
});

// Edge cases

test('invalid classifyResult → failed', () => {
  const result = gateProvider({
    classifyResult: 'maybe',
    buildResult: 'success',
    lintResult: 'success',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Invalid classify result/);
});

test('invalid job result → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'unknown',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Invalid job result/);
});

test('classify=false but not all skipped → failed', () => {
  const result = gateProvider({
    classifyResult: 'false',
    buildResult: 'skipped',
    lintResult: 'skipped',
    golangciLintResult: 'skipped',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
});

test('lint failure → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'failure',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
});

test('golangci-lint failure → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'success',
    golangciLintResult: 'failure',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
});

test('all cancelled → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'cancelled',
    lintResult: 'cancelled',
    golangciLintResult: 'cancelled',
    testResult: 'cancelled',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
});

test('failure takes priority over unexpected skip', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'failure',
    lintResult: 'skipped',
    golangciLintResult: 'success',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /failed or were cancelled/);
  assert.doesNotMatch(result.reason, /Unexpected skip/);
});

test('golangci-lint unexpectedly skipped → failed', () => {
  const result = gateProvider({
    classifyResult: 'true',
    buildResult: 'success',
    lintResult: 'success',
    golangciLintResult: 'skipped',
    testResult: 'success',
  });
  assert.equal(result.passed, false);
  assert.match(result.reason, /Unexpected skip/);
});