import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const updateResearchComment = require('../research-factory/update-research-comment.js');

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
