import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

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
