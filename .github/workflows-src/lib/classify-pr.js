/**
 * Determines archive/push eligibility for a pull request.
 *
 * Same-repository pull requests (head.repo.id === repository.id) have archive/push allowed.
 * Fork pull requests have archive/push disallowed (fork-controlled content must not be pushed
 * to a branch in the trusted repository context).
 *
 * @param {{ headRepoId: number|undefined, baseRepoId: number|undefined }} opts
 * @returns {{
 *   archive_push_allowed: boolean,
 *   archive_push_allowed_reason: string,
 * }}
 */
function classifyPullRequest({ headRepoId, baseRepoId }) {
  if (headRepoId != null && baseRepoId != null && headRepoId === baseRepoId) {
    return {
      archive_push_allowed: true,
      archive_push_allowed_reason:
        'Same-repository pull request: archive and push to PR branch are allowed.',
    };
  }

  return {
    archive_push_allowed: false,
    archive_push_allowed_reason:
      'Fork pull request: archive and push to PR branch are disallowed to prevent pushing fork-controlled content to the trusted repository.',
  };
}

if (typeof module !== 'undefined') {
  module.exports = { classifyPullRequest };
}
