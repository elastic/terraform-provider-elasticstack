function buildCommentBody(selectionReason) {
  const reason = typeof selectionReason === 'string' && selectionReason.trim() !== ''
    ? selectionReason
    : '(no reason provided by classify_and_select)';

  return [
    '**OpenSpec verify skipped** ⚠️',
    '',
    'The `verify-openspec` label was applied but this PR is not eligible for verification:',
    '',
    `> ${reason}`,
    '',
    '**How to fix**',
    '',
    'For the PR to be verified, it must contain exactly one active OpenSpec change directory under `openspec/changes/<id>/` where:',
    '- `<id>` is a single path segment (not `archive`).',
    '- All changed files under that path have status `added` or `modified` only (no renames, deletes, etc.).',
    '- No other OpenSpec change directories (non-archive) appear in the PR.',
    '',
    'See the [OpenSpec authoring guide](../../dev-docs/high-level/openspec-requirements.md) for details.',
  ].join('\n');
}

module.exports = async function ({ github, context, core }) {
  const prNumber = context.payload.pull_request?.number;
  if (prNumber === undefined || prNumber === null) {
    core.info('comment_ineligible: skipped — no pull request number in event payload');
    return;
  }

  const selectionReason = process.env.SELECTION_REASON ?? '';
  const body = buildCommentBody(selectionReason);

  await github.rest.issues.createComment({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: prNumber,
    body,
  });

  core.info(`comment_ineligible: posted ineligibility comment on PR #${prNumber}`);
};

module.exports.buildCommentBody = buildCommentBody;
