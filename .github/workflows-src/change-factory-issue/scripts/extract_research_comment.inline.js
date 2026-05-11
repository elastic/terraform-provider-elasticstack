//include: ../../lib/sanitize-context.js

const marker = '<!-- gha-research-factory -->';
const commentsJson = process.env.INPUT_COMMENTS_JSON || '[]';

let comments;
try {
  comments = JSON.parse(commentsJson);
} catch {
  comments = [];
}

const fs = require('fs');
const crypto = require('crypto');

const researchComment = findResearchComment(comments, marker);
const sanitizedBody = researchComment ? sanitizeUserContent(researchComment.body) : '';
if (researchComment) {
  const eofDelim = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
  const output = `research_comment_body<<${eofDelim}\n${sanitizedBody}\n${eofDelim}\n`;
  fs.appendFileSync(process.env.GITHUB_OUTPUT, output);
  core.info(`Found research comment for issue`);
} else {
  core.setOutput('research_comment_body', '');
  core.info('No research comment found.');
}

// Serialize human comments for agent context
const humanComments = (comments || []).filter(
  (c) => !c.author.endsWith('[bot]'),
);

//include: ../../lib/factory-issue-comments.js

const serialized = serializeIssueComments({ comments: humanComments, truncated: false });
const eofDelim2 = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
const output2 = `human_comments<<${eofDelim2}\n${serialized}\n${eofDelim2}\n`;
fs.appendFileSync(process.env.GITHUB_OUTPUT, output2);

const dir = '/tmp/change-factory-context';
fs.mkdirSync(dir, { recursive: true });
fs.writeFileSync(`${dir}/research_comment.md`, sanitizedBody);
core.info('Wrote research comment to /tmp/change-factory-context/research_comment.md');
