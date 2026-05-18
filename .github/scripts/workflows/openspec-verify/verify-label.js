const { verifyTriggerLabel } = require('../lib/verify-label.js');

module.exports = async function ({ github, context, core }) {
  const labelName = context.payload.label?.name ?? '';
  const result = verifyTriggerLabel(labelName);

  core.setOutput('label_verified', result.label_verified ? 'true' : 'false');
  core.setOutput('label_verified_reason', result.label_verified_reason);

  if (result.label_verified) {
    core.info(`Trigger label verified: ${labelName}`);
  } else {
    core.info(`Trigger label not matched (got: ${labelName}): ${result.label_verified_reason}`);
  }
};
