const fs = require('fs');
const { owner, repo } = context.repo;
const workflowRun = context.payload.workflow_run;
const headSha = workflowRun?.head_sha;
const headBranch = workflowRun?.head_branch;

if (!headSha) {
  core.setFailed('PR_CHANGELOG_GATING: No head_sha in workflow_run payload — cannot resolve PR');
  return;
}

if (!headBranch) {
  core.setFailed('PR_CHANGELOG_GATING: Could not determine head branch from workflow_run event');
  return;
}

const headRepoFullName = workflowRun.head_repository?.full_name;
const baseRepoFullName = `${context.repo.owner}/${context.repo.repo}`;
if (headRepoFullName !== baseRepoFullName) {
  core.info(`Skipping: workflow_run is for a fork PR (${headRepoFullName})`);
  core.setOutput('is_pr_event', 'false');
  return;
}

core.info(`Resolving PR for head_sha=${headSha} head_branch=${headBranch}`);

// List open PRs from the triggering head branch
const { data: matchingPRs } = await github.rest.pulls.list({
  owner,
  repo,
  state: 'open',
  head: `${owner}:${headBranch}`,
  per_page: 10,
});

const candidatePRs = matchingPRs.filter((pr) => pr.head.sha === headSha);

if (candidatePRs.length === 0) {
  core.setFailed(
    `PR_CHANGELOG_GATING: No open pull request found for head_sha=${headSha} head_branch=${headBranch}`
  );
  return;
}

if (candidatePRs.length > 1) {
  const numbers = candidatePRs.map((pr) => `#${pr.number}`).join(', ');
  core.setFailed(
    `PR_CHANGELOG_GATING: Multiple open pull requests found for head_sha=${headSha}: ${numbers} — cannot resolve deterministically`
  );
  return;
}

const pr = candidatePRs[0];
core.info(`Resolved PR #${pr.number}: ${pr.title}`);

const labelNames = (pr.labels ?? []).map((l) => l.name);
const hasNoChangelog = labelNames.includes('no-changelog');
core.info(`PR labels: ${labelNames.join(', ') || '(none)'}`);
core.info(`no-changelog label present: ${hasNoChangelog}`);

const prBody = pr.body ?? '';
const prBodyPath = '/tmp/pr-body.txt';
fs.mkdirSync('/tmp', { recursive: true });
fs.writeFileSync(prBodyPath, prBody, 'utf8');
core.info(`PR body written to ${prBodyPath} (${prBody.length} bytes)`);

core.setOutput('is_pr_event', 'true');
core.setOutput('pr_number', String(pr.number));
core.setOutput('pr_title', pr.title);
core.setOutput('pr_body', prBody);
core.setOutput('pr_body_path', prBodyPath);
core.setOutput('pr_url', pr.html_url);
core.setOutput('has_no_changelog_label', hasNoChangelog ? 'true' : 'false');
