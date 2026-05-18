const { gateProvider } = require('../lib/gate-provider.js');

module.exports = async function ({ github, context, core }) {
  const classifyResult = process.env.PROVIDER_GATE_CLASSIFY_RESULT ?? '';
  const buildResult = process.env.PROVIDER_GATE_BUILD_RESULT ?? '';
  const lintResult = process.env.PROVIDER_GATE_LINT_RESULT ?? '';
  const testResult = process.env.PROVIDER_GATE_TEST_RESULT ?? '';

  const result = gateProvider({ classifyResult, buildResult, lintResult, testResult });

  if (result.passed) {
    core.info(result.reason);
  } else {
    core.setFailed(result.reason);
  }
};
