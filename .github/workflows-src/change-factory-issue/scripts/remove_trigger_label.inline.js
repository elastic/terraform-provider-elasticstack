//include: ../../lib/remove-trigger-label.js

const issueNumber = context.payload.issue?.number;
const result = await removeTriggerLabel({
  github,
  context,
  issueNumber,
  labelName: 'change-factory',
});

core.setOutput('trigger_label_removed', result.trigger_label_removed ? 'true' : 'false');
core.setOutput('trigger_label_removed_reason', result.trigger_label_removed_reason);

if (result.trigger_label_removed) {
  core.info(`Removed trigger label change-factory from issue #${issueNumber}`);
} else {
  core.info(`Trigger label removal skipped: ${result.trigger_label_removed_reason}`);
}
