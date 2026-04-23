const RELEASE_BRANCH_PATTERN = /^prep-release-(.+)$/;
const SEMVER_TAG_PATTERN = /^v\d+\.\d+\.\d+$/;

function resolveReleaseMode({ eventName, headBranch = '', dispatchMode = '', targetVersion = '' }) {
  if (eventName === 'workflow_dispatch') {
    if (dispatchMode === 'release') {
      const normalizedVersion = targetVersion.trim();
      return {
        mode: 'release',
        targetVersion: normalizedVersion,
        targetBranch: normalizedVersion ? `prep-release-${normalizedVersion}` : '',
      };
    }

    return {
      mode: 'unreleased',
      targetVersion: '',
      targetBranch: 'generated-changelog',
    };
  }

  let mode = 'unreleased';
  let resolvedTargetVersion = '';
  let targetBranch = 'generated-changelog';

  if (eventName === 'pull_request' || eventName === 'pull_request_target') {
    const match = headBranch.match(RELEASE_BRANCH_PATTERN);
    if (match) {
      mode = 'release';
      resolvedTargetVersion = match[1];
      targetBranch = headBranch;
    }
  }

  return { mode, targetVersion: resolvedTargetVersion, targetBranch };
}

function parseSemverTags(tagsRaw = '') {
  return tagsRaw
    .split('\n')
    .map((tag) => tag.trim())
    .filter((tag) => SEMVER_TAG_PATTERN.test(tag));
}

function selectPreviousTag({ tags = [], mode = 'unreleased', targetVersion = '' }) {
  const excludedTag = mode === 'release' && targetVersion ? `v${targetVersion}` : '';
  const candidates = excludedTag ? tags.filter((tag) => tag !== excludedTag) : tags;

  return {
    previousTag: candidates[0] ?? '',
    excludedTag,
    excludedCurrentTag: Boolean(excludedTag) && candidates.length < tags.length,
  };
}

function buildCompareRange(previousTag) {
  return previousTag ? `${previousTag}..HEAD` : 'HEAD';
}

function buildReleaseContext({
  eventName,
  headBranch = '',
  dispatchMode = '',
  targetVersion = '',
  tags = [],
}) {
  const modeResult = resolveReleaseMode({ eventName, headBranch, dispatchMode, targetVersion });
  const previousTagResult = selectPreviousTag({
    tags,
    mode: modeResult.mode,
    targetVersion: modeResult.targetVersion,
  });

  return {
    ...modeResult,
    ...previousTagResult,
    compareRange: buildCompareRange(previousTagResult.previousTag),
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    RELEASE_BRANCH_PATTERN,
    SEMVER_TAG_PATTERN,
    buildCompareRange,
    buildReleaseContext,
    parseSemverTags,
    resolveReleaseMode,
    selectPreviousTag,
  };
}
