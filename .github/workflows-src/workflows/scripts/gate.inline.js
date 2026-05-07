//include: ../../lib/gate-workflows.js

const classifyResult = '${{ needs.classify.outputs.workflow_changes }}';
const testResult = '${{ needs.test.result }}';

const result = gateWorkflows({ classifyResult, testResult });

if (result.passed) {
  core.info(result.reason);
} else {
  core.setFailed(result.reason);
}
