//include: ../../lib/set-phase-label.js

// Prefer INPUT_ISSUE_NUMBER (routed by the workflow template based on intake_mode)
// over context.payload.issue?.number to avoid ambiguity across event types.
const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10) || context.payload.issue?.number || undefined;
const result = await setPhaseLabel({
  github,
  context,
  core,
  issueNumber,
  phaseLabelName: 'phase-research',
});

core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
core.setOutput('phase_label_name', result.phase_label_name);

if (result.phase_label_set) {
  core.info(`Set phase label ${result.phase_label_name} on issue #${issueNumber}. ${result.reason}`);
} else {
  core.warning(`Phase label not set: ${result.reason}`);
}
