import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { setPhaseLabel } = require('./set-phase-label.js');

// ---------------------------------------------------------------------------
// set-phase-label.js — setPhaseLabel
// ---------------------------------------------------------------------------

test('setPhaseLabel returns phase_label_set false when issueNumber is undefined', async () => {
  const result = await setPhaseLabel({
    github: {},
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: undefined,
    phaseLabelName: 'phase-research',
  });
  assert.equal(result.phase_label_set, false);
  assert.ok(
    result.reason.includes('No issue number'),
    `expected "No issue number" in reason, got: ${result.reason}`,
  );
});

test('setPhaseLabel returns phase_label_set false when phaseLabelName is empty', async () => {
  const result = await setPhaseLabel({
    github: {},
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: '   ',
  });
  assert.equal(result.phase_label_set, false);
  assert.ok(
    result.reason.includes('No phase label name'),
    `expected "No phase label name" in reason, got: ${result.reason}`,
  );
});

test('setPhaseLabel adds label and returns true when no stale labels exist', async () => {
  const addedLabels = [];
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async (args) => {
          addedLabels.push(...args.labels);
          return {};
        },
        listLabelsOnIssue: async () => ({ data: [{ name: 'bug' }, { name: 'help wanted' }] }),
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-research',
  });
  assert.equal(result.phase_label_set, true);
  assert.equal(result.phase_label_name, 'phase-research');
  assert.deepEqual(result.stale_labels_removed, []);
  assert.ok(result.reason.includes('phase-research'));
  assert.deepEqual(addedLabels, ['phase-research']);
});

test('setPhaseLabel removes stale phase-* labels and keeps the new one', async () => {
  const addedLabels = [];
  const removedLabels = [];
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async (args) => {
          addedLabels.push(...args.labels);
          return {};
        },
        listLabelsOnIssue: async () => ({
          data: [
            { name: 'phase-research' },
            { name: 'phase-specification' },
            { name: 'bug' },
          ],
        }),
        removeLabel: async (args) => {
          removedLabels.push(args.name);
          return {};
        },
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-coding',
  });
  assert.equal(result.phase_label_set, true);
  assert.equal(result.phase_label_name, 'phase-coding');
  assert.deepEqual(result.stale_labels_removed, ['phase-research', 'phase-specification']);
  assert.deepEqual(addedLabels, ['phase-coding']);
  assert.deepEqual(removedLabels, ['phase-research', 'phase-specification']);
  assert.ok(result.reason.includes('phase-coding'));
  assert.ok(result.reason.includes('phase-research'));
});

test('setPhaseLabel treats 404 on removeLabel as success', async () => {
  const err = new Error('Not Found');
  err.status = 404;
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async () => ({}) ,
        listLabelsOnIssue: async () => ({ data: [{ name: 'phase-research' }] }),
        removeLabel: async () => {
          throw err;
        },
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-coding',
  });
  assert.equal(result.phase_label_set, true);
  assert.equal(result.stale_labels_removed.length, 1);
  assert.ok(result.reason.includes('phase-coding'));
});

test('setPhaseLabel returns failure when addLabels fails', async () => {
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async () => {
          throw new Error('Validation Failed');
        },
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-research',
  });
  assert.equal(result.phase_label_set, false);
  assert.ok(result.reason.includes('Validation Failed'));
});

test('setPhaseLabel returns failure when non-404 removeLabel fails', async () => {
  const err = new Error('Internal Server Error');
  err.status = 500;
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async () => ({}) ,
        listLabelsOnIssue: async () => ({ data: [{ name: 'phase-research' }] }),
        removeLabel: async () => {
          throw err;
        },
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-coding',
  });
  assert.equal(result.phase_label_set, true);
  assert.ok(result.reason.includes('Internal Server Error'));
  assert.ok(result.reason.includes('phase-research'));
});

test('setPhaseLabel preserves phase_label_set true and empty stale_labels_removed when listLabelsOnIssue throws', async () => {
  const mockGithub = {
    rest: {
      issues: {
        addLabels: async () => ({}),
        listLabelsOnIssue: async () => {
          throw new Error('API rate limit exceeded');
        },
      },
    },
  };
  const result = await setPhaseLabel({
    github: mockGithub,
    context: { repo: { owner: 'owner', repo: 'repo' } },
    issueNumber: 42,
    phaseLabelName: 'phase-research',
  });
  assert.equal(result.phase_label_set, true);
  assert.equal(result.phase_label_name, 'phase-research');
  assert.deepEqual(result.stale_labels_removed, []);
  assert.ok(result.reason.includes('Added label phase-research but failed to list current labels'));
  assert.ok(result.reason.includes('API rate limit exceeded'));
});
