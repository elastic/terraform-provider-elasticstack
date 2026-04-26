/**
 * Qualifies a GitHub issues event for change-factory intake.
 * @param {{ eventName: string, eventAction: string, labelName: string, issueLabels: string[] }} params
 * @returns {{ event_eligible: boolean, event_eligible_reason: string }}
 */
function qualifyTriggerEvent({ eventName, eventAction, labelName, issueLabels }) {
  if (eventName !== 'issues') {
    return {
      event_eligible: false,
      event_eligible_reason: `Unsupported event '${eventName || '(empty)'}'; expected 'issues'.`,
    };
  }

  if (eventAction === 'labeled') {
    if (labelName === 'change-factory') {
      return {
        event_eligible: true,
        event_eligible_reason: 'Issue labeled event qualifies because the applied label is change-factory.',
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: `Issue labeled event does not qualify because the applied label is '${labelName || '(empty)'}', not 'change-factory'.`,
    };
  }

  if (eventAction === 'opened') {
    if (Array.isArray(issueLabels) && issueLabels.includes('change-factory')) {
      return {
        event_eligible: true,
        event_eligible_reason: 'Issue opened event qualifies because the issue already has the change-factory label.',
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: 'Issue opened event does not qualify because the issue was created without the change-factory label or issue labels were missing.',
    };
  }

  return {
    event_eligible: false,
    event_eligible_reason: `Issue event action '${eventAction || '(empty)'}' is not eligible; expected 'opened' or 'labeled'.`,
  };
}

/**
 * Checks whether the triggering actor is trusted for change-factory intake.
 * @param {{ sender: string, permission: string | null }} params
 * @returns {{ actor_trusted: boolean, actor_trusted_reason: string }}
 */
/**
 * Result when the workflow cannot read a sender login (mirrors check_actor_trust.inline.js).
 * @returns {{ actor_trusted: boolean, actor_trusted_reason: string }}
 */
function actorTrustWhenSenderMissing() {
  return {
    actor_trusted: false,
    actor_trusted_reason: 'Trigger actor could not be identified; sender login is missing from the event payload.',
  };
}

function checkActorTrust({ sender, permission }) {
  if (sender === 'github-actions[bot]') {
    return {
      actor_trusted: true,
      actor_trusted_reason: 'Trigger actor github-actions[bot] is trusted without collaborator permission lookup.',
    };
  }

  if (['write', 'maintain', 'admin'].includes(permission)) {
    return {
      actor_trusted: true,
      actor_trusted_reason: `Trigger actor '${sender || '(empty)'}' is trusted with repository permission '${permission}'.`,
    };
  }

  return {
    actor_trusted: false,
    actor_trusted_reason: `Trigger actor '${sender || '(empty)'}' is not trusted; repository permission '${permission || '(none)'}' does not meet the required write/maintain/admin policy.`,
  };
}

/** GitHub-recognized issue-closing keywords (case-insensitive). See https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests */
const GITHUB_ISSUE_CLOSING_KEYWORDS = '(?:close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)';

/**
 * @param {number} issueNumber
 * @returns {RegExp}
 */
function issueClosingReferencePattern(issueNumber) {
  return new RegExp(
    `\\b${GITHUB_ISSUE_CLOSING_KEYWORDS}\\s*#\\s*${issueNumber}(?![0-9])`,
    'i',
  );
}

/**
 * Checks for an existing open linked change-factory PR for the given issue.
 * @param {{ issueNumber: number, pullRequests: Array<{ number: number, state: string, head_branch: string, labels: string[], body: string, html_url: string }> }} params
 * @returns {{ duplicate_pr_found: boolean, duplicate_pr_url: string | null, gate_reason: string }}
 */
function checkDuplicatePR({ issueNumber, pullRequests }) {
  const expectedBranch = `change-factory/issue-${issueNumber}`;
  const closingExample = `Closes #${issueNumber}`;
  const closingPattern = issueClosingReferencePattern(issueNumber);
  const duplicate = (pullRequests || []).find(pr => (
    pr.state === 'open' &&
    Array.isArray(pr.labels) && pr.labels.includes('change-factory') &&
    pr.head_branch === expectedBranch &&
    closingPattern.test(String(pr.body || ''))
  ));

  if (duplicate) {
    const url = duplicate.html_url ?? null;
    return {
      duplicate_pr_found: true,
      duplicate_pr_url: url,
      gate_reason: `Found existing linked change-factory PR #${duplicate.number} (${url ?? '(unknown URL)'}) for issue #${issueNumber} on branch '${expectedBranch}' with issue-closing reference such as '${closingExample}'.`,
    };
  }

  return {
    duplicate_pr_found: false,
    duplicate_pr_url: null,
    gate_reason: `No open linked change-factory PR found for issue #${issueNumber}; expected label 'change-factory', branch '${expectedBranch}', and issue-closing reference such as '${closingExample}'.`,
  };
}

/**
 * Computes a consolidated gate reason from step outputs.
 * @param {{ eventEligible: boolean, eventEligibleReason: string, actorTrusted: boolean | null, actorTrustedReason: string | null, duplicatePrFound: boolean | null, duplicatePrUrl: string | null, noDuplicateReason: string | null }} params
 * @returns {{ gate_reason: string }}
 */
function computeGateReason({ eventEligible, eventEligibleReason, actorTrusted, actorTrustedReason, duplicatePrFound, duplicatePrUrl, noDuplicateReason }) {
  if (!eventEligible) {
    return { gate_reason: eventEligibleReason };
  }

  if (actorTrusted === false) {
    return { gate_reason: actorTrustedReason || 'Trigger actor is not trusted.' };
  }

  if (actorTrusted == null) {
    return { gate_reason: 'Actor trust could not be determined; the trust check step did not produce an output.' };
  }

  if (duplicatePrFound === true) {
    return {
      gate_reason: noDuplicateReason || `Found existing linked change-factory PR: ${duplicatePrUrl || '(unknown URL)'}.`,
    };
  }

  if (duplicatePrFound == null) {
    return { gate_reason: 'Duplicate PR check did not complete; the check step did not produce an output.' };
  }

  return {
    gate_reason: noDuplicateReason || 'All deterministic gates passed: event eligible, actor trusted, and no linked change-factory PR found.',
  };
}

/**
 * Parses optional tri-state booleans from workflow env (finalize_gate.inline.js).
 * @param {string | undefined} raw
 * @returns {boolean | null}
 */
function parseOptionalTriStateFromEnv(raw) {
  if (raw == null || raw === '') {
    return null;
  }
  return raw === 'true';
}

/**
 * @param {Record<string, string | undefined>} env
 */
function parseFinalizeGateEnv(env) {
  const e = env || {};
  return {
    eventEligible: e.EVENT_ELIGIBLE === 'true',
    eventEligibleReason: e.EVENT_ELIGIBLE_REASON ?? '',
    actorTrusted: parseOptionalTriStateFromEnv(e.ACTOR_TRUSTED),
    actorTrustedReason: e.ACTOR_TRUSTED_REASON ?? null,
    duplicatePrFound: parseOptionalTriStateFromEnv(e.DUPLICATE_PR_FOUND),
    duplicatePrUrl: e.DUPLICATE_PR_URL && e.DUPLICATE_PR_URL !== '' ? e.DUPLICATE_PR_URL : null,
    noDuplicateReason: e.DUPLICATE_GATE_REASON ?? null,
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    qualifyTriggerEvent,
    actorTrustWhenSenderMissing,
    checkActorTrust,
    checkDuplicatePR,
    computeGateReason,
    issueClosingReferencePattern,
    parseOptionalTriStateFromEnv,
    parseFinalizeGateEnv,
  };
}
