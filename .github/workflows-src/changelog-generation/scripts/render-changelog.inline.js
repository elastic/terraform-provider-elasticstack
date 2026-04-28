const fs = require('fs');
//include: ../../lib/changelog-engine-workflow.js

const mergedPRsPath = process.env.MERGED_PRS_PATH || '';
const mode = process.env.MODE || 'unreleased';
const targetVersion = process.env.TARGET_VERSION || '';
const changelogPath = process.env.CHANGELOG_PATH || 'CHANGELOG.md';

if (!mergedPRsPath) {
  core.setFailed('MERGED_PRS_PATH environment variable is required');
  process.exit(1);
}

let manifest;
try {
  manifest = JSON.parse(fs.readFileSync(mergedPRsPath, 'utf8'));
} catch (err) {
  core.setFailed(`Failed to read merged PR manifest from ${mergedPRsPath}: ${err.message}`);
  process.exit(1);
}

const prRecords = manifest.pull_requests || [];
core.info(`Rendering changelog from ${prRecords.length} merged PR(s)`);

const out = runChangelogRenderAndWrite({ core, prRecords, mode, targetVersion, changelogPath, fs });

core.setOutput('section_header', out.sectionHeader);
core.setOutput('has_changes', out.included.length > 0 || out.excluded.length > 0 ? 'true' : 'false');
core.setOutput('has_user_facing_changes', out.hasUserFacingChanges ? 'true' : 'false');
core.info(`Included ${out.included.length} PR(s) with change bullets or breaking changes`);
core.info(`Excluded ${out.excluded.length} PR(s) (no-changelog or Customer impact: none)`);
for (const ex of out.excluded) {
  core.info(`  Excluded PR #${ex.prNumber}: ${ex.reason}`);
}
core.info(`Changelog section rendered: ${out.sectionHeader}`);
