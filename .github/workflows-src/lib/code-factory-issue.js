'use strict';

const { createFactoryIssueIntake, factoryActorTrustWhenSenderMissing, factoryParseOptionalTriStateFromEnv, factoryParseFinalizeGateEnv } = require('./factory-issue-shared.js');
const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../code-factory-issue/intake-constants.js');

const intake = createFactoryIssueIntake({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: 'closes-literal',
});

function actorTrustWhenSenderMissing() {
  return factoryActorTrustWhenSenderMissing();
}

function parseOptionalTriStateFromEnv(raw) {
  return factoryParseOptionalTriStateFromEnv(raw);
}

function parseFinalizeGateEnv(env) {
  return factoryParseFinalizeGateEnv(env);
}

module.exports = {
  qualifyTriggerEvent: intake.qualifyTriggerEvent,
  checkActorTrust: intake.checkActorTrust,
  checkDuplicatePR: intake.checkDuplicatePR,
  computeGateReason: intake.computeGateReason,
  issueBranchName: intake.issueBranchName,
  actorTrustWhenSenderMissing,
  parseOptionalTriStateFromEnv,
  parseFinalizeGateEnv,
};
