/**
 * Qualifies a GitHub issues event for code-factory intake.
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
    if (labelName === 'code-factory') {
      return {
        event_eligible: true,
        event_eligible_reason: 'Issue labeled event qualifies because the applied label is code-factory.',
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: `Issue labeled event does not qualify because the applied label is '${labelName || '(empty)'}', not 'code-factory'.`,
    };
  }

  if (eventAction === 'opened') {
    if (Array.isArray(issueLabels) && issueLabels.includes('code-factory')) {
      return {
        event_eligible: true,
        event_eligible_reason: 'Issue opened event qualifies because the issue already has the code-factory label.',
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: 'Issue opened event does not qualify because the issue was created without the code-factory label.',
    };
  }

  return {
    event_eligible: false,
    event_eligible_reason: `Issue event action '${eventAction || '(empty)'}' is not eligible; expected 'opened' or 'labeled'.`,
  };
}

/**
 * Checks whether the triggering actor is trusted for code-factory intake.
 * @param {{ sender: string, permission: string | null }} params
 * @returns {{ actor_trusted: boolean, actor_trusted_reason: string }}
 */
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

/**
 * Checks for an existing open linked code-factory PR for the given issue.
 * @param {{ issueNumber: number, pullRequests: Array<{ number: number, state: string, head_branch: string, labels: string[], body: string, html_url: string }> }} params
 * @returns {{ duplicate_pr_found: boolean, duplicate_pr_url: string | null, gate_reason: string }}
 */
function checkDuplicatePR({ issueNumber, pullRequests }) {
  const expectedBranch = `code-factory/issue-${issueNumber}`;
  const expectedLink = `Closes #${issueNumber}`;
  const closesPattern = new RegExp(`Closes #${issueNumber}(?![0-9])`);
  const duplicate = (pullRequests || []).find(pr => (
    pr.state === 'open' &&
    Array.isArray(pr.labels) && pr.labels.includes('code-factory') &&
    pr.head_branch === expectedBranch &&
    closesPattern.test(String(pr.body || ''))
  ));

  if (duplicate) {
    return {
      duplicate_pr_found: true,
      duplicate_pr_url: duplicate.html_url,
      gate_reason: `Found existing linked code-factory PR #${duplicate.number} (${duplicate.html_url}) for issue #${issueNumber} on branch '${expectedBranch}' with canonical linkage '${expectedLink}'.`,
    };
  }

  return {
    duplicate_pr_found: false,
    duplicate_pr_url: null,
    gate_reason: `No open linked code-factory PR found for issue #${issueNumber}; expected label 'code-factory', branch '${expectedBranch}', and canonical linkage '${expectedLink}'.`,
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
      gate_reason: noDuplicateReason || `Found existing linked code-factory PR: ${duplicatePrUrl || '(unknown URL)'}.`,
    };
  }

  if (duplicatePrFound == null) {
    return { gate_reason: 'Duplicate PR check did not complete; the check step did not produce an output.' };
  }

  return {
    gate_reason: noDuplicateReason || 'All deterministic gates passed: event eligible, actor trusted, and no linked code-factory PR found.',
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    qualifyTriggerEvent,
    checkActorTrust,
    checkDuplicatePR,
    computeGateReason,
  };
}
