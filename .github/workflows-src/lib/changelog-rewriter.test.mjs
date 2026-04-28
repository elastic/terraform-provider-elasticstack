import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { findSectionEnd, rewriteChangelogSection } = require('./changelog-rewriter.js');

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

test('rewriteChangelogSection release inserts after Unreleased when section missing', () => {
  const before = ['# C', '', '## [Unreleased]', 'work', '', '## [0.9.0]', 'x'].join('\n');
  const newSection = '## [1.0.0] - 2025-01-01\n\n### Changes\n\n- x ([#2](u))';
  const out = rewriteChangelogSection(before, newSection, 'release', '1.0.0');
  const unreleasedIdx = out.indexOf('## [Unreleased]');
  const newIdx = out.indexOf('## [1.0.0]');
  const oldIdx = out.indexOf('## [0.9.0]');
  assert.ok(unreleasedIdx !== -1 && newIdx !== -1 && oldIdx !== -1);
  assert.ok(newIdx > unreleasedIdx && newIdx < oldIdx);
});

test('rewriteChangelogSection prepends unreleased when no Unreleased heading', () => {
  const before = '# T\n\n## [1.0.0]\nx';
  const newSection = '## [Unreleased]\n\n### Changes\n\n- y ([#3](u))';
  const out = rewriteChangelogSection(before, newSection, 'unreleased', '');
  assert.ok(out.startsWith('## [Unreleased]'));
});
