import assert from 'node:assert/strict';
import test from 'node:test';
import { ISSUE_CAP, kibanaSpecImpactGate } from './kibana-spec-impact-gate.js';

test('gate triggers when kbapi symbols changed', () => {
  const g = kibanaSpecImpactGate({
    changed_kbapi_symbols: ['SomeType'],
    transform_schema_hints: [],
    high_confidence_impacts: [],
  });
  assert.equal(g.shouldRun, true);
  assert.equal(g.issueCap, 0);
});

test('gate triggers on transform hints only', () => {
  const g = kibanaSpecImpactGate({
    changed_kbapi_symbols: [],
    transform_schema_hints: ['internal/kibana/foo/transform_schema.go'],
    high_confidence_impacts: [],
  });
  assert.equal(g.shouldRun, true);
});

test('gate caps issues at ISSUE_CAP', () => {
  const hi = Array.from({ length: 20 }, (_, i) => ({ entity_name: `e${i}` }));
  const g = kibanaSpecImpactGate({
    changed_kbapi_symbols: ['A'],
    transform_schema_hints: [],
    high_confidence_impacts: hi,
  });
  assert.equal(g.issueCap, ISSUE_CAP);
});

test('idle repo has no work', () => {
  const g = kibanaSpecImpactGate({
    changed_kbapi_symbols: [],
    transform_schema_hints: [],
    high_confidence_impacts: [],
  });
  assert.equal(g.shouldRun, false);
});
