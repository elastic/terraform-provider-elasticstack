import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);
const {
  qualifyTriggerEvent,
  checkActorTrust,
  checkDuplicatePR,
  computeGateReason,
} = require('./code-factory-issue.js');

const workflowPath = path.resolve(__dirname, '../../workflows/code-factory-issue.md');
const lockPath = path.resolve(__dirname, '../../workflows/code-factory-issue.lock.yml');

function makePullRequest(overrides = {}) {
  return {
    number: 101,
    state: 'open',
    head_branch: 'code-factory/issue-42',
    labels: ['code-factory'],
    body: 'Implements the requested change.\n\nCloses #42',
    html_url: 'https://github.com/elastic/terraform-provider-elasticstack/pull/101',
    ...overrides,
  };
}

test('qualifyTriggerEvent accepts issues.labeled with the code-factory label', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'code-factory',
    issueLabels: [],
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /applied label is code-factory/);
});

test('qualifyTriggerEvent rejects issues.labeled with a different label', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'bug',
    issueLabels: ['bug'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not 'code-factory'/);
});

test('qualifyTriggerEvent accepts issues.opened when code-factory is in initial labels', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
    issueLabels: ['triage', 'code-factory'],
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /already has the code-factory label/);
});

test('qualifyTriggerEvent rejects issues.opened without code-factory in initial labels', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
    issueLabels: ['enhancement'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /created without the code-factory label/);
});

test('qualifyTriggerEvent rejects non-issues events', () => {
  const result = qualifyTriggerEvent({
    eventName: 'pull_request',
    eventAction: 'opened',
    labelName: 'code-factory',
    issueLabels: ['code-factory'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /expected 'issues'/);
});

test('qualifyTriggerEvent rejects unsupported issues actions such as closed', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'closed',
    labelName: 'code-factory',
    issueLabels: ['code-factory'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not eligible/);
});

test('checkActorTrust trusts github-actions[bot] without collaborator permission', () => {
  const result = checkActorTrust({ sender: 'github-actions[bot]', permission: null });

  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /github-actions\[bot\]/);
});

test('checkActorTrust still trusts github-actions[bot] when a permission value is present', () => {
  const result = checkActorTrust({ sender: 'github-actions[bot]', permission: 'read' });

  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /github-actions\[bot\]/);
});

test('checkActorTrust trusts human senders with write permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: 'write' });

  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /permission 'write'/);
});

test('checkActorTrust trusts human senders with maintain permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: 'maintain' });

  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /permission 'maintain'/);
});

test('checkActorTrust trusts human senders with admin permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: 'admin' });

  assert.equal(result.actor_trusted, true);
  assert.match(result.actor_trusted_reason, /permission 'admin'/);
});

test('checkActorTrust rejects human senders with read permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: 'read' });

  assert.equal(result.actor_trusted, false);
  assert.match(result.actor_trusted_reason, /does not meet the required write\/maintain\/admin policy/);
});

test('checkActorTrust rejects human senders with none permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: 'none' });

  assert.equal(result.actor_trusted, false);
  assert.match(result.actor_trusted_reason, /permission 'none'/);
});

test('checkActorTrust rejects human senders with null permission', () => {
  const result = checkActorTrust({ sender: 'alice', permission: null });

  assert.equal(result.actor_trusted, false);
  assert.match(result.actor_trusted_reason, /permission '\(none\)'/);
});

test('checkDuplicatePR reports no duplicate when there are no open PRs', () => {
  const result = checkDuplicatePR({ issueNumber: 42, pullRequests: [] });

  assert.equal(result.duplicate_pr_found, false);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /No open linked code-factory PR found/);
});

test('checkDuplicatePR finds a duplicate when one PR matches all four criteria', () => {
  const pr = makePullRequest();
  const result = checkDuplicatePR({ issueNumber: 42, pullRequests: [pr] });

  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, pr.html_url);
  assert.match(result.gate_reason, /Found existing linked code-factory PR #101/);
});

test('checkDuplicatePR ignores PRs missing the code-factory label', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ labels: ['enhancement'] })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR ignores PRs on the wrong branch name', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ head_branch: 'feature/issue-42' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR ignores PRs missing canonical Closes issue linkage', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Implements the requested change without canonical metadata.' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR does not match a PR whose body has Closes issue linkage followed by more digits', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Implements the requested change.\n\nCloses #420' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR matches when body has canonical Closes issue linkage at end of line', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Implements the requested change.\n\nCloses #42\n' })],
  });

  assert.equal(result.duplicate_pr_found, true);
});

test('checkDuplicatePR ignores PRs that are not open', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ state: 'closed' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR ignores unrelated PRs that mention the issue only in the title', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [
      makePullRequest({
        head_branch: 'feature/unrelated',
        labels: ['maintenance'],
        body: 'Title mentions issue 42 but body does not include the canonical linkage.',
      }),
    ],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR finds the matching duplicate when only one of multiple PRs qualifies', () => {
  const matching = makePullRequest({
    number: 202,
    html_url: 'https://github.com/elastic/terraform-provider-elasticstack/pull/202',
  });
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [
      makePullRequest({ number: 200, labels: ['bug'] }),
      makePullRequest({ number: 201, head_branch: 'feature/issue-42' }),
      matching,
    ],
  });

  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, matching.html_url);
  assert.match(result.gate_reason, /#202/);
});

test('computeGateReason returns the event eligibility failure reason first', () => {
  const result = computeGateReason({
    eventEligible: false,
    eventEligibleReason: 'Event is not eligible.',
    actorTrusted: true,
    actorTrustedReason: 'Actor is trusted.',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    noDuplicateReason: 'No duplicate PR found.',
  });

  assert.equal(result.gate_reason, 'Event is not eligible.');
});

test('computeGateReason returns the actor trust failure when the event is eligible but actor is untrusted', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: false,
    actorTrustedReason: 'Actor is not trusted.',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    noDuplicateReason: 'No duplicate PR found.',
  });

  assert.equal(result.gate_reason, 'Actor is not trusted.');
});

test('computeGateReason mentions the duplicate PR URL when a duplicate is found', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'Actor is trusted.',
    duplicatePrFound: true,
    duplicatePrUrl: 'https://github.com/elastic/terraform-provider-elasticstack/pull/303',
    noDuplicateReason: null,
  });

  assert.match(result.gate_reason, /https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/pull\/303/);
});

test('computeGateReason returns the success reason when all gates pass', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'Actor is trusted.',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    noDuplicateReason: null,
  });

  assert.equal(
    result.gate_reason,
    'All deterministic gates passed: event eligible, actor trusted, and no linked code-factory PR found.',
  );
});

test('code-factory-issue workflow is compiled and exists', () => {
  const source = readFileSync(workflowPath, 'utf8');
  assert.match(source, /code-factory/);
  assert.match(source, /issues/);
});

test('code-factory-issue lock file is compiled and exists', () => {
  const lock = readFileSync(lockPath, 'utf8');
  assert.ok(lock.length > 0);
});

test('computeGateReason returns unknown reason when actorTrusted is null (step skipped)', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: null,
    actorTrustedReason: null,
    duplicatePrFound: null,
    duplicatePrUrl: null,
    noDuplicateReason: null,
  });

  assert.match(result.gate_reason, /Actor trust could not be determined/);
});

test('computeGateReason returns unknown reason when duplicatePrFound is null (step skipped)', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'Actor is trusted.',
    duplicatePrFound: null,
    duplicatePrUrl: null,
    noDuplicateReason: null,
  });

  assert.match(result.gate_reason, /Duplicate PR check did not complete/);
});
