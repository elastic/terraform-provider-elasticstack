const { owner, repo } = context.repo;
const issueNumber = parseInt(process.env.RESEARCH_FACTORY_ISSUE_NUMBER, 10);
const marker = '<!-- gha-research-factory -->';

if (!item) {
  core.setFailed('update-research-comment: no item provided');
  return;
}

let body = item.body || '';

if (!issueNumber || issueNumber <= 0) {
  core.setFailed('update-research-comment: invalid issue number.');
  return;
}

// Prepend the marker automatically; the agent does not need to supply it.
if (body.startsWith(marker + '\n') || body.startsWith(marker + '\r\n')) {
  // body already starts with marker; leave as-is
} else if (body.startsWith(marker)) {
  // marker without newline; normalize
  body = marker + '\n' + body.slice(marker.length);
} else {
  body = marker + '\n' + body;
}

// Find existing research comment by github-actions[bot]
let existingComment = null;
try {
  const comments = await github.paginate(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  });
  for (let i = comments.length - 1; i >= 0; i--) {
    const c = comments[i];
    if (c.user?.login === 'github-actions[bot]' && c.body?.trimStart().startsWith(marker)) {
      existingComment = c;
      break;
    }
  }
} catch (err) {
  core.setFailed(`Could not list comments while searching for existing research comment: ${err.message}`);
  return;
}

if (existingComment) {
  try {
    await github.rest.issues.updateComment({
      owner,
      repo,
      comment_id: existingComment.id,
      body,
    });
    core.info(`Updated research comment ${existingComment.id} on issue #${issueNumber}`);
  } catch (err) {
    core.setFailed(`Failed to update research comment: ${err.message}`);
  }
} else {
  try {
    const { data: newComment } = await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: issueNumber,
      body,
    });
    core.info(`Created research comment ${newComment.id} on issue #${issueNumber}`);
  } catch (err) {
    core.setFailed(`Failed to create research comment: ${err.message}`);
  }
}
