/**
 * Shared deterministic helpers for code-factory and change-factory issue intake workflows.
 * Workflow-specific configuration is passed via {@link createFactoryIssueIntake}.
 */

/** GitHub-recognized issue-closing keywords (case-insensitive). See https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests */
const GITHUB_ISSUE_CLOSING_KEYWORDS = '(?:close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)';

/**
 * GitHub closing-keyword reference: `#` is immediately followed by the issue digits (no whitespace),
 * per GitHub keyword syntax. Case-insensitive keywords; `(?![0-9])` avoids matching `#42` inside `#420`.
 *
 * @param {number} issueNumber
 * @returns {RegExp}
 */
function issueClosingReferencePattern(issueNumber) {
  return new RegExp(
    `\\b${GITHUB_ISSUE_CLOSING_KEYWORDS}\\s*#${issueNumber}(?![0-9])`,
    'i',
  );
}

/**
 * @param {{ eventName: string, eventAction: string, labelName: string, issueLabels: string[] | null | undefined, factoryLabel: string, issueOpenedNotEligibleReason: string }} params
 * @returns {{ event_eligible: boolean, event_eligible_reason: string }}
 */
function factoryQualifyTriggerEvent({
  eventName,
  eventAction,
  labelName,
  issueLabels,
  factoryLabel,
  issueOpenedNotEligibleReason,
}) {
  if (eventName !== 'issues') {
    return {
      event_eligible: false,
      event_eligible_reason: `Unsupported event '${eventName || '(empty)'}'; expected 'issues'.`,
    };
  }

  if (eventAction === 'labeled') {
    if (labelName === factoryLabel) {
      return {
        event_eligible: true,
        event_eligible_reason: `Issue labeled event qualifies because the applied label is ${factoryLabel}.`,
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: `Issue labeled event does not qualify because the applied label is '${labelName || '(empty)'}', not '${factoryLabel}'.`,
    };
  }

  if (eventAction === 'opened') {
    if (Array.isArray(issueLabels) && issueLabels.includes(factoryLabel)) {
      return {
        event_eligible: true,
        event_eligible_reason: `Issue opened event qualifies because the issue already has the ${factoryLabel} label.`,
      };
    }

    return {
      event_eligible: false,
      event_eligible_reason: issueOpenedNotEligibleReason,
    };
  }

  return {
    event_eligible: false,
    event_eligible_reason: `Issue event action '${eventAction || '(empty)'}' is not eligible; expected 'opened' or 'labeled'.`,
  };
}

/**
 * @param {{ sender: string, permission: string | null }} params
 * @returns {{ actor_trusted: boolean, actor_trusted_reason: string }}
 */
function factoryCheckActorTrust({ sender, permission }) {
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
 * @returns {{ actor_trusted: boolean, actor_trusted_reason: string }}
 */
function factoryActorTrustWhenSenderMissing() {
  return {
    actor_trusted: false,
    actor_trusted_reason: 'Trigger actor could not be identified; sender login is missing from the event payload.',
  };
}

/**
 * @param {{ issueNumber: number, pullRequests: Array<{ number: number, state: string, head_branch: string, labels: string[], body: string, html_url: string }>, branchPrefix: string, prLabel: string, duplicateLinkageMode: 'closes-literal' | 'github-keywords' }} params
 * @returns {{ duplicate_pr_found: boolean, duplicate_pr_url: string | null | undefined, gate_reason: string }}
 */
function factoryCheckDuplicatePR({
  issueNumber,
  pullRequests,
  branchPrefix,
  prLabel,
  duplicateLinkageMode,
}) {
  const expectedBranch = `${branchPrefix}${issueNumber}`;
  const expectedClosesExample = `Closes #${issueNumber}`;
  const bodyPattern = duplicateLinkageMode === 'closes-literal'
    ? new RegExp(`Closes #${issueNumber}(?![0-9])`)
    : issueClosingReferencePattern(issueNumber);

  const duplicate = (pullRequests || []).find(pr => (
    pr.state === 'open' &&
    Array.isArray(pr.labels) && pr.labels.includes(prLabel) &&
    pr.head_branch === expectedBranch &&
    bodyPattern.test(String(pr.body || ''))
  ));

  if (duplicate) {
    if (duplicateLinkageMode === 'closes-literal') {
      return {
        duplicate_pr_found: true,
        duplicate_pr_url: duplicate.html_url,
        gate_reason: `Found existing linked ${prLabel} PR #${duplicate.number} (${duplicate.html_url}) for issue #${issueNumber} on branch '${expectedBranch}' with canonical linkage '${expectedClosesExample}'.`,
      };
    }
    const url = duplicate.html_url ?? null;
    return {
      duplicate_pr_found: true,
      duplicate_pr_url: url,
      gate_reason: `Found existing linked ${prLabel} PR #${duplicate.number} (${url ?? '(unknown URL)'}) for issue #${issueNumber} on branch '${expectedBranch}' with issue-closing reference such as '${expectedClosesExample}'.`,
    };
  }

  const linkageTail = duplicateLinkageMode === 'closes-literal'
    ? `canonical linkage '${expectedClosesExample}'`
    : `issue-closing reference such as '${expectedClosesExample}'`;

  return {
    duplicate_pr_found: false,
    duplicate_pr_url: null,
    gate_reason: `No open linked ${prLabel} PR found for issue #${issueNumber}; expected label '${prLabel}', branch '${expectedBranch}', and ${linkageTail}.`,
  };
}

/**
 * @param {{ eventEligible: boolean, eventEligibleReason: string, actorTrusted: boolean | null, actorTrustedReason: string | null, duplicatePrFound: boolean | null, duplicatePrUrl: string | null, duplicateCheckGateReason: string | null }} params
 * @param {string} factoryLabel
 * @returns {{ gate_reason: string }}
 */
function factoryComputeGateReason({
  eventEligible,
  eventEligibleReason,
  actorTrusted,
  actorTrustedReason,
  duplicatePrFound,
  duplicatePrUrl,
  duplicateCheckGateReason,
}, factoryLabel) {
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
      gate_reason: duplicateCheckGateReason || `Found existing linked ${factoryLabel} PR: ${duplicatePrUrl || '(unknown URL)'}.`,
    };
  }

  if (duplicatePrFound == null) {
    return { gate_reason: 'Duplicate PR check did not complete; the check step did not produce an output.' };
  }

  return {
    gate_reason: duplicateCheckGateReason || `All deterministic gates passed: event eligible, actor trusted, and no linked ${factoryLabel} PR found.`,
  };
}

/**
 * @param {string | undefined} raw
 * @returns {boolean | null}
 */
function factoryParseOptionalTriStateFromEnv(raw) {
  if (raw == null || raw === '') {
    return null;
  }
  return raw === 'true';
}

/**
 * @param {Record<string, string | undefined>} env
 */
function factoryParseFinalizeGateEnv(env) {
  const e = env || {};
  return {
    eventEligible: e.EVENT_ELIGIBLE === 'true',
    eventEligibleReason: e.EVENT_ELIGIBLE_REASON ?? '',
    actorTrusted: factoryParseOptionalTriStateFromEnv(e.ACTOR_TRUSTED),
    actorTrustedReason: e.ACTOR_TRUSTED_REASON ?? null,
    duplicatePrFound: factoryParseOptionalTriStateFromEnv(e.DUPLICATE_PR_FOUND),
    duplicatePrUrl: e.DUPLICATE_PR_URL && e.DUPLICATE_PR_URL !== '' ? e.DUPLICATE_PR_URL : null,
    duplicateCheckGateReason: e.DUPLICATE_GATE_REASON ?? null,
  };
}

/**
 * @param {{
 *   branchPrefix: string,
 *   factoryLabel: string,
 *   issueOpenedNotEligibleReason: string,
 *   duplicateLinkageMode: 'closes-literal' | 'github-keywords',
 * }} config
 */
function createFactoryIssueIntake(config) {
  const {
    branchPrefix,
    factoryLabel,
    issueOpenedNotEligibleReason,
    duplicateLinkageMode,
  } = config;

  function issueBranchName(issueNumber) {
    return `${branchPrefix}${issueNumber}`;
  }

  /**
   * @param {{ eventName: string, eventAction: string, labelName: string, issueLabels?: string[] | null | undefined }} params
   */
  function qualifyTriggerEvent(params) {
    return factoryQualifyTriggerEvent({
      ...params,
      factoryLabel,
      issueOpenedNotEligibleReason,
    });
  }

  function checkActorTrust(params) {
    return factoryCheckActorTrust(params);
  }

  function checkDuplicatePR(params) {
    return factoryCheckDuplicatePR({
      ...params,
      branchPrefix,
      prLabel: factoryLabel,
      duplicateLinkageMode,
    });
  }

  function computeGateReason(params) {
    return factoryComputeGateReason(params, factoryLabel);
  }

  return {
    issueBranchName,
    qualifyTriggerEvent,
    checkActorTrust,
    checkDuplicatePR,
    computeGateReason,
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    issueClosingReferencePattern,
    factoryQualifyTriggerEvent,
    factoryCheckActorTrust,
    factoryActorTrustWhenSenderMissing,
    factoryCheckDuplicatePR,
    factoryComputeGateReason,
    factoryParseOptionalTriStateFromEnv,
    factoryParseFinalizeGateEnv,
    createFactoryIssueIntake,
  };
}
