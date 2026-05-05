/**
 * Helpers to parse and validate the PR-body `## Changelog` contract.
 *
 * Contract:
 *   ## Changelog
 *   Customer impact: <none|fix|enhancement|breaking>
 *   Summary: <text>          (required when Customer impact is not "none")
 *
 *   ### Breaking changes
 *   <free-form markdown>     (optional subsection)
 */

const VALID_CUSTOMER_IMPACTS = new Set(['none', 'fix', 'enhancement', 'breaking']);

/**
 * Extract the raw text of the `## Changelog` section from a PR body.
 * Returns the content between `## Changelog` and the next `##`-level heading
 * (or the end of the string).
 *
 * @param {string} body
 * @returns {string|null}
 */
function extractChangelogSection(body) {
  if (!body) return null;

  const lines = body.split('\n');
  let inChangelog = false;
  /** @type {null | '`' | '~'} */
  let fenceType = null;
  const content = [];

  for (const line of lines) {
    if (/^##\s+Changelog/.test(line)) {
      inChangelog = true;
      continue;
    }
    if (inChangelog) {
      if (fenceType === null && /^```/.test(line)) {
        fenceType = '`';
      } else if (fenceType === null && /^~~~/.test(line)) {
        fenceType = '~';
      } else if (fenceType === '`' && /^```/.test(line)) {
        fenceType = null;
      } else if (fenceType === '~' && /^~~~/.test(line)) {
        fenceType = null;
      }
      if (fenceType === null && /^##\s/.test(line)) {
        break;
      }
      content.push(line);
    }
  }

  if (!inChangelog) return null;

  return content.join('\n');
}

/**
 * Extract the raw markdown content of the `### Breaking changes` subsection
 * from a PR body string. Returns everything from the heading line (exclusive)
 * up to the next `##` or `###` heading, or the end of the string.
 *
 * @param {string} body
 * @returns {string|null} Trimmed markdown content, or null if not present or empty.
 */
function extractBreakingChanges(changelogSection) {
  if (!changelogSection) return null;

  const lines = changelogSection.split('\n');
  let inBreaking = false;
  /** @type {null | '`' | '~'} */
  let fenceType = null;
  const content = [];

  for (const line of lines) {
    if (/^###\s+Breaking changes/.test(line)) {
      inBreaking = true;
      continue;
    }
    if (inBreaking) {
      if (fenceType === null && /^```/.test(line)) {
        fenceType = '`';
      } else if (fenceType === null && /^~~~/.test(line)) {
        fenceType = '~';
      } else if (fenceType === '`' && /^```/.test(line)) {
        fenceType = null;
      } else if (fenceType === '~' && /^~~~/.test(line)) {
        fenceType = null;
      }
      if (fenceType === null && /^\s*<!--\s*\/breaking-changes\s*-->\s*$/.test(line)) {
        break;
      }
      if (fenceType === null && /^#{2,3}\s/.test(line)) {
        break;
      }
      content.push(line);
    }
  }

  if (!inBreaking) return null;

  const trimmed = content.join('\n').trimEnd();
  return trimmed.length > 0 ? trimmed : null;
}

/**
 * Parse the `## Changelog` section from a PR body string.
 *
 * @param {string} body - Full PR body text.
 * @returns {{ customerImpact: string|null, summary: string|null, breakingChanges: string|null }|null}
 *   Returns null when no `## Changelog` section is found.
 */
function parseChangelogSection(body) {
  const section = extractChangelogSection(body);
  if (section === null) return null;

  // Parse `Customer impact: <value>`
  const customerImpactMatch = section.match(/^Customer impact:\s*(.+)$/m);
  const customerImpact = customerImpactMatch ? customerImpactMatch[1].trim() : null;

  // Parse `Summary: <text>`
  const summaryMatch = section.match(/^Summary:\s*(.+)$/m);
  const summary = summaryMatch ? summaryMatch[1].trim() : null;

  // Extract breaking changes from the changelog section only
  const breakingChanges = extractBreakingChanges(section);

  return {
    customerImpact,
    summary: summary || null,
    breakingChanges,
  };
}

/**
 * Validate a parsed changelog section.
 *
 * @param {{ customerImpact: string|null, summary: string|null, breakingChanges: string|null }|null} parsed
 * @returns {{ valid: boolean, errors: string[] }}
 */
function validateChangelogSection(parsed) {
  const errors = [];

  if (parsed === null) {
    errors.push('No ## Changelog section found in PR body');
    return { valid: false, errors };
  }

  const { customerImpact, summary } = parsed;

  // Validate Customer impact
  if (!customerImpact) {
    errors.push('Missing required field: Customer impact');
  } else if (!VALID_CUSTOMER_IMPACTS.has(customerImpact)) {
    errors.push(
      `Invalid Customer impact value: "${customerImpact}". Must be one of: ${[...VALID_CUSTOMER_IMPACTS].join(', ')}`
    );
  }

  // Summary is required unless Customer impact is "none"
  if (customerImpact && customerImpact !== 'none' && VALID_CUSTOMER_IMPACTS.has(customerImpact)) {
    if (!summary) {
      errors.push('Missing required field: Summary (required when Customer impact is not "none")');
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Internal variant that also detects whether a `### Breaking changes` heading
 * was present but had no content.
 *
 * @param {string} body
 * @returns {{ customerImpact: string|null, summary: string|null, breakingChanges: string|null, breakingChangesHeadingPresent: boolean }|null}
 */
function parseChangelogSectionFull(body) {
  const section = extractChangelogSection(body);
  if (section === null) return null;

  const customerImpactMatch = section.match(/^Customer impact:\s*(.+)$/m);
  const customerImpact = customerImpactMatch ? customerImpactMatch[1].trim() : null;

  const summaryMatch = section.match(/^Summary:\s*(.+)$/m);
  const summary = summaryMatch ? summaryMatch[1].trim() : null;

  // Detect whether the heading exists within the changelog section (not the full body)
  const lines = section.split('\n');
  const breakingChangesHeadingPresent = lines.some((line) => /^###\s+Breaking changes/.test(line));
  const breakingChanges = extractBreakingChanges(section);

  return {
    customerImpact,
    summary: summary || null,
    breakingChanges,
    breakingChangesHeadingPresent,
  };
}

/**
 * Validate a full parsed changelog section (including empty breaking-changes detection).
 *
 * @param {ReturnType<typeof parseChangelogSectionFull>|null} parsed
 * @param {{ enforceBreakingImpactMatch?: boolean }} [options]
 * @returns {{ valid: boolean, errors: string[] }}
 */
function validateChangelogSectionFull(parsed, options = {}) {
  const base = validateChangelogSection(parsed);
  if (!parsed) return base;

  const { enforceBreakingImpactMatch = true } = options;
  const errors = [...base.errors];

  if (parsed.breakingChangesHeadingPresent && parsed.breakingChanges === null) {
    errors.push('### Breaking changes section is present but contains no content');
  }

  if (parsed.customerImpact === 'breaking' && !parsed.breakingChangesHeadingPresent) {
    errors.push('Customer impact: breaking requires a ### Breaking changes subsection');
  }

  if (
    enforceBreakingImpactMatch &&
    parsed.breakingChangesHeadingPresent &&
    parsed.customerImpact &&
    VALID_CUSTOMER_IMPACTS.has(parsed.customerImpact) &&
    parsed.customerImpact !== 'breaking'
  ) {
    errors.push('### Breaking changes section is only allowed when Customer impact: breaking; change to Customer impact: breaking or remove the ### Breaking changes heading.');
  }

  return { valid: errors.length === 0, errors };
}

if (typeof module !== 'undefined') {
  module.exports = {
    VALID_CUSTOMER_IMPACTS,
    extractBreakingChanges,
    extractChangelogSection,
    parseChangelogSection,
    parseChangelogSectionFull,
    validateChangelogSection,
    validateChangelogSectionFull,
  };
}
