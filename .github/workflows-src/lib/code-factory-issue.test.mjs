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
  issueBranchName,
  actorTrustWhenSenderMissing,
  parseOptionalTriStateFromEnv,
  parseFinalizeGateEnv,
} = require('./code-factory-issue.js');
const { ISSUE_BRANCH_PREFIX, FACTORY_LABEL } = require('../code-factory-issue/intake-constants.js');

const workflowPath = path.resolve(__dirname, '../../workflows/code-factory-issue.md');
const lockPath = path.resolve(__dirname, '../../workflows/code-factory-issue.lock.yml');
const codeFactoryScriptsDir = path.resolve(__dirname, '../code-factory-issue/scripts');
const codeFactoryWorkflowTmplPath = path.resolve(__dirname, '../code-factory-issue/workflow.md.tmpl');

function makePullRequest(overrides = {}) {
  return {
    number: 101,
    state: 'open',
    head_branch: `${ISSUE_BRANCH_PREFIX}42`,
    labels: [FACTORY_LABEL],
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

test('actorTrustWhenSenderMissing matches shared helper', () => {
  const { factoryActorTrustWhenSenderMissing } = require('./factory-issue-shared.js');
  assert.deepEqual(actorTrustWhenSenderMissing(), factoryActorTrustWhenSenderMissing());
});

test('code-factory parseFinalizeGateEnv and tri-state parser match shared implementation', () => {
  const {
    factoryParseFinalizeGateEnv,
    factoryParseOptionalTriStateFromEnv,
  } = require('./factory-issue-shared.js');
  assert.deepEqual(parseFinalizeGateEnv({}), factoryParseFinalizeGateEnv({}));
  assert.equal(parseOptionalTriStateFromEnv('true'), factoryParseOptionalTriStateFromEnv('true'));
});

test('parseFinalizeGateEnv feeds computeGateReason for an all-pass path', () => {
  const parsed = parseFinalizeGateEnv({
    EVENT_ELIGIBLE: 'true',
    EVENT_ELIGIBLE_REASON: 'eligible',
    ACTOR_TRUSTED: 'true',
    ACTOR_TRUSTED_REASON: 'trusted',
    DUPLICATE_PR_FOUND: 'false',
    DUPLICATE_PR_URL: 'https://example.com/pr/1',
    DUPLICATE_GATE_REASON: null,
  });
  const result = computeGateReason(parsed);

  assert.match(result.gate_reason, /All deterministic gates passed/);
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
    duplicateCheckGateReason: 'No duplicate PR found.',
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
    duplicateCheckGateReason: 'No duplicate PR found.',
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
    duplicateCheckGateReason: null,
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
    duplicateCheckGateReason: null,
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
  assert.match(source, /compile-workflow-sources/);
});

test('code-factory-issue lock file is compiled and exists', () => {
  const lock = readFileSync(lockPath, 'utf8');
  assert.ok(lock.length > 0);
  assert.match(lock, /# gh-aw-metadata:/);
  assert.match(lock, /DO NOT EDIT/);
});

test('computeGateReason returns unknown reason when actorTrusted is null (step skipped)', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: null,
    actorTrustedReason: null,
    duplicatePrFound: null,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
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
    duplicateCheckGateReason: null,
  });

  assert.match(result.gate_reason, /Duplicate PR check did not complete/);
});

test('issueBranchName matches deterministic branch naming', () => {
  assert.equal(issueBranchName(42), 'code-factory/issue-42');
});

test('code-factory-issue exports align with shared createFactoryIssueIntake binding', () => {
  const { createFactoryIssueIntake } = require('./factory-issue-shared.js');
  const {
    ISSUE_BRANCH_PREFIX: prefix,
    FACTORY_LABEL: label,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON: openedReason,
  } = require('../code-factory-issue/intake-constants.js');
  const bound = createFactoryIssueIntake({
    branchPrefix: prefix,
    factoryLabel: label,
    issueOpenedNotEligibleReason: openedReason,
    duplicateLinkageMode: 'closes-literal',
  });
  const params = { eventName: 'issues', eventAction: 'labeled', labelName: 'code-factory', issueLabels: [] };
  assert.deepEqual(qualifyTriggerEvent(params), bound.qualifyTriggerEvent(params));
  assert.deepEqual(
    checkActorTrust({ sender: 'alice', permission: 'write' }),
    bound.checkActorTrust({ sender: 'alice', permission: 'write' }),
  );
  assert.deepEqual(
    checkDuplicatePR({ issueNumber: 7, pullRequests: [] }),
    bound.checkDuplicatePR({ issueNumber: 7, pullRequests: [] }),
  );
});

test('code-factory intake constants stay aligned with workflow template branch prefix', () => {
  const workflowTmpl = readFileSync(codeFactoryWorkflowTmplPath, 'utf8');
  const branchExpr = `${ISSUE_BRANCH_PREFIX}\${{ github.event.issue.number }}`;
  assert.ok(
    workflowTmpl.includes(branchExpr),
    'workflow.md.tmpl must express branches with ISSUE_BRANCH_PREFIX + ${{ github.event.issue.number }}',
  );
});

test('code-factory-issue workflow template enables status comments and remove-label pre-activation', () => {
  const workflowTmpl = readFileSync(codeFactoryWorkflowTmplPath, 'utf8');
  assert.match(workflowTmpl, /status-comment:\s*true/);
  assert.match(workflowTmpl, /name: Remove trigger label/);
  assert.match(workflowTmpl, /x-script-include: scripts\/remove_trigger_label\.inline\.js/);
  assert.match(workflowTmpl, /issues:\s*write/);
  assert.match(workflowTmpl, /trigger_label_removed:/);
});

test('code-factory-issue inline scripts include intake constants before shared helpers', () => {
  const expectedHeader = [
    /^\/\/include: \.\.\/intake-constants\.js\n/,
    /^\/\/include: \.\.\/\.\.\/lib\/factory-issue-shared\.js\n/,
    /^\/\/include: \.\.\/\.\.\/lib\/code-factory-issue\.gh\.js\n/,
  ];
  for (const name of [
    'qualify_trigger.inline.js',
    'check_actor_trust.inline.js',
    'check_duplicate_pr.inline.js',
    'finalize_gate.inline.js',
  ]) {
    const source = readFileSync(path.join(codeFactoryScriptsDir, name), 'utf8');
    let offset = 0;
    for (const pat of expectedHeader) {
      const slice = source.slice(offset);
      const m = pat.exec(slice);
      assert.ok(m, `expected include line matching ${pat} in ${name} at offset ${offset}`);
      offset += m.index + m[0].length;
    }
  }
});

test('code-factory-issue finalize_gate.inline.js uses shared parseFinalizeGateEnv path', () => {
  const source = readFileSync(path.join(codeFactoryScriptsDir, 'finalize_gate.inline.js'), 'utf8');
  assert.match(source, /computeGateReason\(parseFinalizeGateEnv\(process\.env\)\)/);
});

test('code-factory-issue check_actor_trust.inline.js uses actorTrustWhenSenderMissing', () => {
  const source = readFileSync(path.join(codeFactoryScriptsDir, 'check_actor_trust.inline.js'), 'utf8');
  assert.match(source, /actorTrustWhenSenderMissing\(\)/);
});
