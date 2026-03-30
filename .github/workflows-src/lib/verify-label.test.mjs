import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { verifyLabel } = require('./verify-label.js');

test('verifyLabel accepts the expected label', () => {
  assert.deepEqual(verifyLabel('verify-openspec'), {
    label_verified: 'true',
    label_reason: 'Label verified: verify-openspec',
    log_message: 'Label verified: verify-openspec',
  });
});

test('verifyLabel rejects unexpected labels', () => {
  assert.deepEqual(verifyLabel('other-label'), {
    label_verified: 'false',
    label_reason: 'Unexpected label: other-label',
    log_message: 'Label check failed: expected verify-openspec, got other-label',
  });
});
