import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../../workflows/semantic-function-refactor.md');
const lockPath = path.resolve(__dirname, '../../workflows/semantic-function-refactor.lock.yml');
const upstreamBaseline = 'https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md';

function workflowSource() {
  return readFileSync(workflowPath, 'utf8');
}

function lockSource() {
  return readFileSync(lockPath, 'utf8');
}

test('semantic-function-refactor workflow references the upstream baseline and deterministic issue-slot gate', () => {
  const source = workflowSource();
  assert.ok(source.includes(upstreamBaseline), 'expected upstream baseline reference in generated workflow');
  assert.match(source, /ISSUE_SLOTS_LABEL:\s*semantic-refactor/);
  assert.match(source, /ISSUE_SLOTS_CAP:\s*"3"/);
  assert.match(source, /open_issues:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.open_issues \}\}/);
  assert.match(source, /issue_slots_available:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.issue_slots_available \}\}/);
  assert.match(source, /gate_reason:\s*\$\{\{ steps\.compute_issue_slots\.outputs\.gate_reason \}\}/);
});

test('semantic-function-refactor workflow encodes the prompt contract for scope and issue creation', () => {
  const source = workflowSource();
  assert.match(source, /\*\*Skip test files\*\*/);
  assert.match(source, /\*\*Test files\*\*/);
  assert.match(source, /\*\*Generated files\*\* and build artifacts/);
  assert.match(source, /\*\*Workflow files\*\*/);
  assert.match(source, /\*\*Create one issue per distinct opportunity\*\*/);
  assert.match(source, /Create separate issues for each distinct actionable refactoring opportunity/);
  assert.match(source, /Never create more than/);
  assert.match(source, /Use Serena for semantic analysis/);
  assert.match(source, /Tool: activate_project/);
  assert.match(source, /Tool: get_symbols_overview/);
  assert.match(source, /Tool: find_symbol/);
});

test('semantic-function-refactor workflow safe outputs and compiled lock keep semantic-refactor issue metadata aligned', () => {
  const source = workflowSource();
  const lock = lockSource();
  assert.match(source, /title-prefix:\s*"\[semantic-refactor\] "/);
  assert.match(source, /labels:\s*\[semantic-refactor, refactoring, code-quality, automated-analysis\]/);
  assert.match(source, /max:\s*3/);
  assert.match(lock, /"create_issue":\{"labels":\["semantic-refactor","refactoring","code-quality","automated-analysis"\],"max":3,"title_prefix":"\[semantic-refactor\] "\}/);
  assert.match(lock, /Maximum 3 issue\(s\) can be created/);
});

test('semantic-function-refactor workflow routes Claude through LiteLLM with secret-backed API key', () => {
  const source = workflowSource();
  assert.match(source, /engine:\s*\n\s*id:\s*claude/m);
  assert.match(source, /model: "?llm-gateway\/claude-sonnet-4-6"?/);
  assert.match(source, /ANTHROPIC_BASE_URL:\s*"?https:\/\/elastic\.litellm-prod\.ai"?/);
  assert.match(source, /ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/);
});

test('compiled lock wires gh-aw anthropic target and Claude env for the agent', () => {
  const lock = lockSource();
  assert.match(
    lock,
    /id: agentic_execution[\s\S]*--anthropic-api-target elastic\.litellm-prod\.ai[\s\S]*--allow-domains[^\n]*elastic\.litellm-prod\.ai[\s\S]*\n\s*ANTHROPIC_BASE_URL:\s*https:\/\/elastic\.litellm-prod\.ai[\s\S]*\n\s*ANTHROPIC_MODEL:\s*llm-gateway\/claude-sonnet-4-6/
  );
});

test('compiled lock excludes ANTHROPIC_API_KEY from AWF --env-all and uses the Claude secret', () => {
  const lock = lockSource();
  assert.match(
    lock,
    /id: agentic_execution[\s\S]*--exclude-env ANTHROPIC_API_KEY[\s\S]*\n\s*ANTHROPIC_API_KEY:\s*\$\{\{\s*secrets\.CLAUDE_LITELLM_PROXY_API_KEY\s*\}\}/
  );
});

test('workflow configures Serena MCP server for semantic Go analysis', () => {
  const source = workflowSource();
  const lock = lockSource();
  assert.match(source, /mcp-servers:/);
  assert.match(source, /container:\s*"ghcr\.io\/github\/serena-mcp-server:latest"/);
  assert.match(source, /entrypoint:\s*"serena"/);
  assert.match(source, /allowed:/);
  assert.match(lock, /"serena":\s*\{/);
  assert.match(lock, /"container":\s*"ghcr\.io\/github\/serena-mcp-server:latest"/);
  assert.match(lock, /"entrypoint":\s*"serena"/);
});

test('compiled lock includes Serena tools in agent allowed-tools', () => {
  const lock = lockSource();
  assert.match(lock, /mcp__serena__activate_project/);
  assert.match(lock, /mcp__serena__get_symbols_overview/);
  assert.match(lock, /mcp__serena__find_symbol/);
  assert.match(lock, /mcp__serena__search_for_pattern/);
  assert.match(lock, /mcp__serena__find_referencing_symbols/);
  assert.match(lock, /mcp__serena__read_file/);
});

test('workflow configures bash tools for Go source navigation', () => {
  const source = workflowSource();
  assert.match(source, /tools:/);
  assert.match(source, /bash:/);
  assert.match(source, /find \. -name '\*\.go' ! -name '\*_test\.go' -type f/);
  assert.match(source, /grep -r '\^func ' \. --include='\*\.go'/);
});

test('compiled lock preserves LiteLLM model and allowed domains', () => {
  const lock = lockSource();
  assert.match(lock, /llm-gateway\/claude-sonnet-4-6/);
  assert.match(lock, /elastic\.litellm-prod\.ai/);
  assert.match(lock, /GH_AW_INFO_ALLOWED_DOMAINS:[\s\S]*elastic\.litellm-prod\.ai/);
});

test('workflow includes dispatch instruction and compiled lock contains dispatch_code_factory job', () => {
  const source = workflowSource();
  const lock = lockSource();
  assert.match(source, /dispatch_code_factory/);
  assert.match(source, /Dispatch/);
  assert.match(lock, /dispatch_code_factory/);
  assert.match(lock, /"dispatch-code-factory":\{"description":"Dispatch code-factory for each created issue"\}/);
  assert.match(lock, /"dispatch_code_factory"/);
});
