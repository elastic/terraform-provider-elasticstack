const { getFactoryConstants, getFactoryModule } = require('./_factory-context.js');

module.exports = async function ({ github, context, core }) {

  const { computeGateReason, parseFinalizeGateEnv } = getFactoryModule();
  const constants = getFactoryConstants();

  const params = parseFinalizeGateEnv(process.env);
  const overrides = constants.DISABLE_DUPLICATE_GATE
    ? { duplicatePrFound: false, duplicatePrUrl: null, duplicateCheckGateReason: null }
    : {};

  const result = computeGateReason({ ...params, ...overrides });

  core.setOutput('gate_reason', result.gate_reason);
  core.info(`Gate reason: ${result.gate_reason}`);
};
