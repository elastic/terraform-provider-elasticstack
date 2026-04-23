import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/duplicate-code-detector.md');
const lockPath = path.resolve(__dirname, '../../workflows/duplicate-code-detector.lock.yml');
const upstreamBaseline = 'https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md';

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

function lockSource() {
  return readFileSync(lockPath, 'utf8');
}

test('duplicate-code detector workflow references the upstream baseline and deterministic issue-slot gate', () => {
  const source = workflowSource();
  assert.ok(source.includes(upstreamBaseline), 'expected upstream baseline reference in generated workflow');
  assert.match(source, /ISSUE_SLOTS_LABEL:\s*duplicate-code/);
  assert.match(source, /ISSUE_SLOTS_CAP:\s*"3"/);
  assert.match(source, /open_issues:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.open_issues \}\}/);
  assert.match(source, /issue_slots_available:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.issue_slots_available \}\}/);
  assert.match(source, /gate_reason:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.gate_reason \}\}/);
});

test('duplicate-code detector workflow encodes the prompt contract for scope and issue creation', () => {
  const source = workflowSource();
  assert.match(source, /\*\*Exclude test files\*\* from analysis/);
  assert.match(source, /\*\*Exclude generated files\*\* and build artifacts/);
  assert.match(source, /\*\*Exclude workflow files\*\* from analysis/);
  assert.match(source, /Only create issues if significant duplication is found \(threshold: >10 lines of duplicated code OR 3\+ instances of similar patterns\)/);
  assert.match(source, /Create separate issues for each distinct duplication pattern found, up to `\$\{\{ needs\.pre_activation\.outputs\.issue_slots_available \}\}` patterns this run/);
  assert.match(source, /Create \*\*one issue per distinct duplication pattern\*\* - do NOT bundle multiple patterns in a single issue/);
});

test('duplicate-code detector workflow safe outputs and compiled lock keep duplicate-code issue metadata aligned', () => {
  const source = workflowSource();
  const lock = lockSource();
  assert.match(source, /title-prefix:\s*"\[duplicate-code\] "/);
  assert.match(source, /labels:\s*\[duplicate-code, code-quality, automated-analysis\]/);
  assert.match(source, /max:\s*3/);
  assert.match(lock, /"create_issue":\{"labels":\["duplicate-code","code-quality","automated-analysis"\],"max":3,"title_prefix":"\[duplicate-code\] "\}/);
  assert.match(lock, /Maximum 3 issue\(s\) can be created\./);
});
