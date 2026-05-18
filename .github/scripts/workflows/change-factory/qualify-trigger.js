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

  const eventName = context.eventName;
  const eventAction = context.payload.action;
  const labelName = context.payload.label?.name ?? '';
  const issueLabels = (context.payload.issue?.labels ?? []).map(l => l.name);

  const result = qualifyTriggerEvent({ eventName, eventAction, labelName, issueLabels });

  core.setOutput('event_eligible', result.event_eligible ? 'true' : 'false');
  core.setOutput('event_eligible_reason', result.event_eligible_reason);

  if (result.event_eligible) {
    core.info(`Event eligible: ${result.event_eligible_reason}`);
  } else {
    core.info(`Event not eligible: ${result.event_eligible_reason}`);
  }
};
