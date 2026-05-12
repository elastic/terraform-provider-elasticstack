/**
 * Research-factory issue intake configuration.
 * Note: ISSUE_BRANCH_PREFIX is not applicable — this workflow does not open PRs or create branches.
 */
'use strict';

const ISSUE_BRANCH_PREFIX = ''; // Unused — research-factory does not create branches
const FACTORY_LABEL = 'research-factory';
const DUPLICATE_LINKAGE_MODE = 'github-keywords'; // Unused — research-factory does not check duplicate PRs
const ISSUE_OPENED_NOT_ELIGIBLE_REASON =
  'Issue opened event does not qualify because the issue was created without the research-factory label or issue labels were missing.';

if (typeof module !== 'undefined') {
  module.exports = {
    ISSUE_BRANCH_PREFIX,
    FACTORY_LABEL,
    DUPLICATE_LINKAGE_MODE,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  };
}
