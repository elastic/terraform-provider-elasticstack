const CHANGE_PATTERN = /^openspec\/changes\/([^/]+)\/.+$/;
const ARCHIVE_PATTERN = /^openspec\/changes\/archive\//;

const ALLOWED_STATUSES = new Set(['added', 'modified']);

function ineligible(selection_reason) {
  return {
    selection_status: 'ineligible',
    selection_reason,
    selected_change: '',
    review_disposition: '',
    disposition_reason: '',
  };
}

function selectChangeFromFiles(files) {
  const relevantFiles = files.filter(
    file => CHANGE_PATTERN.test(file.filename) && !ARCHIVE_PATTERN.test(file.filename)
  );

  if (relevantFiles.length === 0) {
    return ineligible('No files under openspec/changes/ (non-archive) found in this PR');
  }

  const unsupported = relevantFiles.filter(file => !ALLOWED_STATUSES.has(file.status));
  if (unsupported.length > 0) {
    return ineligible(
      `Unsupported file status under openspec/changes/: ${unsupported
        .map(file => `${file.filename} (${file.status})`)
        .join(', ')}`
    );
  }

  const changeIds = new Set(relevantFiles.map(file => file.filename.match(CHANGE_PATTERN)[1]));

  if (changeIds.size > 1) {
    return ineligible(`Multiple active change ids: ${Array.from(changeIds).sort().join(', ')}`);
  }

  const selectedChange = Array.from(changeIds)[0];
  const hasAdded = relevantFiles.some(file => file.status === 'added');
  const reviewDisposition = hasAdded ? 'comment-only' : 'approval-eligible';

  const dispositionReason = hasAdded
    ? 'The selected change includes one or more added files (net-new spec change material). APPROVE is not permitted; submit COMMENT only, even if verification passes with no blocking issues.'
    : 'Every file under the selected change is a modification. APPROVE is permitted when verification finds zero CRITICAL issues and zero unassociated files.';

  return {
    selection_status: 'eligible',
    selection_reason: `Selected change: ${selectedChange}`,
    selected_change: selectedChange,
    review_disposition: reviewDisposition,
    disposition_reason: dispositionReason,
  };
}

function selectChangeForPullRequest({ prNumber, files = [] }) {
  if (!prNumber) {
    return ineligible('No pull request number in event payload');
  }

  return selectChangeFromFiles(files);
}

if (typeof module !== 'undefined') {
  module.exports = {
    ARCHIVE_PATTERN,
    CHANGE_PATTERN,
    selectChangeForPullRequest,
    selectChangeFromFiles,
  };
}
