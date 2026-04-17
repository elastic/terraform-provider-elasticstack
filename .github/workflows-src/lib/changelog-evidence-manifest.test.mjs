import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  DEFAULT_EVIDENCE_ARTIFACT_NAME,
  DEFAULT_EVIDENCE_ARTIFACT_PATH,
  buildEvidenceArtifactPlan,
} = require('./changelog-evidence-manifest.js');

test('buildEvidenceArtifactPlan throws when manifest is missing', () => {
  assert.throws(
    () => buildEvidenceArtifactPlan({ manifest: null }),
    /manifest must be a non-null object/
  );
});

test('buildEvidenceArtifactPlan throws when manifest is not an object', () => {
  assert.throws(
    () => buildEvidenceArtifactPlan({ manifest: [] }),
    /manifest must be a non-null object/
  );
});

test('buildEvidenceArtifactPlan requires artifact metadata', () => {
  assert.throws(
    () => buildEvidenceArtifactPlan({ manifest: {}, artifactName: '' }),
    /artifactName must be provided/
  );
  assert.throws(
    () => buildEvidenceArtifactPlan({ manifest: {}, artifactPath: '' }),
    /artifactPath must be provided/
  );
});

test('buildEvidenceArtifactPlan formats pretty JSON and exposes artifact metadata', () => {
  const result = buildEvidenceArtifactPlan({
    manifest: {
      pr_count: 2,
      pull_requests: [],
    },
  });

  assert.deepEqual(result, {
    artifactName: DEFAULT_EVIDENCE_ARTIFACT_NAME,
    artifactPath: DEFAULT_EVIDENCE_ARTIFACT_PATH,
    directory: '/tmp/gh-aw/pre-activation',
    formattedJson: '{\n  "pr_count": 2,\n  "pull_requests": []\n}',
    prCount: 2,
  });
});
