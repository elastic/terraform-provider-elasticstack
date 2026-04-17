const fs = require('fs');
const { execSync } = require('child_process');

const previousTag = process.env.PREVIOUS_TAG || '';
const compareRange = process.env.COMPARE_RANGE || 'HEAD';
const mode = process.env.MODE || 'unreleased';
const targetVersion = process.env.TARGET_VERSION || '';

const { owner, repo } = context.repo;

function parseCommitShas(raw = '') {
  return raw
    .split('\n')
    .map((sha) => sha.trim())
    .filter(Boolean);
}

function getLabelNames(pr) {
  return (pr.labels ?? []).map((label) => label.name);
}

// Collect commits in range
let commitSHAs = [];
try {
  const range = compareRange || 'HEAD';
  const raw = execSync(`git log --format=%H ${range}`, {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  }).trim();
  commitSHAs = parseCommitShas(raw);
  core.info(`Found ${commitSHAs.length} commit(s) in range ${range}`);
} catch (err) {
  core.warning(`Failed to list commits in range: ${err.message}`);
}

// Find PRs associated with these commits and dedupe merged PRs as we go.
const mergedPullRequestsByNumber = new Map();

for (const sha of commitSHAs) {
  try {
    const { data: prs } = await github.rest.repos.listPullRequestsAssociatedWithCommit({
      owner,
      repo,
      commit_sha: sha,
    });
    for (const pr of prs) {
      if (pr.state === 'closed' && pr.merged_at && !mergedPullRequestsByNumber.has(pr.number)) {
        mergedPullRequestsByNumber.set(pr.number, pr);
      }
    }
  } catch (err) {
    core.warning(`Failed to list PRs for commit ${sha}: ${err.message}`);
  }
}

const mergedPullRequests = Array.from(mergedPullRequestsByNumber.values());
core.info(`Found ${mergedPullRequests.length} unique merged PR(s) in compare range`);

// Build the PR metadata records for the renderer.
// We capture: number, url, merge_commit_sha, labels, body.
// We do NOT need per-file classification here — deterministic rendering uses the
// PR-body changelog contract and the no-changelog label instead.
const prRecords = mergedPullRequests.map((pr) => ({
  number: pr.number,
  title: pr.title,
  url: pr.html_url,
  merge_commit_sha: pr.merge_commit_sha,
  author: pr.user?.login ?? 'unknown',
  labels: getLabelNames(pr),
  body: pr.body || '',
}));

const manifest = {
  generated_at: new Date().toISOString(),
  mode,
  target_version: targetVersion,
  previous_tag: previousTag,
  compare_range: compareRange,
  pr_count: prRecords.length,
  pull_requests: prRecords,
};

const artifactDir = '/tmp/changelog-assembly';
const artifactPath = `${artifactDir}/merged-prs.json`;
fs.mkdirSync(artifactDir, { recursive: true });
fs.writeFileSync(artifactPath, JSON.stringify(manifest, null, 2), 'utf8');

core.setOutput('merged_prs_path', artifactPath);
const hasPRs = prRecords.length > 0;
core.setOutput('has_prs', hasPRs ? 'true' : 'false');
core.info(`Merged PR manifest written to ${artifactPath} (${prRecords.length} PRs)`);
