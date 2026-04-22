/**
 * Evaluate whether the acceptance gate passed or failed.
 *
 * @param {{ preflightShouldRun: string, providerChanges: string, testResult: string }} params
 * @returns {{ passed: boolean, reason: string }}
 */
function validateTestResult({ preflightShouldRun, providerChanges, testResult }) {
  if (preflightShouldRun !== 'true') {
    return {
      passed: true,
      reason: 'Preflight gate intentionally disabled downstream CI; succeeding.',
    };
  }

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
