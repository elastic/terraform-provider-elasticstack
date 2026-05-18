const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../lib/intake/reproducer-factory-constants.js');
const { createFactoryIssueModule } = require('../lib/factory-issue-shared.js');
const { qualifyTriggerEvent } = createFactoryIssueModule({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: DUPLICATE_LINKAGE_MODE,
});

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
