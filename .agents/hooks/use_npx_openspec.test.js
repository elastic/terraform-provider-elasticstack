import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const hookPath = path.resolve(__dirname, 'use_npx_openspec.js');

const require = createRequire(import.meta.url);
const { buildHookResponse, rewriteOpenSpecCommand } = require('./use_npx_openspec.js');

test('rewriteOpenSpecCommand rewrites a plain openspec invocation', () => {
  assert.equal(rewriteOpenSpecCommand('openspec validate'), 'npx openspec validate');
});

test('rewriteOpenSpecCommand rewrites openspec after common shell separators', () => {
  assert.equal(
    rewriteOpenSpecCommand('make setup && openspec validate ; ( openspec check )'),
    'make setup && npx openspec validate ; ( npx openspec check )'
  );
});

test('rewriteOpenSpecCommand leaves non-command mentions unchanged', () => {
  assert.equal(rewriteOpenSpecCommand('echo openspec && printf "%s" openspec'), 'echo openspec && printf "%s" openspec');
});

test('buildHookResponse allows non-shell payloads unchanged', () => {
  assert.deepEqual(
    buildHookResponse({
      tool_name: 'ReadFile',
      tool_input: {
        path: 'openspec/specs/example.md',
      },
    }),
    {
      permission: 'allow',
    }
  );
});

test('buildHookResponse preserves other shell inputs when rewriting', () => {
  assert.deepEqual(
    buildHookResponse({
      tool_name: 'Shell',
      tool_input: {
        command: 'openspec validate',
        working_directory: '/tmp/example',
        description: 'Validate change',
      },
    }),
    {
      permission: 'allow',
      updated_input: {
        command: 'npx openspec validate',
        working_directory: '/tmp/example',
        description: 'Validate change',
      },
    }
  );
});

test('buildHookResponse allows shell payloads without a command', () => {
  assert.deepEqual(
    buildHookResponse({
      tool_name: 'Shell',
      tool_input: {
        working_directory: '/tmp/example',
      },
    }),
    {
      permission: 'allow',
    }
  );
});

test('hook CLI rewrites stdin payloads end-to-end', () => {
  const result = spawnSync(process.execPath, [hookPath], {
    input: JSON.stringify({
      tool_name: 'Shell',
      tool_input: {
        command: 'openspec validate',
        working_directory: '/tmp/example',
      },
    }),
    encoding: 'utf8',
  });

  assert.equal(result.status, 0);
  assert.equal(result.stderr, '');
  assert.deepEqual(JSON.parse(result.stdout), {
    permission: 'allow',
    updated_input: {
      command: 'npx openspec validate',
      working_directory: '/tmp/example',
    },
  });
});

test('hook CLI falls back to allow on invalid JSON', () => {
  const result = spawnSync(process.execPath, [hookPath], {
    input: '{not valid json',
    encoding: 'utf8',
  });

  assert.equal(result.status, 0);
  assert.equal(result.stderr, '');
  assert.deepEqual(JSON.parse(result.stdout), {
    permission: 'allow',
  });
});
