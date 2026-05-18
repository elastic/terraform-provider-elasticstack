import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  buildEvidenceManifest,
  buildPullRequestEvidence,
  buildTargetSection,
  classifyPullRequestForChangelog,
  parseCommitShas,
  selectMergedPullRequests,
} = require('./changelog-pr-evidence.js');

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
    ...overrides,
  };
}

test('classifyPullRequestForChangelog treats user-facing labels as user-facing', () => {
  const result = classifyPullRequestForChangelog(
    makePullRequest({ labels: [{ name: 'bug' }, { name: 'documentation' }] }),
    [{ filename: 'docs/guide.md' }]
  );

  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusionRationale, /bug/);
  assert.equal(result.exclusionRationale, null);
});

test('classifyPullRequestForChangelog treats openspec-only changes as internal', () => {
  const result = classifyPullRequestForChangelog(makePullRequest(), [
    { filename: 'openspec/changes/example/tasks.md' },
  ]);

  assert.equal(result.classification, 'internal');
  assert.match(result.exclusionRationale, /openspec/);
});

test('classifyPullRequestForChangelog treats provider path changes as user-facing', () => {
  const result = classifyPullRequestForChangelog(
    makePullRequest({ labels: [{ name: 'internal' }] }),
    [{ filename: 'internal/provider/resource.go' }]
  );

  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusionRationale, /provider implementation paths/);
});

test('classifyPullRequestForChangelog treats automated pull requests as internal', () => {
  const result = classifyPullRequestForChangelog(
    makePullRequest({ user: { login: 'dependabot[bot]' } }),
    [{ filename: 'go.mod' }]
  );

  assert.equal(result.classification, 'internal');
  assert.match(result.exclusionRationale, /dependabot/);
});

test('parseCommitShas trims and filters blank lines', () => {
  assert.deepEqual(parseCommitShas('\nabc\n\ndef \n'), ['abc', 'def']);
});

test('selectMergedPullRequests keeps only merged PRs and de-duplicates by number', () => {
  const merged = selectMergedPullRequests([
    makePullRequest({ number: 1 }),
    makePullRequest({ number: 2, state: 'open' }),
    makePullRequest({ number: 1, title: 'Duplicate entry' }),
    makePullRequest({ number: 3, merged_at: null }),
  ]);

  assert.deepEqual(
    merged.map((pr) => pr.number),
    [1]
  );
});

test('buildPullRequestEvidence produces normalized evidence records', () => {
  const evidence = buildPullRequestEvidence(
    makePullRequest({
      labels: [{ name: 'enhancement' }],
      user: { login: 'contributor' },
    }),
    [{ filename: 'pkg/example.go' }, { filename: 'docs/guide.md' }]
  );

  assert.deepEqual(evidence, {
    number: 42,
    title: 'Example PR',
    url: 'https://example.test/pr/42',
    merge_commit_sha: 'abc123',
    author: 'contributor',
    labels: ['enhancement'],
    touched_files: ['pkg/example.go', 'docs/guide.md'],
    classification: 'user-facing',
    inclusion_rationale: 'Has user-facing label(s): enhancement',
    exclusion_rationale: null,
  });
});

test('buildTargetSection formats release and unreleased headings', () => {
  assert.equal(
    buildTargetSection({
      mode: 'release',
      targetVersion: '1.2.3',
      date: '2026-04-17T08:00:00.000Z',
    }),
    '## [1.2.3] - 2026-04-17'
  );
  assert.equal(buildTargetSection({ mode: 'unreleased', targetVersion: '1.2.3' }), '## [Unreleased]');
});

test('buildEvidenceManifest calculates counts and target section', () => {
  const manifest = buildEvidenceManifest({
    mode: 'release',
    targetVersion: '1.2.3',
    previousTag: 'v1.2.2',
    compareRange: 'v1.2.2..HEAD',
    generatedAt: '2026-04-17T08:00:00.000Z',
    evidence: [
      { classification: 'user-facing' },
      { classification: 'internal' },
      { classification: 'uncertain' },
    ],
  });

  assert.deepEqual(manifest, {
    generated_at: '2026-04-17T08:00:00.000Z',
    mode: 'release',
    target_section: '## [1.2.3] - 2026-04-17',
    target_section_mode: 'release',
    target_version: '1.2.3',
    previous_tag: 'v1.2.2',
    compare_range: 'v1.2.2..HEAD',
    pr_count: 3,
    user_facing_count: 1,
    internal_count: 1,
    uncertain_count: 1,
    pull_requests: [
      { classification: 'user-facing' },
      { classification: 'internal' },
      { classification: 'uncertain' },
    ],
  });
});
