const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../lib/intake/change-factory-constants.js');
const { createFactoryIssueModule } = require('../lib/factory-issue-shared.js');
const { computeGateReason, parseFinalizeGateEnv } = createFactoryIssueModule({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: DUPLICATE_LINKAGE_MODE,
});

module.exports = async function ({ github, context, core }) {

  const result = computeGateReason(parseFinalizeGateEnv(process.env));

  core.setOutput('gate_reason', result.gate_reason);
  core.info(`Gate reason: ${result.gate_reason}`);
};
