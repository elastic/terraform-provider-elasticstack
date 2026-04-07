//include: ../../lib/select-change.js
//include: ../../lib/classify-pr.js

const prNumber = context.payload.pull_request?.number;
const headRepoId = context.payload.pull_request?.head?.repo?.id;
const baseRepoId = context.payload.repository?.id;

// Classify the pull request trust level
const classification = classifyPullRequest({ headRepoId, baseRepoId });

// Select the active change from PR files
let selectionResult;
if (!prNumber) {
  selectionResult = selectChangeForPullRequest({ prNumber });
} else {
  const files = await github.paginate(github.rest.pulls.listFiles, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: prNumber,
    per_page: 100,
  });

  selectionResult = selectChangeForPullRequest({ prNumber, files });
}

core.setOutput('selection_status', selectionResult.selection_status);
core.setOutput('selection_reason', selectionResult.selection_reason);
core.setOutput('selected_change', selectionResult.selected_change);
core.setOutput('review_disposition', selectionResult.review_disposition ?? '');
core.setOutput('disposition_reason', selectionResult.disposition_reason ?? '');
core.setOutput('verification_mode', classification.verification_mode);
core.setOutput('verification_mode_reason', classification.verification_mode_reason);
core.setOutput('archive_push_allowed', classification.archive_push_allowed ? 'true' : 'false');
core.setOutput('archive_push_reason', classification.archive_push_reason);

if (selectionResult.selection_status === 'eligible') {
  core.info(
    `Selected active change: ${selectionResult.selected_change} (${selectionResult.review_disposition})`
  );
}
core.info(
  `PR classification: ${classification.verification_mode} (archive/push: ${classification.archive_push_allowed})`
);
