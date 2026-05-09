const marker = '<!-- gha-research-factory -->';
const commentsJson = process.env.INPUT_COMMENTS_JSON || '[]';

let comments;
try {
  comments = JSON.parse(commentsJson);
} catch {
  comments = [];
}

// Find research comment (most recent match by github-actions[bot])
const matches = (comments || []).filter(
  (c) => c.author === 'github-actions[bot]' && c.body.includes(marker),
);

const fs = require('fs');
const crypto = require('crypto');

if (matches.length > 0) {
  const latest = matches[matches.length - 1];
  const eofDelim = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
  const output = `research_comment_body<<${eofDelim}\n${latest.body}\n${eofDelim}\n`;
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
