const { execSync } = require('child_process');
const fs = require('fs');
//include: ../../lib/changelog-engine-workflow.js
// Mode is workflow-driven via MODE env only; the inlined bundle still carries resolveReleaseMode (event-based) from changelog-release-context.js for tests/lib reuse—unused on this path until that module is split.

const { owner, repo } = context.repo;

const mode = (process.env.MODE || 'unreleased').trim();
const targetVersion = (process.env.TARGET_VERSION || '').trim();
const targetBranchOverride = (process.env.TARGET_BRANCH || '').trim();

const ctx = resolveChangelogCompareContext({
  mode,
  targetVersion,
  exec: execSync,
  core,
});

const effectiveTargetBranch = targetBranchOverride || ctx.targetBranch;

if (mode === 'release') {
  core.info(`Release mode: version=${targetVersion}, push branch=${effectiveTargetBranch}`);
}
if (ctx.excludedCurrentTag) {
  core.info(
    `Excluded current release tag ${ctx.excludedTag} from previous tag candidates`
  );
}
if (ctx.previousTag) {
  core.info(`Resolved previous tag: ${ctx.previousTag}`);
} else if (ctx.tags.length === 0) {
  core.warning('No semver release tags found; compare range will cover full history');
}

const prRecords = await gatherMergedPRRecordsForRange({
  github,
  owner,
  repo,
  compareRange: ctx.compareRange,
  exec: execSync,
  core,
});

const hasPRs = prRecords.length > 0;
let sectionHeader = '';
let hasUserFacingChanges = false;

const shouldRenderChangelog = mode === 'release' || hasPRs;

if (shouldRenderChangelog) {
  const changelogPath = process.env.CHANGELOG_PATH || 'CHANGELOG.md';
  const out = runChangelogRenderAndWrite({
    core,
    prRecords,
    mode,
    targetVersion,
    previousTag: ctx.previousTag,
    changelogPath,
    fs,
  });
  sectionHeader = out.sectionHeader;
  hasUserFacingChanges = out.hasUserFacingChanges;
  core.info(`Included ${out.included.length} PR(s) with change bullets or breaking changes`);
  core.info(`Excluded ${out.excluded.length} PR(s) (no-changelog or Customer impact: none)`);
  for (const ex of out.excluded) {
    core.info(`  Excluded PR #${ex.prNumber}: ${ex.reason}`);
  }
  core.info(`Changelog section rendered: ${out.sectionHeader}`);
} else {
  core.info('No merged PRs in compare range; skipping changelog file update');
  sectionHeader = '## [Unreleased]';
}

core.setOutput('mode', mode);
core.setOutput('target_version', mode === 'release' ? targetVersion : '');
core.setOutput('previous_tag', ctx.previousTag);
core.setOutput('compare_range', ctx.compareRange);
core.setOutput('target_branch', effectiveTargetBranch);
core.setOutput('has_prs', hasPRs ? 'true' : 'false');
core.setOutput(
  'has_user_facing_changes',
  hasUserFacingChanges ? 'true' : 'false'
);
core.setOutput('section_header', sectionHeader);

core.info(`Mode: ${mode}`);
core.info(`Compare range: ${ctx.compareRange}`);
core.info(`Target branch: ${effectiveTargetBranch}`);
