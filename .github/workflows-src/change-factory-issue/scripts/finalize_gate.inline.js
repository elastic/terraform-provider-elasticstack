//include: ../../lib/change-factory-issue.js

const result = computeGateReason(parseFinalizeGateEnv(process.env));

core.setOutput('gate_reason', result.gate_reason);
core.info(`Gate reason: ${result.gate_reason}`);
