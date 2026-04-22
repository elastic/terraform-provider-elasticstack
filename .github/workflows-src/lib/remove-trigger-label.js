const TRIGGER_LABEL = 'verify-openspec';

/**
 * Removes only the verify-openspec trigger label from the triggering pull request.
 * Does not remove any other labels.
 * @param {{ github: object, context: object, prNumber: number|undefined }} opts
 * @returns {Promise<{ trigger_label_removed: boolean, trigger_label_removed_reason: string }>}
 */
async function removeTriggerLabel({ github, context, prNumber }) {
  if (!prNumber) {
    return {
      trigger_label_removed: false,
      trigger_label_removed_reason: 'No pull request number in event payload',
    };
  }

  try {
    await github.rest.issues.removeLabel({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: prNumber,
      name: TRIGGER_LABEL,
    });
    return {
      trigger_label_removed: true,
      trigger_label_removed_reason: `Removed label: ${TRIGGER_LABEL}`,
    };
  } catch (err) {
    // GitHub returns 404 when the label does not exist on the issue; treat as success
    if (err.status === 404) {
      return {
        trigger_label_removed: true,
        trigger_label_removed_reason: `Label ${TRIGGER_LABEL} was not present (already removed or never applied)`,
      };
    }
    return {
      trigger_label_removed: false,
      trigger_label_removed_reason: `Failed to remove label: ${err.message}`,
    };
  }
}

if (typeof module !== 'undefined') {
  module.exports = { TRIGGER_LABEL, removeTriggerLabel };
}
