//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/factory-issue-module.gh.js

const duplicatePrUrl = process.env.DUPLICATE_PR_URL;
const issueNumber = process.env.ISSUE_NUMBER;
const { owner, repo } = context.repo;

if (duplicatePrUrl && issueNumber) {
  const commentBody = `⚠️ **change-factory skipped** — PR #${extractPrNumber(duplicatePrUrl)} is already open for this issue.\nClose the existing PR, then retry.`;
  
  await github.rest.issues.createComment({
    owner,
    repo,
    issue_number: parseInt(issueNumber, 10),
    body: commentBody,
  });
  
  core.info(`Posted duplicate-blocked comment on issue #${issueNumber} referencing ${duplicatePrUrl}`);
} else {
  core.info('DUPLICATE_PR_URL is empty; skipping duplicate-blocked notification.');
}

/**
 * Extract the PR number from a GitHub PR URL.
 * @param {string} url
 * @returns {string}
 */
function extractPrNumber(url) {
  const match = url.match(/\/(\d+)(?:\/|$)/);
  return match ? match[1] : url;
}