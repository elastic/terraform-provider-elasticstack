/**
 * Changelog engine bundle for GitHub Actions scripts and for `require()` from orchestration modules.
 */
const { execSync } = require('child_process');
const {
  parseSemverTags,
  selectPreviousTag,
  buildCompareRange,
} = require('./changelog-release-context.js');
const { rewriteChangelogSection, rewriteLinkTable } = require('./changelog-rewriter.js');
const { renderChangelogSection } = require('./changelog-renderer.js');
const { createChangelogEngine } = require('./changelog-engine-factory.js');

const _changelogEngine = createChangelogEngine({
  parseSemverTags,
  selectPreviousTag,
  buildCompareRange,
  rewriteChangelogSection,
  rewriteLinkTable,
  renderChangelogSection,
  execSyncDefault: execSync,
});

// `TARGET_VERSION_PATTERN` is intentionally NOT destructured here: it is already declared as a
// `const` by the included `changelog-engine-factory.js`, and a second `const` in the same scope
// would be a SyntaxError when this bundle is inlined into the workflow YAML.
const {
  validateModeAndTargetVersion,
  resolveChangelogCompareContext,
  gatherMergedPRRecordsForRange,
  formatAssemblyFailureMessage,
  runChangelogRenderAndWrite,
  runChangelogEngine,
} = _changelogEngine;

if (typeof module !== 'undefined') {
  module.exports = {
    validateModeAndTargetVersion,
    resolveChangelogCompareContext,
    gatherMergedPRRecordsForRange,
    formatAssemblyFailureMessage,
    runChangelogRenderAndWrite,
    runChangelogEngine,
  };
}
