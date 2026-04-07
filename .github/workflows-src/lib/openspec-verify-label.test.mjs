import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/openspec-verify-label.md');
const lockPath = path.resolve(__dirname, '../../workflows/openspec-verify-label.lock.yml');

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

function lockSource() {
  return readFileSync(lockPath, 'utf8');
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
  // The step name and run command may have an if: condition in between, so just check both appear
  assert.match(source, /name: Setup repository dependencies/);
  assert.match(source, /run: make setup/);
});

test('verify-label workflow uses pull_request_target with labeled type, not label_command', () => {
  const source = workflowSource();
  assert.match(source, /pull_request_target:/);
  assert.match(source, /types:/);
  assert.match(source, /labeled/);
  assert.doesNotMatch(source, /label_command:/);
});

test('verify-label workflow has a deterministic verify_label step', () => {
  const source = workflowSource();
  assert.match(source, /Verify trigger label/);
  assert.match(source, /verify_label/);
  assert.match(source, /label_verified/);
});

test('verify-label workflow has a deterministic remove trigger label step', () => {
  const source = workflowSource();
  assert.match(source, /Remove trigger label/);
  assert.match(source, /remove_trigger_label/);
  assert.match(source, /trigger_label_removed/);
});

test('verify-label workflow does not declare remove-labels safe output', () => {
  const source = workflowSource();
  assert.doesNotMatch(source, /remove-labels:/);
});

test('verify-label workflow prompt does not instruct remove-labels cleanup', () => {
  const source = workflowSource();
  assert.doesNotMatch(source, /## Remove trigger label \(final safe outputs\)/);
  // The agent prompt must instruct NOT to emit remove-labels safe outputs
  assert.match(source, /do \*\*not\*\* emit \*\*`remove-labels`\*\* safe outputs/);
});

test('verify-label workflow compiled lock includes contents write and pull-requests write for safe-output jobs', () => {
  // The gh-aw compiler enforces that write permissions are added at the job level by safe-outputs,
  // not via explicit frontmatter write permissions (strict mode). Check the compiled lock.yml.
  const lock = lockSource();
  assert.match(lock, /contents: write/);
  assert.match(lock, /pull-requests: write/);
});

test('verify-label workflow exposes review disposition and disposition reason to the agent', () => {
  const source = workflowSource();
  assert.match(source, /review_disposition: \$\{\{ steps\.classify_and_select\.outputs\.review_disposition \}\}/);
  assert.match(source, /disposition_reason: \$\{\{ steps\.classify_and_select\.outputs\.disposition_reason \}\}/);
  assert.match(source, /\*\*Review disposition\*\*.*approval-eligible.*comment-only/s);
  assert.match(source, /\*\*Disposition reason\*\*/);
});

test('verify-label workflow exposes verification mode and archive/push outputs to the agent', () => {
  const source = workflowSource();
  assert.match(source, /verification_mode: \$\{\{ steps\.classify_and_select\.outputs\.verification_mode \}\}/);
  assert.match(source, /archive_push_allowed: \$\{\{ steps\.classify_and_select\.outputs\.archive_push_allowed \}\}/);
  assert.match(source, /\*\*Verification mode\*\*/);
  assert.match(source, /workspace.*api-only/s);
  assert.match(source, /\*\*Archive\/push allowed\*\*/);
});

test('verify-label agent prompt interpolates needs.pre_activation review outputs where the agent reads them', () => {
  const source = workflowSource();
  const rd = '${{ needs.pre_activation.outputs.review_disposition }}';
  const dr = '${{ needs.pre_activation.outputs.disposition_reason }}';
  assert.ok(source.includes(rd), 'expected review_disposition interpolation in generated workflow');
  assert.ok(source.includes(dr), 'expected disposition_reason interpolation in generated workflow');
  const pre = source.split('## Pre-activation context')[1].split('## Verification (active change)')[0];
  assert.ok(pre.includes(rd), 'expected review_disposition in Pre-activation context');
  assert.ok(pre.includes(dr), 'expected disposition_reason in Pre-activation context');
  const step5 = source.split('## Review body, inline comments, and decision')[1].split('## Archive and push')[0];
  assert.ok(step5.includes(rd), 'expected review_disposition in review-submission instructions');
});

test('verify-label agent prompt interpolates archive/push outputs where the agent reads them', () => {
  const source = workflowSource();
  const archiveAllowed = '${{ needs.pre_activation.outputs.archive_push_allowed }}';
  const archiveReason = '${{ needs.pre_activation.outputs.archive_push_reason }}';
  assert.ok(source.includes(archiveAllowed), 'expected archive_push_allowed interpolation in generated workflow');
  assert.ok(source.includes(archiveReason), 'expected archive_push_reason interpolation in generated workflow');
  const pre = source.split('## Pre-activation context')[1].split('## Verification (active change)')[0];
  assert.ok(pre.includes(archiveAllowed), 'expected archive_push_allowed in Pre-activation context');
  assert.ok(pre.includes(archiveReason), 'expected archive_push_reason in Pre-activation context');
});

test('verify-label workflow ties APPROVE and archive to approval-eligible disposition', () => {
  const source = workflowSource();
  assert.match(source, /review_disposition.*approval-eligible/s);
  assert.match(source, /Archive and push \(APPROVE only, approval-eligible only, archive-push-allowed only\)/);
  assert.match(source, /comment-only.*net-new spec change/s);
});

test('verify-label workflow gates archive/push on archive_push_allowed being true', () => {
  const source = workflowSource();
  assert.match(source, /archive_push_allowed.*true/s);
  // Archive section should mention archive_push_allowed
  const archiveSection = source.split('## Archive and push')[1];
  assert.ok(
    archiveSection.includes('archive_push_allowed'),
    'expected archive_push_allowed check in archive/push section'
  );
});

test('verify-label workflow states archive_push_allowed false does not force COMMENT review', () => {
  const source = workflowSource();
  assert.match(source, /archive_push_allowed.*false.*does \*\*not\*\* force \*\*`COMMENT`\*\*/s);
});

test('verify-label workflow workspace bootstrap steps are conditional on workspace verification mode', () => {
  const source = workflowSource();
  // The compiled output should show Go/Node/Terraform setup steps are conditional on workspace mode
  assert.match(source, /verification_mode.*workspace/s);
});

test('verify-label compiled lock pre_activation job has issues write for label cleanup', () => {
  const lock = lockSource();
  assert.match(lock, /pre_activation:[\s\S]*?permissions:[\s\S]*?issues: write/);
});

test('verify-label compiled lock preserves workspace-only guards on credential-bearing steps', () => {
  const lock = lockSource();
  const workspaceGuard = /if: needs\.pre_activation\.outputs\.verification_mode == 'workspace'/;
  assert.match(lock, workspaceGuard);
});
