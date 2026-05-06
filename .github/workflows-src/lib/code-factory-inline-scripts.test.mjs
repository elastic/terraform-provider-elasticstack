import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const scriptsDir = path.resolve(__dirname, '../code-factory-issue/scripts');
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

async function runInlineScript(scriptName, { context = {}, github = {}, core = createMockCore(), env = {} } = {}) {
  const scriptPath = path.join(scriptsDir, scriptName);
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

test('validate_dispatch_inputs: accepts valid dispatch inputs', async () => {
  const { core } = await runInlineScript('validate_dispatch_inputs.inline.js', {
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' },
      payload: { inputs: { issue_number: '42' } },
    },
  });

  assert.equal(core.outputs.event_eligible, 'true');
  assert.match(core.outputs.event_eligible_reason, /issue #42/);
  assert.equal(core.outputs.issue_number, '42');
  assert.equal(core.failures.length, 0);
});

test('validate_dispatch_inputs: rejects invalid issue number cleanly', async () => {
  const { core } = await runInlineScript('validate_dispatch_inputs.inline.js', {
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' },
      payload: { inputs: { issue_number: 'abc' } },
    },
  });

  assert.equal(core.outputs.event_eligible, 'false');
  assert.match(core.outputs.event_eligible_reason, /not a valid positive integer/);
  assert.equal(core.failures.length, 0, 'should not call core.setFailed');
});

test('fetch_live_issue: happy path sets outputs from GitHub API response', async () => {
  const mockGithub = {
    rest: {
      issues: {
        get: async () => ({
          data: { number: 42, title: 'Feature request', body: 'Please implement this.\n\nThanks!' },
        }),
      },
    },
  };
  const { core } = await runInlineScript('fetch_live_issue.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.issue_number, '42');
  assert.equal(core.outputs.issue_title, 'Feature request');
  assert.equal(core.outputs.issue_body, 'Please implement this.\n\nThanks!');
  assert.equal(core.outputs.fetch_error, '');
  assert.equal(core.failures.length, 0);
});

test('fetch_live_issue: API 404 sets empty outputs and calls setFailed', async () => {
  const mockGithub = {
    rest: {
      issues: {
        get: async () => {
          const err = new Error('Not Found');
          err.status = 404;
          throw err;
        },
      },
    },
  };
  const { core } = await runInlineScript('fetch_live_issue.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.issue_number, '');
  assert.equal(core.outputs.issue_title, '');
  assert.equal(core.outputs.issue_body, '');
  assert.equal(core.outputs.fetch_error, 'Not Found');
  assert.ok(core.failures.some((f) => f.includes('Failed to fetch issue')));
});

test('fetch_live_issue: API network failure sets fetch_error output and calls setFailed', async () => {
  const mockGithub = {
    rest: {
      issues: {
        get: async () => {
          throw new Error('Network error: ENOTFOUND');
        },
      },
    },
  };
  const { core } = await runInlineScript('fetch_live_issue.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.issue_number, '');
  assert.equal(core.outputs.fetch_error, 'Network error: ENOTFOUND');
  assert.ok(core.failures.some((f) => f.includes('Network error')));
});

test('fetch_live_issue: invalid env issue number sets failure without API call', async () => {
  // When INPUT_ISSUE_NUMBER env is not set, the script falls back to parsing empty string
  const { core } = await runInlineScript('fetch_live_issue.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: {},
    core: createMockCore(),
  });

  assert.equal(core.outputs.issue_number, '');
  assert.equal(core.outputs.fetch_error, 'Invalid issue number in dispatch inputs.');
  assert.ok(core.failures.some((f) => f.includes('Cannot fetch live issue')));
});

test('check_duplicate_pr (dispatch branch): paginates with correct head filter', async () => {
  const captured = [];
  const mockGithub = {
    paginate: async (apiMethod, params) => {
      captured.push(params);
      return [
        {
          number: 101,
          state: 'open',
          head: { ref: 'code-factory/issue-42' },
          labels: [{ name: 'code-factory' }],
          body: 'Closes #42',
          html_url: 'https://github.com/elastic/terraform-provider-elasticstack/pull/101',
        },
      ];
    },
    rest: { pulls: { list: () => {} } },
  };
  const { core } = await runInlineScript('check_duplicate_pr.inline.js', {
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' },
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: '42' } },
    },
    github: mockGithub,
    core: createMockCore(),
  });

  assert.equal(captured.length, 1);
  assert.equal(captured[0].head, 'elastic:code-factory/issue-42');
  assert.equal(captured[0].state, 'open');
  assert.equal(core.outputs.duplicate_pr_found, 'true');
  assert.equal(core.outputs.duplicate_pr_url, 'https://github.com/elastic/terraform-provider-elasticstack/pull/101');
});

test('check_duplicate_pr (dispatch branch): empty issue number skips check gracefully', async () => {
  const { core } = await runInlineScript('check_duplicate_pr.inline.js', {
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' },
      eventName: 'workflow_dispatch',
      payload: { inputs: { issue_number: '' } },
    },
    github: {},
    core: createMockCore(),
  });

  assert.equal(core.outputs.duplicate_pr_found, 'false');
  assert.equal(core.outputs.duplicate_pr_url, '');
  assert.match(core.outputs.gate_reason, /No issue number available/);
});

test('check_duplicate_pr (issue-event branch): uses event payload issue number', async () => {
  const captured = [];
  const mockGithub = {
    paginate: async (apiMethod, params) => {
      captured.push(params);
      return [];
    },
    rest: { pulls: { list: () => {} } },
  };
  const { core } = await runInlineScript('check_duplicate_pr.inline.js', {
    context: {
      repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' },
      eventName: 'issues',
      payload: { issue: { number: 7 } },
    },
    github: mockGithub,
    core: createMockCore(),
  });

  assert.equal(captured.length, 1);
  assert.equal(captured[0].head, 'elastic:code-factory/issue-7');
  assert.equal(core.outputs.duplicate_pr_found, 'false');
  assert.match(core.outputs.gate_reason, /No open linked code-factory PR found/);
});
