const { gateWorkflows } = require('../lib/gate-workflows.js');

module.exports = async function ({ github, context, core }) {
  const classifyResult = process.env.WORKFLOWS_GATE_CLASSIFY_RESULT ?? '';
  const testResult = process.env.WORKFLOWS_GATE_TEST_RESULT ?? '';

  const result = gateWorkflows({ classifyResult, testResult });

  if (result.passed) {
    core.info(result.reason);
  } else {
    core.setFailed(result.reason);
  }
};
