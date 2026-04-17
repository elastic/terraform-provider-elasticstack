/**
 * Evaluate whether the acceptance gate passed or failed.
 *
 * Only called when the preflight gate has allowed downstream CI to run
 * (i.e. `needs.preflight.outputs.should_run == 'true'`). The workflow
 * skips the `test-validation` job entirely when preflight outputs
 * `should_run=false`, so this function only models reachable
 * post-preflight validation states.
 *
 * @param {{ providerChanges: string, testResult: string }} params
 * @returns {{ passed: boolean, reason: string }}
 */
function validateTestResult({ providerChanges, testResult }) {
  if (providerChanges !== 'true' && providerChanges !== 'false') {
    return {
      passed: false,
      reason: `Change classification did not produce a valid output (got '${providerChanges}'). The changes job may have failed or been skipped unexpectedly.`,
    };
  }

  if (providerChanges === 'false') {
    if (testResult === 'skipped') {
      return {
        passed: true,
        reason: 'Change classification reports openspec-only change; test was skipped. Succeeding.',
      };
    }
    if (testResult === 'success') {
      return {
        passed: true,
        reason: 'Change classification reports openspec-only change; test ran and succeeded. Succeeding.',
      };
    }
    return {
      passed: false,
      reason: `Change classification reports openspec-only change but test result is '${testResult}'. Failing.`,
    };
  }

  if (testResult === 'success') {
    return {
      passed: true,
      reason: 'Provider changes detected and test succeeded. Succeeding.',
    };
  }

  return {
    passed: false,
    reason: `Provider changes detected but test result is '${testResult}'. Failing.`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { validateTestResult };
}
