/**
 * Removes a single label from an issue or pull request (GitHub issue API).
 * @param {{ github: object, context: object, issueNumber: number|undefined, labelName: string|undefined }} opts
 * @returns {Promise<{ trigger_label_removed: boolean, trigger_label_removed_reason: string }>}
 */
async function removeTriggerLabel({ github, context, issueNumber, labelName }) {
  if (issueNumber === undefined || issueNumber === null) {
    return {
      trigger_label_removed: false,
      trigger_label_removed_reason: 'No issue number in event payload',
    };
  }

  const label =
    typeof labelName === 'string' && labelName.trim() !== '' ? labelName.trim() : null;
  if (!label) {
    return {
      trigger_label_removed: false,
      trigger_label_removed_reason: 'No label name provided',
    };
  }

  try {
    await github.rest.issues.removeLabel({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      name: label,
    });
    return {
      trigger_label_removed: true,
      trigger_label_removed_reason: `Removed label: ${label}`,
    };
  } catch (err) {
    // GitHub returns 404 when the label does not exist on the issue; treat as success
    if (err.status === 404) {
      return {
        trigger_label_removed: true,
        trigger_label_removed_reason: `Label ${label} was not present (already removed or never applied)`,
      };
    }
    return {
      trigger_label_removed: false,
      trigger_label_removed_reason: `Failed to remove label: ${err.message}`,
    };
  }
}

if (typeof module !== 'undefined') {
  module.exports = { removeTriggerLabel };
}
