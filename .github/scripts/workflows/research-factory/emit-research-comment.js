const fs = require('fs');

const MAX_PART_BYTES = 10000;
const MAX_PARTS = 7;
const MAX_CHARS = 60000;
// The safe-output platform strips HTML comments and trims every string input, so the
// sentinels are plain non-whitespace tokens that protect real newlines at chunk edges.
const PART_START = '%%RF_PART_START%%';
const PART_END = '%%RF_PART_END%%';

const byteLength = (s) => Buffer.byteLength(s, 'utf8');

function splitLongLine(line, budget) {
  const pieces = [];
  let current = '';
  for (const ch of line) {
    if (byteLength(current) + byteLength(ch) > budget) {
      pieces.push(current);
      current = '';
    }
    current += ch;
  }
  pieces.push(current);
  return pieces;
}

function splitBody(text) {
  const chars = Array.from(text).length;
  if (chars > MAX_CHARS) {
    throw new Error(`comment body is ${chars} characters, above the ${MAX_CHARS} character limit`);
  }

  if (byteLength(text) <= MAX_PART_BYTES) {
    return { body: text };
  }

  const budget = MAX_PART_BYTES - byteLength(PART_START) - byteLength(PART_END);
  const chunks = [];
  let current = '';
  for (const line of text.match(/[^\n]*\n|[^\n]+/g) || ['']) {
    for (const piece of byteLength(line) > budget ? splitLongLine(line, budget) : [line]) {
      if (current !== '' && byteLength(current) + byteLength(piece) > budget) {
        chunks.push(current);
        current = '';
      }
      current += piece;
    }
  }
  chunks.push(current);

  if (chunks.length > MAX_PARTS) {
    throw new Error(`comment body needs ${chunks.length} parts, above the maximum of ${MAX_PARTS} parts`);
  }
  const last = chunks.length - 1;
  return Object.fromEntries(
    chunks.map((chunk, i) => [i === 0 ? 'body' : `body_${i + 1}`, `${i > 0 ? PART_START : ''}${chunk}${i < last ? PART_END : ''}`]),
  );
}

function joinParts(item) {
  const keys = ['body', ...Array.from({ length: MAX_PARTS - 1 }, (_, i) => `body_${i + 2}`)];
  return keys
    .map((key) => (typeof item[key] === 'string' ? item[key] : ''))
    .map((part) => (part.startsWith(PART_START) ? part.slice(PART_START.length) : part))
    .map((part) => (part.endsWith(PART_END) ? part.slice(0, -PART_END.length) : part))
    .join('');
}

if (typeof module !== 'undefined') {
  module.exports = { splitBody, joinParts, PART_START, PART_END, MAX_PART_BYTES, MAX_PARTS, MAX_CHARS };
}

if (require.main === module) {
  const file = process.argv[2];
  if (!file) {
    console.error('usage: emit-research-comment.js <markdown-file>');
    process.exit(2);
  }
  try {
    process.stdout.write(JSON.stringify(splitBody(fs.readFileSync(file, 'utf8'))));
  } catch (err) {
    console.error(`emit-research-comment: ${err.message}`);
    process.exit(1);
  }
}
