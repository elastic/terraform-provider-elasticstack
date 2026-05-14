const { owner, repo } = context.repo;
const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10);
const marker = '<!-- gha-reproducer-factory -->';

if (!issueNumber || issueNumber <= 0) {
  core.setOutput('prior_reproducer_comment', '');
  return;
}

try {
  const allComments = await github.paginate(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  });

  const matches = allComments.filter(
    (c) => c.user?.login === 'github-actions[bot]' && c.body?.trimStart().startsWith(marker),
  );

  if (matches.length > 0) {
    const latest = matches[matches.length - 1];
    core.setOutput('prior_reproducer_comment', latest.body);
    core.info(`Found prior reproducer comment ${latest.id} for issue #${issueNumber}`);
  } else {
    core.setOutput('prior_reproducer_comment', '');
    core.info(`No prior reproducer comment found for issue #${issueNumber}`);
  }
} catch (err) {
  core.setOutput('prior_reproducer_comment', '');
  core.warning(`Could not fetch prior reproducer comment for issue #${issueNumber}: ${err.message}`);
}
