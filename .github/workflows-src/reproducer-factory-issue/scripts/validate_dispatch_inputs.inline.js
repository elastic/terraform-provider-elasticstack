//include: ../../lib/code-factory-dispatch.js

const currentRepository = `${context.repo.owner}/${context.repo.repo}`;
const dispatchIssueNumber = context.payload.inputs?.issue_number ?? '';

const result = validateDispatchInputs({
  dispatchIssueNumber,
  currentRepository,
});

core.setOutput('event_eligible', result.event_eligible ? 'true' : 'false');
core.setOutput('event_eligible_reason', result.event_eligible_reason);
if (result.issue_number != null) {
  core.setOutput('issue_number', String(result.issue_number));
}

if (result.event_eligible) {
  core.info(`Dispatch validated: ${result.event_eligible_reason}`);
} else {
  core.info(`Dispatch rejected: ${result.event_eligible_reason}`);
}
