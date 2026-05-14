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
