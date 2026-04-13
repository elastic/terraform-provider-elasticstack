import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/schema-coverage-rotation.md');

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

/** Body between "## Repository toolchain" and the next "## Memory format" (exclusive). */
function extractRepositoryToolchainSection(source) {
  const startMarker = '## Repository toolchain';
  const endMarker = '## Memory format';
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
  assert.match(source, /allowed: \[defaults, node, go\]/);
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
