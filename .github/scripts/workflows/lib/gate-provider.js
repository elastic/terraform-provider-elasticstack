/**
 * Evaluate whether the provider workflow gate passed or failed.
 *
 * @param {{ classifyResult: string, buildResult: string, lintResult: string, golangciLintResult: string, testResult: string, unitTestResult: string }} params
 * @returns {{ passed: boolean, reason: string }}
 */
function gateProvider({ classifyResult, buildResult, lintResult, golangciLintResult, testResult, unitTestResult }) {
  if (classifyResult !== 'true' && classifyResult !== 'false') {
    return {
      passed: false,
      reason: `Invalid classify result '${classifyResult}'. Expected 'true' or 'false'.`,
    };
  }

  const jobResults = [
    { name: 'build', result: buildResult },
    { name: 'lint', result: lintResult },
    { name: 'golangci-lint', result: golangciLintResult },
    { name: 'test', result: testResult },
    { name: 'unit-test', result: unitTestResult },
  ];
  const validResults = ['success', 'skipped', 'failure', 'cancelled'];

  for (const job of jobResults) {
    if (!validResults.includes(job.result)) {
      return {
        passed: false,
        reason: `Invalid job result '${job.result}'. Expected one of: success, skipped, failure, cancelled.`,
      };
    }
  }

  const describe = () => jobResults.map((j) => `${j.name}=${j.result}`).join(', ');

  const allSkipped = jobResults.every((j) => j.result === 'skipped');
  const allSuccess = jobResults.every((j) => j.result === 'success');
  const anyFailureOrCancelled = jobResults.some((j) => j.result === 'failure' || j.result === 'cancelled');

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
      reason: `One or more jobs failed or were cancelled (${describe()}). Gate failed.`,
    };
  }

  const anySkipped = jobResults.some((j) => j.result === 'skipped');
  if (classifyResult === 'true' && anySkipped) {
    return {
      passed: false,
      reason: `Unexpected skip: provider changes detected but one or more jobs were skipped (${describe()}). Gate failed.`,
    };
  }

  // Fallback for any other unexpected combination
  return {
    passed: false,
    reason: `Unexpected job result combination (${describe()}). Gate failed.`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { gateProvider };
}
