const path = require('path');

const DEFAULT_EVIDENCE_ARTIFACT_NAME = 'changelog-release-evidence';
const DEFAULT_EVIDENCE_ARTIFACT_PATH = '/tmp/gh-aw/pre-activation/evidence.json';

function buildEvidenceArtifactPlan({
  manifest,
  artifactName = DEFAULT_EVIDENCE_ARTIFACT_NAME,
  artifactPath = DEFAULT_EVIDENCE_ARTIFACT_PATH,
}) {
  if (!manifest || typeof manifest !== 'object' || Array.isArray(manifest)) {
    throw new Error('manifest must be a non-null object');
  }

  if (!artifactName) {
    throw new Error('artifactName must be provided');
  }

  if (!artifactPath) {
    throw new Error('artifactPath must be provided');
  }

  return {
    artifactName,
    artifactPath,
    directory: path.dirname(artifactPath),
    formattedJson: JSON.stringify(manifest, null, 2),
    prCount: manifest.pr_count ?? '?',
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    DEFAULT_EVIDENCE_ARTIFACT_NAME,
    DEFAULT_EVIDENCE_ARTIFACT_PATH,
    buildEvidenceArtifactPlan,
  };
}
