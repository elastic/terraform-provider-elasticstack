'use strict';
/**
 * rewrite-changelog-section.js
 *
 * Rewrites the target section in CHANGELOG.md.
 *
 * In 'unreleased' mode:
 *   - Replaces the content of the ## [Unreleased] section with the new content.
 *   - Preserves all other sections exactly.
 *
 * In 'release' mode:
 *   - Replaces (or inserts after ## [Unreleased]) the ## [x.y.z] - YYYY-MM-DD section.
 *   - Preserves the ## [Unreleased] section exactly (does not modify it).
 *   - Preserves all other sections exactly.
 *   - Preserves the link footer exactly.
 *
 * Usage (CLI):
 *   node scripts/changelog-generation/rewrite-changelog-section.js \
 *     --mode unreleased|release \
 *     [--target-version 0.14.4] \
 *     [--section-content "### Changes\n\n- Foo (#123)"] \
 *     [--section-file path/to/section.md] \
 *     [--changelog CHANGELOG.md]
 *
 * Env vars (alternative to CLI args):
 *   MODE, TARGET_VERSION, SECTION_CONTENT, SECTION_FILE, CHANGELOG_PATH
 *
 * Outputs JSON: { success: boolean, changelogPath: string, section: string }
 *
 * Exports: rewriteUnreleased, rewriteRelease, rewriteSection (for unit testing)
 */

const fs = require('node:fs');
const path = require('node:path');

// ---------------------------------------------------------------------------
// Core parsing helpers
// ---------------------------------------------------------------------------

/**
 * Splits CHANGELOG.md content into logical blocks:
 *   - 'header':  text before the first ## header (may be empty)
 *   - 'section': { header: string, body: string } — one per ## section
 *   - 'footer':  text after the last section (link definitions etc.)
 *
 * @param {string} content
 * @returns {{ preamble: string, sections: Array<{ header: string, body: string }>, footer: string }}
 */
function parseChangelog(content) {
  const lines = content.split('\n');
  const sections = [];
  let preamble = '';
  let footer = '';

  let currentHeader = null;
  let currentBodyLines = [];
  let inSections = false;

  for (const line of lines) {
    if (line.startsWith('## ')) {
      if (currentHeader !== null) {
        sections.push({ header: currentHeader, body: currentBodyLines.join('\n') });
      } else if (!inSections) {
        // Everything before the first ## section is preamble
        preamble = currentBodyLines.join('\n');
      }
      currentHeader = line;
      currentBodyLines = [];
      inSections = true;
    } else if (inSections) {
      currentBodyLines.push(line);
    } else {
      currentBodyLines.push(line);
    }
  }

  if (currentHeader !== null) {
    sections.push({ header: currentHeader, body: currentBodyLines.join('\n') });
  } else {
    // No sections found at all
    preamble = currentBodyLines.join('\n');
  }

  // The footer is the trailing block of link definitions after the last section.
  // Link definitions look like: [label]: url
  // We extract them from the last section's body tail if they appear after a blank line gap.
  if (sections.length > 0) {
    const lastSection = sections[sections.length - 1];
    const bodyLines = lastSection.body.split('\n');

    // Find where the link footer starts: a line matching [label]: https://...
    let footerStart = -1;
    for (let i = bodyLines.length - 1; i >= 0; i--) {
      if (/^\[.+\]:\s*https?:\/\//.test(bodyLines[i])) {
        footerStart = i;
      } else if (bodyLines[i].trim() !== '') {
        break;
      }
    }

    if (footerStart !== -1) {
      footer = bodyLines.slice(footerStart).join('\n');
      lastSection.body = bodyLines.slice(0, footerStart).join('\n');
    }
  }

  return { preamble, sections, footer };
}

/**
 * Serialises a parsed changelog back to a string.
 *
 * @param {{ preamble: string, sections: Array<{ header: string, body: string }>, footer: string }} parsed
 * @returns {string}
 */
function serialiseChangelog({ preamble, sections, footer }) {
  const parts = [];

  if (preamble && preamble.trim()) {
    parts.push(preamble);
  }

  for (const section of sections) {
    parts.push(section.header);
    if (section.body !== undefined) {
      parts.push(section.body);
    }
  }

  if (footer && footer.trim()) {
    parts.push(footer);
  }

  return parts.join('\n');
}

/**
 * Normalises new section content:
 *   - Ensures it starts with a blank line after the header (will be injected separately).
 *   - Trims trailing blank lines but preserves a single trailing newline.
 *
 * @param {string} content — the body content (without the ## header line)
 * @returns {string}
 */
function normaliseSectionBody(content) {
  // Ensure the body starts with a single newline (header will be joined with '\n' by serialiseChangelog)
  const trimmed = content.replace(/^\n+/, '').trimEnd();
  return '\n' + trimmed + '\n';
}

// ---------------------------------------------------------------------------
// Rewrite operations
// ---------------------------------------------------------------------------

/**
 * Rewrites the ## [Unreleased] section with the given body content.
 *
 * @param {string} changelogContent — current CHANGELOG.md content
 * @param {string} newBody — new content for the section body (without the header line)
 * @returns {string} — updated CHANGELOG.md content
 */
function rewriteUnreleased(changelogContent, newBody) {
  const parsed = parseChangelog(changelogContent);

  const idx = parsed.sections.findIndex((s) => s.header === '## [Unreleased]');
  if (idx === -1) {
    throw new Error('## [Unreleased] section not found in CHANGELOG.md');
  }

  parsed.sections[idx].body = normaliseSectionBody(newBody);

  return serialiseChangelog(parsed);
}

/**
 * Rewrites the ## [x.y.z] - YYYY-MM-DD section for the given version,
 * inserting it after ## [Unreleased] if it doesn't exist yet.
 * The ## [Unreleased] section is left unchanged.
 *
 * @param {string} changelogContent — current CHANGELOG.md content
 * @param {string} targetVersion — e.g. '0.14.4'
 * @param {string} releaseDate — e.g. '2026-04-16'
 * @param {string} newBody — new content for the release section body (without the header line)
 * @returns {string} — updated CHANGELOG.md content
 */
function rewriteRelease(changelogContent, targetVersion, releaseDate, newBody) {
  const parsed = parseChangelog(changelogContent);

  const releaseHeader = `## [${targetVersion}] - ${releaseDate}`;
  const headerPrefix = `## [${targetVersion}]`;

  // Find existing release section (match by version prefix, ignoring date)
  let releaseIdx = parsed.sections.findIndex((s) => s.header.startsWith(headerPrefix));

  if (releaseIdx === -1) {
    // Insert after ## [Unreleased]
    const unreleasedIdx = parsed.sections.findIndex((s) => s.header === '## [Unreleased]');
    const insertAfter = unreleasedIdx !== -1 ? unreleasedIdx : -1;
    const newSection = { header: releaseHeader, body: normaliseSectionBody(newBody) };
    if (insertAfter !== -1) {
      parsed.sections.splice(insertAfter + 1, 0, newSection);
    } else {
      parsed.sections.unshift(newSection);
    }
  } else {
    parsed.sections[releaseIdx].header = releaseHeader;
    parsed.sections[releaseIdx].body = normaliseSectionBody(newBody);
  }

  return serialiseChangelog(parsed);
}

/**
 * Unified entry point: rewrites the appropriate section based on mode.
 *
 * @param {object} params
 * @param {string} params.changelogPath   — path to CHANGELOG.md
 * @param {string} params.mode            — 'unreleased' | 'release'
 * @param {string} [params.targetVersion] — required when mode === 'release'
 * @param {string} [params.releaseDate]   — ISO date string; defaults to today
 * @param {string} params.newSectionBody  — the new body content (without header)
 * @returns {{ success: boolean, changelogPath: string, section: string }}
 */
function rewriteSection({ changelogPath, mode, targetVersion, releaseDate, newSectionBody }) {
  const absPath = path.resolve(changelogPath);
  const changelogContent = fs.readFileSync(absPath, 'utf8');

  const date = releaseDate || new Date().toISOString().split('T')[0];

  let updated;
  let sectionHeader;

  if (mode === 'release') {
    if (!targetVersion) {
      throw new Error('targetVersion is required when mode === "release"');
    }
    updated = rewriteRelease(changelogContent, targetVersion, date, newSectionBody);
    sectionHeader = `## [${targetVersion}] - ${date}`;
  } else {
    updated = rewriteUnreleased(changelogContent, newSectionBody);
    sectionHeader = '## [Unreleased]';
  }

  fs.writeFileSync(absPath, updated, 'utf8');

  return { success: true, changelogPath: absPath, section: sectionHeader };
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

  const mode = argMap.mode || process.env.MODE || 'unreleased';
  const targetVersion = argMap.targetVersion || process.env.TARGET_VERSION || '';
  const changelogPath = argMap.changelog || process.env.CHANGELOG_PATH || 'CHANGELOG.md';
  const sectionFile = argMap.sectionFile || process.env.SECTION_FILE || '';

  let newSectionBody = argMap.sectionContent || process.env.SECTION_CONTENT || '';

  if (!newSectionBody && sectionFile) {
    try {
      newSectionBody = fs.readFileSync(sectionFile, 'utf8');
    } catch (err) {
      const result = { success: false, changelogPath, section: '', error: `Cannot read section file: ${err.message}` };
      process.stdout.write(JSON.stringify(result, null, 2) + '\n');
      process.exit(1);
    }
  }

  if (!newSectionBody) {
    const result = { success: false, changelogPath, section: '', error: 'No section content provided (use --section-content or --section-file)' };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }

  try {
    const result = rewriteSection({
      changelogPath,
      mode,
      targetVersion,
      newSectionBody,
    });
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(0);
  } catch (err) {
    const result = { success: false, changelogPath, section: '', error: err.message };
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
    process.exit(1);
  }
}

module.exports = { parseChangelog, serialiseChangelog, rewriteUnreleased, rewriteRelease, rewriteSection };
