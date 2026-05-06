/**
 * Dispatch-mode helpers for code-factory issue intake.
 */
'use strict';

/**
 * @param {{ dispatchIssueNumber: string, currentRepository: string }} params
 * @returns {{ event_eligible: boolean, event_eligible_reason: string, issue_number?: number }}
 */
function validateDispatchInputs({ dispatchIssueNumber, currentRepository }) {
  const num = parseInt(dispatchIssueNumber, 10);
  if (
    !dispatchIssueNumber ||
    Number.isNaN(num) ||
    num <= 0 ||
    String(num) !== String(dispatchIssueNumber).trim()
  ) {
    return {
      event_eligible: false,
      event_eligible_reason: `Dispatch input issue_number '${dispatchIssueNumber || '(empty)'}' is not a valid positive integer.`,
    };
  }

  return {
    event_eligible: true,
    event_eligible_reason: `Dispatch intake validated: issue #${num} in repository '${currentRepository}'.`,
    issue_number: num,
  };
}

/**
 * @param {{ eventName: string, payload: { issue?: { number?: number, title?: string, body?: string } } }} params
 * @returns {{ issue_number: number | null, issue_title: string, issue_body: string }}
 */
function normalizeIssueEventContext({ eventName, payload }) {
  if (eventName !== 'issues') {
    return { issue_number: null, issue_title: '', issue_body: '' };
  }
  return {
    issue_number: payload.issue?.number ?? null,
    issue_title: payload.issue?.title ?? '',
    issue_body: payload.issue?.body ?? '',
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    validateDispatchInputs,
    normalizeIssueEventContext,
  };
}
