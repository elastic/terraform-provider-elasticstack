const { execSync } = require('child_process');
//include: ../../lib/changelog-pr-evidence.js

const previousTag = core.getInput('previous_tag') || process.env.INPUT_PREVIOUS_TAG || '';
const compareRange = core.getInput('compare_range') || process.env.INPUT_COMPARE_RANGE || 'HEAD';
const mode = core.getInput('mode') || process.env.INPUT_MODE || 'unreleased';
const targetVersion = core.getInput('target_version') || process.env.INPUT_TARGET_VERSION || '';

const { owner, repo } = context.repo;

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

// Find PRs associated with these commits
const associatedPullRequests = [];

for (const sha of commitSHAs) {
  try {
    const { data: prs } = await github.rest.repos.listPullRequestsAssociatedWithCommit({
      owner,
      repo,
      commit_sha: sha,
    });
    associatedPullRequests.push(...prs);
  } catch (err) {
    core.warning(`Failed to list PRs for commit ${sha}: ${err.message}`);
  }
}

const mergedPullRequests = selectMergedPullRequests(associatedPullRequests);
core.info(`Found ${mergedPullRequests.length} unique merged PR(s) in compare range`);

// Enrich each PR with file information
const evidence = [];

for (const pr of mergedPullRequests) {
  const prNumber = pr.number;
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

  evidence.push(buildPullRequestEvidence(pr, files));
}

const manifest = buildEvidenceManifest({
  mode,
  targetVersion,
  previousTag,
  compareRange,
  evidence,
  generatedAt: new Date().toISOString(),
});

core.setOutput('evidence_json', JSON.stringify(manifest));
const hasEvidence = evidence.length > 0;
core.setOutput('has_evidence', hasEvidence ? 'true' : 'false');
core.info(`Evidence manifest built: ${evidence.length} PRs (${manifest.user_facing_count} user-facing, ${manifest.internal_count} internal, ${manifest.uncertain_count} uncertain)`);
