const { execSync } = require('child_process');
//include: ../../lib/changelog-release-context.js

// Determine mode from event
const eventName = context.eventName;
const headBranch =
  context.payload.pull_request?.head?.ref ??
  context.payload.ref?.replace('refs/heads/', '') ??
  process.env.GITHUB_HEAD_REF ??
  process.env.GITHUB_REF_NAME ??
  '';

let tags = [];
try {
  // List all tags matching vX.Y.Z semver pattern, sort by version, pick the latest
  const tagsRaw = execSync(
    'git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-version:refname',
    { encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] }
  ).trim();

  tags = parseSemverTags(tagsRaw);
} catch (err) {
  core.warning(`Failed to list git tags: ${err.message}`);
}

const releaseContext = buildReleaseContext({ eventName, headBranch, tags });

if (releaseContext.mode === 'release') {
  core.info(`Release mode: branch=${headBranch}, version=${releaseContext.targetVersion}`);
}
if (releaseContext.excludedCurrentTag) {
  core.info(
    `Excluded current release tag ${releaseContext.excludedTag} from previous tag candidates`
  );
}
if (releaseContext.previousTag) {
  core.info(`Resolved previous tag: ${releaseContext.previousTag}`);
} else if (tags.length === 0) {
  core.warning('No semver release tags found; compare range will cover full history');
}

core.setOutput('mode', releaseContext.mode);
core.setOutput('target_version', releaseContext.targetVersion);
core.setOutput('previous_tag', releaseContext.previousTag);
core.setOutput('compare_range', releaseContext.compareRange);
core.setOutput('target_branch', releaseContext.targetBranch);

core.info(`Mode: ${releaseContext.mode}`);
core.info(`Compare range: ${releaseContext.compareRange}`);
core.info(`Target branch: ${releaseContext.targetBranch}`);
