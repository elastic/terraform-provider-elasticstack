//include: ../../lib/validate-test-result.js

const providerChanges = '${{ needs.changes.outputs.provider_changes }}';
const testResult = '${{ needs.test.result }}';

const result = validateTestResult({ providerChanges, testResult });

if (result.passed) {
  core.info(result.reason);
} else {
  core.setFailed(result.reason);
}
