//include: ../../lib/pr-changelog-parser.js

const fs = require('fs');

const prNumber = process.env.PR_NUMBER || '';
const prBodyPath = '/tmp/pr-body.txt';
let prBody;
if (fs.existsSync(prBodyPath)) {
  prBody = fs.readFileSync(prBodyPath, 'utf8');
  core.info(`PR body read from ${prBodyPath} (${prBody.length} bytes)`);
} else {
  prBody = process.env.PR_BODY || '';
  core.info(`PR body read from PR_BODY env var (${prBody.length} bytes)`);
}

core.info(`Validating changelog section for PR #${prNumber}`);

const parsed = parseChangelogSectionFull(prBody);

if (parsed === null) {
  // No ## Changelog section found — agent will draft it
  core.info('No ## Changelog section found — agent will draft one');
  core.setOutput('changelog_present', 'false');
  core.setOutput('changelog_valid', 'false');
} else {
  // Section exists — validate it
  const validation = validateChangelogSectionFull(parsed);

  if (validation.valid) {
    core.info(`## Changelog section is valid (Customer impact: ${parsed.customerImpact})`);
    core.setOutput('changelog_present', 'true');
    core.setOutput('changelog_valid', 'true');
  } else {
    const errorList = validation.errors.map((e) => `  - ${e}`).join('\n');
    core.setOutput('changelog_valid', 'false');
    core.setFailed(
      `## Changelog section is malformed in PR #${prNumber}:\n${errorList}\n\nFix the ## Changelog section in the PR body and re-push to re-run this check.`
    );
  }
}
