import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  DEFAULT_EVIDENCE_MEMORY_PATH,
  buildEvidenceManifestWrite,
  resolveEvidenceJsonInput,
} = require('./changelog-evidence-manifest.js');

test('resolveEvidenceJsonInput prefers EVIDENCE_JSON over other sources', () => {
  assert.equal(
    resolveEvidenceJsonInput({
      envEvidenceJson: '{"source":"env"}',
      coreInputEvidenceJson: '{"source":"core"}',
      envInputEvidenceJson: '{"source":"legacy"}',
    }),
    '{"source":"env"}'
  );
});

test('buildEvidenceManifestWrite throws when evidence JSON is missing', () => {
  assert.throws(
    () => buildEvidenceManifestWrite({ evidenceJson: '' }),
    /No evidence JSON provided via EVIDENCE_JSON, the evidence_json input, or INPUT_EVIDENCE_JSON/
  );
});

test('buildEvidenceManifestWrite throws when evidence_json is invalid JSON', () => {
  assert.throws(
    () => buildEvidenceManifestWrite({ evidenceJson: '{not-json}' }),
    /Invalid JSON in evidence_json/
  );
});

test('buildEvidenceManifestWrite throws when evidence_json parses to a non-object', () => {
  assert.throws(
    () => buildEvidenceManifestWrite({ evidenceJson: 'null' }),
    /evidence_json must parse to an object/
  );
  assert.throws(
    () => buildEvidenceManifestWrite({ evidenceJson: '[]' }),
    /evidence_json must parse to an object/
  );
});

test('buildEvidenceManifestWrite formats pretty JSON and exposes write metadata', () => {
  const result = buildEvidenceManifestWrite({
    evidenceJson: '{"pr_count":2,"pull_requests":[]}',
  });

  assert.deepEqual(result, {
    parsed: {
      pr_count: 2,
      pull_requests: [],
    },
    formattedJson: '{\n  "pr_count": 2,\n  "pull_requests": []\n}',
    memoryPath: DEFAULT_EVIDENCE_MEMORY_PATH,
    directory: '/tmp/gh-aw/agent',
    prCount: 2,
  });
});
