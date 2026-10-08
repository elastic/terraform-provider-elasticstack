import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const gate = require('../research-factory/gate.js');

const here = path.dirname(fileURLToPath(import.meta.url));
const rubric = fs.readFileSync(path.join(here, '../research-factory/critic-rubric.md'), 'utf8');
const workflow = fs.readFileSync(path.join(here, '../../../workflows/research-factory-issue.md'), 'utf8');

const { SCORE_THRESHOLD, STABILITY_WINDOW, MAX_ROUNDS } = gate;

const sources = {
  'critic-rubric.md': {
    text: rubric,
    threshold: new RegExp(`Score threshold: \\*\\*${SCORE_THRESHOLD}\\*\\*`),
    window: new RegExp(`Stability window: \\*\\*${STABILITY_WINDOW}\\*\\*`),
    rounds: new RegExp(`Maximum rounds: \\*\\*${MAX_ROUNDS}\\*\\*`),
  },
  'research-factory-issue.md': {
    text: workflow,
    threshold: new RegExp(`score threshold \\*\\*${SCORE_THRESHOLD}\\*\\*`),
    window: new RegExp(`stability window \\*\\*${STABILITY_WINDOW}\\*\\*`),
    rounds: new RegExp(`maximum of \\*\\*${MAX_ROUNDS}\\*\\* rounds`),
  },
};

for (const [file, { text, threshold, window, rounds }] of Object.entries(sources)) {
  test(`${file} states the gate score threshold from gate.js`, () => {
    assert.match(text, threshold);
  });
  test(`${file} states the gate stability window from gate.js`, () => {
    assert.match(text, window);
  });
  test(`${file} states the gate round bound from gate.js`, () => {
    assert.match(text, rounds);
  });
}
