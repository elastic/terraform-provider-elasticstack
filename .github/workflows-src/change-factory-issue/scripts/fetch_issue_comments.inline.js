const MAX_COMMENTS = 200;
const { owner, repo } = context.repo;
const issueNumber = context.payload.issue?.number;

if (!issueNumber) {
  core.setOutput('issue_comments_json', '[]');
  core.info('No issue number in payload; skipping comment fetch.');
} else {
  const allComments = [];
  for await (const { data: page } of github.paginate.iterator(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  })) {
    allComments.push(...page);
    if (allComments.length >= MAX_COMMENTS) {
      break;
    }
  }

  const comments = allComments.slice(0, MAX_COMMENTS).map(c => ({
    author: c.user?.login ?? '',
    createdAt: c.created_at ?? '',
    body: c.body ?? '',
  }));

  core.setOutput('issue_comments_json', JSON.stringify(comments));
  core.info(`Fetched ${comments.length} comments for issue #${issueNumber}${allComments.length > MAX_COMMENTS ? ' (capped at ' + MAX_COMMENTS + ')' : ''}`);
}
