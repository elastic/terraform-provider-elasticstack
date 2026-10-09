import assert from 'node:assert/strict';
import test, { afterEach } from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';
import { sanitizers, sanitizeParts } from './research-factory-sanitizer-helper.mjs';

const require = createRequire(import.meta.url);
const updateResearchComment = require('../research-factory/update-research-comment.js');

const originalEnv = {
  GH_AW_AGENT_OUTPUT: process.env.GH_AW_AGENT_OUTPUT,
  RESEARCH_FACTORY_ISSUE_NUMBER: process.env.RESEARCH_FACTORY_ISSUE_NUMBER,
};

afterEach(() => {
  for (const [key, value] of Object.entries(originalEnv)) {
    if (value === undefined) {
      delete process.env[key];
    } else {
      process.env[key] = value;
    }
  }
});

const READY = 'ready-for-change-factory';
const HUMAN = 'research-needs-human';

function readyMetadata(overrides = {}) {
  return {
    schema_version: '1.1',
    recommendation: { spine: 'spine-a', approach_index: 0 },
    gate: {
      outcome: READY,
      checklist: { grounded: true, mapped: true, compatible: true, versioned: true, testable: true, idiomatic: true },
      score: 88,
      scores: [86, 88],
      converged: true,
      rounds: 2,
      outstanding_feedback: [],
      author_model: 'a',
      critic: { model: 'c', status: 'ok' },
    },
    ...overrides,
  };
}

function body(meta) {
  return `## Implementation research\n\n### Quality gate\n\n**Outcome:** \`${meta.gate.outcome}\` - reported\n\n### References\n\n- a\n\n<details>\n<summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${JSON.stringify(meta, null, 2)}\n\`\`\`\n\n</details>\n`;
}

function setup({ meta, existingComments = [], failLabel = null, rawBody } = {}) {
  const calls = [];
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'urc-'));
  const outputFile = path.join(dir, 'out.json');
  const commentBody = rawBody ?? body(meta);
  fs.writeFileSync(outputFile, JSON.stringify({ items: [{ type: 'update_research_comment', body: commentBody }] }));
  process.env.GH_AW_AGENT_OUTPUT = outputFile;
  process.env.RESEARCH_FACTORY_ISSUE_NUMBER = '42';

  const issues = {
    listComments: 'listComments',
    createComment: async (args) => { calls.push(['createComment', args]); return { data: { id: 1 } }; },
    updateComment: async (args) => { calls.push(['updateComment', args]); return {}; },
    addLabels: async (args) => {
      calls.push(['addLabels', args]);
      if (failLabel === 'add') throw new Error('add boom');
      return {};
    },
    removeLabel: async (args) => {
      calls.push(['removeLabel', args]);
      if (failLabel === 'remove') throw Object.assign(new Error('remove boom'), { status: 500 });
      if (failLabel === 'absent') throw Object.assign(new Error('nf'), { status: 404 });
      return {};
    },
  };
  const github = {
    rest: { issues },
    paginate: async () => existingComments,
  };
  const summary = [];
  const core = {
    infos: [],
    failures: [],
    info(m) { this.infos.push(m); },
    setFailed(m) { this.failures.push(m); },
    summary: {
      addRaw(t) { summary.push(t); return this.chain; },
      chain: null,
      async write() {},
    },
  };
  core.summary.chain = core.summary;
  const context = { repo: { owner: 'o', repo: 'r' } };
  return { calls, github, core, context, summary };
}

const names = (calls) => calls.map(([n]) => n);

test('posts the comment before setting labels and applies exactly the derived label', async () => {
  const env = setup({ meta: readyMetadata() });
  await updateResearchComment(env);

  assert.deepEqual(names(env.calls), ['createComment', 'addLabels', 'removeLabel']);
  assert.deepEqual(env.calls[1][1].labels, [READY]);
  assert.equal(env.calls[2][1].name, HUMAN);
  assert.equal(env.core.failures.length, 0);
});

test('updates an existing sticky comment instead of creating one', async () => {
  const env = setup({
    meta: readyMetadata(),
    existingComments: [{ id: 7, user: { login: 'github-actions[bot]' }, body: '<!-- gha-research-factory -->\nold' }],
  });
  await updateResearchComment(env);
  assert.deepEqual(names(env.calls).slice(0, 1), ['updateComment']);
  assert.equal(env.calls[0][1].comment_id, 7);
});

test('an overridden outcome posts the corrected body and the needs-human label', async () => {
  const meta = readyMetadata();
  meta.gate.checklist.versioned = false;
  const env = setup({ meta });
  await updateResearchComment(env);

  const posted = env.calls[0][1].body;
  assert.match(posted, /\*\*Outcome:\*\* `research-needs-human`/);
  assert.match(posted, /overridden by the gate rule/);
  assert.deepEqual(env.calls[1][1].labels, [HUMAN]);
  assert.equal(env.calls[2][1].name, READY);
});

test('missing metadata yields needs-human and the step summary records the override', async () => {
  const env = setup({ rawBody: '## Implementation research\n\nNo metadata here.' });
  await updateResearchComment(env);
  assert.deepEqual(env.calls[1][1].labels, [HUMAN]);
  assert.match(env.summary.join('\n'), /research-needs-human/);
  assert.match(env.summary.join('\n'), /overridden/i);
});

test('step summary records the derived outcome and reasons', async () => {
  const env = setup({ meta: readyMetadata() });
  await updateResearchComment(env);
  const text = env.summary.join('\n');
  assert.match(text, /ready-for-change-factory/);
  assert.match(text, /all gate conditions satisfied/);
});

test('a 404 when removing the other label is tolerated', async () => {
  const env = setup({ meta: readyMetadata(), failLabel: 'absent' });
  await updateResearchComment(env);
  assert.equal(env.core.failures.length, 0);
});

test('a label add failure fails the job and leaves the comment in place', async () => {
  const env = setup({ meta: readyMetadata(), failLabel: 'add' });
  await updateResearchComment(env);
  assert.equal(names(env.calls)[0], 'createComment');
  assert.equal(env.core.failures.length, 1);
  assert.match(env.core.failures[0], /label/i);
});

test('a non-404 label removal failure fails the job', async () => {
  const env = setup({ meta: readyMetadata(), failLabel: 'remove' });
  await updateResearchComment(env);
  assert.equal(env.core.failures.length, 1);
});

test('only the two outcome labels are ever touched', async () => {
  const env = setup({ meta: readyMetadata() });
  await updateResearchComment(env);
  const touched = env.calls
    .filter(([n]) => n === 'addLabels' || n === 'removeLabel')
    .flatMap(([n, a]) => (n === 'addLabels' ? a.labels : [a.name]));
  assert.deepEqual(touched.sort(), [READY, HUMAN].sort());
});

test('does not touch labels when the comment write fails', async () => {
  const env = setup({ meta: readyMetadata() });
  env.github.rest.issues.createComment = async () => { throw new Error('boom'); };
  await updateResearchComment(env);
  assert.deepEqual(names(env.calls), []);
  assert.equal(env.core.failures.length, 1);
});

const MARKER = '<!-- gha-research-factory -->';

function writeOutput(content) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'urc-'));
  const file = path.join(dir, 'out.json');
  fs.writeFileSync(file, typeof content === 'string' ? content : JSON.stringify(content));
  process.env.GH_AW_AGENT_OUTPUT = file;
}

test('fails when GH_AW_AGENT_OUTPUT is unset', async () => {
  const env = setup({ meta: readyMetadata() });
  delete process.env.GH_AW_AGENT_OUTPUT;
  await updateResearchComment(env);
  assert.equal(env.core.failures.length, 1);
  assert.deepEqual(names(env.calls), []);
});

for (const bad of ['abc', '0', '-3', '']) {
  test(`fails on invalid issue number "${bad}"`, async () => {
    const env = setup({ meta: readyMetadata() });
    process.env.RESEARCH_FACTORY_ISSUE_NUMBER = bad;
    await updateResearchComment(env);
    assert.equal(env.core.failures.length, 1);
    assert.deepEqual(names(env.calls), []);
  });
}

test('does nothing when there are no update_research_comment items', async () => {
  const env = setup({ meta: readyMetadata() });
  writeOutput({ items: [{ type: 'noop', body: 'x' }] });
  await updateResearchComment(env);
  assert.deepEqual(names(env.calls), []);
  assert.equal(env.core.failures.length, 0);
});

test('uses the first update_research_comment item when several are present', async () => {
  const env = setup({ meta: readyMetadata() });
  writeOutput({
    items: [
      { type: 'update_research_comment', body: body(readyMetadata()).replace('Implementation research', 'FIRST') },
      { type: 'update_research_comment', body: 'SECOND' },
    ],
  });
  await updateResearchComment(env);
  assert.match(env.calls[0][1].body, /FIRST/);
  assert.equal(names(env.calls).filter((n) => n === 'createComment').length, 1);
});

test('fails via setFailed on malformed agent output JSON', async () => {
  const env = setup({ meta: readyMetadata() });
  writeOutput('{ not json');
  await updateResearchComment(env);
  assert.equal(env.core.failures.length, 1);
  assert.deepEqual(names(env.calls), []);
});

for (const [name, prefix] of [['already first with newline', `${MARKER}\n`], ['CRLF after marker', `${MARKER}\r\n`], ['marker without newline', MARKER], ['absent', '']]) {
  test(`marker normalisation: ${name}`, async () => {
    const env = setup({ meta: readyMetadata() });
    writeOutput({ items: [{ type: 'update_research_comment', body: prefix + body(readyMetadata()) }] });
    await updateResearchComment(env);
    const posted = env.calls[0][1].body;
    assert.ok(posted.startsWith(MARKER), 'posted body starts with the marker');
    assert.equal(posted.split(MARKER).length, 2, 'marker appears once');
  });
}

test('sticky lookup ignores non-bot and marker-less comments and updates the last match', async () => {
  const env = setup({
    meta: readyMetadata(),
    existingComments: [
      { id: 1, user: { login: 'github-actions[bot]' }, body: `${MARKER}\nfirst` },
      { id: 2, user: { login: 'github-actions[bot]' }, body: `${MARKER}\nlast bot match` },
      { id: 3, user: { login: 'someone' }, body: `${MARKER}\nhuman` },
      { id: 4, user: { login: 'github-actions[bot]' }, body: 'no marker' },
    ],
  });
  await updateResearchComment(env);
  assert.equal(env.calls[0][0], 'updateComment');
  assert.equal(env.calls[0][1].comment_id, 2);
});

test('a listComments failure fails the job and writes nothing', async () => {
  const env = setup({ meta: readyMetadata() });
  env.github.paginate = async () => { throw new Error('list boom'); };
  await updateResearchComment(env);
  assert.equal(env.core.failures.length, 1);
  assert.deepEqual(names(env.calls), []);
});

test('a removeLabel 500 after addLabels still writes the summary', async () => {
  const env = setup({ meta: readyMetadata(), failLabel: 'remove' });
  await updateResearchComment(env);
  assert.deepEqual(names(env.calls), ['createComment', 'addLabels', 'removeLabel']);
  assert.equal(env.summary.length, 1);
  assert.equal(env.core.failures.length, 1);
});

function writeItem(item) {
  writeOutput({ items: [{ type: 'update_research_comment', ...item }] });
}

test('joins body continuation parts in order before posting', async () => {
  const meta = readyMetadata();
  const full = body(meta);
  const a = full.slice(0, 40);
  const b = full.slice(40, 120);
  const c = full.slice(120);
  const env = setup({ meta });
  writeItem({ body: a, body_2: b, body_3: c });
  await updateResearchComment(env);
  assert.equal(env.calls[0][1].body, `${MARKER}\n${full}`);
  assert.deepEqual(env.calls[1][1].labels, [READY]);
});

test('a missing middle part is ignored and the rest stay ordered', async () => {
  const env = setup({ meta: readyMetadata() });
  writeItem({ body: 'A\n', body_3: 'C\n', body_2: '', body_5: 'E\n' });
  await updateResearchComment(env);
  assert.ok(env.calls[0][1].body.startsWith(`${MARKER}\nA\nC\nE\n`));
});

test('unicode parts are joined without alteration', async () => {
  const env = setup({ meta: readyMetadata() });
  writeItem({ body: 'é日本😀\n', body_2: '😀日本é\n' });
  await updateResearchComment(env);
  assert.ok(env.calls[0][1].body.startsWith(`${MARKER}\né日本😀\n😀日本é\n`));
});

test('a single body behaves as before', async () => {
  const env = setup({ meta: readyMetadata() });
  writeItem({ body: 'only\n' });
  await updateResearchComment(env);
  assert.ok(env.calls[0][1].body.startsWith(`${MARKER}\nonly\n`));
});

test('gate derivation runs on the joined body', async () => {
  const meta = readyMetadata();
  const full = body(meta);
  const env = setup({ meta });
  writeItem({ body: full.slice(0, 200), body_2: full.slice(200) });
  await updateResearchComment(env);
  assert.deepEqual(env.calls[1][1].labels, [READY]);
});

test('end to end: emit, platform trim, comment script, gate parses metadata and derives the outcome', async () => {
  const { splitBody } = require('../research-factory/emit-research-comment.js');
  const meta = readyMetadata();
  const filler = Array.from({ length: 400 }, (_, i) => `- bullet ${i}: ${'text '.repeat(12)}`).join('\n');
  const full = `## Implementation research\n\n### Open questions\n\n${filler}\n\n### Quality gate\n\n**Outcome:** \`${READY}\` - ok\n\n### References\n\n- a\n\n<details>\n<summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${JSON.stringify(meta, null, 2)}\n\`\`\`\n\n</details>\n`;
  const parts = splitBody(full);
  assert.ok(Object.keys(parts).length > 1);
  const trimmedParts = Object.fromEntries(Object.entries(parts).map(([k, v]) => [k, v.trim()]));
  const env = setup({ meta });
  writeItem(trimmedParts);
  await updateResearchComment(env);

  const posted = env.calls[0][1].body;
  assert.ok(posted.startsWith(MARKER));
  assert.ok(posted.includes(JSON.stringify(meta, null, 2)));
  assert.doesNotMatch(posted, /RF_PART/);
  assert.doesNotMatch(posted, /overridden by the gate rule/);
  assert.deepEqual(env.calls[1][1].labels, [READY]);
  assert.equal(env.core.failures.length, 0);
});

for (const [name, fn] of sanitizers) {
  test(`${name}: end to end through update-research-comment and the gate`, async () => {
    const meta = readyMetadata();
    const filler = Array.from({ length: 330 }, (_, i) => `- bullet ${i}: ${'text '.repeat(12)}`).join('\n');
    const full = `## Implementation research\n\n### Open questions\n\n${filler}\n\n### Quality gate\n\n**Outcome:** \`${READY}\` - ok\n\n### References\n\n- a\n\n<details>\n<summary>Pipeline metadata</summary>\n\n\`\`\`json\n${JSON.stringify(meta, null, 2)}\n\`\`\`\n\n</details>\n`;
    const parts = sanitizeParts(require('../research-factory/emit-research-comment.js').splitBody(full), fn);
    assert.ok(Object.keys(parts).length > 1);
    const env = setup({ meta });
    writeItem(parts);
    await updateResearchComment(env);
    const posted = env.calls[0][1].body;
    assert.doesNotMatch(posted, /RF_PART/);
    assert.doesNotMatch(posted, /overridden by the gate rule/);
    assert.deepEqual(env.calls[1][1].labels, [READY]);
    assert.equal(env.core.failures.length, 0);
  });
}
