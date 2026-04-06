import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { selectChangeForPullRequest, selectChangeFromFiles } = require('./select-change.js');

const ineligibleBase = {
  selection_status: 'ineligible',
  selected_change: '',
  review_disposition: '',
  disposition_reason: '',
};

const eligibleBase = {
  selection_status: 'eligible',
};

test('selectChangeForPullRequest rejects missing pull request numbers', () => {
  assert.deepEqual(selectChangeForPullRequest({ prNumber: undefined }), {
    ...ineligibleBase,
    selection_reason: 'No pull request number in event payload',
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
      ...ineligibleBase,
      selection_reason: 'No files under openspec/changes/ (non-archive) found in this PR',
    }
  );
});

test('selectChangeFromFiles selects modified-only active change as approval-eligible', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'modified',
      },
    ]),
    {
      ...eligibleBase,
      selection_reason: 'Selected change: example',
      selected_change: 'example',
      review_disposition: 'approval-eligible',
      disposition_reason:
        'Every file under the selected change is a modification. APPROVE is permitted when verification finds zero CRITICAL issues and zero unassociated files.',
    }
  );
});

test('selectChangeFromFiles selects net-new added files as comment-only', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/proposal.md',
        status: 'added',
      },
    ]),
    {
      ...eligibleBase,
      selection_reason: 'Selected change: example',
      selected_change: 'example',
      review_disposition: 'comment-only',
      disposition_reason:
        'The selected change includes one or more added files (net-new spec change material). APPROVE is not permitted; submit COMMENT only, even if verification passes with no blocking issues.',
    }
  );
});

test('selectChangeFromFiles selects mixed added and modified under one change as comment-only', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/proposal.md',
        status: 'added',
      },
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'modified',
      },
    ]),
    {
      ...eligibleBase,
      selection_reason: 'Selected change: example',
      selected_change: 'example',
      review_disposition: 'comment-only',
      disposition_reason:
        'The selected change includes one or more added files (net-new spec change material). APPROVE is not permitted; submit COMMENT only, even if verification passes with no blocking issues.',
    }
  );
});

test('selectChangeFromFiles rejects renamed status under active changes', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'renamed',
      },
    ]),
    {
      ...ineligibleBase,
      selection_reason:
        'Unsupported file status under openspec/changes/: openspec/changes/example/tasks.md (renamed)',
    }
  );
});

test('selectChangeFromFiles rejects removed status under active changes', () => {
  assert.deepEqual(
    selectChangeFromFiles([
      {
        filename: 'openspec/changes/example/tasks.md',
        status: 'removed',
      },
    ]),
    {
      ...ineligibleBase,
      selection_reason:
        'Unsupported file status under openspec/changes/: openspec/changes/example/tasks.md (removed)',
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
      ...eligibleBase,
      selection_reason: 'Selected change: example',
      selected_change: 'example',
      review_disposition: 'approval-eligible',
      disposition_reason:
        'Every file under the selected change is a modification. APPROVE is permitted when verification finds zero CRITICAL issues and zero unassociated files.',
    }
  );
});

test('selectChangeFromFiles rejects multiple active change ids', () => {
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
      ...ineligibleBase,
      selection_reason: 'Multiple active change ids: bar, foo',
    }
  );
});
