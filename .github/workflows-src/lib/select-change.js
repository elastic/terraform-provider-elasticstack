const CHANGE_PATTERN = /^openspec\/changes\/([^/]+)\/.+$/;
const ARCHIVE_PATTERN = /^openspec\/changes\/archive\//;

function ineligible(selection_reason) {
  return {
    selection_status: 'ineligible',
    selection_reason,
    selected_change: '',
  };
}

function selectChangeFromFiles(files) {
  const relevantFiles = files.filter(
    file => CHANGE_PATTERN.test(file.filename) && !ARCHIVE_PATTERN.test(file.filename)
  );

  if (relevantFiles.length === 0) {
    return ineligible('No files under openspec/changes/ (non-archive) found in this PR');
  }

  const addedFiles = relevantFiles.filter(file => file.status === 'added');
  if (addedFiles.length > 0) {
    return ineligible(
      `Added file(s) under openspec/changes/: ${addedFiles.map(file => file.filename).join(', ')}`
    );
  }

  const nonModifiedFiles = relevantFiles.filter(file => file.status !== 'modified');
  if (nonModifiedFiles.length > 0) {
    return ineligible(
      `Non-modified file(s) under openspec/changes/: ${nonModifiedFiles
        .map(file => `${file.filename} (${file.status})`)
        .join(', ')}`
    );
  }

  const modifiedIds = new Set(
    relevantFiles
      .filter(file => file.status === 'modified')
      .map(file => file.filename.match(CHANGE_PATTERN)[1])
  );

  if (modifiedIds.size === 0) {
    return ineligible('No active change id with a modified file found');
  }

  if (modifiedIds.size > 1) {
    return ineligible(
      `Multiple active change ids with modified files: ${Array.from(modifiedIds).join(', ')}`
    );
  }

  const selectedChange = Array.from(modifiedIds)[0];
  return {
    selection_status: 'eligible',
    selection_reason: `Selected change: ${selectedChange}`,
    selected_change: selectedChange,
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
