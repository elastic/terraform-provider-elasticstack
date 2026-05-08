import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  issueClosingReferencePattern,
  factoryQualifyTriggerEvent,
  factoryCheckActorTrust,
  factoryParseOptionalTriStateFromEnv,
  factoryParseFinalizeGateEnv,
  factoryActorTrustWhenSenderMissing,
  factoryCheckDuplicatePR,
  factoryComputeGateReason,
  createFactoryIssueIntake,
  createFactoryIssueModule,
} = require('./factory-issue-shared.js');

const {
  factoryFetchIssueComments,
  serializeIssueComments,
  COMMENT_CONTEXT_BUDGET,
} = require('./factory-issue-comments.js');

test('factoryParseOptionalTriStateFromEnv treats missing and empty as null', () => {
  assert.equal(factoryParseOptionalTriStateFromEnv(undefined), null);
  assert.equal(factoryParseOptionalTriStateFromEnv(''), null);
});

test('factoryParseOptionalTriStateFromEnv parses true only for exact true string', () => {
  assert.equal(factoryParseOptionalTriStateFromEnv('true'), true);
  assert.equal(factoryParseOptionalTriStateFromEnv('false'), false);
  assert.equal(factoryParseOptionalTriStateFromEnv('TRUE'), false);
});

test('factoryParseFinalizeGateEnv matches finalize_gate env semantics', () => {
  assert.deepEqual(
    factoryParseFinalizeGateEnv({}),
    {
      eventEligible: false,
      eventEligibleReason: '',
      actorTrusted: null,
      actorTrustedReason: null,
      duplicatePrFound: null,
      duplicatePrUrl: null,
      duplicateCheckGateReason: null,
    },
  );

  assert.deepEqual(
    factoryParseFinalizeGateEnv({
      EVENT_ELIGIBLE: 'true',
      EVENT_ELIGIBLE_REASON: 'ok',
      ACTOR_TRUSTED: 'true',
      ACTOR_TRUSTED_REASON: '',
      DUPLICATE_PR_FOUND: 'false',
      DUPLICATE_PR_URL: '',
      DUPLICATE_GATE_REASON: 'x',
    }),
    {
      eventEligible: true,
      eventEligibleReason: 'ok',
      actorTrusted: true,
      actorTrustedReason: '',
      duplicatePrFound: false,
      duplicatePrUrl: null,
      duplicateCheckGateReason: 'x',
    },
  );
});

test('factoryParseFinalizeGateEnv feeds factoryComputeGateReason for an all-pass path', () => {
  const parsed = factoryParseFinalizeGateEnv({
    EVENT_ELIGIBLE: 'true',
    EVENT_ELIGIBLE_REASON: 'eligible',
    ACTOR_TRUSTED: 'true',
    ACTOR_TRUSTED_REASON: 'trusted',
    DUPLICATE_PR_FOUND: 'false',
    DUPLICATE_PR_URL: 'https://example.com/pr/1',
    DUPLICATE_GATE_REASON: null,
  });
  const result = factoryComputeGateReason(parsed, 'code-factory');

  assert.match(result.gate_reason, /All deterministic gates passed/);
});

test('factoryComputeGateReason uses generic untrusted text when actorTrusted is false with falsy reason', () => {
  const result = factoryComputeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: false,
    actorTrustedReason: '',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
  }, 'change-factory');

  assert.equal(result.gate_reason, 'Trigger actor is not trusted.');
});

test('factoryComputeGateReason falls back to unknown URL when duplicate found without URL or override', () => {
  const result = factoryComputeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'trusted',
    duplicatePrFound: true,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
  }, 'change-factory');

  assert.match(result.gate_reason, /Found existing linked change-factory PR: \(unknown URL\)\./);
});

test('factoryQualifyTriggerEvent rejects non-issues events', () => {
  const result = factoryQualifyTriggerEvent({
    eventName: 'pull_request',
    eventAction: 'opened',
    labelName: '',
    issueLabels: [],
    factoryLabel: 'demo-factory',
    issueOpenedNotEligibleReason: 'n/a',
  });
  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /expected 'issues'/);
});

test('factoryCheckActorTrust trusts write permission', () => {
  const result = factoryCheckActorTrust({ sender: 'alice', permission: 'write' });
  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /permission 'write'/);
});

test('issueClosingReferencePattern matches GitHub closing keywords but not longer issue numbers', () => {
  const p = issueClosingReferencePattern(42);
  assert.equal(p.test('See fixes #42\n'), true);
  assert.equal(p.test('fixes #420'), false);
});

test('issueClosingReferencePattern does not match whitespace between # and the issue number', () => {
  const p = issueClosingReferencePattern(123);
  assert.equal(p.test('closes #123'), true);
  assert.equal(p.test('closes # 123'), false);
  assert.equal(p.test('Closes # 123'), false);
});

test('factoryActorTrustWhenSenderMissing matches stable contract', () => {
  const result = factoryActorTrustWhenSenderMissing();
  assert.equal(result.actor_trusted, false);
  assert.match(result.actor_trusted_reason, /sender login is missing/);
});

test('factoryCheckDuplicatePR coalesces html_url for closes-literal mode when duplicate html_url is missing', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'open',
      head_branch: 'code-factory/issue-42',
      labels: ['code-factory'],
      body: 'Closes #42',
      html_url: undefined,
    }],
    branchPrefix: 'code-factory/issue-',
    prLabel: 'code-factory',
    duplicateLinkageMode: 'closes-literal',
  });
  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

test('factoryCheckDuplicatePR coalesces html_url for github-keywords mode when duplicate html_url is missing', () => {
  const result = factoryCheckDuplicatePR({
    issueNumber: 42,
    pullRequests: [{
      number: 101,
      state: 'open',
      head_branch: 'change-factory/issue-42',
      labels: ['change-factory'],
      body: 'Closes #42',
      html_url: undefined,
    }],
    branchPrefix: 'change-factory/issue-',
    prLabel: 'change-factory',
    duplicateLinkageMode: 'github-keywords',
  });
  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

test('createFactoryIssueIntake: duplicateLinkageMode selects duplicate PR URL handling', () => {
  const code = createFactoryIssueIntake({
    branchPrefix: 'code-factory/issue-',
    factoryLabel: 'code-factory',
    issueOpenedNotEligibleReason: 'x',
    duplicateLinkageMode: 'closes-literal',
  });
  const change = createFactoryIssueIntake({
    branchPrefix: 'change-factory/issue-',
    factoryLabel: 'change-factory',
    issueOpenedNotEligibleReason: 'y',
    duplicateLinkageMode: 'github-keywords',
  });
  const pr = {
    number: 1,
    state: 'open',
    head_branch: 'code-factory/issue-9',
    labels: ['code-factory'],
    body: 'Closes #9',
    html_url: undefined,
  };
  const c = code.checkDuplicatePR({ issueNumber: 9, pullRequests: [pr] });
  assert.equal(c.duplicate_pr_url, null);
  const p = {
    ...pr,
    head_branch: 'change-factory/issue-9',
    labels: ['change-factory'],
  };
  const ch = change.checkDuplicatePR({ issueNumber: 9, pullRequests: [p] });
  assert.equal(ch.duplicate_pr_url, null);
});

test('factoryFetchIssueComments returns empty array for no comments', async () => {
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => [],
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.deepEqual(result.comments, []);
  assert.equal(result.truncated, false);
});

test('factoryFetchIssueComments filters out all bots', async () => {
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => [
      { user: { login: 'github-actions[bot]' }, created_at: '2024-01-01T00:00:00Z', body: 'bot1' },
      { user: { login: 'dependabot[bot]' }, created_at: '2024-01-02T00:00:00Z', body: 'bot2' },
    ],
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.deepEqual(result.comments, []);
  assert.equal(result.truncated, false);
});

test('factoryFetchIssueComments preserves human comments and excludes bots', async () => {
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => [
      { user: { login: 'alice' }, created_at: '2024-01-01T00:00:00Z', body: 'hello' },
      { user: { login: 'github-actions[bot]' }, created_at: '2024-01-02T00:00:00Z', body: 'bot' },
      { user: { login: 'bob' }, created_at: '2024-01-03T00:00:00Z', body: 'world' },
      { user: { login: 'dependabot[bot]' }, created_at: '2024-01-04T00:00:00Z', body: 'bot2' },
    ],
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.equal(result.comments.length, 2);
  assert.deepEqual(result.comments[0], { author: 'alice', createdAt: '2024-01-01T00:00:00Z', body: 'hello' });
  assert.deepEqual(result.comments[1], { author: 'bob', createdAt: '2024-01-03T00:00:00Z', body: 'world' });
  assert.equal(result.truncated, false);
});

test('factoryFetchIssueComments preserves ordering for multi-item results', async () => {
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => [
      { user: { login: 'alice' }, created_at: '2024-01-01T00:00:00Z', body: 'page1' },
      { user: { login: 'bob' }, created_at: '2024-01-02T00:00:00Z', body: 'page2' },
      { user: { login: 'carol' }, created_at: '2024-01-03T00:00:00Z', body: 'page3' },
    ],
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.equal(result.comments.length, 3);
  assert.deepEqual(result.comments[2], { author: 'carol', createdAt: '2024-01-03T00:00:00Z', body: 'page3' });
  assert.equal(result.truncated, false);
});

test('factoryFetchIssueComments truncates at 200 comments', async () => {
  const manyComments = [];
  for (let i = 0; i < 250; i++) {
    manyComments.push({
      user: { login: `user${i}` },
      created_at: `2024-01-01T00:00:${String(i % 60).padStart(2, '0')}Z`,
      body: `comment ${i}`,
    });
  }
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => manyComments,
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.equal(result.comments.length, 200);
  assert.equal(result.truncated, true);
  assert.equal(result.comments[0].author, 'user0');
  assert.equal(result.comments[199].author, 'user199');
});

test('factoryFetchIssueComments truncates at 200 with interspersed bots', async () => {
  // 198 humans, then 5 bots, then 55 more humans → should get 200 humans, truncated
  const comments = [];
  for (let i = 0; i < 198; i++) {
    comments.push({ user: { login: `user${i}` }, created_at: '2024-01-01T00:00:00Z', body: `h${i}` });
  }
  for (let i = 0; i < 5; i++) {
    comments.push({ user: { login: 'github-actions[bot]' }, created_at: '2024-01-02T00:00:00Z', body: 'bot' });
  }
  for (let i = 0; i < 55; i++) {
    comments.push({ user: { login: `late${i}` }, created_at: '2024-01-03T00:00:00Z', body: `l${i}` });
  }
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => comments,
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  assert.equal(result.comments.length, 200);
  assert.equal(result.truncated, true);
  assert.equal(result.comments[197].author, 'user197');
  assert.equal(result.comments[198].author, 'late0');
});

test('factoryFetchIssueComments handles nullish/ghost comment fields', async () => {
  const github = {
    rest: { issues: { listComments: async () => [] } },
    paginate: async () => [
      { user: null, created_at: undefined, body: null },
    ],
  };
  const result = await factoryFetchIssueComments({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    issueNumber: 1,
  });
  // null user → not a bot (login undefined, doesn't end with '[bot]') → included
  assert.equal(result.comments.length, 1);
  assert.deepEqual(result.comments[0], { author: '', createdAt: '', body: '' });
  assert.equal(result.truncated, false);
});

test('createFactoryIssueModule binds shared exports and branch aliases', () => {
  const mod = createFactoryIssueModule({
    branchPrefix: 'demo-factory/issue-',
    factoryLabel: 'demo-factory',
    issueOpenedNotEligibleReason: 'not eligible',
    duplicateLinkageMode: 'closes-literal',
    issueBranchNameAliases: ['demoFactoryIssueBranchName'],
  });

  assert.equal(mod.issueBranchName(42), 'demo-factory/issue-42');
  assert.equal(mod.demoFactoryIssueBranchName(42), 'demo-factory/issue-42');
  assert.deepEqual(mod.actorTrustWhenSenderMissing(), factoryActorTrustWhenSenderMissing());
  assert.equal(mod.parseOptionalTriStateFromEnv('true'), true);
  assert.deepEqual(mod.parseFinalizeGateEnv({}), factoryParseFinalizeGateEnv({}));

  const event = mod.qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'demo-factory',
    issueLabels: [],
  });
  assert.equal(event.event_eligible, true);
});

test('serializeIssueComments returns empty string for empty input', () => {
  assert.equal(serializeIssueComments({ comments: [], truncated: false }), '');
  assert.equal(serializeIssueComments({ comments: [], truncated: true }), '');
});

test('serializeIssueComments renders a single comment correctly', () => {
  const result = serializeIssueComments({
    comments: [{ author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: 'Hello world.' }],
    truncated: false,
  });
  assert.match(result, /\*\*@alice\*\* \(2024-01-01T12:00:00Z\):/);
  assert.match(result, /Hello world\./);
  assert.match(result, /---/);
});

test('serializeIssueComments renders multiple comments in order', () => {
  const result = serializeIssueComments({
    comments: [
      { author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: 'First.' },
      { author: 'bob', createdAt: '2024-01-02T12:00:00Z', body: 'Second.' },
      { author: 'carol', createdAt: '2024-01-03T12:00:00Z', body: 'Third.' },
    ],
    truncated: false,
  });
  const firstIndex = result.indexOf('First.');
  const secondIndex = result.indexOf('Second.');
  const thirdIndex = result.indexOf('Third.');
  assert.ok(firstIndex < secondIndex);
  assert.ok(secondIndex < thirdIndex);
  assert.equal(result.includes('[... comment history truncated at 200 comments]'), false);
});

test('serializeIssueComments truncates at budget and appends marker', () => {
  const longBody = 'x'.repeat(30_000);
  const result = serializeIssueComments({
    comments: [
      { author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: longBody },
      { author: 'bob', createdAt: '2024-01-02T12:00:00Z', body: longBody },
      { author: 'carol', createdAt: '2024-01-03T12:00:00Z', body: longBody },
    ],
    truncated: false,
  });
  assert.ok(result.includes('alice'));
  // Output must stay within budget (with overhead for markers)
  assert.ok(result.length <= COMMENT_CONTEXT_BUDGET, `Expected output <= ${COMMENT_CONTEXT_BUDGET} chars, got ${result.length}`);
  assert.ok(result.includes('[... 1 more comments truncated for context budget]') || result.includes('[... 2 more comments truncated for context budget]'));
});

test('serializeIssueComments truncates body of single oversized comment', () => {
  // Single comment whose body alone exceeds COMMENT_CONTEXT_BUDGET
  const oversizedBody = 'y'.repeat(COMMENT_CONTEXT_BUDGET + 1_000);
  const result = serializeIssueComments({
    comments: [{ author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: oversizedBody }],
    truncated: false,
  });
  assert.ok(result.length <= COMMENT_CONTEXT_BUDGET, `Expected output <= ${COMMENT_CONTEXT_BUDGET} chars, got ${result.length}`);
  assert.ok(result.includes('alice'));
});

test('serializeIssueComments breaks when remaining budget cannot fit comment frame', () => {
  // Construct a first comment that fills budget to leave less than a frame for the next
  const headerLen = '**@a** (2024-01-01T12:00:00Z):\n\n'.length; // 32
  const footerLen = '\n\n---\n'.length; // 7
  const frameLen = headerLen + footerLen; // 39
  const bodyBudget = COMMENT_CONTEXT_BUDGET - 200; // 49800

  const firstBody = 'x'.repeat(bodyBudget - frameLen - 1);
  const result = serializeIssueComments({
    comments: [
      { author: 'a', createdAt: '2024-01-01T12:00:00Z', body: firstBody },
      { author: 'b', createdAt: '2024-01-02T12:00:00Z', body: 'should not appear' },
    ],
    truncated: false,
  });

  assert.ok(result.length <= COMMENT_CONTEXT_BUDGET, `Expected output <= ${COMMENT_CONTEXT_BUDGET} chars, got ${result.length}`);
  assert.ok(result.includes('a'));
  assert.ok(!result.includes('should not appear'));
  assert.ok(result.includes('[... 1 more comments truncated for context budget]'));
});

test('serializeIssueComments appends both markers when budget and fetch-cap both apply', () => {
  const longBody = 'z'.repeat(30_000);
  const result = serializeIssueComments({
    comments: [
      { author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: longBody },
      { author: 'bob', createdAt: '2024-01-02T12:00:00Z', body: longBody },
      { author: 'carol', createdAt: '2024-01-03T12:00:00Z', body: longBody },
    ],
    truncated: true, // fetch-cap was hit
  });
  assert.ok(result.includes('[... comment history truncated at 200 comments]'), 'fetch-cap marker missing');
  assert.ok(
    result.includes('[... 1 more comments truncated for context budget]') ||
    result.includes('[... 2 more comments truncated for context budget]'),
    'budget marker missing',
  );
  const budgetIdx = result.indexOf('[... ');
  const capIdx = result.lastIndexOf('[... comment history truncated at 200 comments]');
  assert.ok(budgetIdx <= capIdx, 'budget marker should appear before fetch-cap marker');
});

test('serializeIssueComments appends fetch-cap note when truncated is true', () => {
  const result = serializeIssueComments({
    comments: [{ author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: 'Hello.' }],
    truncated: true,
  });
  assert.ok(result.includes('[... comment history truncated at 200 comments]'));
});

test('serializeIssueComments produces stable output for identical input', () => {
  const input = {
    comments: [
      { author: 'alice', createdAt: '2024-01-01T12:00:00Z', body: 'Hello world.' },
      { author: 'bob', createdAt: '2024-01-02T12:00:00Z', body: 'Second comment.' },
    ],
    truncated: false,
  };
  const run1 = serializeIssueComments(input);
  const run2 = serializeIssueComments(input);
  assert.equal(run1, run2);
});
