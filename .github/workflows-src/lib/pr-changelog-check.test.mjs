import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const {
  findExistingComment,
  buildPassCommentBody,
  buildNoChangelogPassCommentBody,
  buildFailureCommentBody,
} = require('./pr-changelog-check.js');

const MARKER = '<!-- pr-changelog-check -->';

// ---------------------------------------------------------------------------
// findExistingComment
// ---------------------------------------------------------------------------

test('findExistingComment returns null when comments array is empty', () => {
  assert.equal(findExistingComment([], MARKER), null);
});

test('findExistingComment returns null when no github-actions[bot] comment contains the marker', () => {
  const comments = [
    { user: { login: 'github-actions[bot]' }, body: 'some other content' },
    { user: { login: 'octocat' }, body: MARKER },
  ];
  assert.equal(findExistingComment(comments, MARKER), null);
});

test('findExistingComment returns the comment when a github-actions[bot] comment contains the marker', () => {
  const matching = { id: 42, user: { login: 'github-actions[bot]' }, body: `${MARKER}\nsome content` };
  const comments = [
    { user: { login: 'octocat' }, body: MARKER },
    matching,
  ];
  assert.equal(findExistingComment(comments, MARKER), matching);
});

test('findExistingComment returns null when a different user comment contains the marker', () => {
  const comments = [
    { user: { login: 'dependabot[bot]' }, body: `${MARKER}\nsome content` },
    { user: { login: 'octocat' }, body: `${MARKER}\nsome content` },
  ];
  assert.equal(findExistingComment(comments, MARKER), null);
});

test('findExistingComment returns null when a github-actions[bot] comment exists but body does not contain the marker', () => {
  const comments = [
    { user: { login: 'github-actions[bot]' }, body: 'no marker here' },
  ];
  assert.equal(findExistingComment(comments, MARKER), null);
});

test('findExistingComment returns the first matching comment when multiple matches exist', () => {
  const first = { id: 1, user: { login: 'github-actions[bot]' }, body: `${MARKER}\nfirst` };
  const second = { id: 2, user: { login: 'github-actions[bot]' }, body: `${MARKER}\nsecond` };
  const comments = [first, second];
  assert.equal(findExistingComment(comments, MARKER), first);
});

test('findExistingComment returns null for a comment with user: null (ghost/deleted account)', () => {
  const comments = [
    { user: null, body: `${MARKER}\nsome content` },
  ];
  assert.equal(findExistingComment(comments, MARKER), null);
});

// ---------------------------------------------------------------------------
// buildPassCommentBody
// ---------------------------------------------------------------------------

test('buildPassCommentBody includes the marker in the body', () => {
  const body = buildPassCommentBody(MARKER);
  assert.ok(body.includes(MARKER), 'body should include the marker');
});

test('buildPassCommentBody contains a pass indicator', () => {
  const body = buildPassCommentBody(MARKER);
  assert.ok(body.includes('passed'), 'body should contain "passed"');
});

test('buildPassCommentBody references ## Changelog', () => {
  const body = buildPassCommentBody(MARKER);
  assert.ok(body.includes('## Changelog'), 'body should reference ## Changelog');
});

// ---------------------------------------------------------------------------
// buildNoChangelogPassCommentBody
// ---------------------------------------------------------------------------

test('buildNoChangelogPassCommentBody includes the marker in the body', () => {
  const body = buildNoChangelogPassCommentBody(MARKER);
  assert.ok(body.includes(MARKER), 'body should include the marker');
});

test('buildNoChangelogPassCommentBody contains a pass indicator', () => {
  const body = buildNoChangelogPassCommentBody(MARKER);
  assert.ok(body.includes('passed'), 'body should contain "passed"');
});

test('buildNoChangelogPassCommentBody references no-changelog', () => {
  const body = buildNoChangelogPassCommentBody(MARKER);
  assert.ok(body.includes('no-changelog'), 'body should reference no-changelog');
});

// ---------------------------------------------------------------------------
// buildFailureCommentBody
// ---------------------------------------------------------------------------

test('buildFailureCommentBody includes the marker in the body', () => {
  const body = buildFailureCommentBody(MARKER, ['some error']);
  assert.ok(body.includes(MARKER), 'body should include the marker');
});

test('buildFailureCommentBody includes each error as a bullet point (single error)', () => {
  const body = buildFailureCommentBody(MARKER, ['Missing Customer impact field']);
  assert.ok(body.includes('- Missing Customer impact field'), 'body should list the error as a bullet');
});

test('buildFailureCommentBody includes each error as a bullet point (multiple errors)', () => {
  const errors = ['Missing Customer impact field', 'Summary is required when impact is not none'];
  const body = buildFailureCommentBody(MARKER, errors);
  for (const error of errors) {
    assert.ok(body.includes(`- ${error}`), `body should include bullet for: ${error}`);
  }
});

test('buildFailureCommentBody contains the expected format hint (details block)', () => {
  const body = buildFailureCommentBody(MARKER, ['some error']);
  assert.ok(body.includes('<details>'), 'body should include <details> block');
  assert.ok(body.includes('Expected format'), 'body should include "Expected format"');
  assert.ok(body.includes('Customer impact:'), 'body should include Customer impact field in example');
});

test('buildFailureCommentBody contains the no-changelog bypass hint', () => {
  const body = buildFailureCommentBody(MARKER, ['some error']);
  assert.ok(body.includes('no-changelog'), 'body should mention no-changelog bypass');
});

test('buildFailureCommentBody works with a single error', () => {
  const body = buildFailureCommentBody(MARKER, ['only one error']);
  assert.ok(body.includes('- only one error'));
  // Should not have extra bullet points from an empty array
  const bulletMatches = body.match(/^- /gm);
  assert.equal(bulletMatches?.length, 1, 'should have exactly one bullet point');
});

test('buildFailureCommentBody works with multiple errors', () => {
  const errors = ['error one', 'error two', 'error three'];
  const body = buildFailureCommentBody(MARKER, errors);
  const bulletMatches = body.match(/^- /gm);
  assert.equal(bulletMatches?.length, 3, 'should have exactly three bullet points');
});
