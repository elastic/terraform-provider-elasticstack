import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { normalizeCompiledLock } from './compiled-lock.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const workflowsDir = path.resolve(__dirname, '../../../workflows');
const fragment = readFileSync(path.join(workflowsDir, 'shared/dispatch-research-factory.md'), 'utf8');

test('fragment dispatches research-factory-issue.lock.yml through producer-dispatch.js', () => {
  assert.match(fragment, /producer-dispatch\.js[\s\S]*research-factory-issue\.lock\.yml/);
});

test('fragment defines the dispatch-research-factory job after safe_outputs with a required boolean input', () => {
  assert.match(fragment, /dispatch-research-factory:\s*\n\s+needs: safe_outputs/);
  assert.match(fragment, /dispatch:\s*\n\s+description:[^\n]*\n\s+required: true\n\s+type: boolean/);
});

test('fragment derives source_workflow from the calling workflow name', () => {
  assert.match(fragment, /GITHUB_WORKFLOW_NAME: \$\{\{ github\.workflow \}\}/);
  assert.match(fragment, /tr '\[:upper:\]' '\[:lower:\]' \| tr ' ' '-'/);
});

test('scanner workflow name slugifies to kibana-api-gap-scanner', () => {
  const source = readFileSync(path.join(workflowsDir, 'kibana-api-gap-scanner.md'), 'utf8');
  const name = source.match(/^name:\s*(.+)$/m)[1].trim();
  assert.equal(name.toLowerCase().split(' ').join('-'), 'kibana-api-gap-scanner');
});

const scanner = readFileSync(path.join(workflowsDir, 'kibana-api-gap-scanner.md'), 'utf8');
const scannerLock = normalizeCompiledLock(
  readFileSync(path.join(workflowsDir, 'kibana-api-gap-scanner.lock.yml'), 'utf8')
);

test('scanner imports the dispatch-research-factory fragment', () => {
  const imports = scanner.match(/^imports:\s*\n((?:\s+- .+\n)+)/m);
  assert.ok(imports, 'expected an imports list');
  assert.match(imports[1], /- shared\/dispatch-research-factory\.md/);
});

test('scanner prompt requests one dispatch_research_factory call with dispatch: true', () => {
  assert.match(scanner, /`dispatch_research_factory` safe output tool once with `dispatch: true`/);
});

test('compiled scanner lock exposes dispatch_research_factory with a required boolean dispatch input', () => {
  assert.match(
    scannerLock,
    /"dispatch-research-factory":\{"description":"[^"]*","inputs":\{"dispatch":\{[^}]*"required":true,"type":"boolean"\}\}\}/
  );
  assert.match(scannerLock, /name": "dispatch_research_factory"/);
});

test('compiled scanner lock dispatches research-factory-issue.lock.yml after safe_outputs', () => {
  const job = scannerLock.match(/\n  dispatch_research_factory:\n[\s\S]*?\n  [a-z_]+:\n/);
  assert.ok(job, 'expected a dispatch_research_factory job');
  assert.match(job[0], /- safe_outputs/);
  assert.match(job[0], /research-factory-issue\.lock\.yml/);
});

test('compiled scanner lock declares schedule, kbapi push filter and dispatch', () => {
  assert.match(scannerLock, /\n  schedule:\n    - cron:/);
  assert.match(scannerLock, /\n  push:\n    branches:\n      - main\n    paths:\n      - generated\/kbapi\/\*\*\n/);
  assert.match(scannerLock, /\n  workflow_dispatch:/);
});
