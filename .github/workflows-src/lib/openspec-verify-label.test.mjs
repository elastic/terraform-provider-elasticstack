import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/openspec-verify-label.md');

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

test('verify-label workflow installs Go from go.mod and exports Go paths for AWF', () => {
  const source = workflowSource();
  assert.match(source, /uses: actions\/setup-go@v6/);
  assert.match(source, /go-version-file: go\.mod/);
  assert.match(source, /Export Go paths for AWF chroot mode/);
  assert.match(source, /GOROOT=\$\(go env GOROOT\)/);
  assert.match(source, /GOPATH=\$\(go env GOPATH\)/);
  assert.match(source, /GOMODCACHE=\$\(go env GOMODCACHE\)/);
  assert.match(source, /allowed: \[defaults, node, go\]/);
});

test('verify-label workflow installs Node from package.json and omits runtimes.go', () => {
  const source = workflowSource();
  assert.match(source, /uses: actions\/setup-node@v6/);
  assert.match(source, /node-version-file: package\.json/);
  // The review workflow must not reintroduce a frontmatter go runtime pin.
  assert.doesNotMatch(source, /runtimes:\s*\n\s*go:/);
});

test('verify-label workflow provisions Terraform with wrapper disabled', () => {
  const source = workflowSource();
  assert.match(source, /uses: hashicorp\/setup-terraform@v4/);
  assert.match(source, /terraform_wrapper: false/);
});

test('verify-label workflow bootstraps the repo with make setup', () => {
  const source = workflowSource();
  assert.match(source, /Setup repository dependencies\s*\n\s*run: make setup/m);
});

test('verify-label workflow declares remove-labels safe output for verify-openspec only', () => {
  const source = workflowSource();
  assert.match(
    source,
    /remove-labels:\s*\n\s*target:\s*triggering\s*\n\s*allowed:\s*\[\s*verify-openspec\s*\]\s*\n\s*max:\s*1/m,
  );
});

test('verify-label workflow does not use completion_cleanup job or inline label-removal script', () => {
  const source = workflowSource();
  assert.doesNotMatch(source, /completion_cleanup/);
  assert.doesNotMatch(source, /remove_verify_label\.inline\.js/);
});

test('verify-label workflow prompt instructs terminal remove-labels cleanup', () => {
  const source = workflowSource();
  assert.match(source, /## Remove trigger label \(final safe outputs\)/);
  assert.match(source, /remove-labels.*terminal/s);
  assert.match(source, /post-agent script or job/);
});

test('verify-label workflow exposes review disposition and disposition reason to the agent', () => {
  const source = workflowSource();
  assert.match(source, /review_disposition: \$\{\{ steps\.select_change\.outputs\.review_disposition \}\}/);
  assert.match(source, /disposition_reason: \$\{\{ steps\.select_change\.outputs\.disposition_reason \}\}/);
  assert.match(source, /\*\*Review disposition\*\*.*approval-eligible.*comment-only/s);
  assert.match(source, /\*\*Disposition reason\*\*/);
});

test('verify-label workflow ties APPROVE and archive to approval-eligible disposition', () => {
  const source = workflowSource();
  assert.match(source, /review_disposition.*approval-eligible/s);
  assert.match(source, /Archive and push \(APPROVE only, approval-eligible only\)/);
  assert.match(source, /comment-only.*net-new spec change/s);
});
