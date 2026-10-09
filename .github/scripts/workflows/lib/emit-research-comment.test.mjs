import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const { splitBody, MAX_PART_BYTES, MAX_PARTS, MAX_CHARS } = require('../research-factory/emit-research-comment.js');

const join = (parts) => Object.values(parts).join('');
const bytes = (s) => Buffer.byteLength(s, 'utf8');

test('constants match the per-input limit and part count', () => {
  assert.equal(MAX_PART_BYTES, 10000);
  assert.equal(MAX_PARTS, 7);
  assert.equal(MAX_CHARS, 60000);
});

test('a small body yields only body', () => {
  assert.deepEqual(splitBody('hello\nworld\n'), { body: 'hello\nworld\n' });
});

test('a large multi-line body splits at line boundaries and round-trips', () => {
  const text = Array.from({ length: 1500 }, (_, i) => `line ${i} with some padding text`).join('\n') + '\n';
  const parts = splitBody(text);
  assert.deepEqual(Object.keys(parts), ['body', 'body_2', 'body_3', 'body_4', 'body_5', 'body_6', 'body_7'].slice(0, Object.keys(parts).length));
  assert.ok(Object.keys(parts).length > 1);
  for (const part of Object.values(parts)) {
    assert.ok(bytes(part) <= MAX_PART_BYTES);
    assert.ok(part.endsWith('\n'));
  }
  assert.equal(join(parts), text);
});

test('multi-byte characters are never split and stay within the byte limit', () => {
  const text = `${'é'.repeat(4000)}\n${'日本語'.repeat(1500)}\n${'😀'.repeat(2000)}\n`;
  const parts = splitBody(text);
  for (const part of Object.values(parts)) {
    assert.ok(bytes(part) <= MAX_PART_BYTES);
    assert.ok(!part.includes('�'));
  }
  assert.equal(join(parts), text);
});

test('a single overlong line is split at a code point boundary', () => {
  const text = '😀'.repeat(6000);
  const parts = splitBody(text);
  assert.ok(Object.keys(parts).length >= 3);
  for (const part of Object.values(parts)) {
    assert.ok(bytes(part) <= MAX_PART_BYTES);
    assert.ok(!part.includes('�'));
  }
  assert.equal(join(parts), text);
});

test('a body needing more than seven parts is rejected', () => {
  const text = `${'a'.repeat(5999)}\n`.repeat(10);
  assert.equal(text.length, MAX_CHARS);
  assert.throws(() => splitBody(text), /parts/i);
});

test('a body above the character cap is rejected', () => {
  assert.throws(() => splitBody('a'.repeat(MAX_CHARS + 1)), /60000|characters/i);
});

test('a body exactly at the character cap is accepted', () => {
  const text = `${'a'.repeat(99)}\n`.repeat(600);
  assert.equal(text.length, MAX_CHARS);
  assert.equal(join(splitBody(text)), text);
});

test('the CLI prints the JSON object for a file', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'emit-'));
  const file = path.join(dir, 'draft.md');
  fs.writeFileSync(file, '# Title\n\nbody\n');
  const script = path.join(path.dirname(fileURLToPath(import.meta.url)), '../research-factory/emit-research-comment.js');
  const result = spawnSync('node', [script, file], { encoding: 'utf8' });
  assert.equal(result.status, 0);
  assert.deepEqual(JSON.parse(result.stdout), { body: '# Title\n\nbody\n' });
});

test('the CLI exits non-zero with a clear message when the body is too large', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'emit-'));
  const file = path.join(dir, 'draft.md');
  fs.writeFileSync(file, 'a'.repeat(MAX_CHARS + 1));
  const script = path.join(path.dirname(fileURLToPath(import.meta.url)), '../research-factory/emit-research-comment.js');
  const result = spawnSync('node', [script, file], { encoding: 'utf8' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /characters/i);
  assert.equal(result.stdout, '');
});

test('the CLI errors when no file is given', () => {
  const script = path.join(path.dirname(fileURLToPath(import.meta.url)), '../research-factory/emit-research-comment.js');
  const result = spawnSync('node', [script], { encoding: 'utf8' });
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /usage/i);
});
