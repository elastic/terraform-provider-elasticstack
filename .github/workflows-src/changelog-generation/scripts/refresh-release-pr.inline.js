//include: ../../lib/changelog-pr-management.js

const { owner, repo } = context.repo;
const compareRange = process.env.COMPARE_RANGE;
const targetVersion = process.env.TARGET_VERSION;
const targetBranch = process.env.TARGET_BRANCH;

await refreshReleasePR({ github, core, owner, repo, targetBranch, compareRange, targetVersion });
