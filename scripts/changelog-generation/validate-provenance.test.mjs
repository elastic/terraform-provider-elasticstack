import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const {
  validateProvenance,
  extractPRReferences,
  extractBulletLines,
  looksLikeCommitNarration,
  extractSectionFromChangelog,
} = require(path.resolve(__dirname, 'validate-provenance.js'));

// ---------------------------------------------------------------------------
// Helper factories
// ---------------------------------------------------------------------------

function makeEvidence(prNumbers = [123, 456]) {
  return {
    target_section: '## [Unreleased]',
    pull_requests: prNumbers.map((n) => ({ number: n, title: `PR #${n}`, url: `https://example.com/pull/${n}` })),
  };
}

function makeProvenance(bullets = []) {
  return { bullets };
}

// ---------------------------------------------------------------------------
// extractPRReferences tests
// ---------------------------------------------------------------------------

test('extractPRReferences: finds single reference', () => {
  const refs = extractPRReferences('- Fix foo (#123)');
  assert.deepEqual(refs, [123]);
});

test('extractPRReferences: finds multiple references', () => {
  const refs = extractPRReferences('- Fix foo (#123) and bar (#456)');
  assert.deepEqual(refs, [123, 456]);
});

test('extractPRReferences: ignores URL path segments', () => {
  // /pull/123 should NOT be captured as a PR ref (preceded by /)
  const refs = extractPRReferences('see https://github.com/owner/repo/pull/789');
  assert.deepEqual(refs, []);
});

test('extractPRReferences: deduplicates', () => {
  const refs = extractPRReferences('- Fix (#123) and also (#123)');
  assert.deepEqual(refs, [123]);
});

// ---------------------------------------------------------------------------
// extractBulletLines tests
// ---------------------------------------------------------------------------

test('extractBulletLines: extracts dash bullets', () => {
  const text = '## [Unreleased]\n\n### Changes\n\n- Foo (#1)\n- Bar (#2)\n';
  const bullets = extractBulletLines(text);
  assert.equal(bullets.length, 2);
  assert.ok(bullets[0].includes('Foo'));
});

test('extractBulletLines: extracts asterisk bullets', () => {
  const text = '* Fix thing (#3)\n* Add widget (#4)';
  const bullets = extractBulletLines(text);
  assert.equal(bullets.length, 2);
});

test('extractBulletLines: skips non-bullet lines', () => {
  const text = '## [Unreleased]\n\n### Changes\n\nSome paragraph text.\n\n- Bullet (#5)';
  const bullets = extractBulletLines(text);
  assert.equal(bullets.length, 1);
});

// ---------------------------------------------------------------------------
// looksLikeCommitNarration tests
// ---------------------------------------------------------------------------

test('looksLikeCommitNarration: 40-char SHA → true', () => {
  // exactly 40 hex chars: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
  assert.equal(looksLikeCommitNarration('- Fix a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2'), true);
});

test('looksLikeCommitNarration: 7-char hex without PR ref → true', () => {
  assert.equal(looksLikeCommitNarration('- Fix abc1234 issue'), true);
});

test('looksLikeCommitNarration: normal bullet with PR ref → false', () => {
  assert.equal(looksLikeCommitNarration('- Fix thing (#123)'), false);
});

test('looksLikeCommitNarration: clean bullet → false', () => {
  assert.equal(looksLikeCommitNarration('- Add new resource for Fleet output'), false);
});

// ---------------------------------------------------------------------------
// extractSectionFromChangelog tests
// ---------------------------------------------------------------------------

const SAMPLE_CHANGELOG = `## [Unreleased]

### Changes

- Foo (#1)
- Bar (#2)

## [0.14.3] - 2026-03-02

### Changes

- Baz (#3)

[Unreleased]: https://example.com
[0.14.3]: https://example.com
`;

test('extractSectionFromChangelog: extracts Unreleased section', () => {
  const section = extractSectionFromChangelog(SAMPLE_CHANGELOG, '## [Unreleased]');
  assert.ok(section !== null);
  assert.ok(section.includes('## [Unreleased]'));
  assert.ok(section.includes('Foo (#1)'));
  assert.ok(!section.includes('0.14.3'));
});

test('extractSectionFromChangelog: extracts versioned section', () => {
  const section = extractSectionFromChangelog(SAMPLE_CHANGELOG, '## [0.14.3]');
  assert.ok(section !== null);
  assert.ok(section.includes('Baz (#3)'));
  assert.ok(!section.includes('Foo'));
});

test('extractSectionFromChangelog: returns null for missing section', () => {
  const section = extractSectionFromChangelog(SAMPLE_CHANGELOG, '## [9.9.9]');
  assert.equal(section, null);
});

// ---------------------------------------------------------------------------
// validateProvenance tests — valid cases
// ---------------------------------------------------------------------------

test('validateProvenance: valid provenance and changelog → passes', () => {
  const evidence = makeEvidence([123, 456]);
  const provenance = makeProvenance([
    { text: 'Fix foo (#123)', pr_numbers: [123], pr_urls: [] },
    { text: 'Add bar (#456)', pr_numbers: [456], pr_urls: [] },
  ]);
  const changelogSection = '## [Unreleased]\n\n### Changes\n\n- Fix foo (#123)\n- Add bar (#456)\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, true);
  assert.equal(result.errors.length, 0);
});

test('validateProvenance: empty bullets and no changelog entries → passes', () => {
  const evidence = makeEvidence([]);
  const provenance = makeProvenance([]);
  const changelogSection = '## [Unreleased]\n\n- No unreleased changes\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, true);
});

// ---------------------------------------------------------------------------
// validateProvenance tests — rejection cases
// ---------------------------------------------------------------------------

test('validateProvenance: provenance references unknown PR → fails', () => {
  const evidence = makeEvidence([123]);
  const provenance = makeProvenance([
    { text: 'Fix foo (#999)', pr_numbers: [999], pr_urls: [] },
  ]);
  const changelogSection = '## [Unreleased]\n\n- Fix foo (#999)\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('#999')));
});

test('validateProvenance: changelog references unknown PR → fails', () => {
  const evidence = makeEvidence([123]);
  const provenance = makeProvenance([
    { text: 'Fix foo (#123)', pr_numbers: [123], pr_urls: [] },
  ]);
  const changelogSection = '## [Unreleased]\n\n- Fix foo (#123)\n- Fabricated (#888)\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('#888')));
});

test('validateProvenance: provenance bullet has no pr_numbers → fails', () => {
  const evidence = makeEvidence([123]);
  const provenance = makeProvenance([
    { text: 'Fix something', pr_numbers: [], pr_urls: [] },
  ]);
  const changelogSection = '## [Unreleased]\n\n- Fix something\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('no pr_numbers')));
});

test('validateProvenance: section header mismatch → fails', () => {
  const evidence = { ...makeEvidence([123]), target_section: '## [Unreleased]' };
  const provenance = makeProvenance([
    { text: 'Fix foo (#123)', pr_numbers: [123], pr_urls: [] },
  ]);
  // Section starts with wrong header
  const changelogSection = '## [0.14.4] - 2026-04-16\n\n- Fix foo (#123)\n';

  const result = validateProvenance({
    evidence,
    provenance,
    changelogSection,
    expectedHeader: '## [Unreleased]',
  });
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('does not match')));
});

test('validateProvenance: commit SHA in bullet → fails', () => {
  const evidence = makeEvidence([123]);
  const provenance = makeProvenance([
    { text: 'Fix (#123)', pr_numbers: [123], pr_urls: [] },
  ]);
  // exactly 40 hex chars
  const sha = 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2';
  const changelogSection = `## [Unreleased]\n\n- Fix via ${sha} (#123)\n`;

  const result = validateProvenance({ evidence, provenance, changelogSection });
  assert.equal(result.valid, false);
  assert.ok(result.errors.some((e) => e.includes('commit-level narration')));
});

test('validateProvenance: bullet without PR ref → warning (not error)', () => {
  const evidence = makeEvidence([]);
  const provenance = makeProvenance([]);
  const changelogSection = '## [Unreleased]\n\n- Some generic improvement\n';

  const result = validateProvenance({ evidence, provenance, changelogSection });
  // No PR refs in changelog to check, so no error from check 2.
  // But bullet has no PR ref → warning
  assert.ok(result.warnings.some((w) => w.includes('no PR reference')));
});
