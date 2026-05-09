import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { stripHtmlComments, findResearchComment } = require('./sanitize-context.js');

// ─────────────────────────────────────────────────────────────
// stripHtmlComments
// ─────────────────────────────────────────────────────────────

test('stripHtmlComments returns empty string for empty input', () => {
  assert.equal(stripHtmlComments(''), '');
});

test('stripHtmlComments leaves text with no comments unchanged', () => {
  const text = 'Hello world\nThis has no comments.';
  assert.equal(stripHtmlComments(text), text);
});

test('stripHtmlComments removes a single comment', () => {
  assert.equal(stripHtmlComments('before<!-- hidden -->after'), 'beforeafter');
});

test('stripHtmlComments removes multiple comments', () => {
  assert.equal(
    stripHtmlComments('a<!-- 1 -->b<!-- 2 -->c'),
    'abc',
  );
});

test('stripHtmlComments removes comment at start of string', () => {
  assert.equal(stripHtmlComments('<!-- leading -->text'), 'text');
});

test('stripHtmlComments removes comment at end of string', () => {
  assert.equal(stripHtmlComments('text<!-- trailing -->'), 'text');
});

test('stripHtmlComments removes from unclosed comment to end of string', () => {
  assert.equal(stripHtmlComments('before<!-- never closed'), 'before');
});

test('stripHtmlComments handles only an unclosed comment', () => {
  assert.equal(stripHtmlComments('<!-- unclosed'), '');
});

test('stripHtmlComments handles nested-looking comment structures safely', () => {
  // Non-greedy: stops at the first -->; remainder is preserved
  assert.equal(
    stripHtmlComments('<!-- outer <!-- inner --> -->'),
    ' -->',
  );
});

test('stripHtmlComments handles multiline comments', () => {
  assert.equal(
    stripHtmlComments('start\n<!-- line1\nline2\nline3 -->\nend'),
    'start\n\nend',
  );
});

test('stripHtmlComments handles comment-only string', () => {
  assert.equal(stripHtmlComments('<!-- everything -->'), '');
});

test('stripHtmlComments preserves text between adjacent comments', () => {
  assert.equal(
    stripHtmlComments('<!-- a -->middle<!-- b -->'),
    'middle',
  );
});

test('stripHtmlComments handles dashed text inside comments', () => {
  assert.equal(
    stripHtmlComments('before<!-- -- dashed -->after'),
    'beforeafter',
  );
});

test('stripHtmlComments strips multiple consecutive unclosed comments', () => {
  assert.equal(
    stripHtmlComments('a<!-- one -->b<!-- two'),
    'ab',
  );
});

// ─────────────────────────────────────────────────────────────
// findResearchComment
// ─────────────────────────────────────────────────────────────

test('findResearchComment returns null for empty array', () => {
  assert.equal(findResearchComment([], 'marker'), null);
});

test('findResearchComment returns null when no comments match', () => {
  const comments = [
    { author: 'alice', body: 'hello' },
    { author: 'github-actions[bot]', body: 'no marker here' },
  ];
  assert.equal(findResearchComment(comments, 'MISSING'), null);
});

test('findResearchComment returns single match', () => {
  const comments = [
    { author: 'alice', body: 'hello' },
    { author: 'github-actions[bot]', body: 'contains MARKER-123 here' },
  ];
  const result = findResearchComment(comments, 'MARKER-123');
  assert.equal(result.author, 'github-actions[bot]');
  assert.equal(result.body, 'contains MARKER-123 here');
});

test('findResearchComment returns most recent match when multiple bot comments match', () => {
  const comments = [
    { author: 'github-actions[bot]', body: 'older MARKER-456' },
    { author: 'alice', body: 'interruption' },
    { author: 'github-actions[bot]', body: 'newer MARKER-456' },
  ];
  const result = findResearchComment(comments, 'MARKER-456');
  assert.equal(result.body, 'newer MARKER-456');
});

test('findResearchComment ignores comments with wrong author', () => {
  const comments = [
    { author: 'dependabot[bot]', body: 'marker-X' },
    { author: 'alice', body: 'marker-X' },
  ];
  assert.equal(findResearchComment(comments, 'marker-X'), null);
});

test('findResearchComment ignores comments missing the marker', () => {
  const comments = [
    { author: 'github-actions[bot]', body: 'some other text' },
    { author: 'github-actions[bot]', body: 'another unrelated body' },
  ];
  assert.equal(findResearchComment(comments, 'SEARCH-FOR-THIS'), null);
});

test('findResearchComment handles null or undefined input', () => {
  assert.equal(findResearchComment(null, 'm'), null);
  assert.equal(findResearchComment(undefined, 'm'), null);
});

test('findResearchComment skips comments with null or undefined body', () => {
  const comments = [
    { author: 'github-actions[bot]', body: null },
    { author: 'github-actions[bot]', body: undefined },
    { author: 'github-actions[bot]', body: 'valid marker-Y body' },
  ];
  const result = findResearchComment(comments, 'marker-Y');
  assert.equal(result.body, 'valid marker-Y body');
});

test('findResearchComment treats marker as substring', () => {
  const comments = [
    { author: 'github-actions[bot]', body: 'abc SUB-789 def' },
  ];
  const result = findResearchComment(comments, 'SUB-789');
  assert.ok(result);
  assert.equal(result.body, 'abc SUB-789 def');
});

test('findResearchComment returns last match even when earlier ones have same marker', () => {
  const comments = [
    { author: 'github-actions[bot]', body: 'first REPEAT' },
    { author: 'github-actions[bot]', body: 'second REPEAT' },
    { author: 'github-actions[bot]', body: 'third REPEAT' },
  ];
  const result = findResearchComment(comments, 'REPEAT');
  assert.equal(result.body, 'third REPEAT');
});
