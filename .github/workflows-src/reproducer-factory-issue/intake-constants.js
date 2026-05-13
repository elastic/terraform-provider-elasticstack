/**
 * Reproducer-factory issue intake configuration. Keep `ISSUE_BRANCH_PREFIX` aligned with
 * `workflow.md.tmpl` (branch pattern `reproducer-factory/issue-{n}`).
 */
'use strict';

const ISSUE_BRANCH_PREFIX = 'reproducer-factory/issue-';
const FACTORY_LABEL = 'reproducer-factory';
const DUPLICATE_LINKAGE_MODE = 'related-literal';
const ISSUE_OPENED_NOT_ELIGIBLE_REASON =
  'Issue opened event does not qualify because the issue was created without the reproducer-factory label.';

if (typeof module !== 'undefined') {
  module.exports = {
    ISSUE_BRANCH_PREFIX,
    FACTORY_LABEL,
    DUPLICATE_LINKAGE_MODE,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  };
}
