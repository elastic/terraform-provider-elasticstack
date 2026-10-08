import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const gate = require('../research-factory/gate.js');

const READY = 'ready-for-change-factory';
const HUMAN = 'research-needs-human';

function metadata(overrides = {}) {
  const base = {
    schema_version: '1.1',
    recommendation: { spine: 'spine-a', approach_index: 0 },
    open_questions: [],
    estimated_scope: 'small',
    gate: {
      outcome: READY,
      checklist: { grounded: true, mapped: true, compatible: true, versioned: true, testable: true, idiomatic: true },
      score: 88,
      scores: [86, 88],
      converged: true,
      rounds: 2,
      outstanding_feedback: ['minor nit'],
      author_model: 'anthropic/claude-sonnet-5',
      critic: { model: 'openai/gpt-6.1-sol', status: 'ok' },
    },
  };
  const { gate: gateOverrides, ...rest } = overrides;
  return { ...base, ...rest, gate: { ...base.gate, ...gateOverrides } };
}

function commentBody(meta, { section = true } = {}) {
  const qualityGate = section
    ? `### Quality gate\n\n**Outcome:** \`${meta?.gate?.outcome}\` - reported reason\n\n| Item | Result |\n|---|---|\n| Grounded | pass |\n\n### References\n\n- a\n\n`
    : `### References\n\n- a\n\n`;
  const json = typeof meta === 'string' ? meta : JSON.stringify(meta, null, 2);
  return `## Implementation research\n\n### Out of scope\n\n- none\n\n${qualityGate}<details>\n<summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${json}\n\`\`\`\n\n</details>\n`;
}

test('exports the gate constants', () => {
  assert.equal(gate.SCORE_THRESHOLD, 85);
  assert.equal(gate.STABILITY_WINDOW, 2);
  assert.equal(gate.MAX_ROUNDS, 3);
});

test('extractMetadata parses the JSON block from the details element', () => {
  const meta = metadata();
  assert.deepEqual(gate.extractMetadata(commentBody(meta)), { metadata: meta });
});

test('extractMetadata reports an error when the block is absent', () => {
  const result = gate.extractMetadata('## Implementation research\n\nno metadata');
  assert.equal(result.metadata, undefined);
  assert.match(result.error, /metadata/i);
});

test('extractMetadata reports an error when the JSON is unparseable', () => {
  const result = gate.extractMetadata(commentBody('{ not json'));
  assert.equal(result.metadata, undefined);
  assert.match(result.error, /parse/i);
});

test('validateGate accepts valid 1.1 metadata', () => {
  assert.deepEqual(gate.validateGate(metadata()), []);
});

test('validateGate accepts the no-critique-round shape (rounds 0, score null)', () => {
  const meta = metadata({
    gate: { score: null, scores: [], rounds: 0, converged: false, outstanding_feedback: [], critic: { model: 'openai/gpt-6.1-sol', status: 'unavailable' } },
  });
  assert.deepEqual(gate.validateGate(meta), []);
});

const invalidCases = {
  'schema_version 1.0': metadata({ schema_version: '1.0' }),
  'missing gate': (() => { const m = metadata(); delete m.gate; return m; })(),
  'bad outcome enum': metadata({ gate: { outcome: 'maybe' } }),
  'checklist item missing': metadata({ gate: { checklist: { grounded: true } } }),
  'checklist non-boolean': metadata({ gate: { checklist: { grounded: 'yes', mapped: true, compatible: true, versioned: true, testable: true, idiomatic: true } } }),
  'score out of range': metadata({ gate: { score: 120 } }),
  'scores not array': metadata({ gate: { scores: 'x' } }),
  'rounds above max': metadata({ gate: { rounds: 4, scores: [90, 90, 90, 90] } }),
  'rounds not equal to scores length': metadata({ gate: { rounds: 1 } }),
  'converged not boolean': metadata({ gate: { converged: 'true' } }),
  'feedback not array': metadata({ gate: { outstanding_feedback: 'x' } }),
  'author_model missing': metadata({ gate: { author_model: '' } }),
  'critic status invalid': metadata({ gate: { critic: { model: 'm', status: 'weird' } } }),
  'critic model missing': metadata({ gate: { critic: { status: 'ok' } } }),
  'recommendation missing': (() => { const m = metadata(); delete m.recommendation; return m; })(),
  'open question blocking not boolean': metadata({ open_questions: [{ id: 'oq-1', text: 't', blocking: 'no' }] }),
};

for (const [name, meta] of Object.entries(invalidCases)) {
  test(`validateGate rejects: ${name}`, () => {
    assert.ok(gate.validateGate(meta).length > 0);
  });
}

test('validateGate rejects non-object metadata', () => {
  assert.ok(gate.validateGate(null).length > 0);
});

const convergenceCases = [
  ['single round 90 with no feedback converges', [90], [], true],
  ['plateau below threshold does not converge', [70], [], false],
  ['80 then 88 with feedback does not converge', [80, 88], ['x'], false],
  ['two consecutive rounds at threshold converge despite feedback', [85, 85], ['x'], true],
  ['high then low does not converge', [90, 80], [], false],
  ['empty scores do not converge', [], [], false],
  ['single high round with feedback does not converge', [90], ['x'], false],
];

for (const [name, scores, feedback, expected] of convergenceCases) {
  test(`isConverged: ${name}`, () => {
    assert.equal(gate.isConverged(scores, feedback), expected);
  });
}

const deriveCases = [
  ['consistent ready metadata', metadata(), READY, false],
  ['failed checklist item despite reported ready', metadata({ gate: { checklist: { grounded: true, mapped: true, compatible: true, versioned: false, testable: true, idiomatic: true } } }), HUMAN, true],
  ['blocking open question', metadata({ open_questions: [{ id: 'oq-1', text: 't', blocking: true }] }), HUMAN, true],
  ['claimed convergence unsupported by scores', metadata({ gate: { scores: [80, 88], score: 88, converged: true, outstanding_feedback: ['x'] } }), HUMAN, true],
  ['critic unavailable with no rounds', metadata({ gate: { outcome: HUMAN, score: null, scores: [], rounds: 0, converged: false, outstanding_feedback: [], critic: { model: 'm', status: 'unavailable' } } }), HUMAN, false],
  ['critic status error but otherwise ready', metadata({ gate: { critic: { model: 'm', status: 'error' } } }), HUMAN, true],
  ['score field inconsistent with last scores entry', metadata({ gate: { score: 99 } }), HUMAN, true],
  ['non-blocking open question still ready', metadata({ open_questions: [{ id: 'oq-1', text: 't', blocking: false }] }), READY, false],
  ['reported needs-human but gate is satisfied is overridden to ready', metadata({ gate: { outcome: HUMAN } }), READY, true],
];

for (const [name, meta, label, overridden] of deriveCases) {
  test(`deriveOutcome: ${name}`, () => {
    const result = gate.deriveOutcome(meta);
    assert.equal(result.label, label);
    assert.equal(result.overridden, overridden);
    if (label === HUMAN) {
      assert.ok(result.reasons.length > 0);
    }
  });
}

test('deriveOutcome fails safe to research-needs-human for missing metadata', () => {
  const result = gate.deriveOutcome(undefined, { error: 'no metadata block' });
  assert.equal(result.label, HUMAN);
  assert.match(result.reasons.join(' '), /no metadata block/);
});

test('deriveOutcome fails safe for schema-invalid metadata and lists the validation errors', () => {
  const result = gate.deriveOutcome(metadata({ schema_version: '1.0' }));
  assert.equal(result.label, HUMAN);
  assert.match(result.reasons.join(' '), /schema_version/);
});

test('deriveOutcome with exactly three rounds is allowed', () => {
  const meta = metadata({ gate: { scores: [70, 86, 90], score: 90, rounds: 3 } });
  assert.equal(gate.deriveOutcome(meta).label, READY);
});

test('evaluateBody derives from the comment body', () => {
  assert.equal(gate.evaluateBody(commentBody(metadata())).label, READY);
  assert.equal(gate.evaluateBody('no metadata').label, HUMAN);
});

test('evaluateBody lists extraction errors as the reason', () => {
  assert.match(gate.evaluateBody('no metadata').reasons.join(' '), /metadata/i);
});

// ---------------------------------------------------------------------------
// applyOverride
// ---------------------------------------------------------------------------

test('applyOverride leaves a non-overridden body untouched', () => {
  const body = commentBody(metadata());
  assert.equal(gate.applyOverride(body, gate.evaluateBody(body)), body);
});

test('applyOverride rewrites the outcome line, the JSON gate.outcome, and adds a visible note', () => {
  const meta = metadata({ gate: { checklist: { grounded: true, mapped: true, compatible: true, versioned: false, testable: true, idiomatic: true } } });
  const body = commentBody(meta);
  const result = gate.evaluateBody(body);
  const corrected = gate.applyOverride(body, result);

  assert.match(corrected, /\*\*Outcome:\*\* `research-needs-human`/);
  assert.doesNotMatch(corrected, /\*\*Outcome:\*\* `ready-for-change-factory`/);
  assert.match(corrected, /overridden by the gate rule/);
  assert.match(corrected, /versioned/i);
  assert.equal(gate.extractMetadata(corrected).metadata.gate.outcome, HUMAN);
  assert.equal(gate.evaluateBody(corrected).overridden, false);
});

test('applyOverride inserts a minimal Quality gate section before References when missing', () => {
  const meta = metadata({ gate: { outcome: READY, critic: { model: 'm', status: 'error' } } });
  const body = commentBody(meta, { section: false });
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));

  const gateIndex = corrected.indexOf('### Quality gate');
  const refsIndex = corrected.indexOf('### References');
  assert.ok(gateIndex !== -1 && gateIndex < refsIndex);
  assert.match(corrected, /\*\*Outcome:\*\* `research-needs-human`/);
  assert.match(corrected, /overridden by the gate rule/);
});

test('applyOverride corrects metadata that was missing by leaving the body and adding a note', () => {
  const body = '## Implementation research\n\n### References\n\n- a\n';
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));
  assert.match(corrected, /### Quality gate/);
  assert.match(corrected, /\*\*Outcome:\*\* `research-needs-human`/);
});

test('applyOverride also corrects a reported needs-human that the gate rule satisfies', () => {
  const body = commentBody(metadata({ gate: { outcome: HUMAN } }));
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));
  assert.match(corrected, /\*\*Outcome:\*\* `ready-for-change-factory`/);
  assert.equal(gate.extractMetadata(corrected).metadata.gate.outcome, READY);
});

// ---------------------------------------------------------------------------
// Review-round additions
// ---------------------------------------------------------------------------

for (const [name, json] of [['null', 'null'], ['an array', '[]'], ['a string', '"x"']]) {
  test(`evaluateBody fails safe when the metadata JSON is ${name}`, () => {
    const result = gate.evaluateBody(commentBody(json));
    assert.equal(result.label, HUMAN);
    assert.ok(result.reasons.length > 0);
  });
}

test('evaluateBody fails safe when gate is null', () => {
  const meta = metadata();
  meta.gate = null;
  assert.equal(gate.evaluateBody(commentBody(meta)).label, HUMAN);
});

test('applyOverride inserts replacement text literally (no $ pattern expansion)', () => {
  const meta = metadata({ open_questions: [{ id: "oq-$&-$1", text: 't', blocking: true }] });
  const body = commentBody(meta);
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));
  const outcomeLine = corrected.split('\n').find((l) => l.startsWith('**Outcome:**'));
  assert.doesNotMatch(outcomeLine, /\$|### Quality gate/);
  assert.equal(corrected.match(/\*\*Outcome:\*\*/g).length, 1);
});

test('reasons are sanitised: no newlines or backticks, bounded length, safe ids', () => {
  const meta = metadata({ open_questions: [{ id: 'oq`\n# Injected `x`', text: 't', blocking: true }] });
  const result = gate.deriveOutcome(meta);
  const joined = result.reasons.join('|');
  assert.doesNotMatch(joined, /[`\n\r]/);
  assert.doesNotMatch(joined, /#\s*Injected/);
  const long = gate.deriveOutcome(undefined, { error: 'x'.repeat(1000) });
  assert.ok(long.reasons[0].length <= 200);
});

test('a Quality gate outcome line that disagrees with the label is corrected even when gate.outcome matches', () => {
  const body = commentBody(metadata()).replace(
    /\*\*Outcome:\*\* `[^`]*`/,
    '**Outcome:** `research-needs-human`',
  );
  const result = gate.evaluateBody(body);
  assert.equal(result.label, READY);
  assert.equal(result.overridden, true);
  const corrected = gate.applyOverride(body, result);
  assert.match(corrected, /\*\*Outcome:\*\* `ready-for-change-factory`/);
  assert.equal(gate.evaluateBody(corrected).overridden, false);
});

test('a matching Quality gate outcome line is not an override', () => {
  assert.equal(gate.evaluateBody(commentBody(metadata())).overridden, false);
});

test('applyOverride adds an outcome line when the Quality gate section has none', () => {
  const meta = metadata({ gate: { critic: { model: 'm', status: 'error' } } });
  const body = commentBody(meta).replace(/\*\*Outcome:\*\* .*\n/, '');
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));
  assert.match(corrected, /### Quality gate\s+\*\*Outcome:\*\* `research-needs-human`/);
});

test('applyOverride appends a section at the end when there is no References heading', () => {
  const body = '## Implementation research\n\nplain';
  const corrected = gate.applyOverride(body, gate.evaluateBody(body));
  assert.ok(corrected.trimEnd().endsWith('.') || corrected.includes('### Quality gate'));
  assert.ok(corrected.indexOf('### Quality gate') > corrected.indexOf('plain'));
});

test('applyOverride is idempotent', () => {
  const meta = metadata({ gate: { checklist: { grounded: false, mapped: true, compatible: true, versioned: true, testable: true, idiomatic: true } } });
  const body = commentBody(meta);
  const once = gate.applyOverride(body, gate.evaluateBody(body));
  const twice = gate.applyOverride(once, gate.evaluateBody(once));
  assert.equal(twice, once);
});

test('isConverged: [70, 90] with feedback does not converge', () => {
  assert.equal(gate.isConverged([70, 90], ['x']), false);
});

test('isConverged: [90, 90] with feedback converges', () => {
  assert.equal(gate.isConverged([90, 90], ['x']), true);
});

test('isConverged: 84 fails and 85 passes the threshold', () => {
  assert.equal(gate.isConverged([84], []), false);
  assert.equal(gate.isConverged([85], []), true);
});

test('extractMetadata takes the last JSON fence', () => {
  const meta = metadata();
  const body = `intro\n\n\`\`\`json\n{"a": 1}\n\`\`\`\n\n${commentBody(meta)}`;
  assert.deepEqual(gate.extractMetadata(body), { metadata: meta });
});

test('validateGate accepts metadata without open_questions', () => {
  const meta = metadata();
  delete meta.open_questions;
  assert.deepEqual(gate.validateGate(meta), []);
  assert.equal(gate.deriveOutcome(meta).label, READY);
});

// ---------------------------------------------------------------------------
// Full schema 1.1 validation and outcome-line forms
// ---------------------------------------------------------------------------

const moreInvalid = {
  'spine not kebab-case': { recommendation: { spine: 'Not Kebab', approach_index: 0 } },
  'approach_index negative': { recommendation: { spine: 'a', approach_index: -1 } },
  'approach_index fractional': { recommendation: { spine: 'a', approach_index: 1.5 } },
  'confidence bad enum': { recommendation: { spine: 'a', approach_index: 0, confidence: 'certain' } },
  'estimated_scope bad enum': { estimated_scope: 'huge' },
  'estimated_scope not string': { estimated_scope: 3 },
  'affected_capabilities numeric entries': { affected_capabilities: [1, 2] },
  'affected_capabilities not kebab': { affected_capabilities: ['Not Kebab'] },
  'affected_capabilities not array': { affected_capabilities: 'x' },
  'references not array': { references: 'x' },
  'reference bad type': { references: [{ type: 'blog', url: 'u' }] },
  'reference missing location': { references: [{ type: 'repo-path' }] },
  'reference empty location': { references: [{ type: 'elastic-docs', url: '' }] },
  'reference not an object': { references: ['x'] },
};

for (const [name, overrides] of Object.entries(moreInvalid)) {
  test(`validateGate rejects: ${name}`, () => {
    assert.ok(gate.validateGate(metadata(overrides)).length > 0);
  });
}

test('validateGate accepts fully populated optional fields', () => {
  const meta = metadata({
    recommendation: { spine: 'new-resource-a1', approach_index: 2, confidence: 'high' },
    estimated_scope: 'unknown',
    affected_capabilities: ['kibana-foo', 'fleet-bar'],
    references: [{ type: 'elastic-docs', url: 'https://x' }, { type: 'repo-path', path: 'a/b.go' }],
  });
  assert.deepEqual(gate.validateGate(meta), []);
});

test('an unbackticked outcome line is recognised and normalised to the backticked form', () => {
  const body = commentBody(metadata()).replace(/\*\*Outcome:\*\* `[^`]*`/, '**Outcome:** research-needs-human');
  const result = gate.evaluateBody(body);
  assert.equal(result.overridden, true);
  const corrected = gate.applyOverride(body, result);
  assert.match(corrected, /\*\*Outcome:\*\* `ready-for-change-factory`/);
  assert.doesNotMatch(corrected, /\*\*Outcome:\*\* research-needs-human/);
});

test('an unbackticked outcome line that matches the label is not an override', () => {
  const body = commentBody(metadata()).replace(/\*\*Outcome:\*\* `([^`]*)`/, '**Outcome:** $1');
  assert.equal(gate.evaluateBody(body).overridden, false);
});
