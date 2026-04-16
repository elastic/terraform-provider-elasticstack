/**
 * validate-prompt-fixture.test.mjs
 *
 * Fixture test that validates the changelog generator agent prompt (workflow.md.tmpl)
 * contains the key structural requirements that ensure PR-based changelog output.
 *
 * This guards against accidental removal of commit-narration prohibitions,
 * PR citation requirements, and provenance/rewriter script references.
 */

import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Path to the workflow template from the scripts/changelog-generation directory
const WORKFLOW_TMPL_PATH = path.resolve(
  __dirname,
  '../../.github/workflows-src/changelog-generation/workflow.md.tmpl'
);

// ---------------------------------------------------------------------------
// Read and parse the prompt
// ---------------------------------------------------------------------------

let promptBody;

test('workflow.md.tmpl exists and is readable', () => {
  assert.ok(
    fs.existsSync(WORKFLOW_TMPL_PATH),
    `Expected workflow template at ${WORKFLOW_TMPL_PATH}`
  );
  const raw = fs.readFileSync(WORKFLOW_TMPL_PATH, 'utf8');
  assert.ok(raw.length > 0, 'workflow.md.tmpl must not be empty');

  // Strip YAML frontmatter (between --- delimiters at the start)
  const frontmatterEnd = raw.indexOf('\n---\n', 3);
  if (frontmatterEnd !== -1) {
    promptBody = raw.slice(frontmatterEnd + 5); // skip '\n---\n'
  } else {
    // No frontmatter found — treat entire file as prompt body
    promptBody = raw;
  }
});

// ---------------------------------------------------------------------------
// Structural checks on the prompt body
// ---------------------------------------------------------------------------

test('prompt prohibits commit-level narration', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // The prompt must contain a clear prohibition on narrating individual commits.
  // Accept variations: "do not narrate individual commits", "not the commit level",
  // "commit-level narration", "not narrate commits", etc.
  const hasCommitNarrationProhibition =
    /do not narrate individual commits/i.test(promptBody) ||
    /commit.level narration/i.test(promptBody) ||
    /not.*commit level/i.test(promptBody) ||
    /strictly.*pull.request level/i.test(promptBody);

  assert.ok(
    hasCommitNarrationProhibition,
    'Prompt must explicitly prohibit commit-level narration'
  );
});

test('prompt requires PR citation format with #NNN', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // The prompt must require bullets to end with ([#NNN](url)) style citation.
  const hasPRCitationRequirement =
    /\(\[#NNN\]/i.test(promptBody) ||
    /#NNN/i.test(promptBody) ||
    /\[#\\d\+\]/i.test(promptBody) ||
    /\(\[#<number>\]/i.test(promptBody);

  assert.ok(
    hasPRCitationRequirement,
    'Prompt must require PR citation format (#NNN or ([#NNN](url)))'
  );
});

test('prompt references validate-provenance.js being called before CHANGELOG is written', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // Step 7 of the prompt must instruct calling validate-provenance.js.
  const hasProvenanceValidationCall = /validate-provenance\.js/i.test(promptBody);
  assert.ok(
    hasProvenanceValidationCall,
    'Prompt must reference validate-provenance.js for pre-write validation'
  );
});

test('prompt references rewrite-changelog-section.js for the CHANGELOG write', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // Step 8 of the prompt must instruct calling rewrite-changelog-section.js.
  const hasRewriterCall = /rewrite-changelog-section\.js/i.test(promptBody);
  assert.ok(
    hasRewriterCall,
    'Prompt must reference rewrite-changelog-section.js for writing CHANGELOG.md'
  );
});

test('prompt references provenance.json for structured provenance output', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // Step 6 of the prompt must require writing provenance.json.
  const hasProvenanceJson = /provenance\.json/i.test(promptBody);
  assert.ok(
    hasProvenanceJson,
    'Prompt must reference provenance.json for machine-readable bullet-to-PR mappings'
  );
});

test('prompt requires provenance validation to pass before CHANGELOG is written', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // The prompt must state that validation must pass before writing CHANGELOG.md.
  const hasValidationGate =
    /validation.*pass.*before.*CHANGELOG/i.test(promptBody) ||
    /passes.*proceed to write.*CHANGELOG/i.test(promptBody) ||
    /pass.*before.*CHANGELOG\.md.*is written/i.test(promptBody) ||
    /MUST pass before CHANGELOG/i.test(promptBody);

  assert.ok(
    hasValidationGate,
    'Prompt must require provenance validation to pass before CHANGELOG.md is written'
  );
});

test('prompt requires every bullet to have a PR backing from the evidence manifest', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // The prompt must require that every bullet comes from the evidence manifest.
  const hasEvidenceRequirement =
    /evidence manifest/i.test(promptBody) &&
    /every.*bullet/i.test(promptBody);

  assert.ok(
    hasEvidenceRequirement,
    'Prompt must require every changelog bullet to be backed by the evidence manifest'
  );
});
