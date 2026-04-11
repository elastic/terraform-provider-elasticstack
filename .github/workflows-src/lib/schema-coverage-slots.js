const SCHEMA_COVERAGE_LABEL = 'schema-coverage';
const ISSUE_CAP = 3;

/**
 * Computes schema-coverage issue slot availability.
 * @param {number} openIssueCount - The number of currently open schema-coverage issues.
 * @returns {{ open_schema_coverage_issues: number, issue_slots_available: number, gate_reason: string }}
 */
function computeIssueSlots(openIssueCount) {
  const slotsAvailable = Math.max(0, ISSUE_CAP - openIssueCount);

  let gateReason;
  if (slotsAvailable === 0) {
    gateReason = `Issue cap reached: ${openIssueCount} open schema-coverage issue(s), cap is ${ISSUE_CAP}. Agent job will be skipped.`;
  } else {
    gateReason = `${slotsAvailable} slot(s) available: ${openIssueCount} open schema-coverage issue(s), cap is ${ISSUE_CAP}.`;
  }

  return {
    open_schema_coverage_issues: openIssueCount,
    issue_slots_available: slotsAvailable,
    gate_reason: gateReason,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { SCHEMA_COVERAGE_LABEL, ISSUE_CAP, computeIssueSlots };
}
