/**
 * Adds a phase label to an issue and removes all other phase-* labels.
 * @param {{ github: object, context: object, issueNumber: number|undefined, phaseLabelName: string|undefined, core?: object }} opts
 * @returns {Promise<{ phase_label_set: boolean, phase_label_name: string, stale_labels_removed: string[], reason: string }>}
 */
async function setPhaseLabel({ github, context, issueNumber, phaseLabelName, core: coreParam }) {
  // Use explicitly passed core, fall back to ambient global, or stub.
  const _core =
    coreParam ||
    (typeof core !== 'undefined' ? core : undefined);

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
    // Use paginate to fetch all labels reliably (no truncation concern).
    const listLabels = github.paginate
      ? github.paginate.bind(github)
      : async (_method, params) => {
          const { data } = await github.rest.issues.listLabelsOnIssue(params);
          return data;
        };
    const currentLabels = await listLabels(github.rest.issues.listLabelsOnIssue, {
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      per_page: 100,
    });

    staleLabels = currentLabels
      .map((l) => l.name)
      .filter((name) => name.startsWith('phase-') && name !== label);
  } catch (err) {
    return {
      phase_label_set: true,
      phase_label_name: label,
      stale_labels_removed: [],
      reason: `Added label ${label} but failed to list current labels: ${err.message}`,
    };
  }

  const removed = [];
  const failed = [];
  for (const staleLabel of staleLabels) {
    try {
      await github.rest.issues.removeLabel({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: issueNumber,
        name: staleLabel,
      });
      removed.push(staleLabel);
    } catch (err) {
      if (err.status === 404) {
        // Label was already absent — treat as successfully removed.
        removed.push(staleLabel);
      } else {
        failed.push({ label: staleLabel, message: err.message });
        if (_core && typeof _core.warning === 'function') {
          _core.warning(`Failed to remove stale label ${staleLabel} from issue #${issueNumber}: ${err.message}`);
        }
      }
    }
  }

  if (failed.length > 0) {
    const failedSummary = failed.map((f) => `${f.label}: ${f.message}`).join('; ');
    return {
      phase_label_set: true,
      phase_label_name: label,
      stale_labels_removed: removed,
      reason: `Added label ${label} but failed to remove some stale labels: ${failedSummary}`,
    };
  }

  const removalMsg =
    staleLabels.length > 0
      ? `Removed stale phase labels: ${staleLabels.join(', ')}`
      : 'No stale phase labels to remove';

  return {
    phase_label_set: true,
    phase_label_name: label,
    stale_labels_removed: removed,
    reason: `Set phase label ${label}. ${removalMsg}`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { setPhaseLabel };
}
