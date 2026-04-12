//include: ../../lib/schema-coverage-slots.js

const { owner, repo } = context.repo;

// Count open schema-coverage issues, excluding pull requests
const issues = await github.paginate(github.rest.issues.listForRepo, {
  owner,
  repo,
  labels: SCHEMA_COVERAGE_LABEL,
  state: 'open',
  per_page: 100,
});

// GitHub issues API may return pull requests — exclude them
const openIssueCount = issues.filter(item => !item.pull_request).length;

const result = computeIssueSlots(openIssueCount);

core.setOutput('open_schema_coverage_issues', String(result.open_schema_coverage_issues));
core.setOutput('issue_slots_available', String(result.issue_slots_available));
core.setOutput('gate_reason', result.gate_reason);

core.info(`Gate reason: ${result.gate_reason}`);
