/**
 * HTML comment sanitisation and research-comment lookup helpers.
 */

/**
 * Removes all HTML comment sequences (<code>&lt;!--</code> through the next <code>--&gt;</code>).
 * If an opening sequence has no closing counterpart, everything from the opener to the end of
 * the string is removed.
 *
 * @param {string} text
 * @returns {string}
 */
function stripHtmlComments(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/<!--[\s\S]*?(?:-->|$)/g, '');
}

/**
 * Finds the most recently created matching research comment written by
 * <code>github-actions[bot]</code> whose body starts with <code>marker</code>.
 *
 * @param {Array<{author: string, body: string}>} comments Ordered oldest-first.
 * @param {string} marker
 * @returns {{author: string, body: string} | null}
 */
function findResearchComment(comments, marker) {
  if (!Array.isArray(comments)) {
    return null;
  }
  const matches = comments.filter(
    (c) =>
      c != null &&
      typeof c.body === 'string' &&
      (c.author ?? c.user?.login) === 'github-actions[bot]' &&
      c.body.trimStart().startsWith(marker),
  );
  return matches.length > 0 ? matches[matches.length - 1] : null;
}

if (typeof module !== 'undefined') {
  module.exports = {
    stripHtmlComments,
    findResearchComment,
  };
}
