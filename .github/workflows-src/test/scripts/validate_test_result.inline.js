//include: ../../lib/validate-test-result.js

const preflightShouldRun = '${{ needs.preflight.outputs.should_run }}';
const providerChanges = '${{ needs.changes.outputs.provider_changes }}';
const testResult = '${{ needs.test.result }}';

const result = validateTestResult({ preflightShouldRun, providerChanges, testResult });

if (result.passed) {
  core.info(result.reason);
} else {
  core.setFailed(result.reason);
}
