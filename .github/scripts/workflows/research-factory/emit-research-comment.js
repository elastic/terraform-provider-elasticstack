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

const FENCE_LINE = /^( {0,3})(`{3,}|~{3,})([^`~\s]*)?(.*)$/;

function fenceOf(line) {
  const m = FENCE_LINE.exec(line.replace(/\r?\n$/, ''));
  if (!m) {
    return null;
  }
  const info = `${m[3] || ''}${m[4] || ''}`;
  if (m[2][0] === '`' && info.includes('`')) {
    return null;
  }
  return { char: m[2][0], length: m[2].length, info: info.trim() };
}

// Groups lines into atoms: a fenced block (opener through closer) is one atom, so a cut never lands inside it.
function toAtoms(text) {
  const atoms = [];
  let fenced = null;
  let open = null;
  for (const line of text.match(/[^\n]*\n|[^\n]+/g) || ['']) {
    const fence = fenceOf(line);
    if (open) {
      open.text += line;
      if (fence && fence.char === open.char && fence.length >= open.length && fence.info === '') {
        atoms.push({ text: open.text, fenced: true });
        open = null;
      }
    } else if (fence) {
      open = { text: line, char: fence.char, length: fence.length };
    } else {
      atoms.push({ text: line, fenced: false });
    }
  }
  if (open) {
    atoms.push({ text: open.text, fenced: true });
  }
  return atoms;
}

function splitBody(text) {
  const chars = Array.from(text).length;
  if (chars > MAX_CHARS) {
    throw new Error(`comment body is ${chars} characters, above the ${MAX_CHARS} character limit`);
  }
  if (byteLength(text) <= MAX_PART_BYTES) {
    return { body: text };
  }

  const budget = MAX_PART_BYTES - byteLength(PART_START) - 1 - byteLength(PART_END);
  const chunks = [];
  let current = '';
  for (const atom of toAtoms(text)) {
    if (atom.fenced && byteLength(atom.text) > budget) {
      throw new Error(
        `a fenced code block of ${byteLength(atom.text)} bytes exceeds the ${budget} byte part budget; shorten that block`,
      );
    }
    for (const piece of byteLength(atom.text) > budget ? splitLongLine(atom.text, budget) : [atom.text]) {
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
    chunks.map((chunk, i) => [
      i === 0 ? 'body' : `body_${i + 1}`,
      `${i > 0 ? `${PART_START}\n` : ''}${chunk}${i < last ? PART_END : ''}`,
    ]),
  );
}

const TRAILING_END = new RegExp(`${PART_END}\\s*(?:(?:\`{3,}|~{3,})\\s*)?$`);

function joinParts(item) {
  const keys = ['body', ...Array.from({ length: MAX_PARTS - 1 }, (_, i) => `body_${i + 2}`)];
  const joined = keys
    .map((key) => (typeof item[key] === 'string' ? item[key] : ''))
    .map((part) => (part.startsWith(`${PART_START}\n`) ? part.slice(PART_START.length + 1) : part.startsWith(PART_START) ? part.slice(PART_START.length) : part))
    .map((part) => part.replace(TRAILING_END, ''))
    .join('');
  if (joined.includes('%%RF_PART_')) {
    throw new Error('a part sentinel survived in the joined comment body; refusing to publish a corrupt comment');
  }
  return joined;
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
