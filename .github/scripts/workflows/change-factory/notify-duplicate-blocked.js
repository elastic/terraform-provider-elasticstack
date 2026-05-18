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

  const duplicatePrUrl = process.env.DUPLICATE_PR_URL;
  const issueNumber = process.env.ISSUE_NUMBER;
  const { owner, repo } = context.repo;

  if (duplicatePrUrl && issueNumber) {
    const commentBody = `⚠️ **change-factory skipped** — PR #${extractPrNumber(duplicatePrUrl)} is already open for this issue.\nClose the existing PR, then retry.`;
    
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: parseInt(issueNumber, 10),
      body: commentBody,
    });
    
    core.info(`Posted duplicate-blocked comment on issue #${issueNumber} referencing ${duplicatePrUrl}`);
  } else {
    core.info('DUPLICATE_PR_URL is empty; skipping duplicate-blocked notification.');
  }

  /**
   * Extract the PR number from a GitHub PR URL.
   * @param {string} url
   * @returns {string}
   */
  function extractPrNumber(url) {
    const match = url.match(/\/(\d+)(?:\/|$)/);
    return match ? match[1] : url;
  }
};
