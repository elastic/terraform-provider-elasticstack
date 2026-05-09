const { owner, repo } = context.repo;
const issueNumber = context.payload.issue?.number;

if (!issueNumber) {
  core.setOutput('issue_comments_json', '[]');
  core.info('No issue number in payload; skipping comment fetch.');
} else {
  try {
    const allComments = await github.paginate(github.rest.issues.listComments, {
      owner,
      repo,
      issue_number: issueNumber,
      per_page: 100,
    });

    const comments = allComments.map(c => ({
      author: c.user?.login ?? '',
      createdAt: c.created_at ?? '',
      body: c.body ?? '',
    }));

    core.setOutput('issue_comments_json', JSON.stringify(comments));
    core.info(`Fetched ${comments.length} comments for issue #${issueNumber}`);
  } catch (err) {
    core.setOutput('issue_comments_json', '[]');
    core.warning(`Failed to fetch comments: ${err.message}`);
  }
}
