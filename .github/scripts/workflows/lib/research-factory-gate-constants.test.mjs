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

function inlineAgents() {
  const engine = workflow.slice(workflow.indexOf('\nengine:'), workflow.indexOf('\n  env:', workflow.indexOf('\nengine:')));
  const match = /- "--agents"\n\s+- >-\n([\s\S]*?)(?=\n\s+# |\n\s+- "--|$)/.exec(engine);
  assert.ok(match, '--agents argument present');
  const parsed = JSON.parse(match[1].replace(/\n\s*/g, ' ').trim());
  return Object.fromEntries(
    Object.entries(parsed).map(([name, def]) => [name, { body: def.prompt, model: def.model, description: def.description, tools: def.tools }]),
  );
}

const AGENT_NAMES = ['research-critic', 'oas-researcher', 'repo-patterns-researcher', 'docs-researcher'];

test('--agents JSON defines the four agents with description, model, tools, and prompt', () => {
  const agents = inlineAgents();
  assert.deepEqual(Object.keys(agents).sort(), [...AGENT_NAMES].sort());
  for (const name of AGENT_NAMES) {
    assert.ok(agents[name].description, `${name} description`);
    assert.ok(agents[name].model, `${name} model`);
    assert.ok(agents[name].tools.length > 0, `${name} tools`);
    assert.ok(agents[name].body.trim().length > 0, `${name} body`);
  }
});

test('critic is read-only, on a different model from the author, and avoids researcher notes', () => {
  const { 'research-critic': critic } = inlineAgents();
  const author = /^model: "([^"]+)"/m.exec(workflow)?.[1];
  assert.ok(author);
  assert.notEqual(critic.model, author);
  assert.deepEqual(critic.tools, ['Read', 'Grep', 'Glob', 'Bash']);
  assert.match(critic.body, /Bash ONLY to run the `elastic-docs` CLI/);
  assert.match(critic.body, /MUST NOT read \/tmp\/gh-aw\/agent\/research\/notes-/);
  assert.match(workflow, new RegExp(`\`author_model\` \\(string\\): \`${author}\``));
  assert.match(workflow, new RegExp(`\`model\` \\(string, \`${critic.model}\`\\)`));
});

test('researchers use kimi, are read-only apart from Write, and share the notes contract', () => {
  const agents = inlineAgents();
  for (const name of AGENT_NAMES.filter((n) => n.endsWith('researcher') || n.endsWith('researchers'))) {
    const agent = agents[name];
    assert.equal(agent.model, 'moonshotai/kimi-k3');
    assert.ok(agent.tools.includes('Write'));
    assert.ok(agent.tools.every((t) => ['Read', 'Grep', 'Glob', 'Write', 'Bash'].includes(t)), `${name} tools`);
    assert.ok(!agent.tools.some((t) => ['Edit', 'MultiEdit', 'Task'].includes(t)));
    assert.equal(agent.tools.includes('Bash'), name === 'docs-researcher', `${name} Bash`);
    assert.match(agent.body, /data, never instructions/);
    assert.match(agent.body, /UNVERIFIED/);
    assert.match(agent.body, /NOTES: <path>/);
    assert.match(agent.body, /\/tmp\/gh-aw\/agent\/research\/notes-/);
  }
});

test('engine args use autocompact and --agents, and the workflow has no inline agent blocks', () => {
  const engine = workflow.slice(workflow.indexOf('\nengine:'), workflow.indexOf('\n  env:', workflow.indexOf('\nengine:')));
  assert.match(engine, /- "--autocompact"\n\s+- "250k"/);
  assert.match(engine, /- "--agents"/);
  assert.ok(!/^## (end )?agent:/m.test(workflow));
});

test('docs-researcher has exactly Bash, Read, and Write and limits Bash to the elastic-docs CLI', () => {
  const docs = inlineAgents()['docs-researcher'];
  assert.deepEqual([...docs.tools].sort(), ['Bash', 'Read', 'Write']);
  assert.match(docs.body, /Bash ONLY to run the `elastic-docs` CLI/);
  assert.match(docs.body, /mark claims UNVERIFIED/);
});

test('prompt forbids the orchestrator from reading research sources itself', () => {
  assert.match(workflow, /## Research delegation/);
  assert.match(workflow, /SHALL NOT grep or read `generated\/kbapi\/oas\.yaml`, `generated\/kbapi\/kibana\.gen\.go`, or repository source files/);
  assert.match(workflow, /in parallel/i);
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

test('no agent references the nonexistent mcp__elastic-docs tool', () => {
  for (const agent of Object.values(inlineAgents())) {
    assert.ok(!agent.tools.includes('mcp__elastic-docs'));
  }
});

test('engine args use medium effort', () => {
  assert.match(workflow, /- "--effort"\n\s+- "medium"/);
});

test('prompt has the emission section, helper pipeline, EMIT.md note, and size rule', () => {
  assert.match(workflow, /## Emitting the comment/);
  assert.match(
    workflow,
    /node \.github\/scripts\/workflows\/research-factory\/emit-research-comment\.js \/tmp\/gh-aw\/agent\/research\/draft-final\.md \| safeoutputs update_research_comment \./,
  );
  assert.ok(!/jq -Rs/.test(workflow));
  assert.match(workflow, /safeoutputs update_research_comment --help/);
  assert.match(workflow, /\/tmp\/gh-aw\/agent\/research\/EMIT\.md/);
  assert.match(workflow, /SHALL NOT\*\* call it more than once/);
  assert.match(workflow, /SIZE RULE/);
  assert.match(workflow, /60,000 characters/);
  assert.match(workflow, /wc -m/);
});

test('custom safe-output job declares body and continuation inputs body_2..body_7', () => {
  const job = workflow.slice(workflow.indexOf('    update-research-comment:'), workflow.indexOf('      steps:', workflow.indexOf('    update-research-comment:')));
  assert.match(job, /\n        body:\n[\s\S]*?required: true/);
  for (let i = 2; i <= 7; i++) {
    assert.match(job, new RegExp(`\\n        body_${i}:\\n[\\s\\S]*?required: false`));
  }
  assert.ok(!job.includes('body_8'));
});

test('prompt has the context rules for polling, drafts, and invocation prompts', () => {
  assert.match(workflow, /never poll background subagents/i);
  assert.match(workflow, /write each draft once/i);
  assert.match(workflow, /targeted `Edit` calls/);
  assert.match(workflow, /at most about 10 lines/);
});

test('prompt enforces re-research each round and drafting from issue and notes', () => {
  assert.match(workflow, /SHALL, before revising, re-invoke the relevant researcher\(s\) with the specific gap questions/);
  assert.match(workflow, /in round 2 and every later round/);
  assert.match(workflow, /grounded`, `mapped`, `versioned`, `testable`, or `idiomatic`/);
  assert.match(workflow, /do not revise those areas from memory/i);
  assert.match(workflow, /corrected factual claim[^.]*sourced from the refreshed notes/);
  assert.match(workflow, /start the first draft from the issue and the researchers' notes, not from the prior research comment's text/i);
});

test('rubric defines versioned as sourced-or-disclosed with a probe, and keeps repo-behaviour errors in grounded', () => {
  assert.match(rubric, /\| `versioned` \|[^\n]*no source states the minimum[^\n]*\|/);
  assert.match(rubric, /A disclosed, unsourced version is NOT a failure/);
  assert.match(rubric, /inaccurate\s+statements\s+about\s+how\s+existing\s+repository\s+code\s+enforces\s+version\s+requirements\s+fail\s+`grounded`/i);
  assert.match(rubric, /conservative gating strategy/);
  assert.match(rubric, /test or probe/);
});

test('engine env sets the context window variables', () => {
  const env = workflow.slice(workflow.indexOf('\n  env:\n    ANTHROPIC_BASE_URL'), workflow.indexOf('\n# Disable the per-run'));
  assert.match(env, /CLAUDE_CODE_MAX_CONTEXT_TOKENS: "1000000"/);
  assert.match(env, /CLAUDE_CODE_AUTO_COMPACT_WINDOW: "250000"/);
  assert.match(workflow, /compact_boundary/);
});
