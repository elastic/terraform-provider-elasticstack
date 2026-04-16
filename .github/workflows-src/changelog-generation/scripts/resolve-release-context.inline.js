const { execSync } = require('child_process');

/**
 * Resolves the release context:
 *   - mode: 'unreleased' | 'release'
 *   - target_version: e.g. '0.14.4' (release mode only)
 *   - previous_tag: e.g. 'v0.14.3'
 *   - compare_range: e.g. 'v0.14.3..HEAD'
 *   - target_branch: branch to push the updated changelog to
 */

// Determine mode from event
const eventName = context.eventName;
const headBranch =
  context.payload.pull_request?.head?.ref ??
  context.payload.ref?.replace('refs/heads/', '') ??
  process.env.GITHUB_HEAD_REF ??
  process.env.GITHUB_REF_NAME ??
  '';

let mode = 'unreleased';
let targetVersion = '';
let targetBranch = 'generated-changelog';

if (eventName === 'pull_request' || eventName === 'pull_request_target') {
  const match = headBranch.match(/^prep-release-(.+)$/);
  if (match) {
    mode = 'release';
    targetVersion = match[1];
    targetBranch = headBranch;
    core.info(`Release mode: branch=${headBranch}, version=${targetVersion}`);
  }
}

// Resolve the previous semver release tag from git
let previousTag = '';
try {
  // List all tags matching vX.Y.Z semver pattern, sort by version, pick the latest
  const tagsRaw = execSync(
    'git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-version:refname',
    { encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] }
  ).trim();

  const tags = tagsRaw
    .split('\n')
    .map((t) => t.trim())
    .filter((t) => /^v\d+\.\d+\.\d+$/.test(t));

  if (tags.length > 0) {
    // In release mode, exclude the current target version tag to avoid picking up the release being prepared
    let candidates = tags;
    if (mode === 'release' && targetVersion) {
      const versionToExclude = `v${targetVersion}`;
      candidates = tags.filter((t) => t !== versionToExclude);
      if (candidates.length < tags.length) {
        core.info(`Excluded current release tag ${versionToExclude} from previous tag candidates`);
      }
    }
    previousTag = candidates[0] ?? '';
    if (previousTag) {
      core.info(`Resolved previous tag: ${previousTag}`);
    }
  } else {
    core.warning('No semver release tags found; compare range will cover full history');
  }
} catch (err) {
  core.warning(`Failed to list git tags: ${err.message}`);
}

const compareRange = previousTag ? `${previousTag}..HEAD` : 'HEAD';

core.setOutput('mode', mode);
core.setOutput('target_version', targetVersion);
core.setOutput('previous_tag', previousTag);
core.setOutput('compare_range', compareRange);
core.setOutput('target_branch', targetBranch);

core.info(`Mode: ${mode}`);
core.info(`Compare range: ${compareRange}`);
core.info(`Target branch: ${targetBranch}`);
