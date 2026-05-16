/**
 * Adds a phase label to an issue and removes all other phase-* labels.
 * @param {{ github: object, context: object, issueNumber: number|undefined, phaseLabelName: string|undefined }} opts
 * @returns {Promise<{ phase_label_set: boolean, phase_label_name: string, stale_labels_removed: string[], reason: string }>}
 */
async function setPhaseLabel({ github, context, issueNumber, phaseLabelName }) {
  if (issueNumber === undefined || issueNumber === null) {
    return {
      phase_label_set: false,
      phase_label_name: phaseLabelName || '',
      stale_labels_removed: [],
      reason: 'No issue number provided',
    };
  }

  const label =
    typeof phaseLabelName === 'string' && phaseLabelName.trim() !== '' ? phaseLabelName.trim() : null;
  if (!label) {
    return {
      phase_label_set: false,
      phase_label_name: '',
      stale_labels_removed: [],
      reason: 'No phase label name provided',
    };
  }

  try {
    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      labels: [label],
    });
  } catch (err) {
    return {
      phase_label_set: false,
      phase_label_name: label,
      stale_labels_removed: [],
      reason: `Failed to add label: ${err.message}`,
    };
  }

  let staleLabels = [];
  try {
    const { data: currentLabels } = await github.rest.issues.listLabelsOnIssue({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      per_page: 100,
    });

    staleLabels = currentLabels
      .map((l) => l.name)
      .filter((name) => name.startsWith('phase-') && name !== label);

    if (currentLabels.length === 100 && typeof core !== 'undefined' && typeof core.warning === 'function') {
      core.warning(`Issue #${issueNumber} has 100 labels; stale phase label removal may be incomplete.`);
    }
  } catch (err) {
    return {
      phase_label_set: true,
      phase_label_name: label,
      stale_labels_removed: [],
      reason: `Added label ${label} but failed to list current labels: ${err.message}`,
    };
  }

  for (const staleLabel of staleLabels) {
    try {
      await github.rest.issues.removeLabel({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: issueNumber,
        name: staleLabel,
      });
    } catch (err) {
      if (err.status !== 404) {
        return {
          phase_label_set: true,
          phase_label_name: label,
          stale_labels_removed: staleLabels.filter((l) => l !== staleLabel),
          reason: `Added label ${label} but failed to remove stale label ${staleLabel}: ${err.message}`,
        };
      }
      // 404 means label was already absent — treat as success
    }
  }

  const removalMsg =
    staleLabels.length > 0
      ? `Removed stale phase labels: ${staleLabels.join(', ')}`
      : 'No stale phase labels to remove';

  return {
    phase_label_set: true,
    phase_label_name: label,
    stale_labels_removed: staleLabels,
    reason: `Set phase label ${label}. ${removalMsg}`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { setPhaseLabel };
}
