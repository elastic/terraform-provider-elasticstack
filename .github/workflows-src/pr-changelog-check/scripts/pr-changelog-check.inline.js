//include: ../../lib/pr-changelog-parser.js
//include: ../../lib/pr-changelog-check.js

const pr = context.payload.pull_request;
const prNumber = pr.number;
const prBody = pr.body || '';
const labels = (pr.labels || []).map((l) => l.name);

const COMMENT_MARKER = '<!-- pr-changelog-check -->';

// Early-exit if no-changelog label is present
if (labels.includes('no-changelog')) {
  core.info('PR has no-changelog label — skipping changelog check');
  const noChangelogComments = await github.paginate(github.rest.issues.listComments, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: prNumber,
    per_page: 100,
  });
  const existingOnNoChangelog = findExistingComment(noChangelogComments, COMMENT_MARKER);
  if (existingOnNoChangelog) {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: existingOnNoChangelog.id,
      body: buildNoChangelogPassCommentBody(COMMENT_MARKER),
    });
  }
  return;
}

// Parse and validate the ## Changelog section
const parsed = parseChangelogSectionFull(prBody);
const { valid, errors } = validateChangelogSectionFull(parsed);

// Find existing bot comment
const allComments = await github.paginate(github.rest.issues.listComments, {
  owner: context.repo.owner,
  repo: context.repo.repo,
  issue_number: prNumber,
  per_page: 100,
});
const existing = findExistingComment(allComments, COMMENT_MARKER);

if (valid) {
  // Pass path: update existing failure comment if present; silent on first-valid push
  if (existing) {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: existing.id,
      body: buildPassCommentBody(COMMENT_MARKER),
    });
  }
  core.info('PR changelog check passed');
} else {
  // Fail path: upsert failure comment then fail the workflow
  const failBody = buildFailureCommentBody(COMMENT_MARKER, errors);
  if (existing) {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: existing.id,
      body: failBody,
    });
  } else {
    await github.rest.issues.createComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: prNumber,
      body: failBody,
    });
  }
  core.setFailed(`PR changelog check failed:\n${errors.join('\n')}`);
}
