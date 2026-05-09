//include: ../../lib/sanitize-context.js

const fs = require('fs');
const crypto = require('crypto');

const body = process.env.ISSUE_BODY || '';
const comments = process.env.HUMAN_COMMENTS || '';

const sanitizedBody = stripHtmlComments(body);
const sanitizedComments = stripHtmlComments(comments);

const eofDelim1 = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
const output1 = `sanitized_issue_body<<${eofDelim1}\n${sanitizedBody}\n${eofDelim1}\n`;
fs.appendFileSync(process.env.GITHUB_OUTPUT, output1);

const eofDelim2 = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
const output2 = `sanitized_issue_comments<<${eofDelim2}\n${sanitizedComments}\n${eofDelim2}\n`;
fs.appendFileSync(process.env.GITHUB_OUTPUT, output2);

const dir = '/tmp/code-factory-context';
fs.mkdirSync(dir, { recursive: true });
fs.writeFileSync(`${dir}/issue_body.md`, sanitizedBody);
fs.writeFileSync(`${dir}/issue_comments.md`, sanitizedComments);
core.info('Wrote sanitized issue context files to /tmp/code-factory-context/');
