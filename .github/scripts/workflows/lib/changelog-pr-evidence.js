const USER_FACING_LABELS = new Set([
  'enhancement',
  'bug',
  'feature',
  'breaking-change',
  'deprecation',
  'new-resource',
  'new-data-source',
]);

const INTERNAL_LABELS = new Set([
  'dependencies',
  'chore',
  'internal',
  'documentation',
  'ci',
  'test',
  'openspec',
]);

const PROVIDER_PATH_PREFIXES = ['internal/', 'pkg/', 'libs/', 'provider/', 'go.mod', 'go.sum'];

function getLabelNames(pr) {
  return (pr.labels ?? []).map((label) => label.name);
}

function classifyPullRequestForChangelog(pr, files = []) {
  const labels = getLabelNames(pr);
  const hasUserFacingLabel = labels.some((label) => USER_FACING_LABELS.has(label));
  const hasInternalLabel = labels.some((label) => INTERNAL_LABELS.has(label));
  const isAutomated =
    pr.user?.login === 'dependabot[bot]' ||
    pr.user?.login === 'dependabot' ||
    pr.user?.login === 'github-actions[bot]';

  const touchesProviderCode = files.some((file) =>
    PROVIDER_PATH_PREFIXES.some((prefix) => file.filename.startsWith(prefix))
  );
  const openspecOnly = files.length > 0 && files.every((file) => file.filename.startsWith('openspec/'));

  let classification;
  let inclusionRationale;
  let exclusionRationale;

  if (isAutomated) {
    classification = 'internal';
    exclusionRationale = `Automated PR by ${pr.user?.login}`;
  } else if (openspecOnly) {
    classification = 'internal';
    exclusionRationale = 'Touches only openspec/ files — no provider code changes';
  } else if (hasUserFacingLabel) {
    classification = 'user-facing';
    inclusionRationale = `Has user-facing label(s): ${labels
      .filter((label) => USER_FACING_LABELS.has(label))
      .join(', ')}`;
  } else if (hasInternalLabel && !touchesProviderCode) {
    classification = 'internal';
    exclusionRationale = `Has internal label(s): ${labels
      .filter((label) => INTERNAL_LABELS.has(label))
      .join(', ')} and does not touch provider code`;
  } else if (touchesProviderCode) {
    classification = 'user-facing';
    inclusionRationale = 'Touches provider implementation paths — presumed user-facing';
  } else {
    classification = 'uncertain';
    inclusionRationale = 'Classification uncertain — agent to decide';
  }

  return {
    classification,
    inclusionRationale: inclusionRationale ?? null,
    exclusionRationale: exclusionRationale ?? null,
  };
}

function parseCommitShas(raw = '') {
  return raw
    .split('\n')
    .map((sha) => sha.trim())
    .filter(Boolean);
}

function selectMergedPullRequests(prs = []) {
  const seen = new Set();
  const merged = [];

  for (const pr of prs) {
    if (pr.state === 'closed' && pr.merged_at && !seen.has(pr.number)) {
      seen.add(pr.number);
      merged.push(pr);
    }
  }

  return merged;
}

function buildPullRequestEvidence(pr, files = []) {
  const { classification, inclusionRationale, exclusionRationale } =
    classifyPullRequestForChangelog(pr, files);

  return {
    number: pr.number,
    title: pr.title,
    url: pr.html_url,
    merge_commit_sha: pr.merge_commit_sha,
    author: pr.user?.login ?? 'unknown',
    labels: getLabelNames(pr),
    touched_files: files.map((file) => file.filename),
    classification,
    inclusion_rationale: inclusionRationale,
    exclusion_rationale: exclusionRationale,
  };
}

function buildTargetSection({ mode = 'unreleased', targetVersion = '', date = new Date().toISOString() }) {
  if (mode === 'release' && targetVersion) {
    return `## [${targetVersion}] - ${String(date).split('T')[0]}`;
  }

  return '## [Unreleased]';
}

function countByClassification(evidence = []) {
  return {
    user_facing_count: evidence.filter((item) => item.classification === 'user-facing').length,
    internal_count: evidence.filter((item) => item.classification === 'internal').length,
    uncertain_count: evidence.filter((item) => item.classification === 'uncertain').length,
  };
}

function buildEvidenceManifest({
  mode = 'unreleased',
  targetVersion = '',
  previousTag = '',
  compareRange = 'HEAD',
  evidence = [],
  generatedAt = new Date().toISOString(),
}) {
  const counts = countByClassification(evidence);

  return {
    generated_at: generatedAt,
    mode,
    target_section: buildTargetSection({ mode, targetVersion, date: generatedAt }),
    target_section_mode: mode,
    target_version: targetVersion,
    previous_tag: previousTag,
    compare_range: compareRange,
    pr_count: evidence.length,
    ...counts,
    pull_requests: evidence,
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    INTERNAL_LABELS,
    PROVIDER_PATH_PREFIXES,
    USER_FACING_LABELS,
    buildEvidenceManifest,
    buildPullRequestEvidence,
    buildTargetSection,
    classifyPullRequestForChangelog,
    countByClassification,
    parseCommitShas,
    selectMergedPullRequests,
  };
}
