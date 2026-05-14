import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync, mkdtempSync, writeFileSync, unlinkSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { createRequire } from 'node:module';

const nodeRequire = createRequire(import.meta.url);

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const scriptsDir = path.resolve(__dirname, '../research-factory-issue/scripts');
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
  const warnings = [];
  const failures = [];
  return {
    outputs,
    logs,
    warnings,
    failures,
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

async function runInlineScript(scriptName, { context = {}, github = {}, core = createMockCore(), env = {}, item = undefined } = {}) {
  const scriptPath = path.join(scriptsDir, scriptName);
  const code = expandIncludes(scriptPath);
  const previousEnv = {};
  for (const key of Object.keys(env)) {
    previousEnv[key] = process.env[key];
    process.env[key] = env[key];
  }
  try {
    const fn = new AsyncFunction('context', 'github', 'core', 'require', 'item', code);
    await fn(context, github, core, nodeRequire, item);
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

function createMockGithubWithComments(comments) {
  return {
    paginate: async () => comments,
    rest: { issues: { listComments: () => {} } },
  };
}

test('fetch_issue_comments: happy path emits serialized comments via GITHUB_OUTPUT', async () => {
  const tmpDir = mkdtempSync(join(tmpdir(), 'gh-output-'));
  const ghOutputFile = join(tmpDir, 'github_output');
  writeFileSync(ghOutputFile, '');

  const mockComments = [
    {
      user: { login: 'alice' },
      created_at: '2024-01-15T10:00:00Z',
      body: 'First comment.',
    },
    {
      user: { login: 'bob' },
      created_at: '2024-01-15T11:30:00Z',
      body: 'Second comment.\nWith a newline.',
    },
  ];

  const { core } = await runInlineScript('fetch_issue_comments.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubWithComments(mockComments),
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '7', GITHUB_OUTPUT: ghOutputFile },
  });

  const outputContent = readFileSync(ghOutputFile, 'utf8');
  assert.ok(outputContent.includes('issue_comments<<EOF_'));
  assert.ok(outputContent.includes('**@alice** (2024-01-15T10:00:00Z):'));
  assert.ok(outputContent.includes('**@bob** (2024-01-15T11:30:00Z):'));
  assert.ok(outputContent.includes('Second comment.\nWith a newline.'));
  assert.equal(core.outputs.comment_count, '2');
  assert.equal(core.failures.length, 0);

  unlinkSync(ghOutputFile);
});

test('fetch_issue_comments: filters out bot comments and counts only humans', async () => {
  const tmpDir = mkdtempSync(join(tmpdir(), 'gh-output-'));
  const ghOutputFile = join(tmpDir, 'github_output');
  writeFileSync(ghOutputFile, '');

  const mockComments = [
    { user: { login: 'github-actions[bot]' }, created_at: '2024-01-15T10:00:00Z', body: 'Bot comment.' },
    { user: { login: 'human-user' }, created_at: '2024-01-15T11:00:00Z', body: 'Human comment.' },
    { user: { login: 'dependabot[bot]' }, created_at: '2024-01-15T12:00:00Z', body: 'Dependabot comment.' },
  ];

  const { core } = await runInlineScript('fetch_issue_comments.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubWithComments(mockComments),
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '7', GITHUB_OUTPUT: ghOutputFile },
  });

  const outputContent = readFileSync(ghOutputFile, 'utf8');
  assert.ok(outputContent.includes('**@human-user**'));
  assert.ok(!outputContent.includes('github-actions[bot]'));
  assert.ok(!outputContent.includes('dependabot[bot]'));
  assert.equal(core.outputs.comment_count, '1');
  assert.equal(core.failures.length, 0);

  unlinkSync(ghOutputFile);
});

test('fetch_issue_comments: API failure sets empty outputs and calls setFailed', async () => {
  const mockGithub = {
    paginate: async () => {
      throw new Error('API rate limit exceeded');
    },
    rest: { issues: { listComments: () => {} } },
  };

  const { core } = await runInlineScript('fetch_issue_comments.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.comment_count, '0');
  assert.equal(core.outputs.issue_comments, '');
  assert.ok(core.failures.some((f) => f.includes('API rate limit exceeded')));
});

test('fetch_issue_comments: invalid env issue number sets failure without API call', async () => {
  const { core } = await runInlineScript('fetch_issue_comments.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: {},
    core: createMockCore(),
  });

  assert.equal(core.outputs.comment_count, '0');
  assert.equal(core.outputs.issue_comments, '');
  assert.ok(core.failures.some((f) => f.includes('Cannot fetch issue comments')));
});

test('finalize_gate: all gates passed returns all-gates-passed reason', async () => {
  const { core } = await runInlineScript('finalize_gate.inline.js', {
    context: {},
    github: {},
    core: createMockCore(),
    env: {
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'Issue labeled event qualifies because the applied label is research-factory.',
      ACTOR_TRUSTED: 'true',
      ACTOR_TRUSTED_REASON: "Trigger actor 'alice' is trusted with repository permission 'admin'.",
    },
  });

  assert.match(core.outputs.gate_reason, /All deterministic gates passed/);
  assert.equal(core.failures.length, 0);
});

test('finalize_gate: event not eligible returns event reason', async () => {
  const { core } = await runInlineScript('finalize_gate.inline.js', {
    context: {},
    github: {},
    core: createMockCore(),
    env: {
      EVENT_ELIGIBLE: 'false',
      EVENT_ELIGIBLE_REASON: "Unsupported event 'pull_request'; expected 'issues'.",
      ACTOR_TRUSTED: 'true',
      ACTOR_TRUSTED_REASON: "Trigger actor 'alice' is trusted.",
    },
  });

  assert.equal(core.outputs.gate_reason, "Unsupported event 'pull_request'; expected 'issues'.");
  assert.equal(core.failures.length, 0);
});

test('fetch_issue_comments: empty comment list emits empty heredoc and count 0', async () => {
  const tmpDir = mkdtempSync(join(tmpdir(), 'gh-output-'));
  const ghOutputFile = join(tmpDir, 'github_output');
  writeFileSync(ghOutputFile, '');

  const { core } = await runInlineScript('fetch_issue_comments.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubWithComments([]),
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42', GITHUB_OUTPUT: ghOutputFile },
  });

  const outputContent = readFileSync(ghOutputFile, 'utf8');
  // Heredoc structure must still be present even for empty content
  assert.ok(outputContent.includes('issue_comments<<EOF_'), 'heredoc opener missing');
  assert.equal(core.outputs.comment_count, '0');
  assert.equal(core.failures.length, 0);

  unlinkSync(ghOutputFile);
});

test('finalize_gate: actor not trusted returns actor reason', async () => {
  const { core } = await runInlineScript('finalize_gate.inline.js', {
    context: {},
    github: {},
    core: createMockCore(),
    env: {
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'Issue labeled event qualifies because the applied label is research-factory.',
      ACTOR_TRUSTED: 'false',
      ACTOR_TRUSTED_REASON: "Trigger actor 'outsider' is not trusted; repository permission 'read' does not meet the required write/maintain/admin policy.",
    },
  });

  assert.equal(
    core.outputs.gate_reason,
    "Trigger actor 'outsider' is not trusted; repository permission 'read' does not meet the required write/maintain/admin policy.",
  );
  assert.equal(core.failures.length, 0);
});

test('finalize_gate: actor_trusted null (missing) returns cannot-determine reason', async () => {
  const { core } = await runInlineScript('finalize_gate.inline.js', {
    context: {},
    github: {},
    core: createMockCore(),
    env: {
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'Issue labeled event qualifies.',
      // ACTOR_TRUSTED intentionally omitted → null
    },
  });

  assert.match(core.outputs.gate_reason, /Actor trust could not be determined/);
  assert.equal(core.failures.length, 0);
});

// --- fetch_prior_research_comment tests ---

test('fetch_prior_research_comment: happy path finds latest matching comment', async () => {
  const marker = '<!-- gha-research-factory -->';
  const mockComments = [
    { user: { login: 'alice' }, body: 'human comment', id: 1 },
    { user: { login: 'github-actions[bot]' }, body: `${marker}\nold research`, id: 2 },
    { user: { login: 'github-actions[bot]' }, body: `${marker}\nlatest research`, id: 3 },
  ];

  const { core } = await runInlineScript('fetch_prior_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubWithComments(mockComments),
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.prior_research_comment, `${marker}\nlatest research`);
  assert.ok(core.logs.some((l) => l.includes('Found prior research comment 3')));
  assert.equal(core.failures.length, 0);
});

test('fetch_prior_research_comment: no match returns empty output', async () => {
  const mockComments = [
    { user: { login: 'alice' }, body: 'human comment', id: 1 },
  ];

  const { core } = await runInlineScript('fetch_prior_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubWithComments(mockComments),
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.prior_research_comment, '');
  assert.ok(core.logs.some((l) => l.includes('No prior research comment found')));
  assert.equal(core.failures.length, 0);
});

test('fetch_prior_research_comment: API failure sets empty output and logs warning', async () => {
  const mockGithub = {
    paginate: async () => {
      throw new Error('API rate limit exceeded');
    },
    rest: { issues: { listComments: () => {} } },
  };

  const { core } = await runInlineScript('fetch_prior_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: mockGithub,
    core: createMockCore(),
    env: { INPUT_ISSUE_NUMBER: '42' },
  });

  assert.equal(core.outputs.prior_research_comment, '');
  assert.ok(core.warnings.some((w) => w.includes('API rate limit exceeded')));
  assert.equal(core.failures.length, 0);
});

// --- update_research_comment tests ---

function createMockGithubForUpdate({ existingComments = [], createdComment = null } = {}) {
  return {
    paginate: async () => existingComments,
    rest: {
      issues: {
        listComments: () => {},
        createComment: async () => ({ data: createdComment || { id: 99 } }),
        updateComment: async () => {},
      },
    },
  };
}

test('update_research_comment: happy path creates comment when none exists', async () => {
  const marker = '<!-- gha-research-factory -->';
  const body = `${marker}\nImplementation research content`;

  const { core } = await runInlineScript('update_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubForUpdate({ existingComments: [] }),
    core: createMockCore(),
    env: { RESEARCH_FACTORY_ISSUE_NUMBER: '42' },
    item: { body },
  });

  assert.equal(core.failures.length, 0);
  assert.ok(core.logs.some((l) => l.includes('Created research comment')));
});

test('update_research_comment: happy path updates existing comment', async () => {
  const marker = '<!-- gha-research-factory -->';
  const body = `${marker}\nUpdated research content`;
  const existingComments = [
    { user: { login: 'github-actions[bot]' }, body: `${marker}\nold`, id: 10 },
  ];

  const { core } = await runInlineScript('update_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubForUpdate({ existingComments }),
    core: createMockCore(),
    env: { RESEARCH_FACTORY_ISSUE_NUMBER: '42' },
    item: { body },
  });

  assert.equal(core.failures.length, 0);
  assert.ok(core.logs.some((l) => l.includes('Updated research comment 10')));
});

test('update_research_comment: missing marker is prepended automatically', async () => {
  const body = 'Missing marker content';

  const { core, github } = await runInlineScript('update_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubForUpdate({ existingComments: [] }),
    core: createMockCore(),
    env: { RESEARCH_FACTORY_ISSUE_NUMBER: '42' },
    item: { body },
  });

  assert.equal(core.failures.length, 0);
  assert.ok(core.logs.some((l) => l.includes('Created research comment')));
});

test('update_research_comment: invalid issue number fails', async () => {
  const marker = '<!-- gha-research-factory -->';
  const body = `${marker}\nContent`;

  const { core } = await runInlineScript('update_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubForUpdate(),
    core: createMockCore(),
    env: {},
    item: { body },
  });

  assert.ok(core.failures.some((f) => f.includes('invalid issue number')));
});

test('update_research_comment: null item fails', async () => {
  const { core } = await runInlineScript('update_research_comment.inline.js', {
    context: { repo: { owner: 'elastic', repo: 'terraform-provider-elasticstack' } },
    github: createMockGithubForUpdate(),
    core: createMockCore(),
    env: { RESEARCH_FACTORY_ISSUE_NUMBER: '42' },
    item: null,
  });

  assert.ok(core.failures.some((f) => f.includes('no item provided')));
});
