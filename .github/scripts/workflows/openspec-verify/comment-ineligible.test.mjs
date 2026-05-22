import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const commentIneligible = require('./comment-ineligible.js');
const { buildCommentBody } = commentIneligible;

function makeCore() {
  const infos = [];
  return {
    infos,
    info(msg) {
      infos.push(msg);
    },
  };
}

function makeGithub(captured) {
  return {
    rest: {
      issues: {
        createComment: async (args) => {
          captured.push(args);
          return {};
        },
      },
    },
  };
}

const baseContext = {
  repo: { owner: 'acme', repo: 'demo' },
  payload: {
    pull_request: { number: 42 },
  },
};

test('posts a comment with selection reason and remediation guidance when PR is ineligible', async () => {
  const captured = [];
  const core = makeCore();
  const reason = 'No files under openspec/changes/ (non-archive) found in this PR';
  const previous = process.env.SELECTION_REASON;
  process.env.SELECTION_REASON = reason;
  try {
    await commentIneligible({
      github: makeGithub(captured),
      context: baseContext,
      core,
    });
  } finally {
    if (previous === undefined) delete process.env.SELECTION_REASON;
    else process.env.SELECTION_REASON = previous;
  }

  assert.equal(captured.length, 1);
  const args = captured[0];
  assert.equal(args.owner, 'acme');
  assert.equal(args.repo, 'demo');
  assert.equal(args.issue_number, 42);
  assert.match(args.body, /OpenSpec verify skipped/);
  assert.ok(args.body.includes(reason), 'comment body includes verbatim selection_reason');
  assert.match(args.body, /How to fix/);
  assert.match(args.body, /openspec\/changes\/<id>\//);
  assert.match(args.body, /OpenSpec authoring guide/);
  assert.ok(core.infos.some((m) => m.includes('posted ineligibility comment on PR #42')));
});

test('short-circuits when the PR number is absent (no API call, no throw)', async () => {
  const captured = [];
  const core = makeCore();
  const previous = process.env.SELECTION_REASON;
  process.env.SELECTION_REASON = 'Multiple active change ids: bar, foo';
  try {
    await commentIneligible({
      github: makeGithub(captured),
      context: { repo: { owner: 'acme', repo: 'demo' }, payload: {} },
      core,
    });
  } finally {
    if (previous === undefined) delete process.env.SELECTION_REASON;
    else process.env.SELECTION_REASON = previous;
  }

  assert.equal(captured.length, 0, 'no comment should be posted');
  assert.ok(
    core.infos.some((m) => m.includes('no pull request number')),
    'expected info log mentioning missing PR number'
  );
});

test('comment body includes the selection_reason verbatim for known ineligibility scenarios', () => {
  const reasons = [
    'No files under openspec/changes/ (non-archive) found in this PR',
    'Multiple active change ids: bar, foo',
    'Unsupported file status under openspec/changes/: openspec/changes/example/tasks.md (renamed)',
  ];
  for (const reason of reasons) {
    const body = buildCommentBody(reason);
    assert.ok(body.includes(reason), `expected body to include verbatim reason: ${reason}`);
    assert.match(body, /How to fix/);
  }
});

test('comment body falls back to a placeholder when SELECTION_REASON is empty', () => {
  const body = buildCommentBody('');
  assert.match(body, /no reason provided by classify_and_select/);
  assert.match(body, /How to fix/);
});
