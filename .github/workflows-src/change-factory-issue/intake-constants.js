/**
 * Change-factory issue intake configuration. Keep `ISSUE_BRANCH_PREFIX` aligned with
 * the branch name used in `workflow.md.tmpl`: change-factory/issue-{n}.
 *
 * The duplicate-linkage mode is `'related-literal'` because the change-factory PR body
 * uses `Related to #N` rather than a GitHub closing keyword; merging a proposal-only PR
 * must not auto-close the source issue.
 */
'use strict';

const ISSUE_BRANCH_PREFIX = 'change-factory/issue-';
const FACTORY_LABEL = 'change-factory';
const DUPLICATE_LINKAGE_MODE = 'related-literal';
const ISSUE_OPENED_NOT_ELIGIBLE_REASON =
  'Issue opened event does not qualify because the issue was created without the change-factory label or issue labels were missing.';

if (typeof module !== 'undefined') {
  module.exports = {
    ISSUE_BRANCH_PREFIX,
    FACTORY_LABEL,
    DUPLICATE_LINKAGE_MODE,
    ISSUE_OPENED_NOT_ELIGIBLE_REASON,
  };
}
