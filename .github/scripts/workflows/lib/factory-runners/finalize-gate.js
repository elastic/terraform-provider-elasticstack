const { getFactoryModule } = require('./_factory-context.js');

module.exports = async function ({ github, context, core }) {

  const { computeGateReason, parseFinalizeGateEnv } = getFactoryModule();

  const result = computeGateReason(parseFinalizeGateEnv(process.env));

  core.setOutput('gate_reason', result.gate_reason);
  core.info(`Gate reason: ${result.gate_reason}`);
};
