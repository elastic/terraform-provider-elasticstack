//include: ../../lib/changelog-pr-management.js

const { owner, repo } = context.repo;
const compareRange = process.env.COMPARE_RANGE;

const result = await manageUnreleasedPR({ github, core, owner, repo, compareRange });
core.setOutput('pr_action', result.prAction);
core.setOutput('pr_number', String(result.prNumber));
core.setOutput('pr_url', result.prUrl);
