import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { verifyTriggerLabel } = require('./verify-label.js');
const { classifyPullRequest } = require('./classify-pr.js');
const { removeTriggerLabel } = require('./remove-trigger-label.js');

// ---------------------------------------------------------------------------
// verify-label.js — verifyTriggerLabel
// ---------------------------------------------------------------------------

test('verifyTriggerLabel returns label_verified true for the expected trigger label', () => {
  const result = verifyTriggerLabel('verify-openspec');
  assert.equal(result.label_verified, true);
  assert.ok(result.label_verified_reason.length > 0, 'expected a non-empty reason');
});

test('verifyTriggerLabel returns label_verified false for a different label', () => {
  const result = verifyTriggerLabel('some-other-label');
  assert.equal(result.label_verified, false);
  assert.ok(
    result.label_verified_reason.includes('some-other-label'),
    `expected received label in reason, got: ${result.label_verified_reason}`
  );
});

test('verifyTriggerLabel returns label_verified false for an empty string with (empty) in reason', () => {
  const result = verifyTriggerLabel('');
  assert.equal(result.label_verified, false);
  assert.ok(
    result.label_verified_reason.includes('(empty)'),
    `expected "(empty)" in reason, got: ${result.label_verified_reason}`
  );
});

// ---------------------------------------------------------------------------
// classify-pr.js — classifyPullRequest
// ---------------------------------------------------------------------------

test('classifyPullRequest allows archive/push when headRepoId matches baseRepoId', () => {
  const result = classifyPullRequest({ headRepoId: 123, baseRepoId: 123 });
  assert.equal(result.archive_push_allowed, true);
  assert.ok(result.archive_push_allowed_reason.length > 0, 'expected a non-empty reason');
});

test('classifyPullRequest disallows archive/push when headRepoId differs from baseRepoId', () => {
  const result = classifyPullRequest({ headRepoId: 456, baseRepoId: 123 });
  assert.equal(result.archive_push_allowed, false);
  assert.ok(result.archive_push_allowed_reason.length > 0, 'expected a non-empty reason');
});

test('classifyPullRequest disallows archive/push when headRepoId is null (safe fallback)', () => {
  const result = classifyPullRequest({ headRepoId: null, baseRepoId: 123 });
  assert.equal(result.archive_push_allowed, false);
});

test('classifyPullRequest disallows archive/push when headRepoId is undefined (deleted fork repo)', () => {
  const result = classifyPullRequest({ headRepoId: undefined, baseRepoId: 123 });
  assert.equal(result.archive_push_allowed, false);
});

test('classifyPullRequest disallows archive/push when both headRepoId and baseRepoId are undefined', () => {
  const result = classifyPullRequest({ headRepoId: undefined, baseRepoId: undefined });
  assert.equal(result.archive_push_allowed, false);
});

// ---------------------------------------------------------------------------
// remove-trigger-label.js — removeTriggerLabel
// ---------------------------------------------------------------------------

test('removeTriggerLabel returns trigger_label_removed false when issueNumber is undefined', async () => {
  const result = await removeTriggerLabel({
    github: {},
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: undefined,
    labelName: 'verify-openspec',
  });
  assert.equal(result.trigger_label_removed, false);
  assert.ok(
    result.trigger_label_removed_reason.includes('No issue number'),
    `expected "No issue number" in reason, got: ${result.trigger_label_removed_reason}`
  );
});

test('removeTriggerLabel returns trigger_label_removed false when labelName is empty', async () => {
  const result = await removeTriggerLabel({
    github: {},
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    labelName: '   ',
  });
  assert.equal(result.trigger_label_removed, false);
  assert.ok(
    result.trigger_label_removed_reason.includes('No label name'),
    `expected "No label name" in reason, got: ${result.trigger_label_removed_reason}`
  );
});

test('removeTriggerLabel returns trigger_label_removed true on successful API call', async () => {
  let removedName;
  const mockGithub = {
    rest: {
      issues: {
        removeLabel: async (args) => {
          removedName = args.name;
          return {};
        },
      },
    },
  };
  const result = await removeTriggerLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    labelName: 'verify-openspec',
  });
  assert.equal(result.trigger_label_removed, true);
  assert.equal(removedName, 'verify-openspec');
});

test('removeTriggerLabel removes a factory label from an issue when API succeeds', async () => {
  let captured;
  const mockGithub = {
    rest: {
      issues: {
        removeLabel: async (args) => {
          captured = args;
          return {};
        },
      },
    },
  };
  const result = await removeTriggerLabel({
    github: mockGithub,
    context: { repo: { owner: 'acme', repo: 'demo' } },
    issueNumber: 99,
    labelName: 'code-factory',
  });
  assert.equal(result.trigger_label_removed, true);
  assert.deepEqual(captured, {
    owner: 'acme',
    repo: 'demo',
    issue_number: 99,
    name: 'code-factory',
  });
});

test('removeTriggerLabel returns trigger_label_removed true on 404 (idempotent — label already removed)', async () => {
  const err = new Error('Not Found');
  err.status = 404;
  const mockGithub = {
    rest: {
      issues: {
        removeLabel: async () => {
          throw err;
        },
      },
    },
  };
  const result = await removeTriggerLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    labelName: 'change-factory',
  });
  assert.equal(result.trigger_label_removed, true);
  assert.match(result.trigger_label_removed_reason, /change-factory/);
});

test('removeTriggerLabel returns trigger_label_removed false on non-404 API error', async () => {
  const err = new Error('Internal Server Error');
  err.status = 500;
  const mockGithub = {
    rest: {
      issues: {
        removeLabel: async () => {
          throw err;
        },
      },
    },
  };
  const result = await removeTriggerLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    labelName: 'verify-openspec',
  });
  assert.equal(result.trigger_label_removed, false);
  assert.ok(
    result.trigger_label_removed_reason.includes('Internal Server Error'),
    `expected error message in reason, got: ${result.trigger_label_removed_reason}`
  );
});
