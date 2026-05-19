/**
 * Classify changed files as provider-impacting or non-impacting.
 *
 * Non-impacting paths are:
 *   - CHANGELOG.md
 *   - Any path under openspec/
 *   - Any path under .agents/
 *   - Any path under .github/ EXCEPT .github/workflows/provider.yml
 *
 * @param {string[]} changedFiles - Flat list of changed file paths for this run.
 * @returns {{ providerChanges: 'true' | 'false' }} - 'true' if any file is provider-impacting, or if the file list is empty (safe default).
 */
function classifyChanges(changedFiles) {
  if (!Array.isArray(changedFiles)) {
    throw new TypeError('changedFiles must be an array');
  }

  if (changedFiles.length === 0) {
    return { providerChanges: 'true' };
  }

  const allNonImpacting = changedFiles.every((f) => {
    if (f === 'CHANGELOG.md') return true;
    if (f.startsWith('openspec/')) return true;
    if (f.startsWith('.agents/')) return true;
    if (f.startsWith('.github/')) {
      return f !== '.github/workflows/provider.yml';
    }
    return false;
  });

  return { providerChanges: allNonImpacting ? 'false' : 'true' };
}

if (typeof module !== 'undefined') {
  module.exports = { classifyChanges };
}
