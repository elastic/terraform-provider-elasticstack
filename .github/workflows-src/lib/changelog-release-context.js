const RELEASE_BRANCH_PATTERN = /^prep-release-(.+)$/;
const SEMVER_TAG_PATTERN = /^v\d+\.\d+\.\d+$/;

function resolveReleaseMode({ eventName, dispatchMode = '', targetVersion = '' }) {
  if (eventName === 'workflow_dispatch' && dispatchMode === 'release') {
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
  dispatchMode = '',
  targetVersion = '',
  tags = [],
}) {
  const modeResult = resolveReleaseMode({ eventName, dispatchMode, targetVersion });
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
