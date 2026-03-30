import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { selectChangeForPullRequest, selectChangeFromFiles } = require('./select-change.js');

test('selectChangeForPullRequest rejects missing pull request numbers', () => {
  assert.deepEqual(selectChangeForPullRequest({ prNumber: undefined }), {
    selection_status: 'ineligible',
    selection_reason: 'No pull request number in event payload',
    selected_change: '',
  });
});

test('selectChangeFromFiles rejects pull requests without active change files', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'README.md',
        status: 'modified',
      },
    ]),
    {
      selection_status: 'ineligible',
      selection_reason: 'No files under openspec/changes/ (non-archive) found in this PR',
      selected_change: '',
    }
  );
});

test('selectChangeFromFiles selects exactly one modified active change', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'modified',
      },
    ]),
    {
      selection_status: 'eligible',
      selection_reason: 'Selected change: example',
      selected_change: 'example',
    }
  );
});

test('selectChangeFromFiles rejects added files under active changes', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/new.md',
        status: 'added',
      },
    ]),
    {
      selection_status: 'ineligible',
      selection_reason: 'Added file(s) under openspec/changes/: openspec/changes/example/new.md',
      selected_change: '',
    }
  );
});

test('selectChangeFromFiles rejects non-modified statuses under active changes', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'renamed',
      },
    ]),
    {
      selection_status: 'ineligible',
      selection_reason: 'Non-modified file(s) under openspec/changes/: openspec/changes/example/tasks.md (renamed)',
      selected_change: '',
    }
  );
});

test('selectChangeFromFiles ignores archive paths', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/archive/2026-01-01-example/tasks.md',
        status: 'modified',
      },
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'modified',
      },
    ]),
    {
      selection_status: 'eligible',
      selection_reason: 'Selected change: example',
      selected_change: 'example',
    }
  );
});

test('selectChangeFromFiles rejects multiple modified change ids', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/foo/tasks.md',
        status: 'modified',
      },
      {
        filename: 'openspec/changes/bar/tasks.md',
        status: 'modified',
      },
    ]),
    {
      selection_status: 'ineligible',
      selection_reason: 'Multiple active change ids with modified files: foo, bar',
      selected_change: '',
    }
  );
});
