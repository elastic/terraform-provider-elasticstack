//include: ../../lib/changelog-pr-management.js

const { owner, repo } = context.repo;
let prNumber = context.payload.pull_request?.number ?? null;
const compareRange = process.env.COMPARE_RANGE;
const targetVersion = process.env.TARGET_VERSION;
const targetBranch = process.env.TARGET_BRANCH;

if (!prNumber && targetBranch) {
  const { data: prs } = await github.rest.pulls.list({
    owner,
    repo,
    state: 'open',
    head: `${owner}:${targetBranch}`,
    base: 'main',
  });
  prNumber = prs[0]?.number ?? null;

  if (prNumber) {
    core.info(`Resolved release PR #${prNumber} from branch ${targetBranch}`);
  } else {
    core.info(`No open release PR found for branch ${targetBranch}; skipping PR metadata refresh`);
  }
}

await refreshReleasePR({ github, core, owner, repo, prNumber, compareRange, targetVersion });
