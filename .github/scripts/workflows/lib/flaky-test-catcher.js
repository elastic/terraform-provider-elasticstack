/**
 * Pure helper functions for the flaky-test-catcher workflow.
 * These are extracted from check_ci_failures.inline.js so they can be unit-tested.
 */

/**
 * Conclusions that represent meaningful completed runs for the purpose of
 * computing fail rates. Cancelled and skipped runs are excluded because they
 * don't reflect a real test outcome and would dilute the denominator.
 */
const COUNTABLE_CONCLUSIONS = new Set(['success', 'failure', 'timed_out', 'neutral', 'action_required']);

/**
 * Classifies an array of workflow run objects into failed runs and total count.
 *
 * Only runs with countable conclusions (success, failure, timed_out, neutral,
 * action_required) are included in `totalRunCount`. Cancelled and skipped runs
 * are omitted from the total to avoid understating fail rates.
 *
 * Runs are pre-filtered by date via the GitHub API `created` parameter, so
 * this function does not re-filter by date.
 *
 * @param {Array<{ id: number, conclusion: string | null }>} runs - Workflow run objects
 * @returns {{ failedRunIds: number[], totalRunCount: number }}
 */
function classifyRuns(runs) {
  const failedRunIds = [];
  let totalRunCount = 0;
  for (const run of runs) {
    if (COUNTABLE_CONCLUSIONS.has(run.conclusion)) {
      totalRunCount++;
    }
    if (run.conclusion === 'failure') {
      failedRunIds.push(run.id);
    }
  }
  return { failedRunIds, totalRunCount };
}

/**
 * Filters a list of GitHub issues/PR items, returning only real issues.
 * The GitHub issues API may return pull requests; this removes them.
 *
 * @param {Array<{ pull_request?: object }>} items - Items from listForRepo
 * @returns {Array<{ pull_request?: object }>}
 */
function filterIssues(items) {
  return items.filter(item => !item.pull_request);
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
  module.exports = { classifyRuns, computeGate, filterIssues };
}
