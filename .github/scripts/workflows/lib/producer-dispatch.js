/**
 * Producer-side dispatch helpers for fanning out code-factory runs
 * from safe-output temporary issue ID maps.
 */
'use strict';

const fs = require('fs');
const { spawnSync } = require('child_process');

/**
 * Parse and validate a temporary issue ID map produced by gh-aw safe outputs.
 *
 * Expected shape:
 *   {
 *     "<temporary-id>": { "repo": "owner/repo", "number": 123 },
 *     ...
 *   }
 *
 * @param {string} filePath
 * @returns {{ repo: string, number: number }[]}
 * @throws {Error} when the file is missing, malformed, or contains invalid entries
 */
function parseTemporaryIdMap(filePath) {
  if (!fs.existsSync(filePath)) {
    throw new Error(`Temporary ID map not found at ${filePath}`);
  }

  let raw;
  try {
    raw = JSON.parse(fs.readFileSync(filePath, 'utf8'));
  } catch (err) {
    throw new Error(`Failed to parse temporary ID map: ${err.message}`);
  }

  if (raw === null || typeof raw !== 'object' || Array.isArray(raw)) {
    throw new Error(
      `Temporary ID map must be a JSON object, got ${raw === null ? 'null' : typeof raw}`
    );
  }

  const entries = [];
  for (const [tempId, value] of Object.entries(raw)) {
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
      throw new Error(
        `Entry "${tempId}" must be an object, got ${value === null ? 'null' : typeof value}`
      );
    }
    if (typeof value.repo !== 'string' || !value.repo.includes('/')) {
      throw new Error(
        `Entry "${tempId}" has invalid repo: ${value.repo}`
      );
    }
    const num = Number(value.number);
    if (!Number.isInteger(num) || num <= 0) {
      throw new Error(
        `Entry "${tempId}" has invalid number: ${value.number}`
      );
    }
    entries.push({ repo: value.repo, number: num });
  }

  return entries;
}

/**
 * Dispatch the code-factory-issue workflow for each issue entry.
 *
 * @param {{ repo: string, number: number }[]} entries
 * @param {string} sourceWorkflow – e.g. "semantic-function-refactor"
 * @param {string} [workflowFile='code-factory-issue.lock.yml']
 */
function dispatchCodeFactory(entries, sourceWorkflow, workflowFile = 'code-factory-issue.lock.yml') {
  const ghToken = process.env.GH_TOKEN || process.env.GITHUB_TOKEN;
  if (!ghToken) {
    throw new Error('GH_TOKEN or GITHUB_TOKEN environment variable is required');
  }

  const allowedRepo = process.env.GITHUB_REPOSITORY;
  for (const entry of entries) {
    if (allowedRepo && entry.repo !== allowedRepo) {
      throw new Error(`Refusing to dispatch to ${entry.repo}: not the current repository (${allowedRepo})`);
    }
  }

  const env = { ...process.env, GH_TOKEN: ghToken };

  for (const entry of entries) {
    const cmd = [
      'gh', 'workflow', 'run', workflowFile,
      '--repo', entry.repo,
      '--field', `issue_number=${entry.number}`,
      '--field', `source_workflow=${sourceWorkflow}`,
    ];

    // eslint-disable-next-line no-console
    console.log(`Dispatching ${workflowFile} for issue #${entry.number} in ${entry.repo}`);
    const result = spawnSync(cmd[0], cmd.slice(1), { env, stdio: 'inherit' });
    if (result.status !== 0 || result.error) {
      const message = result.error?.message
        || result.stderr?.toString()
        || `Process exited with code ${result.status}`;
      throw new Error(
        `Failed to dispatch for issue #${entry.number} in ${entry.repo}: ${message}`
      );
    }
  }
}

function main() {
  const [mapPath, sourceWorkflow, workflowFile] = process.argv.slice(2);
  if (!mapPath || !sourceWorkflow) {
    // eslint-disable-next-line no-console
    console.error('Usage: node producer-dispatch.js <temporary-id-map.json> <source-workflow> [workflow-file]');
    process.exit(1);
  }

  const entries = parseTemporaryIdMap(mapPath);
  if (entries.length === 0) {
    // eslint-disable-next-line no-console
    console.log('No created issues found; nothing to dispatch.');
    return;
  }

  // eslint-disable-next-line no-console
  console.log(`Dispatching code-factory for ${entries.length} issue(s)`);
  dispatchCodeFactory(entries, sourceWorkflow, workflowFile);
}

if (require.main === module) {
  main();
}

module.exports = {
  parseTemporaryIdMap,
  dispatchCodeFactory,
};
