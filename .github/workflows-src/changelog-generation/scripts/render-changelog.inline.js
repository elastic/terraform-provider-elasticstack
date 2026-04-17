const fs = require('fs');
//include: ../../lib/changelog-renderer.js

/**
 * Find the index of the next `##`-level section header after startIndex,
 * returning it as the exclusive end of the current section.
 * If no next section is found, returns the index of the last non-blank line + 1.
 *
 * @param {string[]} lines
 * @param {number} startIndex - Index of the current section header.
 * @returns {number}
 */
function findSectionEnd(lines, startIndex) {
  for (let i = startIndex + 1; i < lines.length; i++) {
    if (/^## /.test(lines[i])) {
      return i;
    }
  }
  // No next section — return end of file
  return lines.length;
}

/**
 * Rewrite only the target section of CHANGELOG.md.
 * Preserves all other sections and the link footer exactly.
 *
 * @param {string} content - Current CHANGELOG.md content.
 * @param {string} sectionHeader - The target section header line.
 * @param {string} newSectionContent - The full replacement (header + body).
 * @param {string} mode - 'unreleased' | 'release'
 * @param {string} targetVersion - Version string (release mode only).
 * @returns {string}
 */
function rewriteChangelogSection(content, sectionHeader, newSectionContent, mode, targetVersion) {
  const lines = content.split('\n');

  // Find the start of the target section
  let targetStart = -1;

  if (mode === 'unreleased') {
    targetStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
  } else {
    // For release mode, look for the exact version header or the Unreleased section
    // (we insert after Unreleased when no existing release section is found)
    targetStart = lines.findIndex((line) =>
      line.startsWith(`## [${targetVersion}]`)
    );
  }

  if (targetStart === -1) {
    // Section not found — insert appropriately
    if (mode === 'release') {
      // Insert after ## [Unreleased] section if present, otherwise at the top
      const unreleasedStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
      if (unreleasedStart !== -1) {
        // Find the end of the Unreleased section
        const insertAfter = findSectionEnd(lines, unreleasedStart);
        const before = lines.slice(0, insertAfter);
        const after = lines.slice(insertAfter);
        return [...before, '', newSectionContent, ...after].join('\n');
      } else {
        // No Unreleased section — prepend after any top-level heading
        return newSectionContent + '\n\n' + content;
      }
    } else {
      // unreleased: prepend after any top-level heading
      return newSectionContent + '\n\n' + content;
    }
  }

  // Find the end of the target section (start of the next ## section)
  const sectionEnd = findSectionEnd(lines, targetStart);

  const before = lines.slice(0, targetStart);
  const after = lines.slice(sectionEnd);

  // Remove trailing blank lines from 'before' that were padding before the old section
  while (before.length > 0 && before[before.length - 1] === '') {
    before.pop();
  }

  // Rebuild: before content, blank separator, new section, blank separator, after content
  const parts = [...before];
  if (parts.length > 0) parts.push('');
  parts.push(newSectionContent);

  // Normalize leading blank lines in 'after'
  let afterStart = 0;
  while (afterStart < after.length && after[afterStart] === '') {
    afterStart++;
  }

  if (afterStart < after.length) {
    parts.push('');
    parts.push(...after.slice(afterStart));
  }

  return parts.join('\n');
}

const mergedPRsPath = process.env.MERGED_PRS_PATH || '';
const mode = process.env.MODE || 'unreleased';
const targetVersion = process.env.TARGET_VERSION || '';
const changelogPath = process.env.CHANGELOG_PATH || 'CHANGELOG.md';

if (!mergedPRsPath) {
  core.setFailed('MERGED_PRS_PATH environment variable is required');
  process.exit(1);
}

if (mode === 'release' && !targetVersion) {
  core.setFailed('Release mode requires target_version from resolve_release_context');
  process.exit(1);
}

// Read the merged PR manifest
let manifest;
try {
  manifest = JSON.parse(fs.readFileSync(mergedPRsPath, 'utf8'));
} catch (err) {
  core.setFailed(`Failed to read merged PR manifest from ${mergedPRsPath}: ${err.message}`);
  process.exit(1);
}

const prRecords = manifest.pull_requests || [];
core.info(`Rendering changelog from ${prRecords.length} merged PR(s)`);

// Run deterministic rendering
const result = renderChangelogSection(prRecords);

if (!result.success) {
  // Hard fail: report all assembly errors clearly
  const errorMessages = result.errors.map((e) => `  - ${e.reason}`).join('\n');
  core.setFailed(
    `Changelog assembly failed. The following pull requests are missing a required ## Changelog section or Summary field:\n${errorMessages}\n\n` +
    'Each merged PR must either:\n' +
    "  1. Have a '## Changelog' section with 'Customer impact' and (when not 'none') a 'Summary' field, OR\n" +
    "  2. Be labeled 'no-changelog'"
  );
  process.exit(1);
}

// Log what was included/excluded
core.info(`Included ${result.included.length} PR(s) with change bullets or breaking changes`);
core.info(`Excluded ${result.excluded.length} PR(s) (no-changelog or Customer impact: none)`);
for (const ex of result.excluded) {
  core.info(`  Excluded PR #${ex.prNumber}: ${ex.reason}`);
}

// Read current CHANGELOG.md
let currentChangelog = '';
try {
  currentChangelog = fs.readFileSync(changelogPath, 'utf8');
} catch (err) {
  core.warning(`Could not read ${changelogPath}: ${err.message}. Will create a new file.`);
}

// Rewrite the target section in CHANGELOG.md
const today = new Date().toISOString().split('T')[0];
let sectionHeader;
if (mode === 'release' && targetVersion) {
  sectionHeader = `## [${targetVersion}] - ${today}`;
} else {
  sectionHeader = '## [Unreleased]';
}

const sectionBody = result.sectionBody;
const newSectionContent = sectionBody
  ? `${sectionHeader}\n\n${sectionBody}`
  : `${sectionHeader}`;

const updatedChangelog = rewriteChangelogSection(currentChangelog, sectionHeader, newSectionContent, mode, targetVersion);

// Write updated CHANGELOG.md
try {
  fs.writeFileSync(changelogPath, updatedChangelog, 'utf8');
  core.info(`CHANGELOG.md updated with section: ${sectionHeader}`);
} catch (err) {
  core.setFailed(`Failed to write ${changelogPath}: ${err.message}`);
  process.exit(1);
}

core.setOutput('section_header', sectionHeader);
core.setOutput('has_changes', result.included.length > 0 || result.excluded.length > 0 ? 'true' : 'false');
core.setOutput('has_user_facing_changes', result.included.length > 0 ? 'true' : 'false');
core.info(`Changelog section rendered: ${sectionHeader}`);
