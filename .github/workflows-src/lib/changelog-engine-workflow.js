/**
 * Expanded into GitHub Actions inline scripts. Do not require() repository lib paths from here.
 * Node tests load changelog-engine.js instead.
 */
//include: ./changelog-release-context.js
//include: ./changelog-rewriter.js
//include: ./changelog-renderer.js
//include: ./changelog-engine-factory.js

const { execSync } = require('child_process');

const _changelogEngine = createChangelogEngine({
  parseSemverTags,
  selectPreviousTag,
  buildCompareRange,
  rewriteChangelogSection,
  renderChangelogSection,
  execSyncDefault: execSync,
});

const {
  TARGET_VERSION_PATTERN,
  validateModeAndTargetVersion,
  resolveChangelogCompareContext,
  gatherMergedPRRecordsForRange,
  formatAssemblyFailureMessage,
  runChangelogRenderAndWrite,
  runChangelogEngine,
} = _changelogEngine;
