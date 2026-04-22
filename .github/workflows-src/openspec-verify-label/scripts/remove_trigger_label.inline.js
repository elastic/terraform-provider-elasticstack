//include: ../../lib/remove-trigger-label.js

const prNumber = context.payload.pull_request?.number;
const result = await removeTriggerLabel({ github, context, prNumber });

core.setOutput('trigger_label_removed', result.trigger_label_removed ? 'true' : 'false');
core.setOutput('trigger_label_removed_reason', result.trigger_label_removed_reason);

if (result.trigger_label_removed) {
  core.info(`Removed trigger label verify-openspec from PR #${prNumber}`);
} else {
  core.info(`Trigger label removal skipped: ${result.trigger_label_removed_reason}`);
}
