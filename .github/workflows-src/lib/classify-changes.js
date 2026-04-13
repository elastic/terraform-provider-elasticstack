/**
 * Classify changed files as provider-impacting or openspec-only.
 *
 * @param {string[]} changedFiles - Flat list of changed file paths for this run.
 * @returns {{ providerChanges: 'true' | 'false' }}
 */
function classifyChanges(changedFiles) {
  const allOpenspec =
    changedFiles.length > 0 && changedFiles.every((f) => f.startsWith('openspec/'));
  return { providerChanges: allOpenspec ? 'false' : 'true' };
}

if (typeof module !== 'undefined') {
  module.exports = { classifyChanges };
}
