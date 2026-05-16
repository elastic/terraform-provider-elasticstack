//include: ../../lib/set-phase-label.js

// Change-factory is issue-event only (no dispatch), but use the same
// INPUT_ISSUE_NUMBER-first pattern for consistency across all factories.
const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10) || context.payload.issue?.number || undefined;
const result = await setPhaseLabel({
  github,
  context,
  core,
  issueNumber,
  phaseLabelName: 'phase-specification',
});

core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
core.setOutput('phase_label_name', result.phase_label_name);

if (result.phase_label_set) {
  core.info(`Set phase label ${result.phase_label_name} on issue #${issueNumber}. ${result.reason}`);
} else {
  core.warning(`Phase label not set: ${result.reason}`);
}
