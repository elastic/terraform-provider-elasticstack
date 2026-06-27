const { getFactoryName, getFactoryContextDir } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const factoryName = getFactoryName();
  const dir = getFactoryContextDir();
  const priorCommentFilename = `prior_${factoryName.replace(/-factory$/, '')}_comment.md`;

  fs.mkdirSync(dir, { recursive: true });
  const issueBody = sanitizeUserContent(process.env.ISSUE_BODY);
  const issueComments = sanitizeUserContent(process.env.ISSUE_COMMENTS);
  const priorComment = sanitizeUserContent(process.env.PRIOR_FACTORY_COMMENT || '');

  if (typeof issueBody !== 'string' || typeof issueComments !== 'string' || typeof priorComment !== 'string') {
    throw new Error('sanitizeUserContent must return a string');
  }

  fs.writeFileSync(`${dir}/issue_body.md`, issueBody);
  fs.writeFileSync(`${dir}/issue_comments.md`, issueComments);
  fs.writeFileSync(`${dir}/${priorCommentFilename}`, priorComment);
  core.info(`Wrote sanitized issue context files to ${dir}/`);
};
