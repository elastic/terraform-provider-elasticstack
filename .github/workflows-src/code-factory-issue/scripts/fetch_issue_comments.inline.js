//include: ../../lib/factory-issue-comments.js

const { owner, repo } = context.repo;
const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10);

if (!issueNumber || issueNumber <= 0) {
  core.setOutput('human_comments', '');
  core.info('No issue number provided; skipping comment fetch.');
} else {
  const { comments, truncated } = await factoryFetchIssueComments({ github, owner, repo, issueNumber });
  const serialized = serializeIssueComments({ comments, truncated });
  core.setOutput('human_comments', serialized);
  core.info(`Fetched and serialized ${comments.length} human comments for issue #${issueNumber}`);
}
