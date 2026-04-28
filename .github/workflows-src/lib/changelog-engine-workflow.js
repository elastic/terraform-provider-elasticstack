/**
 * Expanded into GitHub Actions inline scripts. Do not require() repository lib paths from here.
 * Node tests load changelog-engine.js instead.
 */
//include: ./changelog-release-context.js
//include: ./changelog-rewriter.js
//include: ./changelog-renderer.js
//include: ./changelog-engine-factory.js

// `execSync` MUST be declared in the including entry script (e.g. run-changelog-engine.inline.js)
// before this file is expanded, so we do not re-declare it here (duplicate const breaks the bundle).

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
