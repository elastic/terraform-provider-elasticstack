const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../lib/intake/change-factory-constants.js');
const { createFactoryIssueModule } = require('../lib/factory-issue-shared.js');
const factoryIssueModule = createFactoryIssueModule({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: DUPLICATE_LINKAGE_MODE,
});
const qualifyTriggerEvent = factoryIssueModule.qualifyTriggerEvent;
const checkActorTrust = factoryIssueModule.checkActorTrust;
const checkDuplicatePR = factoryIssueModule.checkDuplicatePR;
const computeGateReason = factoryIssueModule.computeGateReason;
const issueBranchName = factoryIssueModule.issueBranchName;
const actorTrustWhenSenderMissing = factoryIssueModule.actorTrustWhenSenderMissing;
const parseFinalizeGateEnv = factoryIssueModule.parseFinalizeGateEnv;

module.exports = async function ({ github, context, core }) {

  const { owner, repo } = context.repo;
  const issueNumber = context.payload.issue?.number;
  const expectedBranch = issueBranchName(issueNumber);

  const pulls = await github.paginate(github.rest.pulls.list, {
    owner,
    repo,
    state: 'open',
    head: `${owner}:${expectedBranch}`,
    per_page: 100,
  });

  const pullRequests = pulls.map(pr => ({
    number: pr.number,
    state: pr.state,
    head_branch: pr.head.ref,
    labels: pr.labels.map(l => l.name),
    body: pr.body ?? '',
    html_url: pr.html_url,
  }));

  const result = checkDuplicatePR({ issueNumber, pullRequests });

  core.setOutput('duplicate_pr_found', result.duplicate_pr_found ? 'true' : 'false');
  core.setOutput('duplicate_pr_url', result.duplicate_pr_url ?? '');
  core.setOutput('gate_reason', result.gate_reason);

  if (result.duplicate_pr_found) {
    core.info(`Duplicate PR found: ${result.gate_reason}`);
  } else {
    core.info(`No duplicate PR: ${result.gate_reason}`);
  }
};
