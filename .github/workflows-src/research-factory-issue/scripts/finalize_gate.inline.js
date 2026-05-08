//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/factory-issue-module.gh.js

const params = parseFinalizeGateEnv(process.env);
const result = computeGateReason({
  ...params,
  duplicatePrFound: false,
  duplicatePrUrl: null,
  duplicateCheckGateReason: null,
});

core.setOutput('gate_reason', result.gate_reason);
core.info(`Gate reason: ${result.gate_reason}`);
