//include: ../../lib/issue-slots.js

const ISSUE_LABEL = process.env.ISSUE_SLOTS_LABEL;
const ISSUE_CAP = process.env.ISSUE_SLOTS_CAP;
const { owner, repo } = context.repo;

// Count open issues for this workflow label, excluding pull requests
const issues = await github.paginate(github.rest.issues.listForRepo, {
  owner,
  repo,
  labels: ISSUE_LABEL,
  state: 'open',
  per_page: 100,
});

// GitHub issues API may return pull requests — exclude them
const openIssueCount = issues.filter(item => !item.pull_request).length;

const result = computeIssueSlots({
  label: ISSUE_LABEL,
  issueCap: ISSUE_CAP,
  openIssueCount,
});

core.setOutput('open_issues', String(result.open_issues));
core.setOutput('issue_slots_available', String(result.issue_slots_available));
core.setOutput('gate_reason', result.gate_reason);

core.info(`Gate reason: ${result.gate_reason}`);
