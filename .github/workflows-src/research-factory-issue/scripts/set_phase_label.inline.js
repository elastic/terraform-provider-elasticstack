//include: ../../lib/set-phase-label.js

const issueNumber = context.payload.issue?.number || parseInt(process.env.INPUT_ISSUE_NUMBER, 10) || undefined;
const result = await setPhaseLabel({
  github,
  context,
  issueNumber,
  phaseLabelName: 'phase-research',
});

core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
core.setOutput('phase_label_name', result.phase_label_name);

if (result.phase_label_set) {
  core.info(`Set phase label ${result.phase_label_name} on issue #${issueNumber}. ${result.reason}`);
} else {
  core.info(`Phase label not set: ${result.reason}`);
}
