import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { sanitizers, sanitizeParts } from './research-factory-sanitizer-helper.mjs';

const require = createRequire(import.meta.url);
const { splitBody, joinParts, PART_START, PART_END, MAX_PART_BYTES, MAX_PARTS, MAX_CHARS } = require('../research-factory/emit-research-comment.js');

const script = path.join(path.dirname(fileURLToPath(import.meta.url)), '../research-factory/emit-research-comment.js');
const bytes = (s) => Buffer.byteLength(s, 'utf8');

const trimParts = (parts) => sanitizeParts(parts, (v) => v.trim());
const roundTrip = (text, fn = (v) => v.trim()) => joinParts(sanitizeParts(splitBody(text), fn));
const normalise = (t) => t.replace(/\r\n/g, '\n').trim();

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

// ---------------------------------------------------------------------------
// Fence-aware splitting and sanitizer round trips
// ---------------------------------------------------------------------------

function fencedDoc({ fence = '```', leadBytes }) {
  const meta = JSON.stringify({ schema_version: '1.1', gate: { outcome: 'research-needs-human', note: 'x'.repeat(300) } }, null, 2);
  const lead = `${'filler line to push the fence to the boundary\n'.repeat(Math.ceil(leadBytes / 46))}`;
  return `${lead}\n### Open questions\n\n- **OQ-1** q\n\n<details>\n<summary>Pipeline metadata</summary>\n\n${fence}json\n${meta}\n${fence}\n\n</details>\n`;
}

for (const [name, fn] of sanitizers) {
  test(`${name}: a metadata fence landing at every possible boundary round-trips`, () => {
    for (let leadBytes = 9800; leadBytes <= 10100; leadBytes += 7) {
      const text = fencedDoc({ leadBytes }) + 'tail\n'.repeat(5);
      assert.equal(roundTrip(text, fn), normalise(text), `leadBytes=${leadBytes}`);
    }
  });

  test(`${name}: tilde fences, multiple fences and a long fenced block round-trip`, () => {
    const block = (f, n) => `${f}text\n${Array.from({ length: n }, (_, i) => `code ${i}`).join('\n')}\n${f}\n\n`;
    const text = `${'prose line\n'.repeat(700)}\n${block('~~~', 300)}${'more prose\n'.repeat(500)}\n${block('```', 400)}${block('````', 50)}end\n`;
    assert.ok(Object.keys(splitBody(text)).length > 1);
    assert.equal(roundTrip(text, fn), normalise(text));
  });

  test(`${name}: no part ends inside a fenced block`, () => {
    const text = fencedDoc({ leadBytes: 9900 });
    for (const part of Object.values(splitBody(text))) {
      const opens = (part.match(/^```/gm) || []).length;
      assert.equal(opens % 2, 0, 'balanced fences in every part');
    }
  });
}

test('a fenced block larger than a part is rejected with a clear message', () => {
  const text = `intro\n\n\`\`\`json\n${'x'.repeat(11000)}\n\`\`\`\n`;
  assert.throws(() => splitBody(text), /fenced code block[\s\S]*shorten/i);
});

test('the CLI reports an oversize fenced block and emits nothing', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'emit-'));
  const file = path.join(dir, 'draft.md');
  fs.writeFileSync(file, `intro\n\n\`\`\`json\n${'x'.repeat(11000)}\n\`\`\`\n`);
  const result = spawnSync('node', [script, file], { encoding: 'utf8' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /shorten/i);
  assert.equal(result.stdout, '');
});

test('joinParts tolerates a sanitizer-appended closing fence after the end sentinel', () => {
  assert.equal(joinParts({ body: `a\n${PART_END}\n\`\`\``, body_2: `${PART_START}\nb` }), 'a\nb');
  assert.equal(joinParts({ body: `a\n${PART_END}  `, body_2: `${PART_START}\nb` }), 'a\nb');
});

test('joinParts throws when a sentinel survives in the joined body', () => {
  assert.throws(() => joinParts({ body: `a ${PART_END} mid`, body_2: 'b' }), /sentinel/i);
});
