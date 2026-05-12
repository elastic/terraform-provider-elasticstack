/**
 * Rewrite a single section in CHANGELOG.md without affecting other sections.
 * Extracted from the changelog-generation inline render step for reuse and tests.
 */

/**
 * @param {string[]} lines
 * @param {number} startIndex
 * @returns {number}
 */
function findSectionEnd(lines, startIndex) {
  for (let i = startIndex + 1; i < lines.length; i++) {
    if (/^## /.test(lines[i])) {
      return i;
    }
  }
  return lines.length;
}

/**
 * @param {string[]} lines
 * @param {{ start: number, end: number }[]} ranges - sorted ascending by start;
 *    each end is exclusive. First range replaced by `newSectionContent`; others dropped.
 * @param {string} newSectionContent
 * @returns {string}
 */
function spliceReleaseSectionRanges(lines, ranges, newSectionContent) {
  const first = ranges[0];

  let before = lines.slice(0, first.start);
  while (before.length > 0 && before[before.length - 1] === '') {
    before.pop();
  }

  const parts = [...before];
  if (parts.length > 0) parts.push('');
  parts.push(newSectionContent);

  let cursor = first.end;

  for (let i = 1; i < ranges.length; i++) {
    const r = ranges[i];
    parts.push(...lines.slice(cursor, r.start));
    cursor = r.end;
  }

  const after = lines.slice(cursor);

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

/**
 * @param {string} content - Current CHANGELOG.md content.
 * @param {string} newSectionContent - Full replacement (header + body).
 * @param {'unreleased'|'release'} mode
 * @param {string} targetVersion - Version without leading v (release mode only).
 * @returns {string}
 */
function rewriteChangelogSection(content, newSectionContent, mode, targetVersion) {
  const lines = content.split('\n');

  let targetStart = -1;

  if (mode === 'unreleased') {
    targetStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
  } else {
    targetStart = lines.findIndex((line) => line.startsWith(`## [${targetVersion}]`));
  }

  if (mode === 'release') {
    const unreleasedStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
    const ranges = [];
    if (unreleasedStart !== -1) {
      ranges.push({
        start: unreleasedStart,
        end: findSectionEnd(lines, unreleasedStart),
      });
    }
    if (targetStart !== -1) {
      ranges.push({
        start: targetStart,
        end: findSectionEnd(lines, targetStart),
      });
    }
    ranges.sort((a, b) => a.start - b.start);

    if (ranges.length === 0) {
      return newSectionContent + '\n\n' + content;
    }

    return spliceReleaseSectionRanges(lines, ranges, newSectionContent);
  }

  if (targetStart === -1) {
    return newSectionContent + '\n\n' + content;
  }

  const sectionEnd = findSectionEnd(lines, targetStart);

  const before = lines.slice(0, targetStart);
  const after = lines.slice(sectionEnd);

  while (before.length > 0 && before[before.length - 1] === '') {
    before.pop();
  }

  const parts = [...before];
  if (parts.length > 0) parts.push('');
  parts.push(newSectionContent);

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

if (typeof module !== 'undefined') {
  module.exports = {
    findSectionEnd,
    rewriteChangelogSection,
  };
}
