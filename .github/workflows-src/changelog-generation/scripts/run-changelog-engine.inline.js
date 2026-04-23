//include: ../../lib/changelog-engine.js

const mode = process.env.CHANGELOG_MODE || (context.eventName === 'schedule' ? 'unreleased' : 'unreleased');
const targetVersion = process.env.TARGET_VERSION || '';

const { owner, repo } = context.repo;

const result = await runChangelogEngine({
  github,
  owner,
  repo,
  mode,
  targetVersion,
  changelogPath: process.env.CHANGELOG_PATH || 'CHANGELOG.md',
});

core.setOutput('mode', result.mode);
core.setOutput('target_version', result.targetVersion);
core.setOutput('target_branch', result.targetBranch);
core.setOutput('previous_tag', result.previousTag);
core.setOutput('compare_range', result.compareRange);
core.setOutput('section_header', result.sectionHeader);
core.setOutput('has_user_facing_changes', String(result.hasUserFacingChanges));
core.setOutput('has_pull_requests', String(result.pullRequests.length > 0));
