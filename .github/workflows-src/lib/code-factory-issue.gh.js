const intake = createFactoryIssueIntake({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: 'closes-literal',
});

const qualifyTriggerEvent = intake.qualifyTriggerEvent;
const checkActorTrust = intake.checkActorTrust;
const checkDuplicatePR = intake.checkDuplicatePR;
const computeGateReason = intake.computeGateReason;
const issueBranchName = intake.issueBranchName;

function actorTrustWhenSenderMissing() {
  return factoryActorTrustWhenSenderMissing();
}

function parseFinalizeGateEnv(env) {
  return factoryParseFinalizeGateEnv(env);
}
