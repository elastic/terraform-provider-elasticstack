/**
 * Evaluate whether the provider workflow gate passed or failed.
 *
 * @param {{ classifyResult: string, buildResult: string, lintResult: string, testResult: string }} params
 * @returns {{ passed: boolean, reason: string }}
 */
function gateProvider({ classifyResult, buildResult, lintResult, testResult }) {
  if (classifyResult !== 'true' && classifyResult !== 'false') {
    return {
      passed: false,
      reason: `Invalid classify result '${classifyResult}'. Expected 'true' or 'false'.`,
    };
  }

  const jobResults = [buildResult, lintResult, testResult];
  const validResults = ['success', 'skipped', 'failure', 'cancelled'];

  for (const result of jobResults) {
    if (!validResults.includes(result)) {
      return {
        passed: false,
        reason: `Invalid job result '${result}'. Expected one of: success, skipped, failure, cancelled.`,
      };
    }
  }

  const allSkipped = jobResults.every((r) => r === 'skipped');
  const allSuccess = jobResults.every((r) => r === 'success');
  const anyFailureOrCancelled = jobResults.some((r) => r === 'failure' || r === 'cancelled');

  if (classifyResult === 'false' && allSkipped) {
    return {
      passed: true,
      reason: 'Non-provider changes detected; all jobs legitimately skipped. Gate passed.',
    };
  }

  if (allSuccess) {
    return {
      passed: true,
      reason: 'Provider changes detected; all jobs succeeded. Gate passed.',
    };
  }

  if (anyFailureOrCancelled) {
    return {
      passed: false,
      reason: `One or more jobs failed or were cancelled (build=${buildResult}, lint=${lintResult}, test=${testResult}). Gate failed.`,
    };
  }

  const anySkipped = jobResults.some((r) => r === 'skipped');
  if (classifyResult === 'true' && anySkipped) {
    return {
      passed: false,
      reason: `Unexpected skip: provider changes detected but one or more jobs were skipped (build=${buildResult}, lint=${lintResult}, test=${testResult}). Gate failed.`,
    };
  }

  // Fallback for any other unexpected combination
  return {
    passed: false,
    reason: `Unexpected job result combination (build=${buildResult}, lint=${lintResult}, test=${testResult}). Gate failed.`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { gateProvider };
}
