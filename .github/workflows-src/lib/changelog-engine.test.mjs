import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const require = createRequire(import.meta.url);
const {
  validateModeAndTargetVersion,
  resolveChangelogCompareContext,
  gatherMergedPRRecordsForRange,
  formatAssemblyFailureMessage,
  runChangelogRenderAndWrite,
  runChangelogEngine,
} = require('./changelog-engine.js');

function mockCore() {
  return {
    failed: [],
    warnings: [],
    infos: [],
    setFailed(msg) {
      this.failed.push(msg);
    },
    warning(msg) {
      this.warnings.push(msg);
    },
    info(msg) {
      this.infos.push(msg);
    },
  };
}

test('validateModeAndTargetVersion rejects invalid mode', () => {
  assert.throws(
    () => validateModeAndTargetVersion('staging', '', null),
    /Invalid changelog mode/
  );
});

test('validateModeAndTargetVersion rejects release without targetVersion', () => {
  assert.throws(() => validateModeAndTargetVersion('release', '', null), /targetVersion/);
});

test('validateModeAndTargetVersion rejects release with leading v', () => {
  assert.throws(() => validateModeAndTargetVersion('release', 'v1.0.0', null), /targetVersion/);
});

test('validateModeAndTargetVersion accepts release semver', () => {
  assert.doesNotThrow(() => validateModeAndTargetVersion('release', '1.0.0', null));
});

test('resolveChangelogCompareContext unreleased: compare HEAD when no tags', () => {
  const exec = (cmd) => {
    if (cmd.startsWith('git tag')) return '';
    throw new Error(`unexpected: ${cmd}`);
  };
  const ctx = resolveChangelogCompareContext({
    mode: 'unreleased',
    targetVersion: '',
    exec,
    core: null,
  });
  assert.equal(ctx.compareRange, 'HEAD');
  assert.equal(ctx.targetBranch, 'generated-changelog');
});

test('resolveChangelogCompareContext release: excludes current version tag', () => {
  const exec = (cmd) => {
    if (cmd.startsWith('git tag')) return 'v2.0.0\nv1.9.0\n';
    throw new Error(`unexpected: ${cmd}`);
  };
  const ctx = resolveChangelogCompareContext({
    mode: 'release',
    targetVersion: '2.0.0',
    exec,
    core: null,
  });
  assert.equal(ctx.previousTag, 'v1.9.0');
  assert.equal(ctx.compareRange, 'v1.9.0..HEAD');
  assert.equal(ctx.targetBranch, 'prep-release-2.0.0');
});

test('gatherMergedPRRecordsForRange dedupes merged PRs from commits', async () => {
  const calls = [];
  const github = {
    rest: {
      repos: {
        listPullRequestsAssociatedWithCommit: async ({ commit_sha }) => {
          calls.push(commit_sha);
          return {
            data: [
              {
                number: 10,
                state: 'closed',
                merged_at: '2025-01-01',
                title: 'T',
                html_url: 'https://example/pull/10',
                merge_commit_sha: commit_sha,
                user: { login: 'u' },
                labels: [{ name: 'bug' }],
                body: '## Changelog\nCustomer impact: fix\nSummary: fix it\n',
              },
            ],
          };
        },
      },
    },
  };
  const exec = (cmd) => {
    assert.match(cmd, /git log --format=%H/);
    return 'aaa\nbbb\n';
  };
  const records = await gatherMergedPRRecordsForRange({
    github,
    owner: 'o',
    repo: 'r',
    compareRange: 'v1..HEAD',
    exec,
    core: null,
  });
  assert.equal(records.length, 1);
  assert.equal(records[0].number, 10);
  assert.equal(records[0].labels[0], 'bug');
  assert.equal(calls.length, 2);
});

test('formatAssemblyFailureMessage lists reasons', () => {
  const msg = formatAssemblyFailureMessage([{ reason: 'bad PR' }]);
  assert.ok(msg.includes('bad PR'));
  assert.ok(msg.includes('Changelog assembly failed'));
});

test('runChangelogRenderAndWrite sets hasUserFacingChanges when PR included', () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(
    changelogPath,
    ['# L', '', '## [Unreleased]', 'old', '', '## [0.1.0]', 'x'].join('\n'),
    'utf8'
  );
  const fs = require('node:fs');
  const prRecords = [
    {
      number: 1,
      url: 'https://github.com/o/r/pull/1',
      labels: [],
      body: '## Changelog\nCustomer impact: fix\nSummary: hello\n',
    },
  ];
  const out = runChangelogRenderAndWrite({
    core,
    prRecords,
    mode: 'unreleased',
    targetVersion: '',
    changelogPath,
    fs,
  });
  assert.equal(out.hasPRs, true);
  assert.equal(out.hasUserFacingChanges, true);
  const text = readFileSync(changelogPath, 'utf8');
  assert.ok(text.includes('hello'));
  assert.ok(!text.includes('\nold\n'));
});

test('runChangelogRenderAndWrite hasUserFacingChanges is true when only breaking-changes content rendered', () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(changelogPath, '# L\n\n## [Unreleased]\nold\n', 'utf8');
  const fs = require('node:fs');
  const prRecords = [
    {
      number: 7,
      url: 'https://github.com/o/r/pull/7',
      labels: [],
      body: [
        '## Changelog',
        'Customer impact: none',
        '',
        '### Breaking changes',
        'A new required env var `FOO` must be set.',
      ].join('\n'),
    },
  ];
  const out = runChangelogRenderAndWrite({
    core,
    prRecords,
    mode: 'release',
    targetVersion: '1.0.0',
    changelogPath,
    fs,
  });
  assert.equal(out.included.length, 0);
  assert.equal(out.hasPRs, true);
  assert.equal(out.hasUserFacingChanges, true);
  const text = readFileSync(changelogPath, 'utf8');
  assert.ok(text.includes('### Breaking changes'));
  assert.ok(text.includes('FOO'));
});

test('runChangelogRenderAndWrite release inserts section after Unreleased', () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(
    changelogPath,
    ['# L', '', '## [Unreleased]', 'pending', '', '## [0.9.0]', 'z'].join('\n'),
    'utf8'
  );
  const fs = require('node:fs');
  const prRecords = [
    {
      number: 2,
      url: 'https://github.com/o/r/pull/2',
      labels: [],
      body: '## Changelog\nCustomer impact: enhancement\nSummary: ship\n',
    },
  ];
  const out = runChangelogRenderAndWrite({
    core,
    prRecords,
    mode: 'release',
    targetVersion: '1.0.0',
    changelogPath,
    fs,
  });
  assert.match(out.sectionHeader, /^## \[1\.0\.0\] - \d{4}-\d{2}-\d{2}$/);
  const text = readFileSync(changelogPath, 'utf8');
  const u = text.indexOf('## [Unreleased]');
  const r = text.indexOf('## [1.0.0]');
  const old = text.indexOf('## [0.9.0]');
  assert.ok(u !== -1 && r !== -1 && r > u && old > r);
});

test('runChangelogRenderAndWrite release with zero PRs writes header-only section', () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(
    changelogPath,
    ['# L', '', '## [Unreleased]', 'pending', '', '## [0.9.0]', 'z'].join('\n'),
    'utf8'
  );
  const fs = require('node:fs');
  const out = runChangelogRenderAndWrite({
    core,
    prRecords: [],
    mode: 'release',
    targetVersion: '1.0.0',
    changelogPath,
    fs,
  });
  assert.match(out.sectionHeader, /^## \[1\.0\.0\] - \d{4}-\d{2}-\d{2}$/);
  assert.equal(out.hasPRs, false);
  assert.equal(out.hasUserFacingChanges, false);
  assert.equal(out.included.length, 0);
  const text = readFileSync(changelogPath, 'utf8');
  const u = text.indexOf('## [Unreleased]');
  const r = text.indexOf('## [1.0.0]');
  const old = text.indexOf('## [0.9.0]');
  assert.ok(u !== -1 && r !== -1 && r > u && old > r);
  assert.ok(text.includes('## [1.0.0]'));
});

test('runChangelogRenderAndWrite assembly failure calls setFailed and throws', () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(changelogPath, '# x\n', 'utf8');
  const fs = require('node:fs');
  const prRecords = [
    {
      number: 99,
      url: 'https://github.com/o/r/pull/99',
      labels: [],
      body: 'no changelog block',
    },
  ];
  assert.throws(
    () =>
      runChangelogRenderAndWrite({
        core,
        prRecords,
        mode: 'unreleased',
        targetVersion: '',
        changelogPath,
        fs,
      }),
    /Changelog assembly failed/
  );
  assert.equal(core.failed.length, 1);
  assert.ok(core.failed[0].includes('missing a required ## Changelog'));
});

test('runChangelogEngine end-to-end with mocks', async () => {
  const core = mockCore();
  const dir = mkdtempSync(path.join(os.tmpdir(), 'clog-'));
  const changelogPath = path.join(dir, 'CHANGELOG.md');
  writeFileSync(
    changelogPath,
    ['# L', '', '## [Unreleased]', 'x', '', '## [0.1.0]', 'y'].join('\n'),
    'utf8'
  );
  const fs = require('node:fs');

  const github = {
    rest: {
      repos: {
        listPullRequestsAssociatedWithCommit: async () => ({
          data: [
            {
              number: 5,
              state: 'closed',
              merged_at: '2025-02-01',
              title: 'Feat',
              html_url: 'https://github.com/o/r/pull/5',
              merge_commit_sha: 'sha1',
              user: { login: 'dev' },
              labels: [],
              body: '## Changelog\nCustomer impact: enhancement\nSummary: feat done\n',
            },
          ],
        }),
      },
    },
  };

  const exec = (cmd, _opts) => {
    if (cmd.startsWith('git tag')) return 'v0.1.0\n';
    if (cmd.startsWith('git log')) return 'cafe\n';
    throw new Error(cmd);
  };

  const result = await runChangelogEngine({
    github,
    core,
    mode: 'unreleased',
    targetVersion: '',
    owner: 'o',
    repo: 'r',
    changelogPath,
    exec,
    fs,
  });

  assert.equal(result.mode, 'unreleased');
  assert.equal(result.previousTag, 'v0.1.0');
  assert.equal(result.compareRange, 'v0.1.0..HEAD');
  assert.equal(result.targetBranch, 'generated-changelog');
  assert.equal(result.hasPRs, true);
  assert.equal(result.hasUserFacingChanges, true);
  assert.equal(result.errors.length, 0);
  assert.ok(readFileSync(changelogPath, 'utf8').includes('feat done'));
});
