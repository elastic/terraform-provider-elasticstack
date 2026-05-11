//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/factory-issue-module.gh.js

const params = parseFinalizeGateEnv(process.env);
// research-factory does not create branches or PRs, so the duplicate-PR gate is intentionally
// disabled by passing duplicatePrFound: false. This deviates from sibling factories by design.
const result = computeGateReason({
  ...params,
  duplicatePrFound: false,
  duplicatePrUrl: null,
  duplicateCheckGateReason: null,
});

core.setOutput('gate_reason', result.gate_reason);
core.info(`Gate reason: ${result.gate_reason}`);
