//include: ../../lib/select-change.js

const prNumber = context.payload.pull_request?.number;
let result;

if (!prNumber) {
  result = selectChangeForPullRequest({ prNumber });
} else {
  const files = await github.paginate(github.rest.pulls.listFiles, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: prNumber,
    per_page: 100,
  });

  result = selectChangeForPullRequest({ prNumber, files });
}

core.setOutput('selection_status', result.selection_status);
core.setOutput('selection_reason', result.selection_reason);
core.setOutput('selected_change', result.selected_change);

if (result.selection_status === 'eligible') {
  core.info(`Selected active change: ${result.selected_change}`);
}
