const fs = require('fs');

const MAX_PART_BYTES = 10000;
const MAX_PARTS = 7;
const MAX_CHARS = 60000;

const byteLength = (s) => Buffer.byteLength(s, 'utf8');

function splitLongLine(line) {
  const pieces = [];
  let current = '';
  for (const ch of line) {
    if (byteLength(current) + byteLength(ch) > MAX_PART_BYTES) {
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

  const chunks = [];
  let current = '';
  for (const line of text.match(/[^\n]*\n|[^\n]+/g) || ['']) {
    for (const piece of byteLength(line) > MAX_PART_BYTES ? splitLongLine(line) : [line]) {
      if (current !== '' && byteLength(current) + byteLength(piece) > MAX_PART_BYTES) {
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
  return Object.fromEntries(chunks.map((chunk, i) => [i === 0 ? 'body' : `body_${i + 1}`, chunk]));
}

if (typeof module !== 'undefined') {
  module.exports = { splitBody, MAX_PART_BYTES, MAX_PARTS, MAX_CHARS };
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
