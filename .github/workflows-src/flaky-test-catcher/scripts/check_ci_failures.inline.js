//include: ../../lib/issue-slots.js
//include: ../../lib/flaky-test-catcher.js

const ISSUE_LABEL = 'flaky-test';
const ISSUE_CAP = 3;

const { owner, repo } = context.repo;

// Calculate the cutoff date for the 3-day window
const cutoffDate = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000);

// Query workflow runs for test.yml on main in the last 3 days.
// The `created` filter uses a date (not datetime) prefix so we may get some
// runs from just before the cutoff hour; that is acceptable.
const allRuns = await github.paginate(github.rest.actions.listWorkflowRuns, {
  owner,
  repo,
  workflow_id: 'test.yml',
  branch: 'main',
  status: 'completed',
  created: `>=${cutoffDate.toISOString().split('T')[0]}`,
  per_page: 100,
});

const { failedRunIds, totalRunCount } = classifyRuns(allRuns);

// Count open GitHub issues labelled `flaky-test`, excluding pull requests
const issues = await github.paginate(github.rest.issues.listForRepo, {
  owner,
  repo,
  labels: ISSUE_LABEL,
  state: 'open',
  per_page: 100,
});

const openIssueCount = filterIssues(issues).length;

const issueSlots = computeIssueSlots({
  label: ISSUE_LABEL,
  issueCap: ISSUE_CAP,
  openIssueCount,
});

const gate = computeGate(failedRunIds, issueSlots);

core.setOutput('has_ci_failures', gate.has_ci_failures);
core.setOutput('failed_run_ids', JSON.stringify(failedRunIds));
core.setOutput('total_run_count', String(totalRunCount));
core.setOutput('open_issues', String(issueSlots.open_issues));
core.setOutput('issue_slots_available', String(issueSlots.issue_slots_available));
core.setOutput('gate_reason', gate.gate_reason);

core.info(`Gate reason: ${gate.gate_reason}`);
core.info(`Failed run IDs: ${JSON.stringify(failedRunIds)}`);
core.info(`Total runs in window: ${totalRunCount}`);
