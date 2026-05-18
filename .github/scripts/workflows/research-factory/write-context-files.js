const { sanitizeUserContent } = require('../lib/sanitize-context.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const dir = '/tmp/research-factory-context';

  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/issue_body.md`, sanitizeUserContent(process.env.ISSUE_BODY));
  fs.writeFileSync(`${dir}/issue_comments.md`, sanitizeUserContent(process.env.ISSUE_COMMENTS));
  fs.writeFileSync(`${dir}/prior_research_comment.md`, sanitizeUserContent(process.env.PRIOR_RESEARCH_COMMENT || ''));
  core.info('Wrote sanitized issue context files to /tmp/research-factory-context/');
};
