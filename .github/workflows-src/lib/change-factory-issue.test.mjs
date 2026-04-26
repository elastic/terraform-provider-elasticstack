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
  CHANGE_FACTORY_ISSUE_BRANCH_PREFIX,
  changeFactoryIssueBranchName,
  qualifyTriggerEvent,
  actorTrustWhenSenderMissing,
  checkActorTrust,
  checkDuplicatePR,
  computeGateReason,
  parseOptionalTriStateFromEnv,
  parseFinalizeGateEnv,
} = require('./change-factory-issue.js');

const scriptsDir = path.resolve(__dirname, '../change-factory-issue/scripts');
const workflowTemplatePath = path.resolve(__dirname, '../change-factory-issue/workflow.md.tmpl');
const inlineScripts = [
  'qualify_trigger.inline.js',
  'check_actor_trust.inline.js',
  'check_duplicate_pr.inline.js',
  'finalize_gate.inline.js',
];

function makePullRequest(overrides = {}) {
  return {
    number: 101,
    state: 'open',
    head_branch: changeFactoryIssueBranchName(42),
    labels: ['change-factory'],
    body: 'Proposes the OpenSpec change.\n\nCloses #42',
    html_url: 'https://github.com/elastic/terraform-provider-elasticstack/pull/101',
    ...overrides,
  };
}

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

test('checkDuplicatePR matches lowercase closes keyword', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'See description.\n\ncloses #42' })],
  });

  assert.equal(result.duplicate_pr_found, true);
});

test('checkDuplicatePR matches alternate GitHub closing keywords case-insensitively', () => {
  for (const body of [
    'FIXES #42',
    'Fixed #42',
    'Resolve #42',
    'RESOLVED #42',
  ]) {
    const result = checkDuplicatePR({
      issueNumber: 42,
      pullRequests: [makePullRequest({ body })],
    });
    assert.equal(result.duplicate_pr_found, true, `expected match for body: ${body}`);
  }
});

test('checkDuplicatePR does not match a PR whose body has Closes issue linkage followed by more digits', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Proposes the change.\n\nCloses #420' })],
  });

  assert.equal(result.duplicate_pr_found, false);
});

test('checkDuplicatePR matches when body has canonical Closes issue linkage at end of line', () => {
  const result = checkDuplicatePR({
    issueNumber: 42,
    pullRequests: [makePullRequest({ body: 'Proposes the change.\n\nCloses #42\n' })],
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

test('computeGateReason uses generic untrusted text when actorTrusted is false with a falsy reason', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: false,
    actorTrustedReason: '',
    duplicatePrFound: false,
    duplicatePrUrl: null,
    noDuplicateReason: null,
  });

  assert.equal(result.gate_reason, 'Trigger actor is not trusted.');
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

test('computeGateReason falls back to unknown URL when duplicate is found without URL or override reason', () => {
  const result = computeGateReason({
    eventEligible: true,
    eventEligibleReason: 'Event is eligible.',
    actorTrusted: true,
    actorTrustedReason: 'Actor is trusted.',
    duplicatePrFound: true,
    duplicatePrUrl: null,
    noDuplicateReason: null,
  });

  assert.match(result.gate_reason, /\(unknown URL\)/);
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
    'All deterministic gates passed: event eligible, actor trusted, and no linked change-factory PR found.',
  );
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

test('changeFactoryIssueBranchName stays aligned with workflow template prefix', () => {
  assert.equal(CHANGE_FACTORY_ISSUE_BRANCH_PREFIX, 'change-factory/issue-');
  assert.equal(changeFactoryIssueBranchName(42), 'change-factory/issue-42');

  const workflowTmpl = readFileSync(workflowTemplatePath, 'utf8');
  const branchExpr = `${CHANGE_FACTORY_ISSUE_BRANCH_PREFIX}\${{ github.event.issue.number }}`;
  assert.ok(
    workflowTmpl.includes(branchExpr),
    'workflow.md.tmpl must express branches with CHANGE_FACTORY_ISSUE_BRANCH_PREFIX + ${{ github.event.issue.number }}',
  );
});

test('change-factory-issue workflow.md.tmpl wiring matches intake contract', () => {
  const workflowTmpl = readFileSync(workflowTemplatePath, 'utf8');

  assert.match(workflowTmpl, /\non:\n  issues:\n    types: \[opened, labeled\]/);

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
    /- name: Check duplicate PR\n      id: check_duplicate_pr\n      if: >-\n        steps\.qualify_trigger\.outputs\.event_eligible == 'true' &&\n        steps\.check_actor_trust\.outputs\.actor_trusted == 'true'/,
  );

  const scriptIncludes = [
    'x-script-include: scripts/qualify_trigger.inline.js',
    'x-script-include: scripts/check_actor_trust.inline.js',
    'x-script-include: scripts/check_duplicate_pr.inline.js',
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
    /- name: Finalize gate reason\n      id: finalize_gate\n      if: always\(\)/,
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
    /create-pull-request:\s*\n\s*labels: \[change-factory\]\s*\n\s*max: 1/,
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
  const requiredPhrases = [
    'sole authoritative source',
    'openspec/changes/<change-id>/',
    'proposal.md',
    'design.md',
    'tasks.md',
    'specs/<capability>/spec.md',
    'OPENSPEC_TELEMETRY=0',
    'openspec validate <change-id> --type change',
    'openspec status --change',
    'Terraform acceptance tests',
    'TF_ACC',
    'Elastic Stack',
    'Fleet',
    'API key',
    'noop',
    'exploration loop',
    'Do not open GitHub comment',
  ];
  for (const phrase of requiredPhrases) {
    assert.ok(
      workflowTmpl.includes(phrase),
      `workflow.md.tmpl agent prompt must include contract phrase: ${phrase}`,
    );
  }
});

test('check_duplicate_pr.inline.js resolves expected branch via changeFactoryIssueBranchName', () => {
  const source = readFileSync(path.join(scriptsDir, 'check_duplicate_pr.inline.js'), 'utf8');
  assert.match(source, /const expectedBranch = changeFactoryIssueBranchName\(issueNumber\);/);
});

test('change-factory-issue inline scripts include the deterministic helper library', () => {
  for (const name of inlineScripts) {
    const source = readFileSync(path.join(scriptsDir, name), 'utf8');
    assert.match(source, /^\/\/include: \.\.\/\.\.\/lib\/change-factory-issue\.js\n/);
  }
});

test('parseOptionalTriStateFromEnv treats missing and empty as null', () => {
  assert.equal(parseOptionalTriStateFromEnv(undefined), null);
  assert.equal(parseOptionalTriStateFromEnv(''), null);
});

test('parseOptionalTriStateFromEnv parses true only for exact true string', () => {
  assert.equal(parseOptionalTriStateFromEnv('true'), true);
  assert.equal(parseOptionalTriStateFromEnv('false'), false);
  assert.equal(parseOptionalTriStateFromEnv('TRUE'), false);
});

test('parseFinalizeGateEnv matches finalize_gate env semantics', () => {
  assert.deepEqual(
    parseFinalizeGateEnv({}),
    {
      eventEligible: false,
      eventEligibleReason: '',
      actorTrusted: null,
      actorTrustedReason: null,
      duplicatePrFound: null,
      duplicatePrUrl: null,
      noDuplicateReason: null,
    },
  );

  assert.deepEqual(
    parseFinalizeGateEnv({
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
      noDuplicateReason: 'x',
    },
  );
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
