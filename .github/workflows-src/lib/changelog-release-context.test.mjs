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

test('resolveReleaseMode defaults non-dispatch events to unreleased', () => {
  assert.deepEqual(
    resolveReleaseMode({
      eventName: 'schedule',
    }),
    {
      mode: 'unreleased',
      targetVersion: '',
      targetBranch: 'generated-changelog',
    }
  );
});

test('resolveReleaseMode defaults workflow_dispatch to unreleased', () => {
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

test('resolveReleaseMode uses explicit workflow_dispatch release inputs', () => {
  assert.deepEqual(
    resolveReleaseMode({
      eventName: 'workflow_dispatch',
      dispatchMode: 'release',
      targetVersion: '2.0.0',
    }),
    {
      mode: 'release',
      targetVersion: '2.0.0',
      targetBranch: 'prep-release-2.0.0',
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

test('buildReleaseContext combines workflow_dispatch release mode and previous tag selection', () => {
  assert.deepEqual(
    buildReleaseContext({
      eventName: 'workflow_dispatch',
      dispatchMode: 'release',
      targetVersion: '2.0.0',
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

test('buildReleaseContext keeps current release tag when it is not present locally', () => {
  assert.deepEqual(
    buildReleaseContext({
      eventName: 'workflow_dispatch',
      dispatchMode: 'release',
      targetVersion: '2.0.0',
      tags: ['v1.9.0', 'v1.8.0'],
    }),
    {
      mode: 'release',
      targetVersion: '2.0.0',
      targetBranch: 'prep-release-2.0.0',
      previousTag: 'v1.9.0',
      excludedTag: 'v2.0.0',
      excludedCurrentTag: false,
      compareRange: 'v1.9.0..HEAD',
    }
  );
});
