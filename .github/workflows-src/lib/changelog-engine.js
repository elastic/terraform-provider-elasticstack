/**
 * Shared changelog assembly engine (Node entry). GitHub Actions loads changelog-engine-workflow.js (inlined).
 */

const { execSync } = require('child_process');
const { createChangelogEngine } = require('./changelog-engine-factory.js');
const crc = require('./changelog-release-context.js');
const rew = require('./changelog-rewriter.js');
Object.assign(globalThis, require('./pr-changelog-parser.js'));
const { renderChangelogSection } = require('./changelog-renderer.js');

module.exports = createChangelogEngine({
  parseSemverTags: crc.parseSemverTags,
  selectPreviousTag: crc.selectPreviousTag,
  buildCompareRange: crc.buildCompareRange,
  rewriteChangelogSection: rew.rewriteChangelogSection,
  renderChangelogSection,
  execSyncDefault: execSync,
});
