import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const AsyncFunction = Object.getPrototypeOf(async function () {}).constructor;

function expandIncludes(scriptPath) {
  let content = readFileSync(scriptPath, 'utf8');
  const includePattern = /^(\s*)\/\/include:\s*(.+?)\s*$/gm;
  let match;
  while ((match = includePattern.exec(content)) !== null) {
    const indent = match[1];
    const includeRel = match[2];
    const includedPath = path.resolve(path.dirname(scriptPath), includeRel);
    let included = readFileSync(includedPath, 'utf8');
    if (indent) {
      included = included
        .split('\n')
        .map((line) => (line ? indent + line : line))
        .join('\n');
    }
    content = content.replace(match[0], included);
  }
  return content;
}

function createMockCore() {
  const outputs = {};
  const logs = [];
  const failures = [];
  return {
    outputs,
    logs,
    failures,
    setOutput(key, value) {
      outputs[key] = value;
    },
    info(msg) {
      logs.push(msg);
    },
    setFailed(msg) {
      failures.push(msg);
    },
  };
}

async function runInlineScript(scriptPath, { context = {}, github = {}, core = createMockCore(), env = {} } = {}) {
  const code = expandIncludes(scriptPath);
  const previousEnv = {};
  for (const key of Object.keys(env)) {
    previousEnv[key] = process.env[key];
    process.env[key] = env[key];
  }
  try {
    const fn = new AsyncFunction('context', 'github', 'core', code);
    await fn(context, github, core);
  } finally {
    for (const key of Object.keys(env)) {
      if (previousEnv[key] === undefined) {
        delete process.env[key];
      } else {
        process.env[key] = previousEnv[key];
      }
    }
  }
  return { core };
}

function makeMockGithub() {
  return {
    rest: {
      issues: {
        addLabels: async () => ({}),
        listLabelsOnIssue: async () => ({ data: [] }),
      },
    },
  };
}

test('research-factory set_phase_label.inline.js emits correct phase label outputs', async () => {
  const scriptPath = path.resolve(__dirname, '../research-factory-issue/scripts/set_phase_label.inline.js');
  const { core } = await runInlineScript(scriptPath, {
    context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: makeMockGithub(),
    core: createMockCore(),
  });
  assert.equal(core.outputs.phase_label_set, 'true');
  assert.equal(core.outputs.phase_label_name, 'phase-research');
});

test('reproducer-factory set_phase_label.inline.js emits correct phase label outputs', async () => {
  const scriptPath = path.resolve(__dirname, '../reproducer-factory-issue/scripts/set_phase_label.inline.js');
  const { core } = await runInlineScript(scriptPath, {
    context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: makeMockGithub(),
    core: createMockCore(),
  });
  assert.equal(core.outputs.phase_label_set, 'true');
  assert.equal(core.outputs.phase_label_name, 'phase-reproduction');
});

test('change-factory set_phase_label.inline.js emits correct phase label outputs', async () => {
  const scriptPath = path.resolve(__dirname, '../change-factory-issue/scripts/set_phase_label.inline.js');
  const { core } = await runInlineScript(scriptPath, {
    context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: makeMockGithub(),
    core: createMockCore(),
  });
  assert.equal(core.outputs.phase_label_set, 'true');
  assert.equal(core.outputs.phase_label_name, 'phase-specification');
});

test('code-factory set_phase_label.inline.js emits correct phase label outputs', async () => {
  const scriptPath = path.resolve(__dirname, '../code-factory-issue/scripts/set_phase_label.inline.js');
  const { core } = await runInlineScript(scriptPath, {
    context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: makeMockGithub(),
    core: createMockCore(),
  });
  assert.equal(core.outputs.phase_label_set, 'true');
  assert.equal(core.outputs.phase_label_name, 'phase-coding');
});
