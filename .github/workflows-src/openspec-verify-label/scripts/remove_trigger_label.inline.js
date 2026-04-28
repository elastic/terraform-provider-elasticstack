//include: ../../lib/remove-trigger-label.js

const verifyTriggerLabel = 'verify-openspec';
const issueNumber = context.payload.pull_request?.number;
const result = await removeTriggerLabel({
  github,
  context,
  issueNumber,
  labelName: verifyTriggerLabel,
});

core.setOutput('trigger_label_removed', result.trigger_label_removed ? 'true' : 'false');
core.setOutput('trigger_label_removed_reason', result.trigger_label_removed_reason);

if (result.trigger_label_removed) {
  core.info(`Removed trigger label ${verifyTriggerLabel} from PR #${issueNumber}`);
} else {
  core.info(`Trigger label removal skipped: ${result.trigger_label_removed_reason}`);
}
