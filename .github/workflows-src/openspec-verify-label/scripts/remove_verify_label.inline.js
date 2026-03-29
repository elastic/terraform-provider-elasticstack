const issueNumber = context.payload.pull_request?.number;
if (!issueNumber) {
  core.info('No pull request number found in the event payload; skipping cleanup.');
  return;
}

try {
  await github.rest.issues.removeLabel({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: issueNumber,
    name: 'verify-openspec',
  });
  core.info('Removed verify-openspec from the triggering pull request.');
} catch (error) {
  if (error.status === 404) {
    core.info('verify-openspec was already absent on the triggering pull request.');
    return;
  }

  throw error;
}
