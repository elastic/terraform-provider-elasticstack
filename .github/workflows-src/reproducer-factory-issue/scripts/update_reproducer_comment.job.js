const fs = require('fs');
const { owner, repo } = context.repo;
const issueNumber = parseInt(process.env.REPRODUCER_FACTORY_ISSUE_NUMBER, 10);
const marker = '<!-- gha-reproducer-factory -->';

const outputFile = process.env.GH_AW_AGENT_OUTPUT;
if (!outputFile) {
  core.setFailed('update-reproducer-comment: GH_AW_AGENT_OUTPUT environment variable is not set');
  return;
}

if (!issueNumber || issueNumber <= 0) {
  core.setFailed('update-reproducer-comment: invalid issue number.');
  return;
}

const fileContent = fs.readFileSync(outputFile, 'utf8');
const agentOutput = JSON.parse(fileContent);
const items = (agentOutput.items || []).filter(i => i.type === 'update_reproducer_comment');

if (items.length === 0) {
  core.info('update-reproducer-comment: no update_reproducer_comment items in agent output; nothing to do.');
  return;
}

const item = items[0];
let body = item.body || '';

// Prepend the marker automatically; the agent does not need to supply it.
if (body.startsWith(marker + '\n') || body.startsWith(marker + '\r\n')) {
  // body already starts with marker; leave as-is
} else if (body.startsWith(marker)) {
  // marker without newline; normalize
  body = marker + '\n' + body.slice(marker.length);
} else {
  body = marker + '\n' + body;
}

// Find existing reproducer comment by github-actions[bot]
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
  core.setFailed(`Could not list comments while searching for existing reproducer comment: ${err.message}`);
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
    core.info(`Updated reproducer comment ${existingComment.id} on issue #${issueNumber}`);
  } catch (err) {
    core.setFailed(`Failed to update reproducer comment: ${err.message}`);
  }
} else {
  try {
    const { data: newComment } = await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: issueNumber,
      body,
    });
    core.info(`Created reproducer comment ${newComment.id} on issue #${issueNumber}`);
  } catch (err) {
    core.setFailed(`Failed to create reproducer comment: ${err.message}`);
  }
}
