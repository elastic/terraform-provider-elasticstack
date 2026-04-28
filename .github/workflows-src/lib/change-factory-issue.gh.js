const intake = createFactoryIssueIntake({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: 'github-keywords',
});

const qualifyTriggerEvent = intake.qualifyTriggerEvent;
const checkActorTrust = intake.checkActorTrust;
const checkDuplicatePR = intake.checkDuplicatePR;
const computeGateReason = intake.computeGateReason;
const changeFactoryIssueBranchName = intake.issueBranchName;

function actorTrustWhenSenderMissing() {
  return factoryActorTrustWhenSenderMissing();
}

function parseFinalizeGateEnv(env) {
  return factoryParseFinalizeGateEnv(env);
}
