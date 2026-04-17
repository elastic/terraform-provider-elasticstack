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
// Read and parse the prompt at module scope
// ---------------------------------------------------------------------------

const rawPrompt = fs.readFileSync(WORKFLOW_TMPL_PATH, 'utf8');
const frontmatterEnd = rawPrompt.indexOf('\n---\n', 3);
const promptBody =
  frontmatterEnd !== -1
    ? rawPrompt.slice(frontmatterEnd + 5)
    : rawPrompt;

test('workflow.md.tmpl exists and is readable', () => {
  assert.ok(
    fs.existsSync(WORKFLOW_TMPL_PATH),
    `Expected workflow template at ${WORKFLOW_TMPL_PATH}`
  );
  assert.ok(rawPrompt.length > 0, 'workflow.md.tmpl must not be empty');
  assert.ok(promptBody.length > 0, 'prompt body must not be empty');
});

test('workflow passes release context into gather-pr-evidence via supported env vars', () => {
  assert.match(
    rawPrompt,
    /- name: Gather PR evidence\n\s+id: gather_pr_evidence\n\s+uses: actions\/github-script@v8\n\s+env:\n\s+PREVIOUS_TAG: \$\{\{ steps\.resolve_release_context\.outputs\.previous_tag \}\}\n\s+COMPARE_RANGE: \$\{\{ steps\.resolve_release_context\.outputs\.compare_range \}\}\n\s+MODE: \$\{\{ steps\.resolve_release_context\.outputs\.mode \}\}\n\s+TARGET_VERSION: \$\{\{ steps\.resolve_release_context\.outputs\.target_version \}\}/,
    'workflow should pass release context through env vars before gather-pr-evidence runs'
  );
  assert.doesNotMatch(
    rawPrompt,
    /with:\n(?:.*\n)*?\s+(previous_tag|compare_range|mode|target_version): \$\{\{ steps\.resolve_release_context\.outputs\.(previous_tag|compare_range|mode|target_version) \}\}/,
    'workflow should not rely on custom github-script inputs for gather-pr-evidence release context'
  );
});

test('workflow uploads and downloads evidence as an artifact instead of cross-job JSON', () => {
  assert.match(
    rawPrompt,
    /- name: Upload release evidence artifact\n\s+if: steps\.gather_pr_evidence\.outputs\.has_evidence == 'true'\n\s+uses: actions\/upload-artifact@v4\n\s+with:\n\s+name: changelog-release-evidence\n\s+path: \$\{\{ steps\.gather_pr_evidence\.outputs\.evidence_file_path \}\}\n\s+if-no-files-found: error/,
    'workflow should upload the gathered evidence file as an artifact'
  );
  assert.match(
    rawPrompt,
    /- name: Download release evidence artifact\n\s+uses: actions\/download-artifact@v4\n\s+with:\n\s+name: changelog-release-evidence\n\s+path: \/tmp\/gh-aw\/agent\n\s+- name: Verify evidence manifest path\n\s+run: test -f \/tmp\/gh-aw\/agent\/evidence\.json/,
    'workflow should download the artifact into the agent memory path and verify evidence.json exists'
  );
  assert.doesNotMatch(
    rawPrompt,
    /needs\.pre_activation\.outputs\.evidence_json|EVIDENCE_JSON|Write evidence manifest for agent/,
    'workflow should not transport the full manifest through evidence_json or a bridge step'
  );
});

test('workflow listens to pull_request_target for main', () => {
  assert.match(
    rawPrompt,
    /pull_request_target:\n\s+branches:\n\s+- main\n\s+types: \[opened, synchronize, reopened\]/,
    'workflow should listen to pull_request_target events targeting main'
  );
  assert.doesNotMatch(
    rawPrompt,
    /pull_request:\n\s+branches:\n\s+- main\n\s+types: \[opened, synchronize, reopened\]/,
    'workflow should not listen to pull_request events targeting main'
  );
});

test('workflow restricts pull_request_target to prep-release branches before agent activation', () => {
  assert.match(
    rawPrompt,
    /\(github\.event_name != 'pull_request_target' \|\|\n\s+startsWith\(github\.head_ref, 'prep-release-'\)\) &&/,
    'workflow should require prep-release-* for pull_request_target before agent activation'
  );
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

test('prompt references validate-provenance being called before CHANGELOG is written', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // Step 7 of the prompt must instruct calling validate-provenance (Go subcommand).
  const hasProvenanceValidationCall =
    /validate-provenance/i.test(promptBody);
  assert.ok(
    hasProvenanceValidationCall,
    'Prompt must reference validate-provenance for pre-write validation'
  );
});

test('prompt references rewrite-changelog-section for the CHANGELOG write', () => {
  assert.ok(promptBody, 'prompt body must be loaded');
  // Step 8 of the prompt must instruct calling rewrite-changelog-section (Go subcommand).
  const hasRewriterCall = /rewrite-changelog-section/i.test(promptBody);
  assert.ok(
    hasRewriterCall,
    'Prompt must reference rewrite-changelog-section for writing CHANGELOG.md'
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
