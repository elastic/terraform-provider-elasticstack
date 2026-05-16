/**
 * Deterministic changelog renderer.
 *
 * Converts an array of merged PR metadata records (each with a parsed
 * `## Changelog` section) into a rendered changelog section body.
 *
 * Rendering rules:
 * - PRs with the `no-changelog` label are silently excluded.
 * - PRs with `Customer impact: none` produce no change bullet (but may
 *   still contribute breaking-change blocks).
 * - PRs that have neither the `no-changelog` label nor a parseable
 *   `## Changelog` section cause a hard assembly failure.
 * - Breaking-change blocks from included PRs are collected and placed
 *   under a `### Breaking changes` subsection.
 * - Regular change bullets are placed under `### Changes`.
 * - Normalization is minimal: standard `- ` bullet prefix, citation
 *   shape `([#NNN](url))`, and consistent whitespace. Author-provided
 *   summary text is preserved verbatim.
 */

//include: ./pr-changelog-parser.js
/* global parseChangelogSectionFull, validateChangelogSectionFull */

const NO_CHANGELOG_LABEL = 'no-changelog';

/**
 * Result from assembling PR changelog data.
 *
 * @typedef {Object} RenderResult
 * @property {boolean} success
 * @property {string|null} sectionBody - Rendered markdown body (without section header line), or null on failure.
 * @property {AssemblyError[]} errors - Assembly errors (non-empty only when success is false).
 * @property {IncludedPR[]} included - PRs that contributed at least one bullet or breaking-change block.
 * @property {ExcludedPR[]} excluded - PRs that were explicitly excluded (no-changelog or Customer impact: none).
 */

/**
 * @typedef {Object} AssemblyError
 * @property {number} prNumber
 * @property {string} prUrl
 * @property {string} reason
 */

/**
 * @typedef {Object} IncludedPR
 * @property {number} prNumber
 * @property {string} prUrl
 * @property {string} summary
 * @property {string|null} breakingChanges
 */

/**
 * @typedef {Object} ExcludedPR
 * @property {number} prNumber
 * @property {string} prUrl
 * @property {string} reason
 * @property {string|null} [breakingChanges] - Present when a `Customer impact: none` PR also had a ### Breaking changes block.
 */

/**
 * Normalize a bullet line: ensure it starts with `- ` (strip leading `-`, `*`, `+` and spaces).
 *
 * @param {string} line
 * @returns {string}
 */
function normalizeBulletPrefix(line) {
  return '- ' + line.replace(/^[-*+]\s*/, '').replace(/^\s+/, '');
}

/**
 * Build a citation string for a PR.
 *
 * @param {number} prNumber
 * @param {string} prUrl
 * @returns {string} e.g. `([#123](https://github.com/.../pull/123))`
 */
function buildCitation(prNumber, prUrl) {
  return `([#${prNumber}](${prUrl}))`;
}

/**
 * Build a normalized change bullet from a PR summary and citation.
 *
 * @param {string} summary - Author-provided summary text (preserved verbatim).
 * @param {number} prNumber
 * @param {string} prUrl
 * @returns {string}
 */
function buildChangeBullet(summary, prNumber, prUrl) {
  // Normalize the bullet prefix on the summary line, then append the citation.
  const normalizedSummary = normalizeBulletPrefix(summary.trim());
  const citation = buildCitation(prNumber, prUrl);
  return `${normalizedSummary} ${citation}`;
}

/**
 * Render a changelog section body from an array of merged PR records.
 *
 * Each PR record must have:
 *   - `number` {number}
 *   - `url` {string}
 *   - `labels` {string[]}
 *   - `body` {string|null}
 *
 * @param {Array<{number: number, url: string, labels: string[], body: string|null}>} mergedPRs
 * @returns {RenderResult}
 */
function renderChangelogSection(mergedPRs) {
  const errors = [];
  const included = [];
  const excluded = [];

  const changeBullets = [];
  const breakingChangeBlocks = [];

  for (const pr of mergedPRs) {
    const { number: prNumber, url: prUrl, labels, body } = pr;
    const labelNames = Array.isArray(labels) ? labels : [];

    // Exclude PRs with the no-changelog label
    if (labelNames.includes(NO_CHANGELOG_LABEL)) {
      excluded.push({ prNumber, prUrl, reason: 'no-changelog label' });
      continue;
    }

    // Parse and validate the changelog section from the PR body
    const parsed = parseChangelogSectionFull(body || '');

    if (parsed === null) {
      // No parseable ## Changelog section and no no-changelog label — hard fail
      errors.push({
        prNumber,
        prUrl,
        reason:
          `PR #${prNumber} (${prUrl}) has no parseable ## Changelog section and is not labeled 'no-changelog'. ` +
          'Add a ## Changelog section to the PR body or apply the no-changelog label.',
      });
      continue;
    }

    // Validate structural correctness (invalid customerImpact, empty breaking-changes heading, etc.)
    // Release-time rendering skips the breaking-impact match check (enforced at PR time only).
    const { valid, errors: validationErrors } = validateChangelogSectionFull(parsed, { enforceBreakingImpactMatch: false });
    if (!valid) {
      if (parsed.customerImpact === null) {
        errors.push({
          prNumber,
          prUrl,
          reason: `PR #${prNumber}: ## Changelog section is missing the required Customer impact field`,
        });
      } else {
        errors.push({
          prNumber,
          prUrl,
          reason: `PR #${prNumber}: ## Changelog section failed validation: ${validationErrors.join('; ')}`,
        });
      }
      continue;
    }

    const { customerImpact, summary, breakingChanges } = parsed;

    // Collect breaking-change block regardless of customerImpact
    if (breakingChanges) {
      breakingChangeBlocks.push({ prNumber, prUrl, breakingChanges });
    }

    // customerImpact === null case is already caught by validateChangelogSectionFull above
    if (customerImpact.trim().toLowerCase() === 'none') {
      const excludedEntry = { prNumber, prUrl, reason: 'Customer impact: none' };
      if (breakingChanges !== null) {
        excludedEntry.breakingChanges = breakingChanges;
      }
      excluded.push(excludedEntry);
      continue;
    }

    if (!summary) {
      // Missing summary for a non-none impact — treat as assembly error
      errors.push({
        prNumber,
        prUrl,
        reason:
          `PR #${prNumber} (${prUrl}) has Customer impact: ${customerImpact} but is missing the required Summary field.`,
      });
      continue;
    }

    const bullet = buildChangeBullet(summary, prNumber, prUrl);
    changeBullets.push(bullet);
    included.push({ prNumber, prUrl, summary, breakingChanges: breakingChanges || null });
  }

  if (errors.length > 0) {
    return { success: false, sectionBody: null, errors, included, excluded };
  }

  // Render the section body
  const sectionParts = [];

  if (breakingChangeBlocks.length > 0) {
    sectionParts.push('### Breaking changes');
    sectionParts.push('');
    for (const { breakingChanges } of breakingChangeBlocks) {
      // Preserve the author-provided breaking-change prose verbatim, trimming
      // only trailing whitespace from the block as a whole.
      sectionParts.push(breakingChanges.trimEnd());
      sectionParts.push('');
    }
  }

  if (changeBullets.length > 0) {
    sectionParts.push('### Changes');
    sectionParts.push('');
    for (const bullet of changeBullets) {
      sectionParts.push(bullet);
    }
    sectionParts.push('');
  }

  // Trim trailing blank lines
  while (sectionParts.length > 0 && sectionParts[sectionParts.length - 1] === '') {
    sectionParts.pop();
  }

  const sectionBody = sectionParts.length > 0 ? sectionParts.join('\n') : '';

  return { success: true, sectionBody, errors: [], included, excluded };
}

if (typeof module !== 'undefined') {
  module.exports = {
    NO_CHANGELOG_LABEL,
    buildChangeBullet,
    buildCitation,
    normalizeBulletPrefix,
    renderChangelogSection,
  };
}
