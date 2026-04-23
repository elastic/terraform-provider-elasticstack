//include: ../../lib/code-factory-issue.js

const eventEligible = process.env.EVENT_ELIGIBLE === 'true';
const eventEligibleReason = process.env.EVENT_ELIGIBLE_REASON ?? '';
const actorTrustedRaw = process.env.ACTOR_TRUSTED;
const actorTrustedReason = process.env.ACTOR_TRUSTED_REASON ?? null;
const duplicatePrFoundRaw = process.env.DUPLICATE_PR_FOUND;
const duplicatePrUrl = process.env.DUPLICATE_PR_URL || null;
const noDuplicateReason = process.env.DUPLICATE_GATE_REASON ?? null;

const result = computeGateReason({
  eventEligible,
  eventEligibleReason,
  actorTrusted: actorTrustedRaw != null && actorTrustedRaw !== '' ? actorTrustedRaw === 'true' : null,
  actorTrustedReason,
  duplicatePrFound: duplicatePrFoundRaw != null && duplicatePrFoundRaw !== '' ? duplicatePrFoundRaw === 'true' : null,
  duplicatePrUrl,
  noDuplicateReason,
});

core.setOutput('gate_reason', result.gate_reason);
core.info(`Gate reason: ${result.gate_reason}`);
