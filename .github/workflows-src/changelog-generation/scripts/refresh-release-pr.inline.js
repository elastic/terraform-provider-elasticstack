//include: ../../lib/changelog-pr-management.js

const { owner, repo } = context.repo;
const prNumber = context.payload.pull_request?.number ?? null;
const compareRange = process.env.COMPARE_RANGE;
const targetVersion = process.env.TARGET_VERSION;

await refreshReleasePR({ github, core, owner, repo, prNumber, compareRange, targetVersion });
