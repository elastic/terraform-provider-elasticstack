import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../../workflows/schema-coverage-rotation.md');
const lockPath = path.resolve(__dirname, '../../../workflows/schema-coverage-rotation.lock.yml');

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

function lockSource() {
  return readFileSync(lockPath, 'utf8');
}

/** Body between "## Repository toolchain" and the next "## Execution steps" (exclusive). */
function extractRepositoryToolchainSection(source) {
  const startMarker = '## Repository toolchain';
  const endMarker = '## Execution steps';
  const startIdx = source.indexOf(startMarker);
  assert.ok(startIdx !== -1, `missing heading ${JSON.stringify(startMarker)} in ${workflowPath}`);
  const afterStart = source.slice(startIdx + startMarker.length);
  const endRel = afterStart.indexOf(endMarker);
  assert.ok(
    endRel !== -1,
    `missing heading ${JSON.stringify(endMarker)} after ${JSON.stringify(startMarker)} in ${workflowPath} (misordered or renamed section?)`,
  );
  return afterStart.slice(0, endRel);
}

test('schema-coverage rotation workflow installs Go from go.mod and exports Go paths for AWF', () => {
  // Setup steps are injected via the shared setup-dev.md import; verify in the lock file.
  const lock = lockSource();
  assert.match(lock, /setup-go/);
  assert.match(lock, /go-version-file: go\.mod/);
  assert.match(lock, /Export Go and Terraform paths for AWF chroot mode/);
  assert.match(lock, /GOROOT=\$\(go env GOROOT\)/);
  assert.match(lock, /GOPATH=\$\(go env GOPATH\)/);
  assert.match(lock, /GOMODCACHE=\$\(go env GOMODCACHE\)/);
  // Source workflow declares the import.
  const source = workflowSource();
  assert.match(source, /shared\/setup-dev\.md/);
  assert.match(source, /shared\/dispatch-code-factory\.md/);
});

test('schema-coverage rotation workflow installs Node from package.json and allows bootstrap ecosystems', () => {
  // Setup steps are injected via the shared setup-dev.md import; verify in the lock file.
  const lock = lockSource();
  assert.match(lock, /setup-node/);
  assert.match(lock, /node-version-file: package\.json/);
  // Network config remains in the source workflow.
  const source = workflowSource();
  assert.match(source, /allowed: \[defaults, node, go, elastic\.litellm-prod\.ai\]/);
});

test('schema-coverage rotation workflow uses Claude through LiteLLM with secret-backed API key and tool timeout', () => {
  const source = workflowSource();
  assert.match(source, /engine:\s*\n\s*id:\s*claude/m);
  assert.match(source, /model: "?llm-gateway\/claude-sonnet-5"?/);
  assert.match(source, /ANTHROPIC_BASE_URL:\s*"?https:\/\/elastic\.litellm-prod\.ai\/?"?/);
  assert.match(source, /ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/);
  assert.match(source, /tools:[\s\S]*?\n\s+timeout:\s*300/m);
});

test('schema-coverage rotation source workflow configures engine env with base URL and model', () => {
  const source = workflowSource();
  assert.match(source, /engine:\s*\n\s*id:\s*claude/m);
  assert.match(source, /model: "?llm-gateway\/claude-sonnet-5"?/);
  assert.match(source, /ANTHROPIC_BASE_URL:\s*"?https:\/\/elastic\.litellm-prod\.ai\/?"?/);
  assert.match(source, /ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/);
  assert.match(source, /tools:[\s\S]*?\n\s+timeout:\s*300/m);
});

test('schema-coverage rotation workflow bootstraps the repo with make setup', () => {
  // Setup steps are injected via the shared setup-dev.md import; verify in the lock file.
  const lock = lockSource();
  assert.match(lock, /name: Setup repository dependencies/);
  assert.match(lock, /run: make setup/);
});

test('schema-coverage rotation prompt documents deterministic toolchain without self-install', () => {
  const source = workflowSource();
  const section = extractRepositoryToolchainSection(source);
  assert.match(section, /Deterministic workflow steps have already/);
  assert.match(section, /Do \*\*not\*\* install alternate Go or Node versions/);
  assert.doesNotMatch(section, /scripts\/schema-coverage-rotation/);
});

test('workflow includes dispatch instruction and compiled lock contains dispatch_code_factory job', () => {
  const source = workflowSource();
  const lock = lockSource();
  assert.match(source, /shared\/dispatch-code-factory\.md/);
  assert.match(source, /dispatch_code_factory/);
  assert.match(source, /Dispatch/);
  assert.doesNotMatch(source, /safe-outputs:[\s\S]*?jobs:[\s\S]*?dispatch-code-factory:/);
  assert.match(lock, /dispatch_code_factory/);
  assert.match(lock, /"dispatch-code-factory":\{"description":"Dispatch code-factory for each created issue"\}/);
  assert.match(lock, /"dispatch_code_factory"/);
  assert.match(lock, /SOURCE_WORKFLOW=\$\(echo "\$GITHUB_WORKFLOW_NAME"/);
  assert.doesNotMatch(lock, /SOURCE_WORKFLOW: (?:flaky-test-catcher|semantic-function-refactor|schema-coverage-rotation|duplicate-code-detector)\b/);
  assert.match(lock, /"labels":\["testing","acceptance-tests","schema-coverage","triaged"\]/);
});
