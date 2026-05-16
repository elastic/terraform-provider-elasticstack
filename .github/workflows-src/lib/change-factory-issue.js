'use strict';

const { createFactoryIssueModule } = require('./factory-issue-shared.js');
const {
  ISSUE_BRANCH_PREFIX,
  FACTORY_LABEL,
  DUPLICATE_LINKAGE_MODE,
  ISSUE_OPENED_NOT_ELIGIBLE_REASON,
} = require('../change-factory-issue/intake-constants.js');

module.exports = createFactoryIssueModule({
  branchPrefix: ISSUE_BRANCH_PREFIX,
  factoryLabel: FACTORY_LABEL,
  issueOpenedNotEligibleReason: ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  duplicateLinkageMode: DUPLICATE_LINKAGE_MODE,
  issueBranchNameAliases: ['changeFactoryIssueBranchName'],
});
