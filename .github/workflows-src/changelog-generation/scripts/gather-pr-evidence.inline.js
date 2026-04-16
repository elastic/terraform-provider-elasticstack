const { execSync } = require('child_process');

const previousTag = core.getInput('previous_tag') || process.env.INPUT_PREVIOUS_TAG || '';
const compareRange = core.getInput('compare_range') || process.env.INPUT_COMPARE_RANGE || 'HEAD';
const mode = core.getInput('mode') || process.env.INPUT_MODE || 'unreleased';
const targetVersion = core.getInput('target_version') || process.env.INPUT_TARGET_VERSION || '';

const { owner, repo } = context.repo;

/**
 * Labels that indicate a PR is user-facing / CHANGELOG-worthy.
 */
const USER_FACING_LABELS = new Set([
  'enhancement',
  'bug',
  'feature',
  'breaking-change',
  'deprecation',
  'new-resource',
  'new-data-source',
]);

/**
 * Labels / patterns that indicate a PR is internal / maintenance (likely excluded).
 */
const INTERNAL_LABELS = new Set([
  'dependencies',
  'chore',
  'internal',
  'documentation',
  'ci',
  'test',
  'openspec',
]);

/**
 * File path prefixes that indicate provider-impacting changes.
 */
const PROVIDER_PATH_PREFIXES = ['internal/', 'pkg/', 'libs/', 'provider/', 'go.mod', 'go.sum'];

function classifyPR(pr, files) {
  const labels = (pr.labels ?? []).map((l) => l.name);

  // Explicit user-facing labels
  const hasUserFacingLabel = labels.some((l) => USER_FACING_LABELS.has(l));
  // Explicit internal labels
  const hasInternalLabel = labels.some((l) => INTERNAL_LABELS.has(l));

  // Dependabot / automated PRs
  const isAutomated =
    pr.user?.login === 'dependabot[bot]' ||
    pr.user?.login === 'dependabot' ||
    pr.user?.login === 'github-actions[bot]';

  // Check if PR touches provider code (not just openspec/, docs/, or tests)
  const touchesProviderCode = (files ?? []).some((f) =>
    PROVIDER_PATH_PREFIXES.some((prefix) => f.filename.startsWith(prefix))
  );

  const openspecOnly =
    (files ?? []).length > 0 && (files ?? []).every((f) => f.filename.startsWith('openspec/'));

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
    inclusionRationale = `Has user-facing label(s): ${labels.filter((l) => USER_FACING_LABELS.has(l)).join(', ')}`;
  } else if (hasInternalLabel && !touchesProviderCode) {
    classification = 'internal';
    exclusionRationale = `Has internal label(s): ${labels.filter((l) => INTERNAL_LABELS.has(l)).join(', ')} and does not touch provider code`;
  } else if (touchesProviderCode) {
    classification = 'user-facing';
    inclusionRationale = 'Touches provider implementation paths — presumed user-facing';
  } else {
    classification = 'uncertain';
    inclusionRationale = 'Classification uncertain — agent to decide';
  }

  return { classification, inclusionRationale: inclusionRationale ?? null, exclusionRationale: exclusionRationale ?? null };
}

// Collect commits in range
let commitSHAs = [];
try {
  const range = compareRange || 'HEAD';
  const raw = execSync(`git log --format=%H ${range}`, {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  }).trim();
  commitSHAs = raw ? raw.split('\n').map((s) => s.trim()).filter(Boolean) : [];
  core.info(`Found ${commitSHAs.length} commit(s) in range ${range}`);
} catch (err) {
  core.warning(`Failed to list commits in range: ${err.message}`);
}

// Find PRs associated with these commits
const prMap = new Map();

for (const sha of commitSHAs) {
  try {
    const { data: prs } = await github.rest.repos.listPullRequestsAssociatedWithCommit({
      owner,
      repo,
      commit_sha: sha,
    });
    for (const pr of prs) {
      if (pr.state === 'closed' && pr.merged_at && !prMap.has(pr.number)) {
        prMap.set(pr.number, pr);
      }
    }
  } catch (err) {
    core.warning(`Failed to list PRs for commit ${sha}: ${err.message}`);
  }
}

core.info(`Found ${prMap.size} unique merged PR(s) in compare range`);

// Enrich each PR with file information
const evidence = [];

for (const [prNumber, pr] of prMap) {
  let files = [];
  try {
    files = await github.paginate(github.rest.pulls.listFiles, {
      owner,
      repo,
      pull_number: prNumber,
      per_page: 100,
    });
  } catch (err) {
    core.warning(`Failed to list files for PR #${prNumber}: ${err.message}`);
  }

  const { classification, inclusionRationale, exclusionRationale } = classifyPR(pr, files);

  evidence.push({
    number: pr.number,
    title: pr.title,
    url: pr.html_url,
    merge_commit_sha: pr.merge_commit_sha,
    author: pr.user?.login ?? 'unknown',
    labels: (pr.labels ?? []).map((l) => l.name),
    touched_files: files.map((f) => f.filename),
    classification,
    inclusion_rationale: inclusionRationale,
    exclusion_rationale: exclusionRationale,
  });
}

const targetSection =
  mode === 'release' && targetVersion
    ? `## [${targetVersion}] - ${new Date().toISOString().split('T')[0]}`
    : '## [Unreleased]';

const manifest = {
  generated_at: new Date().toISOString(),
  mode,
  target_section: targetSection,
  target_section_mode: mode,
  target_version: targetVersion,
  previous_tag: previousTag,
  compare_range: compareRange,
  pr_count: evidence.length,
  user_facing_count: evidence.filter((e) => e.classification === 'user-facing').length,
  internal_count: evidence.filter((e) => e.classification === 'internal').length,
  uncertain_count: evidence.filter((e) => e.classification === 'uncertain').length,
  pull_requests: evidence,
};

core.setOutput('evidence_json', JSON.stringify(manifest));
const hasEvidence = evidence.length > 0;
core.setOutput('has_evidence', hasEvidence ? 'true' : 'false');
core.info(`Evidence manifest built: ${evidence.length} PRs (${manifest.user_facing_count} user-facing, ${manifest.internal_count} internal, ${manifest.uncertain_count} uncertain)`);
