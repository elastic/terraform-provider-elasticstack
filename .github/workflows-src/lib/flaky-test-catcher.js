/**
 * Pure helper functions for the flaky-test-catcher workflow.
 * These are extracted from check_ci_failures.inline.js so they can be unit-tested.
 */

/**
 * Classifies an array of workflow run objects into failed runs and total count.
 *
 * Runs are pre-filtered by date via the GitHub API `created` parameter, so
 * this function does not re-filter by date. It filters only by conclusion.
 *
 * @param {Array<{ id: number, conclusion: string | null }>} runs - Workflow run objects
 * @returns {{ failedRunIds: number[], totalRunCount: number }}
 */
function classifyRuns(runs) {
  const failedRunIds = runs
    .filter(run => run.conclusion === 'failure')
    .map(run => run.id);

  return {
    failedRunIds,
    totalRunCount: runs.length,
  };
}

/**
 * Computes the gate decision for the flaky-test-catcher workflow.
 *
 * @param {number[]} failedRunIds - IDs of failed workflow runs
 * @param {{ open_issues: number, issue_slots_available: number, gate_reason: string }} issueSlots - Output from computeIssueSlots
 * @returns {{ has_ci_failures: string, gate_reason: string }}
 */
function computeGate(failedRunIds, issueSlots) {
  const hasCiFailures = failedRunIds.length > 0;

  if (!hasCiFailures) {
    return {
      has_ci_failures: 'false',
      gate_reason: `No CI failures detected on main in the last 3 days. Agent job will be skipped.`,
    };
  }

  if (issueSlots.issue_slots_available === 0) {
    return {
      has_ci_failures: 'true',
      gate_reason: `CI failures detected (${failedRunIds.length} failed run(s)), but ${issueSlots.gate_reason}`,
    };
  }

  return {
    has_ci_failures: 'true',
    gate_reason: `CI failures detected: ${failedRunIds.length} failed run(s) on main. ${issueSlots.gate_reason}`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { classifyRuns, computeGate };
}
