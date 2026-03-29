function verifyLabel(label, expectedLabel = 'verify-openspec') {
  if (label !== expectedLabel) {
    return {
      label_verified: 'false',
      label_reason: `Unexpected label: ${label || '(none)'}`,
      log_message: `Label check failed: expected ${expectedLabel}, got ${label || '(none)'}`,
    };
  }

  return {
    label_verified: 'true',
    label_reason: `Label verified: ${expectedLabel}`,
    log_message: `Label verified: ${expectedLabel}`,
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    verifyLabel,
  };
}
