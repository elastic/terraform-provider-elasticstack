import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/schema-coverage-rotation.md');
const lockPath = path.resolve(__dirname, '../../workflows/schema-coverage-rotation.lock.yml');

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
  const source = workflowSource();
  assert.match(source, /uses: actions\/setup-go@v6/);
  assert.match(source, /go-version-file: go\.mod/);
  assert.match(source, /Export Go paths for AWF chroot mode/);
  assert.match(source, /GOROOT=\$\(go env GOROOT\)/);
  assert.match(source, /GOPATH=\$\(go env GOPATH\)/);
  assert.match(source, /GOMODCACHE=\$\(go env GOMODCACHE\)/);
});

test('schema-coverage rotation workflow installs Node from package.json and allows bootstrap ecosystems', () => {
  const source = workflowSource();
  assert.match(source, /uses: actions\/setup-node@v6/);
  assert.match(source, /node-version-file: package\.json/);
  assert.match(source, /allowed: \[defaults, node, go, elastic\.litellm-prod\.ai\]/);
});

test('schema-coverage rotation workflow uses Claude through LiteLLM with secret-backed API key and tool timeout', () => {
  const source = workflowSource();
  assert.match(source, /engine:\s*\n\s*id:\s*claude/m);
  assert.match(source, /model: "?llm-gateway\/claude-sonnet-4-6"?/);
  assert.match(source, /ANTHROPIC_BASE_URL:\s*"?https:\/\/elastic\.litellm-prod\.ai\/?"?/);
  assert.match(source, /ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/);
  assert.match(source, /tools:\s*\n\s*timeout:\s*300/m);
});

test('compiled schema-coverage rotation lock sets Claude tool budget and Anthropic proxy for main and threat-detection agentic execution', () => {
  const lock = lockSource();
  assert.match(lock, /GH_AW_TOOL_TIMEOUT:\s*300/);
  assert.match(
    lock,
    /id: agentic_execution[\s\S]*--anthropic-api-target elastic\.litellm-prod\.ai[\s\S]*ANTHROPIC_BASE_URL:\s*https:\/\/elastic\.litellm-prod\.ai\/[\s\S]*ANTHROPIC_MODEL:\s*llm-gateway\/claude-sonnet-4-6/
  );
  assert.match(
    lock,
    /id: agentic_execution[\s\S]*--exclude-env ANTHROPIC_API_KEY[\s\S]*\n\s*ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/
  );
  assert.match(
    lock,
    /id: detection_agentic_execution[\s\S]*--anthropic-api-target elastic\.litellm-prod\.ai[\s\S]*\n\s*ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/
  );
  assert.match(
    lock,
    /id: detection_agentic_execution[\s\S]*ANTHROPIC_BASE_URL:\s*https:\/\/elastic\.litellm-prod\.ai\/[\s\S]*ANTHROPIC_MODEL:\s*llm-gateway\/claude-sonnet-4-6/
  );
});

test('schema-coverage rotation workflow bootstraps the repo with make setup', () => {
  const source = workflowSource();
  assert.match(source, /name: Setup repository dependencies/);
  assert.match(source, /run: make setup/);
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
  assert.match(source, /dispatch_code_factory/);
  assert.match(source, /Dispatch/);
  assert.match(lock, /dispatch_code_factory/);
  assert.match(lock, /"dispatch-code-factory":\{"description":"Dispatch code-factory for each created issue"\}/);
  assert.match(lock, /"dispatch_code_factory"/);
  assert.match(lock, /"labels":\["testing","acceptance-tests","schema-coverage"\]/);
});
