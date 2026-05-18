const { gateWorkflows } = require('../lib/gate-workflows.js');

module.exports = async function ({ github, context, core }) {
  const classifyResult = '${{ needs.classify.outputs.workflow_changes }}';
  const testResult = '${{ needs.test.result }}';

  const result = gateWorkflows({ classifyResult, testResult });

  if (result.passed) {
    core.info(result.reason);
  } else {
    core.setFailed(result.reason);
  }
};
