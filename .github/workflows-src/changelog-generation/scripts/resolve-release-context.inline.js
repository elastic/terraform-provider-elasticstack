const { execSync } = require('child_process');
//include: ../../lib/changelog-engine-workflow.js

const eventName = context.eventName;
const headBranch =
  context.payload.pull_request?.head?.ref ??
  context.payload.ref?.replace('refs/heads/', '') ??
  process.env.GITHUB_HEAD_REF ??
  process.env.GITHUB_REF_NAME ??
  '';

const explicitModeRaw = process.env.MODE ?? process.env.INPUT_MODE ?? '';
let mode;
let targetVersion = process.env.TARGET_VERSION ?? process.env.INPUT_TARGET_VERSION ?? '';

if (explicitModeRaw === 'unreleased' || explicitModeRaw === 'release') {
  mode = explicitModeRaw;
} else {
  const legacy = resolveReleaseMode({ eventName, headBranch });
  mode = legacy.mode;
  if (legacy.targetVersion) {
    targetVersion = legacy.targetVersion;
  }
}

const ctx = resolveChangelogCompareContext({ mode, targetVersion, exec: execSync, core });

if (mode === 'release') {
  core.info(`Release mode: branch=${headBranch}, version=${targetVersion}`);
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

core.setOutput('mode', mode);
core.setOutput('target_version', mode === 'release' ? targetVersion : '');
core.setOutput('previous_tag', ctx.previousTag);
core.setOutput('compare_range', ctx.compareRange);
core.setOutput('target_branch', ctx.targetBranch);

core.info(`Mode: ${mode}`);
core.info(`Compare range: ${ctx.compareRange}`);
core.info(`Target branch: ${ctx.targetBranch}`);
