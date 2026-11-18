const { removeTriggerLabel } = require('../lib/remove-trigger-label.js');

const OUTCOME_LABELS = ['ready-for-change-factory', 'research-needs-human'];

module.exports = async function ({ github, context, core }) {
  const issueNumber =
    parseInt(process.env.INPUT_ISSUE_NUMBER, 10) || context.payload.issue?.number || undefined;

  const results = [];
  for (const labelName of OUTCOME_LABELS) {
    results.push(await removeTriggerLabel({ github, context, issueNumber, labelName }));
  }

  const allRemoved = results.every((r) => r.trigger_label_removed);
  const reason = results.map((r) => r.trigger_label_removed_reason).join('; ');
  core.setOutput('stale_outcome_labels_removed', allRemoved ? 'true' : 'false');
  core.setOutput('stale_outcome_labels_removed_reason', reason);
  (allRemoved ? core.info : core.warning)(`Stale outcome label removal: ${reason}`);
};
