const intake = createFactoryIssueIntake({
  branchPrefix: 'code-factory/issue-',
  factoryLabel: 'code-factory',
  issueOpenedNotEligibleReason:
    'Issue opened event does not qualify because the issue was created without the code-factory label.',
  duplicateLinkageMode: 'closes-literal',
  duplicatePrUrlCoalesceNull: false,
});

const qualifyTriggerEvent = intake.qualifyTriggerEvent;
const checkActorTrust = intake.checkActorTrust;
const checkDuplicatePR = intake.checkDuplicatePR;
const computeGateReason = intake.computeGateReason;
const issueBranchName = intake.issueBranchName;
