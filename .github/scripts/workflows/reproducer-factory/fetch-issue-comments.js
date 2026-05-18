const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../lib/intake/reproducer-factory-constants.js');
const { createFactoryIssueModule } = require('../lib/factory-issue-shared.js');
const factoryIssueModule = createFactoryIssueModule({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: DUPLICATE_LINKAGE_MODE,
});
const qualifyTriggerEvent = factoryIssueModule.qualifyTriggerEvent;
const checkActorTrust = factoryIssueModule.checkActorTrust;
const checkDuplicatePR = factoryIssueModule.checkDuplicatePR;
const computeGateReason = factoryIssueModule.computeGateReason;
const issueBranchName = factoryIssueModule.issueBranchName;
const actorTrustWhenSenderMissing = factoryIssueModule.actorTrustWhenSenderMissing;
const parseFinalizeGateEnv = factoryIssueModule.parseFinalizeGateEnv;

const { factoryFetchIssueComments, serializeIssueComments } = require('../lib/factory-issue-comments.js');

module.exports = async function ({ github, context, core }) {

  const { owner, repo } = context.repo;
  const issueNumber = parseInt(process.env.INPUT_ISSUE_NUMBER, 10);

  if (!issueNumber || issueNumber <= 0) {
    core.setOutput('comment_count', '0');
    core.setOutput('issue_comments', '');
    core.setFailed('Cannot fetch issue comments: invalid issue number.');
  } else {
    try {
      const fetchResult = await factoryFetchIssueComments({ github, owner, repo, issueNumber });
      const serialized = serializeIssueComments(fetchResult);

      const fs = require('fs');
      const crypto = require('crypto');
      const eofDelim = `EOF_${crypto.randomUUID().replace(/-/g, '')}`;
      const output = `issue_comments<<${eofDelim}\n${serialized}\n${eofDelim}\n`;
      fs.appendFileSync(process.env.GITHUB_OUTPUT, output);

      core.setOutput('comment_count', String(fetchResult.comments.length));
      core.info(`Fetched ${fetchResult.comments.length} comments for issue #${issueNumber}${fetchResult.truncated ? ' (truncated)' : ''}`);
    } catch (err) {
      core.setOutput('comment_count', '0');
      core.setOutput('issue_comments', '');
      core.setFailed(`Failed to fetch issue comments for #${issueNumber}: ${err.message}`);
    }
  }
};
