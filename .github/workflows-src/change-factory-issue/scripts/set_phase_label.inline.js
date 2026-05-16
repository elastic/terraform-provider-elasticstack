//include: ../../lib/set-phase-label.js

const issueNumber = context.payload.issue?.number;
const result = await setPhaseLabel({
  github,
  context,
  issueNumber,
  phaseLabelName: 'phase-specification',
});

core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
core.setOutput('phase_label_name', result.phase_label_name);

if (result.phase_label_set) {
  core.info(`Set phase label ${result.phase_label_name} on issue #${issueNumber}. ${result.reason}`);
} else {
  core.info(`Phase label not set: ${result.reason}`);
}
