import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const gate = require('../research-factory/gate.js');

const here = path.dirname(fileURLToPath(import.meta.url));
const rubric = fs.readFileSync(path.join(here, '../research-factory/critic-rubric.md'), 'utf8');
const workflow = fs.readFileSync(path.join(here, '../../../workflows/research-factory-issue.md'), 'utf8');

const { SCORE_THRESHOLD, STABILITY_WINDOW, MAX_ROUNDS } = gate;

const sources = {
  'critic-rubric.md': {
    text: rubric,
    threshold: new RegExp(`Score threshold: \\*\\*${SCORE_THRESHOLD}\\*\\*`),
    window: new RegExp(`Stability window: \\*\\*${STABILITY_WINDOW}\\*\\*`),
    rounds: new RegExp(`Maximum rounds: \\*\\*${MAX_ROUNDS}\\*\\*`),
  },
  'research-factory-issue.md': {
    text: workflow,
    threshold: new RegExp(`score threshold \\*\\*${SCORE_THRESHOLD}\\*\\*`),
    window: new RegExp(`stability window \\*\\*${STABILITY_WINDOW}\\*\\*`),
    rounds: new RegExp(`maximum of \\*\\*${MAX_ROUNDS}\\*\\* rounds`),
  },
};

for (const [file, { text, threshold, window, rounds }] of Object.entries(sources)) {
  test(`${file} states the gate score threshold from gate.js`, () => {
    assert.match(text, threshold);
  });
  test(`${file} states the gate stability window from gate.js`, () => {
    assert.match(text, window);
  });
  test(`${file} states the gate round bound from gate.js`, () => {
    assert.match(text, rounds);
  });
}

test('workflow prompt prose repeats the gate numbers from gate.js', () => {
  assert.match(
    workflow,
    new RegExp(`the final score is at least ${SCORE_THRESHOLD}\\s+and either the last ${STABILITY_WINDOW} rounds both scored at least ${SCORE_THRESHOLD}`),
  );
  assert.match(workflow, new RegExp(`run another round, up to ${MAX_ROUNDS} rounds`));
  assert.match(workflow, new RegExp(`plateau below ${SCORE_THRESHOLD} is not converged`));
});

test('rubric prose repeats the gate numbers from gate.js', () => {
  assert.match(
    rubric,
    new RegExp(`final score is at least ${SCORE_THRESHOLD} and either the last ${STABILITY_WINDOW} rounds both scored at\\s+least ${SCORE_THRESHOLD}`),
  );
});

test('author and critic models agree across frontmatter, --agents JSON, and prompt, and differ', () => {
  const author = /^model: "([^"]+)"/m.exec(workflow)?.[1];
  const critic = /"model": "([^"]+)"\}\}/.exec(workflow)?.[1];
  assert.ok(author && critic);
  assert.notEqual(author, critic);
  assert.match(workflow, new RegExp(`\`author_model\` \\(string\\): \`${author}\``));
  assert.match(workflow, new RegExp(`\`model\` \\(string, \`${critic}\`\\)`));
});

test('workflow source: timeout, budget, safe outputs, and stale-label step', () => {
  assert.match(workflow, /^timeout-minutes: 60$/m);
  assert.match(workflow, /approximately 50 minutes of agentic work/);
  const safeOutputs = workflow.slice(workflow.indexOf('\nsafe-outputs:'), workflow.indexOf('\n---', workflow.indexOf('\nsafe-outputs:')));
  for (const forbidden of ['add-labels', 'remove-labels', 'add-comment', 'create-issue', 'update-issue', 'create-pull-request']) {
    assert.ok(!safeOutputs.includes(forbidden), `safe-outputs must not enable ${forbidden}`);
  }
  assert.match(workflow, /name: Remove stale outcome labels[\s\S]*?research-factory\/remove-stale-outcome-labels\.js/);
});

test('workflow source downloads the pinned Kibana OAS before the agent runs', () => {
  const stepsBlock = workflow.slice(workflow.indexOf('\nsteps:\n'), workflow.indexOf('\nmodel:'));
  assert.match(stepsBlock, /name: Download Kibana OpenAPI spec\n\s+continue-on-error: true\n\s+run: make -C generated\/kbapi download/);
  assert.match(workflow, /oas\.yaml` is missing[^]*unavailable/i);
  assert.match(rubric, /oas\.yaml` is missing[^]*unavailable/i);
  assert.ok(stepsBlock.indexOf('Download issue context artifact') < stepsBlock.indexOf('Download Kibana OpenAPI spec'));
  assert.match(workflow, /generated\/kbapi\/oas\.yaml/);
  assert.match(rubric, /generated\/kbapi\/oas\.yaml/);
});
