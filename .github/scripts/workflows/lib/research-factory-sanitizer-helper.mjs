import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

// Faithful mini-sanitizer: what gh-aw's sanitize_content_core does to a string input that matters here
// (strip XML comments, balance unclosed fences by appending a closing fence, trim). Optionally the
// real sanitizer is used when RF_REAL_SANITIZER_DIR points at a checkout of gh-aw's actions/setup/js.
export function miniSanitize(value) {
  // Repeat until stable: a single pass can leave a `<!--` behind for nested input such as
  // `<!<!---->-->` (CodeQL js/incomplete-multi-character-sanitization). The real sanitizer uses a
  // depth-tracking scan; this fixed-point loop also drops any leftover opener, which is stricter than
  // needed for these tests and cannot leave a `<!--` in the output.
  let out = value;
  for (let prev = null; prev !== out; ) {
    prev = out;
    out = out.replace(/<!--[\s\S]*?-->/g, '').replace(/<!--/g, '');
  }
  const fence = /^( {0,3})(`{3,}|~{3,})([^`~\s]*)?(.*)$/;
  let open = null;
  for (const line of out.replace(/\r\n/g, '\n').split('\n')) {
    const m = fence.exec(line);
    if (!m) continue;
    if (!open) open = { char: m[2][0], length: m[2].length };
    else if (m[2][0] === open.char && m[2].length >= open.length && !(m[3] || m[4])) open = null;
  }
  if (open) out += `\n${open.char.repeat(open.length)}`;
  return out.trim();
}
if (process.env.RF_REAL_SANITIZER_DIR) {
  globalThis.core ??= { info() {}, warning() {}, debug() {} };
}
const realDir = process.env.RF_REAL_SANITIZER_DIR;
const realSanitize = realDir ? require(path.join(realDir, 'sanitize_content_core.cjs')).sanitizeContentCore : null;
export const sanitizers = [['mini sanitizer', miniSanitize], ...(realSanitize ? [['real sanitizer', (v) => realSanitize(v, 100000, 10)]] : [])];

export const sanitizeParts = (parts, fn) => Object.fromEntries(Object.entries(parts).map(([k, v]) => [k, fn(v)]));
