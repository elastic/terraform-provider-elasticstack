import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const AsyncFunction = Object.getPrototypeOf(async function () {}).constructor;

function createMockCore() {
  const outputs = {};
  const logs = [];
  const failures = [];
  const warnings = [];
  return {
    outputs,
    logs,
    failures,
    warnings,
    setOutput(key, value) {
      outputs[key] = value;
    },
    info(msg) {
      logs.push(msg);
    },
    warning(msg) {
      warnings.push(msg);
    },
    setFailed(msg) {
      failures.push(msg);
    },
  };
}

async function runScript({ phaseLabelName, context = {}, github = {}, core = createMockCore(), env = {} } = {}) {
  const code = [
    readFileSync(path.resolve(__dirname, './set-phase-label.js'), 'utf8'),
    readFileSync(path.resolve(__dirname, './set-phase-label-run.js'), 'utf8'),
  ].join('\n');

  const mergedEnv = { ...env };
  if (phaseLabelName !== undefined) {
    mergedEnv.PHASE_LABEL_NAME = phaseLabelName;
  }

  const previousEnv = {};
  for (const key of Object.keys(mergedEnv)) {
    previousEnv[key] = process.env[key];
    process.env[key] = mergedEnv[key];
  }
  try {
    const fn = new AsyncFunction('context', 'github', 'core', code);
    await fn(context, github, core);
  } finally {
    for (const key of Object.keys(mergedEnv)) {
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

test('set-phase-label run script emits correct phase label outputs for each factory phase', async () => {
  const phases = [
    'phase-research',
    'phase-reproduction',
    'phase-specification',
    'phase-coding',
  ];

  for (const phase of phases) {
    const { core } = await runScript({
      phaseLabelName: phase,
      context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
      github: makeMockGithub(),
      core: createMockCore(),
    });
    assert.equal(core.outputs.phase_label_set, 'true');
    assert.equal(core.outputs.phase_label_name, phase);
  }
});

test('run script prefers INPUT_ISSUE_NUMBER over context.payload.issue.number', async () => {
  const addedIssues = [];
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async (args) => { addedIssues.push(args.issue_number); return {}; },
        listLabelsOnIssue: async () => ({ data: [] }),
      },
    },
  };

  await runScript({
    phaseLabelName: 'phase-research',
    context: { payload: { issue: { number: 99 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '77' },
  });

  assert.equal(addedIssues[0], 77, 'Should use INPUT_ISSUE_NUMBER (77), not context.payload (99)');
});

test('run script emits core.warning when phase label is not set', async () => {
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async () => { throw new Error('Validation Failed'); },
      },
    },
  };

  const { core } = await runScript({
    phaseLabelName: 'phase-research',
    context: { payload: { issue: { number: 42 } }, repo: { owner: 'elastic', repo: 'repo' } },
    github: mockGithub,
    core: createMockCore(),
  });

  assert.equal(core.outputs.phase_label_set, 'false');
  assert.equal(core.warnings.length, 1);
  assert.ok(core.warnings[0].includes('Phase label not set'));
});
