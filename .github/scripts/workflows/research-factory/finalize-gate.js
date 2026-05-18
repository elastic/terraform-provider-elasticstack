const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../lib/intake/research-factory-constants.js');
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

  const params = parseFinalizeGateEnv(process.env);
  // research-factory does not create branches or PRs, so the duplicate-PR gate is intentionally
  // disabled by passing duplicatePrFound: false. This deviates from sibling factories by design.
  const result = computeGateReason({
    ...params,
    duplicatePrFound: false,
    duplicatePrUrl: null,
    duplicateCheckGateReason: null,
  });

  core.setOutput('gate_reason', result.gate_reason);
  core.info(`Gate reason: ${result.gate_reason}`);
};
