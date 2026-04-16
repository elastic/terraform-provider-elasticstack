import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const {
  parseChangelog,
  serialiseChangelog,
  rewriteUnreleased,
  rewriteRelease,
} = require(path.resolve(__dirname, 'rewrite-changelog-section.js'));

// ---------------------------------------------------------------------------
// Sample CHANGELOG content
// ---------------------------------------------------------------------------

const SAMPLE = `## [Unreleased]

### Changes

- Existing unreleased entry (#100)

## [0.14.3] - 2026-03-02

### Changes

- Stable release entry (#99)

## [0.14.2] - 2026-02-19

### Changes

- Older entry (#98)

[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.3...HEAD
[0.14.3]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.2...v0.14.3
[0.14.2]: https://github.com/elastic/terraform-provider-elasticstack/releases/tag/v0.14.2
`;

// ---------------------------------------------------------------------------
// parseChangelog tests
// ---------------------------------------------------------------------------

test('parseChangelog: identifies all sections', () => {
  const parsed = parseChangelog(SAMPLE);
  assert.equal(parsed.sections.length, 3);
  assert.equal(parsed.sections[0].header, '## [Unreleased]');
  assert.equal(parsed.sections[1].header, '## [0.14.3] - 2026-03-02');
  assert.equal(parsed.sections[2].header, '## [0.14.2] - 2026-02-19');
});

test('parseChangelog: footer is separated from last section', () => {
  const parsed = parseChangelog(SAMPLE);
  assert.ok(parsed.footer.includes('[Unreleased]:'));
  assert.ok(parsed.footer.includes('[0.14.3]:'));
  // Footer should not be part of any section body
  for (const section of parsed.sections) {
    assert.ok(!section.body.includes('[Unreleased]: https'));
  }
});

test('parseChangelog: round-trips cleanly (serialise after parse)', () => {
  const parsed = parseChangelog(SAMPLE);
  const roundTripped = serialiseChangelog(parsed);
  // Should contain all the same sections and footer
  assert.ok(roundTripped.includes('## [Unreleased]'));
  assert.ok(roundTripped.includes('## [0.14.3] - 2026-03-02'));
  assert.ok(roundTripped.includes('[Unreleased]: https'));
});

// ---------------------------------------------------------------------------
// rewriteUnreleased tests
// ---------------------------------------------------------------------------

test('rewriteUnreleased: replaces Unreleased body content', () => {
  const newBody = '### Changes\n\n- New entry (#200)';
  const result = rewriteUnreleased(SAMPLE, newBody);

  assert.ok(result.includes('## [Unreleased]'));
  assert.ok(result.includes('New entry (#200)'));
  assert.ok(!result.includes('Existing unreleased entry'));
});

test('rewriteUnreleased: preserves other release sections', () => {
  const newBody = '### Changes\n\n- New entry (#200)';
  const result = rewriteUnreleased(SAMPLE, newBody);

  assert.ok(result.includes('## [0.14.3] - 2026-03-02'));
  assert.ok(result.includes('Stable release entry (#99)'));
  assert.ok(result.includes('## [0.14.2] - 2026-02-19'));
  assert.ok(result.includes('Older entry (#98)'));
});

test('rewriteUnreleased: preserves link footer', () => {
  const newBody = '### Changes\n\n- New entry (#200)';
  const result = rewriteUnreleased(SAMPLE, newBody);

  assert.ok(result.includes('[Unreleased]: https://'));
  assert.ok(result.includes('[0.14.3]: https://'));
});

test('rewriteUnreleased: throws when Unreleased section missing', () => {
  const noUnreleased = `## [0.14.3] - 2026-03-02\n\n- Entry (#1)\n`;
  assert.throws(
    () => rewriteUnreleased(noUnreleased, '- New (#2)'),
    /Unreleased.*not found/
  );
});

// ---------------------------------------------------------------------------
// rewriteRelease tests — insert new section
// ---------------------------------------------------------------------------

test('rewriteRelease: inserts new release section after Unreleased', () => {
  const newBody = '### Changes\n\n- Release feature (#201)';
  const result = rewriteRelease(SAMPLE, '0.14.4', '2026-04-16', newBody);

  assert.ok(result.includes('## [0.14.4] - 2026-04-16'));
  assert.ok(result.includes('Release feature (#201)'));
});

test('rewriteRelease: new section appears before existing versioned sections', () => {
  const newBody = '### Changes\n\n- Release feature (#201)';
  const result = rewriteRelease(SAMPLE, '0.14.4', '2026-04-16', newBody);

  const pos0144 = result.indexOf('## [0.14.4]');
  const pos0143 = result.indexOf('## [0.14.3]');
  assert.ok(pos0144 < pos0143, 'New section should appear before 0.14.3');
});

test('rewriteRelease: clears Unreleased section body', () => {
  const newBody = '### Changes\n\n- Release feature (#201)';
  const result = rewriteRelease(SAMPLE, '0.14.4', '2026-04-16', newBody);

  assert.ok(result.includes('## [Unreleased]'));
  assert.ok(result.includes('No unreleased changes'));
  assert.ok(!result.includes('Existing unreleased entry'));
});

test('rewriteRelease: preserves other release sections', () => {
  const newBody = '### Changes\n\n- Release feature (#201)';
  const result = rewriteRelease(SAMPLE, '0.14.4', '2026-04-16', newBody);

  assert.ok(result.includes('Stable release entry (#99)'));
  assert.ok(result.includes('Older entry (#98)'));
});

test('rewriteRelease: preserves link footer', () => {
  const newBody = '### Changes\n\n- Release feature (#201)';
  const result = rewriteRelease(SAMPLE, '0.14.4', '2026-04-16', newBody);

  assert.ok(result.includes('[Unreleased]: https://'));
  assert.ok(result.includes('[0.14.3]: https://'));
});

// ---------------------------------------------------------------------------
// rewriteRelease tests — replace existing section
// ---------------------------------------------------------------------------

const SAMPLE_WITH_EXISTING_RELEASE = `## [Unreleased]

- No unreleased changes

## [0.14.4] - 2026-04-01

### Changes

- Old release content (#150)

## [0.14.3] - 2026-03-02

### Changes

- Stable release entry (#99)

[Unreleased]: https://example.com
[0.14.4]: https://example.com
[0.14.3]: https://example.com
`;

test('rewriteRelease: replaces existing release section when version matches', () => {
  const newBody = '### Changes\n\n- Updated release content (#160)';
  const result = rewriteRelease(SAMPLE_WITH_EXISTING_RELEASE, '0.14.4', '2026-04-16', newBody);

  assert.ok(result.includes('## [0.14.4] - 2026-04-16'));
  assert.ok(result.includes('Updated release content (#160)'));
  assert.ok(!result.includes('Old release content (#150)'));
});

test('rewriteRelease: does not duplicate the release section', () => {
  const newBody = '### Changes\n\n- Updated release content (#160)';
  const result = rewriteRelease(SAMPLE_WITH_EXISTING_RELEASE, '0.14.4', '2026-04-16', newBody);

  const count = (result.match(/## \[0\.14\.4\]/g) ?? []).length;
  assert.equal(count, 1, 'Release section should appear exactly once');
});

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

test('rewriteUnreleased: handles CHANGELOG with only Unreleased section', () => {
  const minimal = `## [Unreleased]\n\n- Old entry (#1)\n\n[Unreleased]: https://example.com\n`;
  const result = rewriteUnreleased(minimal, '### Changes\n\n- New entry (#2)');
  assert.ok(result.includes('## [Unreleased]'));
  assert.ok(result.includes('New entry (#2)'));
  assert.ok(!result.includes('Old entry'));
});
