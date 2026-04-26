'use strict';

const { createFactoryIssueIntake } = require('./factory-issue-shared.js');

const intake = createFactoryIssueIntake({
  branchPrefix: 'code-factory/issue-',
  factoryLabel: 'code-factory',
  issueOpenedNotEligibleReason:
    'Issue opened event does not qualify because the issue was created without the code-factory label.',
  duplicateLinkageMode: 'closes-literal',
  duplicatePrUrlCoalesceNull: false,
});

module.exports = {
  qualifyTriggerEvent: intake.qualifyTriggerEvent,
  checkActorTrust: intake.checkActorTrust,
  checkDuplicatePR: intake.checkDuplicatePR,
  computeGateReason: intake.computeGateReason,
  issueBranchName: intake.issueBranchName,
};
