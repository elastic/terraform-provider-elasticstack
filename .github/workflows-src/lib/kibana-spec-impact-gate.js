const ISSUE_CAP = 5;

/**
 * Derives workflow gate outputs from a deterministic kibana-spec-impact report object.
 * @param {object} report - Parsed JSON from `go run ./scripts/kibana-spec-impact report`.
 * @returns {{ shouldRun: boolean, issueCap: number, kbapiChanged: boolean, transformHints: boolean, highConfidenceCount: number, gate_reason: string }}
 */
function kibanaSpecImpactGate(report) {
  const kbapiChanged = (report.changed_kbapi_symbols || []).length > 0;
  const transformHints = (report.transform_schema_hints || []).length > 0;
  const hi = report.high_confidence_impacts || [];
  const shouldRun = kbapiChanged || transformHints;
  const issueCap = Math.min(ISSUE_CAP, hi.length);
  const gateReason = `kbapi_changed=${kbapiChanged} transform_hints=${transformHints} high_confidence=${hi.length}`;
  return {
    shouldRun,
    issueCap,
    kbapiChanged,
    transformHints,
    highConfidenceCount: hi.length,
    gate_reason: gateReason,
  };
}

if (typeof module !== 'undefined') {
  module.exports = { ISSUE_CAP, kibanaSpecImpactGate };
}
