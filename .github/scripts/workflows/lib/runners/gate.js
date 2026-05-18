'use strict';

const { gateProvider } = require('../gate-provider.js');
const { gateWorkflows } = require('../gate-workflows.js');

const GATES = {
  provider: {
    envPrefix: 'PROVIDER_GATE_',
    fields: ['CLASSIFY_RESULT', 'BUILD_RESULT', 'LINT_RESULT', 'TEST_RESULT'],
    evaluate: (env) => gateProvider({
      classifyResult: env.CLASSIFY_RESULT,
      buildResult: env.BUILD_RESULT,
      lintResult: env.LINT_RESULT,
      testResult: env.TEST_RESULT,
    }),
  },
  workflows: {
    envPrefix: 'WORKFLOWS_GATE_',
    fields: ['CLASSIFY_RESULT', 'TEST_RESULT'],
    evaluate: (env) => gateWorkflows({
      classifyResult: env.CLASSIFY_RESULT,
      testResult: env.TEST_RESULT,
    }),
  },
};

function readGateEnv(spec, processEnv) {
  const env = {};
  for (const field of spec.fields) {
    env[field] = processEnv[`${spec.envPrefix}${field}`] ?? '';
  }
  return env;
}

module.exports = async function ({ github, context, core }) {

  const gateName = process.env.GATE_NAME;
  if (!gateName) {
    core.setFailed('GATE_NAME environment variable is required.');
    return;
  }
  const spec = GATES[gateName];
  if (!spec) {
    core.setFailed(`Unknown GATE_NAME: ${gateName}. Expected one of: ${Object.keys(GATES).join(', ')}.`);
    return;
  }

  const env = readGateEnv(spec, process.env);
  const result = spec.evaluate(env);

  if (result.passed) {
    core.info(result.reason);
  } else {
    core.setFailed(result.reason);
  }
};
