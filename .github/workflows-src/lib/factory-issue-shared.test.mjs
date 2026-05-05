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
