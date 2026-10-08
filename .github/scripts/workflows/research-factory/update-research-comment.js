const fs = require('fs');
const { evaluateBody, applyOverride, READY, NEEDS_HUMAN } = require('./gate.js');

module.exports = async function ({ github, context, core }) {
  const { owner, repo } = context.repo;
  const issueNumber = parseInt(process.env.RESEARCH_FACTORY_ISSUE_NUMBER, 10);
  const marker = '<!-- gha-research-factory -->';

  const outputFile = process.env.GH_AW_AGENT_OUTPUT;
  if (!outputFile) {
    core.setFailed('update-research-comment: GH_AW_AGENT_OUTPUT environment variable is not set');
    return;
  }

  if (!issueNumber || issueNumber <= 0) {
    core.setFailed('update-research-comment: invalid issue number.');
    return;
  }

  const fileContent = fs.readFileSync(outputFile, 'utf8');
  const agentOutput = JSON.parse(fileContent);
  const items = (agentOutput.items || []).filter((i) => i.type === 'update_research_comment');

  if (items.length === 0) {
    core.info('update-research-comment: no update_research_comment items in agent output; nothing to do.');
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

  const result = evaluateBody(body);
  body = applyOverride(body, result);

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

  try {
    if (existingComment) {
      await github.rest.issues.updateComment({
        owner,
        repo,
        comment_id: existingComment.id,
        body,
      });
      core.info(`Updated research comment ${existingComment.id} on issue #${issueNumber}`);
    } else {
      const { data: newComment } = await github.rest.issues.createComment({
        owner,
        repo,
        issue_number: issueNumber,
        body,
      });
      core.info(`Created research comment ${newComment.id} on issue #${issueNumber}`);
    }
  } catch (err) {
    core.setFailed(`Failed to write research comment: ${err.message}`);
    return;
  }

  const otherLabel = result.label === READY ? NEEDS_HUMAN : READY;
  try {
    await github.rest.issues.addLabels({ owner, repo, issue_number: issueNumber, labels: [result.label] });
    try {
      await github.rest.issues.removeLabel({ owner, repo, issue_number: issueNumber, name: otherLabel });
    } catch (err) {
      if (err.status !== 404) {
        throw err;
      }
    }
  } catch (err) {
    core.setFailed(`Failed to set outcome label ${result.label}: ${err.message}`);
  }

  await core.summary
    .addRaw(
      `### Research gate outcome\n\n- Derived outcome: \`${result.label}\`\n- Reasons: ${result.reasons.join('; ')}\n- Reported outcome overridden: ${result.overridden ? 'yes' : 'no'}\n`,
    )
    .write();
};
