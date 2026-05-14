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
  changeFactoryIssueBranchName,
  issueBranchName,
  qualifyTriggerEvent,
  actorTrustWhenSenderMissing,
  checkActorTrust,
  checkDuplicatePR,
  computeGateReason,
  parseOptionalTriStateFromEnv,
  parseFinalizeGateEnv,
} = require('./change-factory-issue.js');
const { ISSUE_BRANCH_PREFIX, DUPLICATE_LINKAGE_MODE } = require('../change-factory-issue/intake-constants.js');

const scriptsDir = path.resolve(__dirname, '../change-factory-issue/scripts');
const workflowTemplatePath = path.resolve(__dirname, '../change-factory-issue/workflow.md.tmpl');
const workflowCompiledPath = path.resolve(__dirname, '../../workflows/change-factory-issue.md');
const lockCompiledPath = path.resolve(__dirname, '../../workflows/change-factory-issue.lock.yml');
const inlineScripts = [
  'qualify_trigger.inline.js',
  'capture_command_text.inline.js',
  'check_actor_trust.inline.js',
  'check_duplicate_pr.inline.js',
  'notify_duplicate_blocked.inline.js',
  'finalize_gate.inline.js',
];

function makePullRequest(overrides = {}) {
  return {
    number: 101,
    state: 'open',
    head_branch: changeFactoryIssueBranchName(42),
    labels: ['change-factory'],
    body: 'Proposes the OpenSpec change.\n\nRelated to #42',
    html_url: 'https://github.com/elastic/terraform-provider-elasticstack/pull/101',
    ...overrides,
  };
}

test('change-factory-issue exports align with shared createFactoryIssueModule binding', () => {
  const { createFactoryIssueModule } = require('./factory-issue-shared.js');
  const {
    ISSUE_BRANCH_PREFIX: prefix,
    FACTORY_LABEL: label,
    DUPLICATE_LINKAGE_MODE: duplicateLinkageMode,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON: openedReason,
  } = require('../change-factory-issue/intake-constants.js');
  const bound = createFactoryIssueModule({
    branchPrefix: prefix,
    factoryLabel: label,
    issueOpenedNotEligibleReason: openedReason,
    duplicateLinkageMode,
    issueBranchNameAliases: ['changeFactoryIssueBranchName'],
  });
  const params = { eventName: 'issues', eventAction: 'labeled', labelName: 'change-factory', issueLabels: [] };
  assert.deepEqual(qualifyTriggerEvent(params), bound.qualifyTriggerEvent(params));
  assert.deepEqual(
    checkActorTrust({ sender: 'alice', permission: 'write' }),
    bound.checkActorTrust({ sender: 'alice', permission: 'write' }),
  );
  assert.deepEqual(
    checkDuplicatePR({ issueNumber: 7, pullRequests: [] }),
    bound.checkDuplicatePR({ issueNumber: 7, pullRequests: [] }),
  );
  assert.equal(changeFactoryIssueBranchName(7), bound.changeFactoryIssueBranchName(7));
});

test('qualifyTriggerEvent accepts issues.labeled with the change-factory label', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'change-factory',
    issueLabels: [],
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /applied label is change-factory/);
});

test('qualifyTriggerEvent rejects issues.labeled with a different label', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: 'bug',
    issueLabels: ['bug'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /not 'change-factory'/);
});

test('qualifyTriggerEvent rejects issues.labeled when the applied label name is empty', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'labeled',
    labelName: '',
    issueLabels: [],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /applied label is '\(empty\)'/);
});

test('qualifyTriggerEvent accepts issues.opened when change-factory is in initial labels', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
    issueLabels: ['triage', 'change-factory'],
  });

  assert.equal(result.event_eligible, true);
  assert.match(result.event_eligible_reason, /already has the change-factory label/);
});

test('qualifyTriggerEvent rejects issues.opened without change-factory in initial labels', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
    issueLabels: ['enhancement'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /created without the change-factory label/);
});

test('qualifyTriggerEvent rejects issues.opened when issue labels are null', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
    issueLabels: null,
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /missing/);
});

test('qualifyTriggerEvent rejects issues.opened when issue labels are undefined', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'opened',
    labelName: '',
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /missing/);
});

test('qualifyTriggerEvent rejects non-issues events', () => {
  const result = qualifyTriggerEvent({
    eventName: 'pull_request',
    eventAction: 'opened',
    labelName: 'change-factory',
    issueLabels: ['change-factory'],
  });

  assert.equal(result.event_eligible, false);
  assert.match(result.event_eligible_reason, /expected 'issues'/);
});

test('qualifyTriggerEvent rejects unsupported issues actions such as closed', () => {
  const result = qualifyTriggerEvent({
    eventName: 'issues',
    eventAction: 'closed',
    labelName: 'change-factory',
    issueLabels: ['change-factory'],
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

test('actorTrustWhenSenderMissing matches check_actor_trust inline missing-sender path', () => {
  const result = actorTrustWhenSenderMissing();

  assert.equal(result.actor_trusted, false);
  assert.match(result.actor_trusted_reason, /sender login is missing/);
});

test('checkDuplicatePR reports no duplicate when there are no open PRs', () => {
  const result = checkDuplicatePR({ issueNumber: 42, pullRequests: [] });

  assert.equal(result.duplicate_pr_found, false);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /No open linked change-factory PR found/);
});

test('checkDuplicatePR finds a duplicate when one PR matches all four criteria', () => {
  const pr = makePullRequest();
  const result = checkDuplicatePR({ issueNumber: 42, pullRequests: [pr] });

  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, pr.html_url);
  assert.match(result.gate_reason, /Found existing linked change-factory PR #101/);
});

test('checkDuplicatePR treats missing PR html_url as unknown in gate_reason and duplicate_pr_url', () => {
  const pr = makePullRequest({ html_url: undefined });
  const result = checkDuplicatePR({ issueNumber: 42, pullRequests: [pr] });

  assert.equal(result.duplicate_pr_found, true);
  assert.equal(result.duplicate_pr_url, null);
  assert.match(result.gate_reason, /\(unknown URL\)/);
});

test('checkDuplicatePR ignores PRs missing the change-factory label', () => {
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

test('checkDuplicatePR ignores PRs missing issue-closing reference', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Proposes the change without canonical metadata.' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR does not match GitHub closing keywords (related-literal mode)', () => {
  for (const body of [
    'See description.\n\ncloses #42',
    'FIXES #42',
    'Fixed #42',
    'Resolve #42',
    'RESOLVED #42',
    'Closes #42',
  ]) {
    const result = checkDuplicatePR({
      issueNumber: 42,
      pullRequests: [makePullRequest({ body })],
    });
    assert.equal(
      result.duplicate_pr_found,
      false,
      `expected no match for closing-keyword body in related-literal mode: ${body}`,
    );
  }
});

test('checkDuplicatePR ignores issue linkage with whitespace between # and the issue number', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'See description.\n\nRelated to # 42' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR does not match a PR whose body has Related to linkage followed by more digits', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Proposes the change.\n\nRelated to #420' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR matches when body has canonical Related to linkage at end of line', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Proposes the change.\n\nRelated to #42\n' })],
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
    'All deterministic gates passed: event eligible, actor trusted, and no linked change-factory PR found.',
  );
});

test('change-factory-issue workflow is compiled and exists', () => {
  const source = readFileSync(workflowCompiledPath, 'utf8');
  assert.match(source, /change-factory/);
  assert.match(source, /issues/);
  assert.match(source, /compile-workflow-sources/);
});

test('change-factory-issue lock file is compiled and exists', () => {
  const lock = readFileSync(lockCompiledPath, 'utf8');
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

test('changeFactoryIssueBranchName stays aligned with workflow template prefix', () => {
  assert.equal(ISSUE_BRANCH_PREFIX, 'change-factory/issue-');
  assert.equal(DUPLICATE_LINKAGE_MODE, 'related-literal');
  assert.equal(issueBranchName(42), 'change-factory/issue-42');
  assert.equal(changeFactoryIssueBranchName(42), 'change-factory/issue-42');

  const workflowTmpl = readFileSync(workflowTemplatePath, 'utf8');
  const branchExpr = `${ISSUE_BRANCH_PREFIX}\${{ github.event.issue.number }}`;
  assert.ok(
    workflowTmpl.includes(branchExpr),
    'workflow.md.tmpl must express branches with ISSUE_BRANCH_PREFIX + ${{ github.event.issue.number }}',
  );
});

test('change-factory-issue workflow.md.tmpl wiring matches intake contract', () => {
  const workflowTmpl = readFileSync(workflowTemplatePath, 'utf8');

  assert.match(workflowTmpl, /\non:\n  issues:\n    types: \[labeled\]/);
  assert.match(
    workflowTmpl,
    /slash_command:\n    name: change-factory\n    events: \[issue_comment\]/,
  );
  assert.match(workflowTmpl, /status-comment:\s*true/);
  assert.match(workflowTmpl, /issues:\s*write/);

  assert.match(
    workflowTmpl,
    /event_eligible: \$\{\{ steps\.qualify_trigger\.outputs\.event_eligible \}\}/,
  );
  assert.match(
    workflowTmpl,
    /event_eligible_reason: \$\{\{ steps\.qualify_trigger\.outputs\.event_eligible_reason \}\}/,
  );

  assert.match(
    workflowTmpl,
    /- name: Check actor trust\n      id: check_actor_trust\n      if: steps\.qualify_trigger\.outputs\.event_eligible == 'true'/,
  );
  assert.match(
    workflowTmpl,
    /- name: Capture command text\n      id: capture_command_text\n      if: steps\.qualify_trigger\.outputs\.event_eligible == 'true'/,
  );
  assert.match(
    workflowTmpl,
    /- name: Check duplicate PR\n      id: check_duplicate_pr\n      if: >-\n        steps\.qualify_trigger\.outputs\.event_eligible == 'true' &&\n        steps\.check_actor_trust\.outputs\.actor_trusted == 'true'/,
  );
  assert.match(
    workflowTmpl,
    /- name: Notify duplicate blocked\n      id: notify_duplicate_blocked\n      if: >-\n        steps\.qualify_trigger\.outputs\.event_eligible == 'true' &&\n        steps\.check_actor_trust\.outputs\.actor_trusted == 'true' &&\n        steps\.check_duplicate_pr\.outputs\.duplicate_pr_found == 'true'/,
  );

  const scriptIncludes = [
    'x-script-include: scripts/qualify_trigger.inline.js',
    'x-script-include: scripts/capture_command_text.inline.js',
    'x-script-include: scripts/check_actor_trust.inline.js',
    'x-script-include: scripts/check_duplicate_pr.inline.js',
    'x-script-include: scripts/notify_duplicate_blocked.inline.js',
    'x-script-include: scripts/remove_trigger_label.inline.js',
    'x-script-include: scripts/finalize_gate.inline.js',
  ];
  let lastIdx = -1;
  for (const line of scriptIncludes) {
    const idx = workflowTmpl.indexOf(line, lastIdx + 1);
    assert.ok(idx > lastIdx, `expected ordered script include: ${line}`);
    lastIdx = idx;
  }

  assert.match(
    workflowTmpl,
    /DUPLICATE_GATE_REASON: \$\{\{ steps\.check_duplicate_pr\.outputs\.gate_reason \}\}/,
  );

  assert.match(
    workflowTmpl,
    /- name: Remove trigger label\n      id: remove_trigger_label\n      if: >-\n        steps\.qualify_trigger\.outputs\.event_eligible == 'true' &&\n        steps\.check_actor_trust\.outputs\.actor_trusted == 'true' &&\n        steps\.check_duplicate_pr\.outputs\.duplicate_pr_found != 'true'/,
  );

  assert.match(
    workflowTmpl,
    /- name: Finalize gate reason\n      id: finalize_gate\n      if: always\(\)/,
  );

  assert.match(workflowTmpl, /trigger_label_removed:/);
  assert.match(workflowTmpl, /trigger_label_removed_reason:/);

  assert.match(
    workflowTmpl,
    /human_direction: \$\{\{ steps\.capture_command_text\.outputs\.human_direction \}\}/,
  );

  assert.match(
    workflowTmpl,
    /DUPLICATE_PR_URL: \$\{\{ steps\.check_duplicate_pr\.outputs\.duplicate_pr_url \}\}/,
  );

  assert.match(
    workflowTmpl,
    /issue_title: \$\{\{ steps\.capture_issue_context\.outputs\.issue_title \}\}/,
  );
  assert.match(
    workflowTmpl,
    /issue_body: \$\{\{ steps\.capture_issue_context\.outputs\.issue_body \}\}/,
  );
  assert.match(
    workflowTmpl,
    /gate_reason: \$\{\{ steps\.finalize_gate\.outputs\.gate_reason \}\}/,
  );
  assert.match(
    workflowTmpl,
    /actor_trusted: \$\{\{ steps\.check_actor_trust\.outputs\.actor_trusted \}\}/,
  );
  assert.match(
    workflowTmpl,
    /actor_trusted_reason: \$\{\{ steps\.check_actor_trust\.outputs\.actor_trusted_reason \}\}/,
  );
  assert.match(
    workflowTmpl,
    /duplicate_pr_found: \$\{\{ steps\.check_duplicate_pr\.outputs\.duplicate_pr_found \}\}/,
  );
  assert.match(
    workflowTmpl,
    /duplicate_pr_url: \$\{\{ steps\.check_duplicate_pr\.outputs\.duplicate_pr_url \}\}/,
  );

  assert.match(
    workflowTmpl,
    /if: >-\s*\n\s*needs\.pre_activation\.outputs\.event_eligible == 'true' &&\s*\n\s*needs\.pre_activation\.outputs\.actor_trusted == 'true' &&\s*\n\s*needs\.pre_activation\.outputs\.duplicate_pr_found != 'true'/,
  );

  assert.match(
    workflowTmpl,
    /- name: Setup Node\.js\n    uses: actions\/setup-node@v6\n    with:\n      node-version-file: package\.json/,
  );
  assert.match(workflowTmpl, /- name: Install npm dependencies\n    run: npm ci/);

  assert.match(
    workflowTmpl,
    /create-pull-request:\s*\n\s*labels: \[change-factory, no-changelog\]\s*\n\s*max: 1/,
  );
  assert.match(
    workflowTmpl,
    /add-comment:\s*\n\s*max: 1\s*\n\s*target: triggering/,
  );
  assert.match(workflowTmpl, /noop:\s*\n\s*max: 1\s*\n\s*report-as-issue: false/);

  const forbiddenFragments = [
    'docker-fleet',
    'create-es-api-key',
    'setup-kibana-fleet',
    'set-kibana-password',
    'Setup Elastic Stack',
    'Setup Fleet',
    'hashicorp/setup-terraform',
    'actions/setup-go@v',
  ];
  for (const fragment of forbiddenFragments) {
    assert.ok(
      !workflowTmpl.includes(fragment),
      `workflow template must not include ${fragment}`,
    );
  }
});

test('change-factory-issue agent prompt matches stable OpenSpec proposal contract', () => {
  const workflowTmpl = readFileSync(workflowTemplatePath, 'utf8');
  const promptMarker = '# Change Factory issue proposal worker';
  const promptStart = workflowTmpl.indexOf(promptMarker);
  assert.ok(promptStart >= 0, `expected ${promptMarker} in workflow.md.tmpl`);
  const prompt = workflowTmpl.slice(promptStart);

  assert.match(
    prompt,
    /Related to #\$\{\{\s*github\.event\.issue\.number\s*\}\}/,
    'expected canonical PR Related to linkage expression',
  );
  assert.doesNotMatch(
    prompt,
    /\bCloses #\$\{\{\s*github\.event\.issue\.number\s*\}\}/,
    'change-factory PR body must not auto-close the source issue (no Closes #N)',
  );
  assert.match(prompt, /exactly one/, 'expected exactly-one change / PR guidance');
  assert.match(prompt, /second change directory/, 'expected no second change directory');
  assert.match(prompt, /split the issue across multiple change ids/, 'expected no split across ids');
  assert.match(prompt, /\.openspec\.yaml/, 'expected change metadata file');
  assert.match(prompt, /openspec new change/, 'expected scaffold command');
  assert.match(prompt, /make build/, 'expected make build prohibition');
  assert.match(prompt, /`go test`/, 'expected go test prohibition');
  assert.match(prompt, /TestAcc/, 'expected TestAcc prohibition');
  assert.match(prompt, /TF_ACC/, 'expected TF_ACC prohibition');
  assert.match(
    prompt,
    /outside `openspec\/changes\/<change-id>\//,
    'expected provider edits confined to OpenSpec change tree wording',
  );
  assert.match(
    prompt,
    /speculative `openspec\/changes\/` files/,
    'expected no speculative change files on noop path',
  );
  assert.match(
    prompt,
    /\*\*must\*\* post \*\*exactly one\*\* `add-comment`[\s\S]*\*\*before\*\* any\s*`noop`/,
    'expected mandatory add-comment before noop on ambiguous path',
  );
  assert.match(
    prompt,
    /\*\*only\*\* `noop`[\s\S]*\*\*not\*\* allowed/,
    'expected noop-only completion forbidden when issue is ambiguous',
  );
  assert.match(prompt, /\*\*concise\*\*/, 'expected concise list of required facts in add-comment');
  assert.match(
    prompt,
    /contain \*\*only\*\* the OpenSpec change tree under `openspec\/changes\/<change-id>\/` for v1/,
    'expected v1 PR scope limited to OpenSpec change tree',
  );
  assert.match(prompt, /sole authoritative source/);
  assert.match(prompt, /proposal\.md/);
  assert.match(prompt, /design\.md/);
  assert.match(prompt, /tasks\.md/);
  assert.match(prompt, /specs\/<capability>\/spec\.md/);
  assert.match(prompt, /openspec validate <change-id> --type change/);
  assert.match(prompt, /OPENSPEC_TELEMETRY=0/);

  assert.match(
    prompt,
    /assume an \*\*Elastic Stack\*\*, \*\*Fleet\*\*/,
    'expected Elastic Stack and Fleet prohibitions in agent prompt',
  );
  assert.match(
    prompt,
    /\*\*Elasticsearch\s+API key\*\* creation flows/,
    'expected API key prohibition in agent prompt',
  );

  assert.match(
    prompt,
    /\*\*do not\*\* open a pull\s+request/,
    'expected noop path to forbid opening a PR',
  );
  assert.match(
    prompt,
    /back-and-forth comment thread/,
    'expected ambiguous path to forbid multi-reply comment threads',
  );
  assert.match(
    prompt,
    /exploration loop/,
    'expected guardrail against interactive comment exploration',
  );
  assert.match(prompt, /`add-comment`/, 'expected add-comment safe output for ambiguous issues');
  assert.match(prompt, /`no-changelog`/, 'expected no-changelog label in agent prompt');

  assert.match(
    prompt,
    /be the only open `change-factory` pull request for this issue/,
    'expected PR contract: single open change-factory PR for the issue',
  );

  assert.match(
    prompt,
    /open new issues/,
    'expected prompt to forbid opening new issues',
  );
  assert.match(
    prompt,
    /Docker compose stack services/,
    'expected Docker compose stack services prohibition in agent prompt',
  );
  assert.match(
    prompt,
    /add CI steps that provision runtime Elastic services/,
    'expected CI Elastic provisioning prohibition in agent prompt',
  );

  assert.doesNotMatch(prompt, /\u2026|\u2014|\u2018|\u2019|\u201c|\u201d/, 'prompt must use ASCII punctuation');
});

test('check_duplicate_pr.inline.js resolves expected branch via shared issueBranchName', () => {
  const source = readFileSync(path.join(scriptsDir, 'check_duplicate_pr.inline.js'), 'utf8');
  assert.match(source, /const expectedBranch = issueBranchName\(issueNumber\);/);
});

test('change-factory-issue inline scripts include shared intake helpers in dependency order', () => {
  const expectedHeader = [
    /^\/\/include: \.\.\/intake-constants\.js\n/,
    /^\/\/include: \.\.\/\.\.\/lib\/factory-issue-shared\.js\n/,
    /^\/\/include: \.\.\/\.\.\/lib\/factory-issue-module\.gh\.js\n/,
  ];
  for (const name of inlineScripts) {
    const source = readFileSync(path.join(scriptsDir, name), 'utf8');
    let offset = 0;
    for (const pat of expectedHeader) {
      const slice = source.slice(offset);
      const m = pat.exec(slice);
      assert.ok(m, `expected include line matching ${pat} in ${name} at offset ${offset}`);
      offset += m.index + m[0].length;
    }
  }
});

test('change-factory-issue finalize_gate.inline.js uses shared parseFinalizeGateEnv path', () => {
  const source = readFileSync(path.join(scriptsDir, 'finalize_gate.inline.js'), 'utf8');
  assert.match(source, /computeGateReason\(parseFinalizeGateEnv\(process\.env\)\)/);
});

test('change-factory-issue check_actor_trust.inline.js uses actorTrustWhenSenderMissing', () => {
  const source = readFileSync(path.join(scriptsDir, 'check_actor_trust.inline.js'), 'utf8');
  assert.match(source, /actorTrustWhenSenderMissing\(\)/);
});

test('change-factory-issue finalize helpers match shared implementation', () => {
  const {
    factoryParseFinalizeGateEnv,
    factoryParseOptionalTriStateFromEnv,
    factoryActorTrustWhenSenderMissing,
  } = require('./factory-issue-shared.js');
  assert.deepEqual(parseFinalizeGateEnv({}), factoryParseFinalizeGateEnv({}));
  assert.equal(parseOptionalTriStateFromEnv('true'), factoryParseOptionalTriStateFromEnv('true'));
  assert.deepEqual(actorTrustWhenSenderMissing(), factoryActorTrustWhenSenderMissing());
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
