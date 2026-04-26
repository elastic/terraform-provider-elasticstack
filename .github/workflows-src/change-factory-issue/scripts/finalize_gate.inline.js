//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/change-factory-issue.gh.js

const result = computeGateReason(parseFinalizeGateEnv(process.env));

core.setOutput('gate_reason', result.gate_reason);
core.info(`Gate reason: ${result.gate_reason}`);
