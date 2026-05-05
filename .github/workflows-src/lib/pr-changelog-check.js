/**
 * Helper functions for the PR Changelog Check workflow.
 */

/**
 * Find an existing bot comment that contains the given marker.
 * @param {Array} comments - flat array of comment objects (from github.paginate)
 * @param {string} marker - HTML marker string to search for
 * @returns {object|null}
 */
function findExistingComment(comments, marker) {
  return (
    comments.find(
      (c) => c.user?.login === 'github-actions[bot]' && c.body.includes(marker)
    ) ?? null
  );
}

/**
 * Build the body for a "check passed" comment.
 * @param {string} marker
 * @returns {string}
 */
function buildPassCommentBody(marker) {
  return `${marker}\n:white_check_mark: **PR Changelog Check passed** — the \`## Changelog\` section looks good.`;
}

/**
 * Build the body for a "no-changelog label" pass comment.
 * @param {string} marker
 * @returns {string}
 */
function buildNoChangelogPassCommentBody(marker) {
  return `${marker}\n:white_check_mark: **PR Changelog Check passed** — \`no-changelog\` label is set.`;
}

/**
 * Build the body for a failure comment listing validation errors.
 * @param {string} marker
 * @param {string[]} errors
 * @returns {string}
 */
function buildFailureCommentBody(marker, errors) {
  const errorList = errors.map((e) => `- ${e}`).join('\n');
  return [
    marker,
    ':x: **PR Changelog Check failed**',
    '',
    'The following issues were found with the `## Changelog` section:',
    '',
    errorList,
    '',
    '<details>',
    '<summary>Expected format</summary>',
    '',
    '```',
    '## Changelog',
    'Customer impact: <none|fix|enhancement|breaking>',
    'Summary: <one-line description>  (required when Customer impact is not "none")',
    '',
    '### Breaking changes',
    '<free-form markdown>  (required when Customer impact is "breaking")',
    '<!-- /breaking-changes -->  (optional — ends the block early)',
    '```',
    '',
    'Or add the `no-changelog` label to bypass this check.',
    '</details>',
  ].join('\n');
}

if (typeof module !== 'undefined') {
  module.exports = {
    findExistingComment,
    buildPassCommentBody,
    buildNoChangelogPassCommentBody,
    buildFailureCommentBody,
  };
}
