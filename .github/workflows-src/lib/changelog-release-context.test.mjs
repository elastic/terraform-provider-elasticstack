import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  buildCompareRange,
  buildReleaseContext,
  parseSemverTags,
  resolveReleaseMode,
  selectPreviousTag,
} = require('./changelog-release-context.js');

test('resolveReleaseMode detects prep-release pull request branches', () => {
  assert.deepEqual(
    resolveReleaseMode({
      eventName: 'pull_request',
      headBranch: 'prep-release-1.2.3',
    }),
    {
      mode: 'release',
      targetVersion: '1.2.3',
      targetBranch: 'prep-release-1.2.3',
    }
  );
});

test('resolveReleaseMode defaults to unreleased for non-release branches', () => {
  assert.deepEqual(
    resolveReleaseMode({
      eventName: 'workflow_dispatch',
      headBranch: 'main',
    }),
    {
      mode: 'unreleased',
      targetVersion: '',
      targetBranch: 'generated-changelog',
    }
  );
});

test('parseSemverTags keeps only strict vX.Y.Z tags', () => {
  assert.deepEqual(
    parseSemverTags('v1.2.3\nv1.2.3-rc1\nfoo\nv2.0.0\n'),
    ['v1.2.3', 'v2.0.0']
  );
});

test('selectPreviousTag excludes the current release tag when present', () => {
  assert.deepEqual(
    selectPreviousTag({
      tags: ['v1.2.3', 'v1.2.2', 'v1.2.1'],
      mode: 'release',
      targetVersion: '1.2.3',
    }),
    {
      previousTag: 'v1.2.2',
      excludedTag: 'v1.2.3',
      excludedCurrentTag: true,
    }
  );
});

test('buildCompareRange falls back to HEAD when no previous tag exists', () => {
  assert.equal(buildCompareRange(''), 'HEAD');
  assert.equal(buildCompareRange('v1.2.2'), 'v1.2.2..HEAD');
});

test('buildReleaseContext combines mode and previous tag selection', () => {
  assert.deepEqual(
    buildReleaseContext({
      eventName: 'pull_request_target',
      headBranch: 'prep-release-2.0.0',
      tags: ['v2.0.0', 'v1.9.0'],
    }),
    {
      mode: 'release',
      targetVersion: '2.0.0',
      targetBranch: 'prep-release-2.0.0',
      previousTag: 'v1.9.0',
      excludedTag: 'v2.0.0',
      excludedCurrentTag: true,
      compareRange: 'v1.9.0..HEAD',
    }
  );
});
