import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

const {
  buildRendererPullRequestRecord,
  resolveEngineContext,
  rewriteChangelogSection,
  runChangelogEngine,
} = require('./changelog-engine.js');

function makeGithubClient({ pullRequestsByCommit = {} } = {}) {
  return {
    rest: {
      repos: {
        async listPullRequestsAssociatedWithCommit({ commit_sha }) {
          return { data: pullRequestsByCommit[commit_sha] ?? [] };
        },
      },
    },
  };
}

function makePullRequest(overrides = {}) {
  return {
    number: 42,
    title: 'Example PR',
    html_url: 'https://example.test/pr/42',
    merge_commit_sha: 'abc123',
    state: 'closed',
    merged_at: '2026-04-17T12:00:00Z',
    user: { login: 'octocat' },
    labels: [],
    body: '## Changelog\nCustomer impact: enhancement\nSummary: added a useful feature\n',
    ...overrides,
  };
}

function makeFs(initialFiles = {}) {
  const files = new Map(Object.entries(initialFiles));
  return {
    readFileSync(path, encoding) {
      assert.equal(encoding, 'utf8');
      if (!files.has(path)) {
        const error = new Error(`ENOENT: no such file or directory, open '${path}'`);
        error.code = 'ENOENT';
        throw error;
      }
      return files.get(path);
    },
    writeFileSync(path, content, encoding) {
      assert.equal(encoding, 'utf8');
      files.set(path, content);
    },
    get(path) {
      return files.get(path);
    },
  };
}

const baseChangelog = `## [Unreleased]\n\n### Changes\n\n- Existing unreleased entry (#100)\n\n## [0.14.3] - 2026-03-02\n\n### Changes\n\n- Stable release entry (#99)\n`;

test('resolveEngineContext supports explicit unreleased mode', () => {
  assert.deepEqual(
    resolveEngineContext({
      mode: 'unreleased',
      tags: ['v1.2.3', 'v1.2.2'],
    }),
    {
      mode: 'unreleased',
      targetVersion: '',
      targetBranch: 'generated-changelog',
      previousTag: 'v1.2.3',
      excludedTag: '',
      excludedCurrentTag: false,
      compareRange: 'v1.2.3..HEAD',
    }
  );
});

test('resolveEngineContext supports explicit release mode', () => {
  assert.deepEqual(
    resolveEngineContext({
      mode: 'release',
      targetVersion: '1.2.3',
      tags: ['v1.2.3', 'v1.2.2'],
    }),
    {
      mode: 'release',
      targetVersion: '1.2.3',
      targetBranch: 'prep-release-1.2.3',
      previousTag: 'v1.2.2',
      excludedTag: 'v1.2.3',
      excludedCurrentTag: true,
      compareRange: 'v1.2.2..HEAD',
    }
  );
});

test('buildRendererPullRequestRecord normalizes renderer inputs', () => {
  assert.deepEqual(
    buildRendererPullRequestRecord(
      makePullRequest({ labels: [{ name: 'bug' }, { name: 'enhancement' }] })
    ),
    {
      number: 42,
      title: 'Example PR',
      url: 'https://example.test/pr/42',
      merge_commit_sha: 'abc123',
      author: 'octocat',
      labels: ['bug', 'enhancement'],
      body: '## Changelog\nCustomer impact: enhancement\nSummary: added a useful feature\n',
    }
  );
});

test('rewriteChangelogSection replaces unreleased section only', () => {
  const updated = rewriteChangelogSection(
    baseChangelog,
    '## [Unreleased]\n\n### Changes\n\n- New entry (#200)',
    'unreleased',
    ''
  );

  assert.match(updated, /## \[Unreleased\][\s\S]*New entry \(#200\)/);
  assert.doesNotMatch(updated, /Existing unreleased entry/);
  assert.match(updated, /## \[0.14.3\] - 2026-03-02/);
});

test('runChangelogEngine assembles unreleased changelog from GitHub-backed merged PRs', async () => {
  const fs = makeFs({ 'CHANGELOG.md': baseChangelog });
  const github = makeGithubClient({
    pullRequestsByCommit: {
      sha1: [makePullRequest({ number: 101, html_url: 'https://example.test/pr/101' })],
      sha2: [makePullRequest({ number: 101, html_url: 'https://example.test/pr/101' })],
    },
  });
  const execCalls = [];
  const exec = (command) => {
    execCalls.push(command);
    if (command.startsWith('git tag --list')) return 'v1.2.3\nv1.2.2\n';
    if (command === 'git log --format=%H v1.2.3..HEAD') return 'sha1\nsha2\n';
    throw new Error(`unexpected exec command: ${command}`);
  };

  const result = await runChangelogEngine({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    mode: 'unreleased',
    changelogPath: 'CHANGELOG.md',
    generatedAt: '2026-04-20T12:00:00.000Z',
    fsImpl: fs,
    exec,
  });

  assert.equal(result.compareRange, 'v1.2.3..HEAD');
  assert.equal(result.sectionHeader, '## [Unreleased]');
  assert.equal(result.hasUserFacingChanges, true);
  assert.equal(result.manifest.pr_count, 1);
  assert.equal(result.pullRequests.length, 1);
  assert.match(fs.get('CHANGELOG.md'), /added a useful feature \(\[#101\]\(https:\/\/example.test\/pr\/101\)\)/);
  assert.deepEqual(execCalls, [
    'git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-version:refname',
    'git log --format=%H v1.2.3..HEAD',
  ]);
});

test('runChangelogEngine assembles release changelog and rewrites only targeted section', async () => {
  const fs = makeFs({ 'CHANGELOG.md': baseChangelog });
  const github = makeGithubClient({
    pullRequestsByCommit: {
      sha1: [makePullRequest({ number: 202, html_url: 'https://example.test/pr/202' })],
    },
  });
  const exec = (command) => {
    if (command.startsWith('git tag --list')) return 'v1.2.3\nv1.2.2\n';
    if (command === 'git log --format=%H v1.2.2..HEAD') return 'sha1\n';
    throw new Error(`unexpected exec command: ${command}`);
  };

  const result = await runChangelogEngine({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    mode: 'release',
    targetVersion: '1.2.3',
    changelogPath: 'CHANGELOG.md',
    generatedAt: '2026-04-20T12:00:00.000Z',
    fsImpl: fs,
    exec,
  });

  assert.equal(result.compareRange, 'v1.2.2..HEAD');
  assert.equal(result.sectionHeader, '## [1.2.3] - 2026-04-20');
  assert.match(fs.get('CHANGELOG.md'), /## \[Unreleased\][\s\S]*Existing unreleased entry/);
  assert.match(fs.get('CHANGELOG.md'), /## \[1.2.3\] - 2026-04-20[\s\S]*added a useful feature \(\[#202\]/);
});

test('runChangelogEngine replaces an existing release section without disturbing adjacent sections', async () => {
  const existingReleaseChangelog = `## [Unreleased]\n\n### Changes\n\n- Existing unreleased entry (#100)\n\n## [1.2.3] - 2026-04-20\n\n### Changes\n\n- Old generated entry (#150)\n\n## [0.14.3] - 2026-03-02\n\n### Changes\n\n- Stable release entry (#99)\n`;
  const fs = makeFs({ 'CHANGELOG.md': existingReleaseChangelog });
  const github = makeGithubClient({
    pullRequestsByCommit: {
      sha1: [makePullRequest({ number: 303, html_url: 'https://example.test/pr/303' })],
    },
  });
  const exec = (command) => {
    if (command.startsWith('git tag --list')) return 'v1.2.3\nv1.2.2\n';
    if (command === 'git log --format=%H v1.2.2..HEAD') return 'sha1\n';
    throw new Error(`unexpected exec command: ${command}`);
  };

  await runChangelogEngine({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    mode: 'release',
    targetVersion: '1.2.3',
    changelogPath: 'CHANGELOG.md',
    generatedAt: '2026-04-20T12:00:00.000Z',
    fsImpl: fs,
    exec,
  });

  const updated = fs.get('CHANGELOG.md');
  assert.match(updated, /## \[Unreleased\][\s\S]*Existing unreleased entry/);
  assert.match(updated, /## \[1.2.3\] - 2026-04-20[\s\S]*added a useful feature \(\[#303\]/);
  assert.doesNotMatch(updated, /Old generated entry/);
  assert.match(updated, /## \[0.14.3\] - 2026-03-02[\s\S]*Stable release entry/);
});

test('runChangelogEngine requires explicit mode', async () => {
  await assert.rejects(
    runChangelogEngine({
      github: makeGithubClient(),
      owner: 'elastic',
      repo: 'terraform-provider-elasticstack',
      fsImpl: makeFs({ 'CHANGELOG.md': baseChangelog }),
      exec: () => '',
    }),
    /mode is required/
  );
});

test('runChangelogEngine rejects unsupported mode', async () => {
  await assert.rejects(
    runChangelogEngine({
      github: makeGithubClient(),
      owner: 'elastic',
      repo: 'terraform-provider-elasticstack',
      mode: 'auto',
      fsImpl: makeFs({ 'CHANGELOG.md': baseChangelog }),
      exec: () => '',
    }),
    /unsupported changelog mode: auto/
  );
});

test('runChangelogEngine requires targetVersion in release mode', async () => {
  await assert.rejects(
    runChangelogEngine({
      github: makeGithubClient(),
      owner: 'elastic',
      repo: 'terraform-provider-elasticstack',
      mode: 'release',
      fsImpl: makeFs({ 'CHANGELOG.md': baseChangelog }),
      exec: () => '',
    }),
    /release mode requires targetVersion/
  );
});

test('runChangelogEngine reports no user-facing changes when all merged PRs are excluded', async () => {
  const fs = makeFs({ 'CHANGELOG.md': baseChangelog });
  const github = makeGithubClient({
    pullRequestsByCommit: {
      sha1: [makePullRequest({
        number: 404,
        html_url: 'https://example.test/pr/404',
        labels: [{ name: 'no-changelog' }],
      })],
    },
  });
  const exec = (command) => {
    if (command.startsWith('git tag --list')) return 'v1.2.3\nv1.2.2\n';
    if (command === 'git log --format=%H v1.2.3..HEAD') return 'sha1\n';
    throw new Error(`unexpected exec command: ${command}`);
  };

  const result = await runChangelogEngine({
    github,
    owner: 'elastic',
    repo: 'terraform-provider-elasticstack',
    mode: 'unreleased',
    changelogPath: 'CHANGELOG.md',
    generatedAt: '2026-04-20T12:00:00.000Z',
    fsImpl: fs,
    exec,
  });

  assert.equal(result.hasUserFacingChanges, false);
  assert.equal(result.includedPullRequests.length, 0);
  assert.equal(result.excludedPullRequests.length, 1);
  assert.equal(result.pullRequests.length, 1);
});

