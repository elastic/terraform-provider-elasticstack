/**
 * Evaluate whether the workflows workflow gate passed or failed.
 *
 * @param {{ classifyResult: string, testResult: string }} params
 * @returns {{ passed: boolean, reason: string }}
 */
function gateWorkflows({ classifyResult, testResult }) {
  if (classifyResult !== 'true' && classifyResult !== 'false') {
    return {
      passed: false,
      reason: `Invalid classify result '${classifyResult}'. Expected 'true' or 'false'.`,
    };
  }

  const validResults = ['success', 'skipped', 'failure', 'cancelled'];
  if (!validResults.includes(testResult)) {
    return {
      passed: false,
      reason: `Invalid test result '${testResult}'. Expected one of: success, skipped, failure, cancelled.`,
    };
  }

  if (classifyResult === 'false' && testResult === 'skipped') {
    return {
      passed: true,
      reason: 'No workflow changes detected; test legitimately skipped. Gate passed.',
    };
  }

  if (testResult === 'success') {
    return {
      passed: true,
      reason: 'Workflow tests succeeded. Gate passed.',
    };
  }

  if (classifyResult === 'true' && testResult === 'skipped') {
    return {
      passed: false,
      reason: 'Unexpected skip: workflow changes detected but test was skipped. Gate failed.',
    };
  }

  if (testResult === 'failure' || testResult === 'cancelled') {
    return {
      passed: false,
      reason: `Workflow test ${testResult}. Gate failed.`,
    };
  }

  // Fallback for any other unexpected combination
  return {
    passed: false,
    reason: `Unexpected result combination (classify=${classifyResult}, test=${testResult}). Gate failed.`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { gateWorkflows };
}
