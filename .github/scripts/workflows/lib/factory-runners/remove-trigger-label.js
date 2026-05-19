const { getFactoryName } = require('./_factory-context.js');
const { removeTriggerLabel } = require('../remove-trigger-label.js');

module.exports = async function ({ github, context, core }) {

  const labelName = getFactoryName();
  const issueNumber = context.payload.issue?.number;
  const result = await removeTriggerLabel({
    github,
    context,
    issueNumber,
    labelName,
  });

  core.setOutput('trigger_label_removed', result.trigger_label_removed ? 'true' : 'false');
  core.setOutput('trigger_label_removed_reason', result.trigger_label_removed_reason);

  if (result.trigger_label_removed) {
    core.info(`Removed trigger label ${labelName} from issue #${issueNumber}`);
  } else {
    core.info(`Trigger label removal skipped: ${result.trigger_label_removed_reason}`);
  }
};
