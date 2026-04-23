//include: ../../lib/code-factory-issue.js

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
