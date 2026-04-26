/**
 * Code-factory issue intake configuration. Keep `ISSUE_BRANCH_PREFIX` aligned with
 * `workflow.md.tmpl` (`code-factory/issue-${{ github.event.issue.number }}`).
 */
'use strict';

const ISSUE_BRANCH_PREFIX = 'code-factory/issue-';
const FACTORY_LABEL = 'code-factory';
const ISSUE_OPENED_NOT_ELIGIBLE_REASON =
  'Issue opened event does not qualify because the issue was created without the code-factory label.';

if (typeof module !== 'undefined') {
  module.exports = {
    ISSUE_BRANCH_PREFIX,
    FACTORY_LABEL,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  };
}
