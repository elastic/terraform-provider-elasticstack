//include: ../../lib/gate-provider.js

const classifyResult = '${{ needs.classify.outputs.provider_changes }}';
const buildResult = '${{ needs.build.result }}';
const lintResult = '${{ needs.lint.result }}';
const testResult = '${{ needs.test.result }}';

const result = gateProvider({ classifyResult, buildResult, lintResult, testResult });

if (result.passed) {
  core.info(result.reason);
} else {
  core.setFailed(result.reason);
}
