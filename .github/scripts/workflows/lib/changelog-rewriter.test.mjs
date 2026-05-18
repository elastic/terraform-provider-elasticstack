import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { findSectionEnd, rewriteLinkTable, rewriteChangelogSection } = require('./changelog-rewriter.js');

test('findSectionEnd stops at next ## heading', () => {
  const lines = ['## [Unreleased]', 'a', '## [1.0.0] - x', 'tail'];
  assert.equal(findSectionEnd(lines, 0), 2);
});

test('findSectionEnd returns lines.length when no following section', () => {
  const lines = ['## [Unreleased]', 'only'];
  assert.equal(findSectionEnd(lines, 0), 2);
});

test('rewriteChangelogSection unreleased replaces Unreleased section body', () => {
  const before = [
    '# Changelog',
    '',
    '## [Unreleased]',
    'old',
    '',
    '## [1.0.0] - 2020-01-01',
    'released',
  ].join('\n');
  const newSection = '## [Unreleased]\n\n### Changes\n\n- fresh ([#1](u))';
  const out = rewriteChangelogSection(before, newSection, 'unreleased', '');
  assert.ok(out.includes('### Changes'));
  assert.ok(!out.includes('\nold\n'));
  assert.ok(out.includes('## [1.0.0]'));
});

test('rewriteChangelogSection release replaces Unreleased section with new versioned section', () => {
  const before = ['# C', '', '## [Unreleased]', 'work', '', '## [0.9.0]', 'x'].join('\n');
  const newSection = '## [1.0.0] - 2025-01-01\n\n### Changes\n\n- x ([#2](u))';
  const out = rewriteChangelogSection(before, newSection, 'release', '1.0.0');
  assert.ok(!out.includes('## [Unreleased]'));
  const newIdx = out.indexOf('## [1.0.0]');
  const oldIdx = out.indexOf('## [0.9.0]');
  assert.ok(newIdx !== -1 && oldIdx !== -1 && newIdx < oldIdx);
});

test('rewriteChangelogSection release prepends new section when neither Unreleased nor target heading exists', () => {
  const before = ['# Changelog', '', '## [0.9.0]', 'prior release'].join('\n');
  const newSection = '## [1.0.0] - 2026-06-01\n\n### Changes\n\n- leap ([#501](https://example/501))';
  const out = rewriteChangelogSection(before, newSection, 'release', '1.0.0');
  assert.ok(out.startsWith('## [1.0.0]'));
  assert.ok(!out.includes('## [Unreleased]'));
  assert.match(out, /# Changelog/);
  const vNew = out.indexOf('## [1.0.0]');
  const vPrev = out.indexOf('## [0.9.0]');
  assert.ok(vNew !== -1 && vPrev !== -1 && vNew < vPrev);
  assert.ok(out.includes('prior release'));
});

test('rewriteChangelogSection release replaces version block only when target exists without Unreleased', () => {
  const before = [
    '# Changelog',
    '',
    '## [1.0.0] - stale-date',
    '',
    '- stale bullet ([#11](https://example/11))',
    '',
    '## [0.9.0]',
    'older',
  ].join('\n');
  const newSection =
    '## [1.0.0] - 2026-06-15\n\n### Changes\n\n- current ([#502](https://example/502))';
  const out = rewriteChangelogSection(before, newSection, 'release', '1.0.0');
  assert.ok(!out.includes('## [Unreleased]'));
  assert.equal([...out.matchAll(/^## \[1\.0\.0\]/gm)].length, 1);
  assert.ok(out.includes('- current'));
  assert.ok(!out.includes('- stale bullet'));
  const vTen = out.indexOf('## [1.0.0]');
  const vNine = out.indexOf('## [0.9.0]');
  assert.ok(vTen !== -1 && vNine !== -1 && vTen < vNine);
});

// Release re-run collapses lingering Unreleased: dual-range splice is exercised
// here directly (see design.md) rather than via the engine, so positioning and
// removal order stay independent of rendered section bodies / PR harness data.
test('rewriteChangelogSection release re-run drops Unreleased when target version already exists', () => {
  const before = [
    '# Changelog',
    '',
    '## [Unreleased]',
    'stale unreleased',
    '',
    '## [1.0.0] - 2020-06-01',
    '',
    '### Changes',
    '',
    '- obsolete ([#10](https://example/10))',
    '',
    '## [0.9.0]',
    'prior',
  ].join('\n');

  const newSection =
    '## [1.0.0] - 2026-05-12\n\n### Changes\n\n- refreshed ([#999](https://example/999))';
  const out = rewriteChangelogSection(before, newSection, 'release', '1.0.0');

  assert.ok(!out.includes('## [Unreleased]'));
  assert.equal([...out.matchAll(/^## \[1\.0\.0\]/gm)].length, 1);
  assert.ok(out.includes('refreshed'));
  assert.ok(!out.includes('stale unreleased'));
  assert.ok(!out.includes('- obsolete'));
  const newIdx = out.indexOf('## [1.0.0]');
  const prevIdx = out.indexOf('## [0.9.0]');
  assert.ok(newIdx !== -1 && prevIdx !== -1 && newIdx < prevIdx);
});

// Observed duplicate Unreleased vs release bullets in elastic/terraform-provider-elasticstack#2857 —
// rewriter must collapse to one versioned section so each bullet appears once.
test('rewriteChangelogSection release dedupes when Unreleased body mirrors rendered release (#2857)', () => {
  const releaseBody =
    '\n### Changes\n\n' +
    '- First ship ([#2840](https://github.com/elastic/terraform-provider-elasticstack/pull/2840))\n' +
    '- Second ship ([#2841](https://github.com/elastic/terraform-provider-elasticstack/pull/2841))\n';

  const unreleasedFixture = `# Log

## [Unreleased]
${releaseBody}
## [0.14.0] - older
prior
`;

  const version = '0.15.0';
  const header = `## [${version}] - 2026-05-11`;
  const newSection = `${header}\n${releaseBody}`;
  const out = rewriteChangelogSection(unreleasedFixture, newSection, 'release', version);

  assert.ok(!out.includes('## [Unreleased]'));
  assert.equal(out.match(/First ship/g)?.length ?? 0, 1);
  assert.equal(out.match(/Second ship/g)?.length ?? 0, 1);
  assert.equal(out.match(/^## \[0\.15\.0\]/gm)?.length ?? 0, 1);
});

test('rewriteChangelogSection prepends unreleased when no Unreleased heading', () => {
  const before = '# T\n\n## [1.0.0]\nx';
  const newSection = '## [Unreleased]\n\n### Changes\n\n- y ([#3](u))';
  const out = rewriteChangelogSection(before, newSection, 'unreleased', '');
  assert.ok(out.startsWith('## [Unreleased]'));
});

test('rewriteLinkTable standard release updates Unreleased URL and inserts new version entry', () => {
  const before = [
    '# Changelog',
    '',
    '[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD',
    '[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5',
  ].join('\n');

  const out = rewriteLinkTable(before, '0.15.0', 'v0.14.5');

  assert.match(
    out,
    /^\[Unreleased\]: https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.15\.0\.\.\.HEAD$/m,
  );
  assert.match(
    out,
    /^\[0\.15\.0\]: https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.14\.5\.\.\.v0\.15\.0$/m,
  );
  assert.match(
    out,
    /\[Unreleased\]: https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.15\.0\.\.\.HEAD\n\[0\.15\.0\]: https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.14\.5\.\.\.v0\.15\.0\n\[0\.14\.5\]:/,
  );
});

test('rewriteLinkTable is idempotent when release entry already exists', () => {
  const before = [
    '# Changelog',
    '',
    '[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.15.0...HEAD',
    '[0.15.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...v0.15.0',
    '[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5',
  ].join('\n');

  const out = rewriteLinkTable(before, '0.15.0', 'v0.14.5');

  assert.equal(out.match(/^\[0\.15\.0\]:/gm)?.length ?? 0, 1);
  assert.match(
    out,
    /^\[Unreleased\]: https:\/\/github.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.15\.0\.\.\.HEAD$/m,
  );
});

test('rewriteLinkTable is a no-op when Unreleased link line is absent', () => {
  const before = [
    '# Changelog',
    '',
    '[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5',
  ].join('\n');

  assert.equal(rewriteLinkTable(before, '0.15.0', 'v0.14.5'), before);
});

test('rewriteLinkTable is a no-op when previousTag is empty', () => {
  const before = [
    '# Changelog',
    '',
    '[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD',
    '[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5',
  ].join('\n');

  assert.equal(rewriteLinkTable(before, '0.15.0', ''), before);
});

test('rewriteLinkTable is a no-op when targetVersion is empty', () => {
  const before = [
    '# Changelog',
    '',
    '[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD',
    '[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5',
  ].join('\n');

  assert.equal(rewriteLinkTable(before, '', 'v0.14.5'), before);
});
