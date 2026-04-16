'use strict';
/**
 * validate-provenance.js
 *
 * Validates the agent's changelog markdown and provenance JSON against the
 * evidence manifest to ensure every changelog bullet is backed by a real PR
 * present in the evidence.
 *
 * Usage (CLI):
 *   node scripts/changelog-generation/validate-provenance.js \
 *     --evidence  memory/changelog-generation/evidence.json \
 *     --provenance memory/changelog-generation/provenance.json \
 *     --changelog  CHANGELOG.md \
 *     [--section-header "## [Unreleased]"|"## [0.14.4] - 2026-04-16"]
 *
 * Env vars (alternative to CLI args):
 *   EVIDENCE_PATH, PROVENANCE_PATH, CHANGELOG_PATH, SECTION_HEADER
 *
 * Exits 0 on success, 1 on validation failure.
 * Outputs a JSON result: { valid: boolean, errors: string[], warnings: string[] }
 *
 * Exports: validateProvenance (for unit testing)
 */

const fs = require('node:fs');
const path = require('node:path');

// ---------------------------------------------------------------------------
// Heuristics
// ---------------------------------------------------------------------------

/** Regex that matches a 40-character hex SHA (commit hash). */
const COMMIT_SHA_RE = /\b[0-9a-f]{40}\b/i;

/** Regex that matches a short (7-12 char) hex SHA. */
const SHORT_SHA_RE = /\b[0-9a-f]{7,12}\b/i;

/**
 * Extracts all #NNN PR number references from a block of markdown text.
 * Only matches references of the form #NNN (not inside URLs).
 *
 * @param {string} text
 * @returns {number[]}
 */
function extractPRReferences(text) {
  // Match #NNN not preceded by / (which would be a URL path)
  const matches = [...text.matchAll(/(?<!\/)#(\d+)/g)];
  return [...new Set(matches.map((m) => parseInt(m[1], 10)))];
}

/**
 * Extracts bullet lines from a changelog section block (the part between
 * the section header and the next ## header).
 *
 * @param {string} sectionContent — text of the section (may include header)
 * @returns {string[]}
 */
function extractBulletLines(sectionContent) {
  return sectionContent
    .split('\n')
    .filter((line) => /^\s*[*-]\s/.test(line));
}

/**
 * Detects whether a line looks like commit-level narration:
 *   - Contains a full 40-char SHA
 *   - Is a short generic sentence without any #NNN PR reference and
 *     contains a 7-12 char hex string
 *
 * @param {string} line
 * @returns {boolean}
 */
function looksLikeCommitNarration(line) {
  if (COMMIT_SHA_RE.test(line)) return true;
  // Short hex SHA without PR reference is suspicious
  if (SHORT_SHA_RE.test(line) && !/#\d+/.test(line)) return true;
  return false;
}

// ---------------------------------------------------------------------------
// Core validation
// ---------------------------------------------------------------------------

/**
 * Validates provenance JSON and changelog markdown against the evidence manifest.
 *
 * @param {object} params
 * @param {object} params.evidence   — parsed evidence manifest JSON
 * @param {object} params.provenance — parsed provenance JSON
 * @param {string} params.changelogSection — the raw markdown of the target section
 *   (including the section header line, e.g. "## [Unreleased]\n...")
 * @param {string} [params.expectedHeader] — e.g. "## [Unreleased]" or "## [0.14.4] - 2026-04-16"
 *   If omitted, derived from evidence.target_section.
 * @returns {{ valid: boolean, errors: string[], warnings: string[] }}
 */
function validateProvenance({ evidence, provenance, changelogSection, expectedHeader }) {
  const errors = [];
  const warnings = [];

  // Build a set of known PR numbers from the evidence manifest
  const evidencePRNumbers = new Set(
    (evidence.pull_requests ?? []).map((pr) => pr.number)
  );

  // --- Check 1: provenance bullets map to known PRs ---
  const bullets = provenance.bullets ?? [];
  for (const bullet of bullets) {
    const prNumbers = bullet.pr_numbers ?? [];
    if (prNumbers.length === 0) {
      errors.push(
        `Provenance bullet has no pr_numbers: "${bullet.text ?? '(no text)'}"`
      );
    }
    for (const prNum of prNumbers) {
      if (!evidencePRNumbers.has(prNum)) {
        errors.push(
          `Provenance bullet references PR #${prNum} which is NOT in the evidence manifest: "${bullet.text ?? ''}"`
        );
      }
    }
  }

  // --- Check 2: every #NNN in the changelog markdown is backed by evidence ---
  const changelogPRRefs = extractPRReferences(changelogSection);
  for (const prNum of changelogPRRefs) {
    if (!evidencePRNumbers.has(prNum)) {
      errors.push(
        `Changelog references PR #${prNum} which is NOT in the evidence manifest (fabricated reference?)`
      );
    }
  }

  // --- Check 3: no commit-level narration ---
  const bulletLines = extractBulletLines(changelogSection);
  for (const line of bulletLines) {
    if (looksLikeCommitNarration(line)) {
      errors.push(
        `Changelog bullet appears to contain commit-level narration (SHA found): "${line.trim()}"`
      );
    }
    // Warn about bullets without any PR reference
    if (!/#\d+/.test(line)) {
      warnings.push(
        `Changelog bullet has no PR reference (#NNN): "${line.trim()}"`
      );
    }
  }

  // --- Check 4: section header matches expected ---
  const resolvedExpectedHeader = expectedHeader ?? evidence.target_section ?? '';
  if (resolvedExpectedHeader) {
    const lines = changelogSection.split('\n');
    const firstNonEmpty = lines.find((l) => l.trim() !== '');
    if (firstNonEmpty && !firstNonEmpty.startsWith(resolvedExpectedHeader)) {
      errors.push(
        `Changelog section header "${firstNonEmpty.trim()}" does not match expected "${resolvedExpectedHeader}"`
      );
    }
  }

  // --- Check 5: markdown format — bullets use "- " or "* " format ---
  for (const line of bulletLines) {
    if (!/^\s*[-*]\s/.test(line)) {
      warnings.push(`Unexpected bullet format (expected "- " or "* "): "${line.trim()}"`);
    }
  }

  const valid = errors.length === 0;
  return { valid, errors, warnings };
}

/**
 * Reads and parses a JSON file. Returns the parsed object.
 *
 * @param {string} filePath
 * @returns {object}
 */
function readJSON(filePath) {
  const raw = fs.readFileSync(filePath, 'utf8');
  return JSON.parse(raw);
}

/**
 * Extracts a named section from CHANGELOG.md content.
 * Returns the content from the section header up to (but not including) the next ## header.
 *
 * @param {string} changelogContent
 * @param {string} header — e.g. "## [Unreleased]" or "## [0.14.4]"
 * @returns {string|null}
 */
function extractSectionFromChangelog(changelogContent, header) {
  const lines = changelogContent.split('\n');
  let inSection = false;
  const sectionLines = [];

  for (const line of lines) {
    if (!inSection) {
      if (line.startsWith(header)) {
        inSection = true;
        sectionLines.push(line);
      }
    } else {
      // Stop at the next ## header (but not the same one)
      if (line.startsWith('## ') && !line.startsWith(header)) {
        break;
      }
      sectionLines.push(line);
    }
  }

  return inSection ? sectionLines.join('\n') : null;
}

// ---------------------------------------------------------------------------
// CLI entry point
// ---------------------------------------------------------------------------

if (require.main === module) {
  // Parse CLI args
  const args = process.argv.slice(2);
  const argMap = {};
  for (let i = 0; i < args.length; i += 2) {
    const key = args[i].replace(/^--/, '').replace(/-([a-z])/g, (_, c) => c.toUpperCase());
    argMap[key] = args[i + 1];
  }

  const evidencePath =
    argMap.evidence ||
    process.env.EVIDENCE_PATH ||
    'memory/changelog-generation/evidence.json';

  const provenancePath =
    argMap.provenance ||
    process.env.PROVENANCE_PATH ||
    'memory/changelog-generation/provenance.json';

  const changelogPath =
    argMap.changelog || process.env.CHANGELOG_PATH || 'CHANGELOG.md';

  const sectionHeaderArg =
    argMap.sectionHeader || process.env.SECTION_HEADER || '';

  let evidence, provenance, changelogContent;

  try {
    evidence = readJSON(evidencePath);
  } catch (err) {
    const result = { valid: false, errors: [`Cannot read evidence manifest: ${err.message}`], warnings: [] };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }

  try {
    provenance = readJSON(provenancePath);
  } catch (err) {
    const result = { valid: false, errors: [`Cannot read provenance file: ${err.message}`], warnings: [] };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }

  try {
    changelogContent = fs.readFileSync(changelogPath, 'utf8');
  } catch (err) {
    const result = { valid: false, errors: [`Cannot read CHANGELOG.md: ${err.message}`], warnings: [] };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }

  const targetHeader = sectionHeaderArg || evidence.target_section || '## [Unreleased]';
  // Extract just the section heading prefix for matching (strip date part for lookup)
  const headerPrefix = targetHeader.split(' - ')[0];
  const changelogSection = extractSectionFromChangelog(changelogContent, headerPrefix);

  if (!changelogSection) {
    const result = {
      valid: false,
      errors: [`Section "${targetHeader}" not found in ${changelogPath}`],
      warnings: [],
    };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }

  const result = validateProvenance({
    evidence,
    provenance,
    changelogSection,
    expectedHeader: targetHeader,
  });

  process.stdout.write(JSON.stringify(result, null, 2) + '\n');

  if (!result.valid) {
    process.stderr.write(`Validation failed: ${result.errors.length} error(s)\n`);
    process.exit(1);
  }

  process.stderr.write(`Validation passed${result.warnings.length > 0 ? ` (${result.warnings.length} warning(s))` : ''}\n`);
  process.exit(0);
}

module.exports = {
  classifyPR: undefined, // re-exported for convenience from build-evidence-manifest
  validateProvenance,
  extractPRReferences,
  extractBulletLines,
  looksLikeCommitNarration,
  extractSectionFromChangelog,
  readJSON,
};
