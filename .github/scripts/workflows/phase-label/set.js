const { setPhaseLabel } = require('../lib/phase-label.js');

module.exports = async function ({ github, context, core }) {
  const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10) || context.payload.issue?.number || undefined;
  const result = await setPhaseLabel({
    github,
    context,
    core,
    issueNumber,
    phaseLabelName: process.env.PHASE_LABEL_NAME,
  });

  core.setOutput('phase_label_set', result.phase_label_set ? 'true' : 'false');
  core.setOutput('phase_label_name', result.phase_label_name);

  const logMessage = result.phase_label_set
    ? `Set phase label ${result.phase_label_name} on issue #${issueNumber}. ${result.reason}`
    : `Phase label not set: ${result.reason}`;

  (result.phase_label_set ? core.info : core.warning)(logMessage);
};
