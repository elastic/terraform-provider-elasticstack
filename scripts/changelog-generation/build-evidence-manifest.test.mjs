import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

const { classifyPR } = require(path.resolve(__dirname, 'build-evidence-manifest.js'));

// ---------------------------------------------------------------------------
// classifyPR tests
// ---------------------------------------------------------------------------

test('classifyPR: dependabot PR → internal', () => {
  const pr = { user: { login: 'dependabot[bot]' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /dependabot/);
  assert.equal(result.inclusion_rationale, null);
});

test('classifyPR: github-actions bot PR → internal', () => {
  const pr = { user: { login: 'github-actions[bot]' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /github-actions/);
});

test('classifyPR: openspec-only files → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [
    { filename: 'openspec/specs/foo.md' },
    { filename: 'openspec/specs/bar.md' },
  ];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /openspec/);
});

test('classifyPR: enhancement label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'enhancement' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /enhancement/);
  assert.equal(result.exclusion_rationale, null);
});

test('classifyPR: bug label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'bug' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /bug/);
});

test('classifyPR: breaking-change label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'breaking-change' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: internal label, no provider code → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'chore' }] };
  const files = [{ filename: 'docs/guide.md' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /chore/);
});

test('classifyPR: internal label but touches provider code → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'chore' }] };
  const files = [{ filename: 'internal/provider/resource.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: no labels, touches provider code → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'pkg/utils/helper.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /provider implementation/);
});

test('classifyPR: no labels, no provider code → uncertain', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'docs/some-doc.md' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'uncertain');
  assert.match(result.inclusion_rationale, /uncertain/);
});

test('classifyPR: go.mod file → user-facing (provider path)', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'go.mod' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: new-resource label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'new-resource' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: mixed openspec and provider files → not openspec-only', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [
    { filename: 'openspec/specs/foo.md' },
    { filename: 'internal/provider/resource.go' },
  ];
  const result = classifyPR(pr, files);
  // Should NOT be classified as internal due to openspec-only rule
  // Has provider code → user-facing
  assert.equal(result.classification, 'user-facing');
});

// ---------------------------------------------------------------------------
// classifyPR — additional label coverage
// ---------------------------------------------------------------------------

test('classifyPR: feature label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'feature' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /feature/);
});

test('classifyPR: deprecation label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'deprecation' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: new-data-source label → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'new-data-source' }] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: documentation label, no provider code → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'documentation' }] };
  const files = [{ filename: 'docs/resources/foo.md' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /documentation/);
});

test('classifyPR: dependencies label, no provider code → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'dependencies' }] };
  const files = [{ filename: 'package.json' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
});

test('classifyPR: ci label, no provider code → internal', () => {
  const pr = { user: { login: 'user123' }, labels: [{ name: 'ci' }] };
  const files = [{ filename: '.github/workflows/test.yml' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'internal');
});

test('classifyPR: libs/ path → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'libs/elasticstack/client.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
  assert.match(result.inclusion_rationale, /provider implementation/);
});

test('classifyPR: provider/ path → user-facing', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'provider/provider.go' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: go.sum file → user-facing (provider path)', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const files = [{ filename: 'go.sum' }];
  const result = classifyPR(pr, files);
  assert.equal(result.classification, 'user-facing');
});

test('classifyPR: no labels, no files → uncertain', () => {
  const pr = { user: { login: 'user123' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'uncertain');
  assert.match(result.inclusion_rationale, /uncertain/);
});

test('classifyPR: dependabot (non-bot suffix) → internal', () => {
  const pr = { user: { login: 'dependabot' }, labels: [] };
  const result = classifyPR(pr, []);
  assert.equal(result.classification, 'internal');
  assert.match(result.exclusion_rationale, /dependabot/);
});

// ---------------------------------------------------------------------------
// Manifest target section and mode logic — tested via classifyPR + inline assembly
// (buildManifest itself requires live git + GitHub API; we test the pure logic here)
// ---------------------------------------------------------------------------

// Helper: assemble a manifest structure the same way buildManifest does,
// but from pre-classified PR data so we can verify the manifest schema.
function assembleManifest({ mode, targetVersion, previousTag, compareRange, evidence }) {
  const today = new Date().toISOString().split('T')[0];
  const targetSection =
    mode === 'release'
      ? `## [${targetVersion}] - ${today}`
      : '## [Unreleased]';

  return {
    generated_at: new Date().toISOString(),
    mode,
    target_section: targetSection,
    target_section_mode: mode,
    target_version: targetVersion,
    previous_tag: previousTag,
    compare_range: compareRange,
    pr_count: evidence.length,
    user_facing_count: evidence.filter((e) => e.classification === 'user-facing').length,
    internal_count: evidence.filter((e) => e.classification === 'internal').length,
    uncertain_count: evidence.filter((e) => e.classification === 'uncertain').length,
    pull_requests: evidence,
  };
}

test('manifest assembly: unreleased mode → target_section is ## [Unreleased]', () => {
  const manifest = assembleManifest({
    mode: 'unreleased',
    targetVersion: '',
    previousTag: 'v0.14.3',
    compareRange: 'v0.14.3..HEAD',
    evidence: [],
  });

  assert.equal(manifest.mode, 'unreleased');
  assert.equal(manifest.target_section, '## [Unreleased]');
  assert.equal(manifest.target_section_mode, 'unreleased');
  assert.equal(manifest.previous_tag, 'v0.14.3');
  assert.equal(manifest.compare_range, 'v0.14.3..HEAD');
  assert.equal(manifest.pr_count, 0);
  assert.equal(manifest.user_facing_count, 0);
  assert.equal(manifest.internal_count, 0);
  assert.equal(manifest.uncertain_count, 0);
  assert.ok(typeof manifest.generated_at === 'string');
  assert.ok(Array.isArray(manifest.pull_requests));
});

test('manifest assembly: release mode → target_section includes version and today date', () => {
  const today = new Date().toISOString().split('T')[0];
  const manifest = assembleManifest({
    mode: 'release',
    targetVersion: '0.14.4',
    previousTag: 'v0.14.3',
    compareRange: 'v0.14.3..HEAD',
    evidence: [],
  });

  assert.equal(manifest.mode, 'release');
  assert.equal(manifest.target_section, `## [0.14.4] - ${today}`);
  assert.equal(manifest.target_section_mode, 'release');
  assert.equal(manifest.target_version, '0.14.4');
});

test('manifest assembly: counts reflect classified evidence correctly', () => {
  const prs = [
    { number: 1, title: 'Feature A', url: 'https://x', merge_commit_sha: 'abc', author: 'user', labels: ['enhancement'], touched_files: [], classification: 'user-facing', inclusion_rationale: 'label', exclusion_rationale: null },
    { number: 2, title: 'Chore B', url: 'https://x', merge_commit_sha: 'def', author: 'user', labels: ['chore'], touched_files: [], classification: 'internal', inclusion_rationale: null, exclusion_rationale: 'chore label' },
    { number: 3, title: 'Unknown C', url: 'https://x', merge_commit_sha: 'ghi', author: 'user', labels: [], touched_files: [], classification: 'uncertain', inclusion_rationale: 'uncertain', exclusion_rationale: null },
    { number: 4, title: 'Feature D', url: 'https://x', merge_commit_sha: 'jkl', author: 'user', labels: ['bug'], touched_files: [], classification: 'user-facing', inclusion_rationale: 'label', exclusion_rationale: null },
  ];

  const manifest = assembleManifest({
    mode: 'unreleased',
    targetVersion: '',
    previousTag: 'v0.14.3',
    compareRange: 'v0.14.3..HEAD',
    evidence: prs,
  });

  assert.equal(manifest.pr_count, 4);
  assert.equal(manifest.user_facing_count, 2);
  assert.equal(manifest.internal_count, 1);
  assert.equal(manifest.uncertain_count, 1);
  // Counts must add up
  assert.equal(
    manifest.user_facing_count + manifest.internal_count + manifest.uncertain_count,
    manifest.pr_count
  );
});

test('manifest assembly: pull_request entries have required fields', () => {
  // Validate that each PR entry in the manifest has all required schema fields.
  const pr = {
    number: 42,
    title: 'Add cool feature',
    url: 'https://github.com/org/repo/pull/42',
    merge_commit_sha: 'a1b2c3d4',
    author: 'contributor',
    labels: ['enhancement'],
    touched_files: ['internal/provider/resource.go'],
    classification: 'user-facing',
    inclusion_rationale: 'Has user-facing label(s): enhancement',
    exclusion_rationale: null,
  };

  const manifest = assembleManifest({
    mode: 'unreleased',
    targetVersion: '',
    previousTag: '',
    compareRange: 'HEAD',
    evidence: [pr],
  });

  assert.equal(manifest.pull_requests.length, 1);
  const entry = manifest.pull_requests[0];
  assert.equal(typeof entry.number, 'number');
  assert.equal(typeof entry.title, 'string');
  assert.equal(typeof entry.url, 'string');
  assert.equal(typeof entry.merge_commit_sha, 'string');
  assert.equal(typeof entry.author, 'string');
  assert.ok(Array.isArray(entry.labels));
  assert.ok(Array.isArray(entry.touched_files));
  assert.ok(['user-facing', 'internal', 'uncertain'].includes(entry.classification));
});

test('manifest assembly: classifyPR integrates correctly with manifest counts', () => {
  // Test classifyPR + manifest assembly together to verify the full evidence pipeline
  const testPRs = [
    { pr: { user: { login: 'user1' }, labels: [{ name: 'enhancement' }] }, files: [] },
    { pr: { user: { login: 'dependabot[bot]' }, labels: [] }, files: [] },
    { pr: { user: { login: 'user2' }, labels: [] }, files: [{ filename: 'docs/guide.md' }] },
  ];

  const evidence = testPRs.map(({ pr, files }, i) => {
    const { classification, inclusion_rationale, exclusion_rationale } = classifyPR(pr, files);
    return {
      number: i + 100,
      title: `PR #${i + 100}`,
      url: `https://example.com/${i + 100}`,
      merge_commit_sha: `sha${i}`,
      author: pr.user.login,
      labels: (pr.labels ?? []).map((l) => l.name),
      touched_files: files.map((f) => f.filename),
      classification,
      inclusion_rationale,
      exclusion_rationale,
    };
  });

  const manifest = assembleManifest({
    mode: 'unreleased',
    targetVersion: '',
    previousTag: 'v0.14.0',
    compareRange: 'v0.14.0..HEAD',
    evidence,
  });

  // PR 100: enhancement → user-facing
  // PR 101: dependabot → internal
  // PR 102: no labels, docs only → uncertain
  assert.equal(manifest.user_facing_count, 1);
  assert.equal(manifest.internal_count, 1);
  assert.equal(manifest.uncertain_count, 1);
  assert.equal(manifest.pr_count, 3);
});
