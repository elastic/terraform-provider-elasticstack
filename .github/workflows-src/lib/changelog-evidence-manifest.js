const path = require('path');

const DEFAULT_EVIDENCE_MEMORY_PATH = '/tmp/gh-aw/agent/evidence.json';

function resolveEvidenceJsonInput({
  envEvidenceJson = '',
  coreInputEvidenceJson = '',
  envInputEvidenceJson = '',
}) {
  return envEvidenceJson || coreInputEvidenceJson || envInputEvidenceJson || '';
}

function buildEvidenceManifestWrite({
  evidenceJson,
  memoryPath = DEFAULT_EVIDENCE_MEMORY_PATH,
}) {
  if (!evidenceJson) {
    throw new Error(
      'No evidence JSON provided via EVIDENCE_JSON, the evidence_json input, or INPUT_EVIDENCE_JSON'
    );
  }

  let parsed;
  try {
    parsed = JSON.parse(evidenceJson);
  } catch (err) {
    throw new Error(`Invalid JSON in evidence_json: ${err.message}`);
  }

  return {
    parsed,
    formattedJson: JSON.stringify(parsed, null, 2),
    memoryPath,
    directory: path.dirname(memoryPath),
    prCount: parsed.pr_count ?? '?',
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    DEFAULT_EVIDENCE_MEMORY_PATH,
    buildEvidenceManifestWrite,
    resolveEvidenceJsonInput,
  };
}
