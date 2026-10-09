import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const { splitBody, joinParts, PART_START, PART_END, MAX_PART_BYTES, MAX_PARTS, MAX_CHARS } = require('../research-factory/emit-research-comment.js');

const script = path.join(path.dirname(fileURLToPath(import.meta.url)), '../research-factory/emit-research-comment.js');
const bytes = (s) => Buffer.byteLength(s, 'utf8');

// The safe-output platform trims every string input; simulate it. Only the whole
// body's own leading/trailing whitespace is lost, which is irrelevant for markdown.
const trimParts = (parts) => Object.fromEntries(Object.entries(parts).map(([k, v]) => [k, v.trim()]));
const roundTrip = (text) => joinParts(trimParts(splitBody(text)));

test('constants match the per-input limit and part count', () => {
  assert.equal(MAX_PART_BYTES, 10000);
  assert.equal(MAX_PARTS, 7);
  assert.equal(MAX_CHARS, 60000);
});

test('sentinels are not HTML comments, which the platform strips', () => {
  assert.doesNotMatch(PART_START + PART_END, /<!--|-->/);
  assert.equal(PART_START, PART_START.trim());
  assert.equal(PART_END, PART_END.trim());
});

test('a small body yields only body with no sentinels', () => {
  assert.deepEqual(splitBody('hello\nworld\n'), { body: 'hello\nworld\n' });
  assert.equal(roundTrip('hello\nworld\n'), 'hello\nworld');
});

function bigDoc(lines = 1500) {
  return Array.from({ length: lines }, (_, i) => `line ${i} with some padding text`).join('\n') + '\n';
}

test('a large body splits within the byte budget, counting sentinels, and survives trimming', () => {
  const text = bigDoc();
  const parts = splitBody(text);
  const names = Object.keys(parts);
  assert.ok(names.length > 1);
  assert.deepEqual(names, ['body', 'body_2', 'body_3', 'body_4', 'body_5', 'body_6', 'body_7'].slice(0, names.length));
  names.forEach((name, i) => {
    assert.ok(bytes(parts[name]) <= MAX_PART_BYTES, `${name} within budget`);
    assert.equal(parts[name].startsWith(PART_START), i > 0);
    assert.equal(parts[name].endsWith(PART_END), i < names.length - 1);
  });
  assert.equal(roundTrip(text), text.trim());
});

test('blank lines at chunk boundaries are preserved', () => {
  const paragraph = `${'p'.repeat(4000)}\n\n`;
  const text = paragraph.repeat(6);
  assert.equal(roundTrip(text), text.trim());
});

test('chunks that end with only a newline are preserved', () => {
  const text = `${'a'.repeat(9000)}\n${'b'.repeat(9000)}\n${'c'.repeat(100)}\n`;
  assert.equal(roundTrip(text), text.trim());
});

test('CRLF line endings are preserved', () => {
  const text = Array.from({ length: 1200 }, (_, i) => `row ${i} some content here`).join('\r\n') + '\r\n';
  assert.equal(roundTrip(text), text.trim());
});

test('unicode and emoji are never split and stay within budget', () => {
  const text = `${'é'.repeat(4000)}\n\n${'日本語'.repeat(1500)}\n\n${'😀'.repeat(2000)}\n`;
  const parts = splitBody(text);
  for (const part of Object.values(parts)) {
    assert.ok(bytes(part) <= MAX_PART_BYTES);
    assert.ok(!part.includes('�'));
  }
  assert.equal(roundTrip(text), text.trim());
});

test('a single overlong line is split at a code point boundary', () => {
  const text = '😀'.repeat(6000);
  const parts = splitBody(text);
  assert.ok(Object.keys(parts).length >= 3);
  for (const part of Object.values(parts)) {
    assert.ok(bytes(part) <= MAX_PART_BYTES);
    assert.ok(!part.includes('�'));
  }
  assert.equal(roundTrip(text), text);
});

function syntheticDraft() {
  const meta = JSON.stringify({ schema_version: '1.1', gate: { outcome: 'research-needs-human' } }, null, 2);
  const filler = (n) => Array.from({ length: n }, (_, i) => `- bullet ${i}: ${'text '.repeat(12)}`).join('\n');
  return `## Implementation research\n\n### Problem framing\n\n${filler(120)}\n\n### Open questions\n\n- **OQ-1** question\n\n### Out of scope\n\n${filler(80)}\n\n### References\n\n${filler(60)}\n\n<details>\n<summary>🤖 Pipeline metadata</summary>\n\n\`\`\`json\n${meta}\n\`\`\`\n\n</details>\n`;
}

test('a code fence split across chunks is preserved', () => {
  const body = Array.from({ length: 900 }, (_, i) => `    "key_${i}": "value ${i}"`).join(',\n');
  const text = `intro\n\n\`\`\`json\n{\n${body}\n}\n\`\`\`\n\noutro\n`;
  assert.ok(Object.keys(splitBody(text)).length > 1);
  assert.equal(roundTrip(text), text.trim());
});

test('a realistic draft with headings and a metadata fence at a boundary survives', () => {
  const text = syntheticDraft();
  assert.ok(text.length > 10000);
  assert.equal(roundTrip(text), text.trim());
});

test('headings and the json fence opener survive when the cut lands right before them', () => {
  const lead = `${'x'.repeat(9900 - PART_START.length - PART_END.length - 1)}\n`;
  const text = `${lead}\`\`\`json\n{"a": 1}\n\`\`\`\n### Heading\n\ntail\n`;
  assert.equal(roundTrip(text), text.trim());
});

test('joinParts tolerates parts without sentinels', () => {
  assert.equal(joinParts({ body: 'a\n', body_2: 'b\n' }), 'a\nb\n');
});

test('joinParts orders by part number and ignores empty parts', () => {
  assert.equal(joinParts({ body_3: 'C', body: 'A', body_2: '', body_5: 'E' }), 'ACE');
});

test('a body needing more than seven parts is rejected', () => {
  const text = `${'a'.repeat(5999)}\n`.repeat(10);
  assert.equal(text.length, MAX_CHARS);
  assert.throws(() => splitBody(text), /parts/i);
});

test('a body above the character cap is rejected', () => {
  assert.throws(() => splitBody('a'.repeat(MAX_CHARS + 1)), /60000|characters/i);
});

test('the CLI prints the JSON object for a file', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'emit-'));
  const file = path.join(dir, 'draft.md');
  fs.writeFileSync(file, '# Title\n\nbody\n');
  const result = spawnSync('node', [script, file], { encoding: 'utf8' });
  assert.equal(result.status, 0);
  assert.deepEqual(JSON.parse(result.stdout), { body: '# Title\n\nbody\n' });
});

test('the CLI exits non-zero with a clear message when the body is too large', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'emit-'));
  const file = path.join(dir, 'draft.md');
  fs.writeFileSync(file, 'a'.repeat(MAX_CHARS + 1));
  const result = spawnSync('node', [script, file], { encoding: 'utf8' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /characters/i);
  assert.equal(result.stdout, '');
});

test('the CLI errors when no file is given', () => {
  const result = spawnSync('node', [script], { encoding: 'utf8' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /usage/i);
});
