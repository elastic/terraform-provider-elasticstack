const TRIGGER_LABEL = 'verify-openspec';

/**
 * Verifies that the event label matches the expected trigger label.
 * @param {string} labelName - The label name from the event payload.
 * @returns {{ label_verified: boolean, label_verified_reason: string }}
 */
function verifyTriggerLabel(labelName) {
  if (labelName === TRIGGER_LABEL) {
    return {
      label_verified: true,
      label_verified_reason: `Label matches trigger: ${TRIGGER_LABEL}`,
    };
  }
  return {
    label_verified: false,
    label_verified_reason: `Label does not match trigger (expected: ${TRIGGER_LABEL}, got: ${labelName || '(empty)'})`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { TRIGGER_LABEL, verifyTriggerLabel };
}
