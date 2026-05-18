const { getFactoryName } = require('./_factory-context.js');
const { sanitizeUserContent } = require('../sanitize-context.js');

module.exports = async function ({ github, context, core }) {

  const fs = require('fs');
  const factoryName = getFactoryName();
  const dir = `/tmp/${factoryName}-context`;
  const priorCommentFilename = `prior_${factoryName.replace(/-factory$/, '')}_comment.md`;

  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(`${dir}/issue_body.md`, sanitizeUserContent(process.env.ISSUE_BODY));
  fs.writeFileSync(`${dir}/issue_comments.md`, sanitizeUserContent(process.env.ISSUE_COMMENTS));
  fs.writeFileSync(`${dir}/${priorCommentFilename}`, sanitizeUserContent(process.env.PRIOR_FACTORY_COMMENT || ''));
  core.info(`Wrote sanitized issue context files to ${dir}/`);
};
