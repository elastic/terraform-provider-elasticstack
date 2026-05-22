function buildDocsGuideUrl(owner, repo, defaultBranch) {
  const branch = defaultBranch && String(defaultBranch).trim() !== '' ? defaultBranch : 'main';
  return `https://github.com/${owner}/${repo}/blob/${branch}/dev-docs/high-level/openspec-requirements.md`;
}

function buildCommentBody(selectionReason, { owner, repo, defaultBranch } = {}) {
  const reason = typeof selectionReason === 'string' && selectionReason.trim() !== ''
    ? selectionReason
    : '(no reason provided by classify_and_select)';

  const docsUrl = owner && repo
    ? buildDocsGuideUrl(owner, repo, defaultBranch)
    : 'https://github.com/elastic/terraform-provider-elasticstack/blob/main/dev-docs/high-level/openspec-requirements.md';

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
    `See the [OpenSpec authoring guide](${docsUrl}) for details.`,
  ].join('\n');
}

module.exports = async function ({ github, context, core }) {
  const prNumber = context.payload.pull_request?.number;
  if (prNumber === undefined || prNumber === null) {
    core.info('comment_ineligible: skipped — no pull request number in event payload');
    return;
  }

  const selectionReason = process.env.SELECTION_REASON ?? '';
  const defaultBranch = context.payload.repository?.default_branch;
  const body = buildCommentBody(selectionReason, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    defaultBranch,
  });

  await github.rest.issues.createComment({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: prNumber,
    body,
  });

  core.info(`comment_ineligible: posted ineligibility comment on PR #${prNumber}`);
};

module.exports.buildCommentBody = buildCommentBody;
