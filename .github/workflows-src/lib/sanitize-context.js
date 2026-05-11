/**
 * HTML comment sanitisation, control/invisible-char removal, and research-comment lookup helpers.
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
 * Removes non-printable ASCII control characters while preserving tab, newline,
 * and carriage return. Also strips Unicode line/paragraph separators.
 *
 * Stripped ASCII:  \x00-\x08, \x0B, \x0C, \x0E-\x1F, \x7F
 * Preserved:       \x09 (tab), \x0A (LF), \x0D (CR)
 * Unicode:         \u2028, \u2029
 *
 * @param {string} text
 * @returns {string}
 */
function stripControlChars(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\u2028\u2029]/g, '');
}

/**
 * Removes invisible Unicode characters that have no legitimate use in issue content:
 * zero-width spaces/joiners, bidirectional marks, word/function/invisible operators,
 * and the BOM (byte order mark).
 *
 * Ranges stripped:
 *   \u200B-\u200F  — zero-width space, non-joiner, joiner, LTR mark, RTL mark
 *   \u2060-\u2064  — word joiner, function application, invisible times/separator/plus
 *   \uFEFF           — BOM / zero-width no-break space
 *
 * @param {string} text
 * @returns {string}
 */
function stripInvisibleUnicode(text) {
  if (typeof text !== 'string') return '';
  return text.replace(/[\u200B-\u200F\u2060-\u2064\uFEFF]/g, '');
}

/**
 * Composed sanitisation pipeline. Runs all three filters in sequence:
 *
 *   input → stripHtmlComments → stripControlChars → stripInvisibleUnicode → output
 *
 * Idempotent: applying twice produces the same result as applying once.
 *
 * @param {string} text
 * @returns {string}
 */
function sanitizeUserContent(text) {
  if (typeof text !== 'string') return '';
  return stripInvisibleUnicode(stripControlChars(stripHtmlComments(text)));
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
    stripControlChars,
    stripInvisibleUnicode,
    sanitizeUserContent,
    findResearchComment,
  };
}
