/**
 * Classifies a pull request as workspace (same-repository) or api-only (fork),
 * and determines archive/push eligibility.
 *
 * Same-repository pull requests (head.repo.id === repository.id) are classified as
 * workspace mode with archive/push allowed. Fork pull requests are classified as
 * api-only mode with archive/push disallowed.
 *
 * @param {{ headRepoId: number|undefined, baseRepoId: number|undefined }} opts
 * @returns {{
 *   verification_mode: 'workspace' | 'api-only',
 *   verification_mode_reason: string,
 *   archive_push_allowed: boolean,
 *   archive_push_reason: string,
 * }}
 */
function classifyPullRequest({ headRepoId, baseRepoId }) {
  if (headRepoId != null && baseRepoId != null && headRepoId === baseRepoId) {
    return {
      verification_mode: 'workspace',
      verification_mode_reason:
        'Pull request head repository matches base repository (same-repository PR). Full workspace toolchain is available.',
      archive_push_allowed: true,
      archive_push_reason:
        'Same-repository pull request: archive and push to PR branch are allowed.',
    };
  }

  return {
    verification_mode: 'api-only',
    verification_mode_reason:
      'Pull request head repository differs from base repository (fork PR). Verification uses PR metadata and diffs only; trusted workspace bootstrap is not available.',
    archive_push_allowed: false,
    archive_push_reason:
      'Fork pull request: archive and push to PR branch are disallowed to prevent execution of fork-controlled content in the trusted workflow context.',
  };
}

if (typeof module !== 'undefined') {
  module.exports = { classifyPullRequest };
}
