const fs = require('fs');
const path = require('path');

const evidenceJson = core.getInput('evidence_json') || process.env.INPUT_EVIDENCE_JSON || '';
const memoryPath = '/tmp/gh-aw/agent/evidence.json';

if (!evidenceJson) {
  core.setFailed('No evidence_json input provided');
  process.exit(1);
}

// Validate JSON
let parsed;
try {
  parsed = JSON.parse(evidenceJson);
} catch (err) {
  core.setFailed(`Invalid JSON in evidence_json: ${err.message}`);
  process.exit(1);
}

// Ensure the target directory exists
const dir = path.dirname(memoryPath);
fs.mkdirSync(dir, { recursive: true });

// Write the evidence manifest to workflow memory
fs.writeFileSync(memoryPath, JSON.stringify(parsed, null, 2), 'utf8');

core.info(`Evidence manifest written to ${memoryPath} (${parsed.pr_count ?? '?'} PRs)`);
core.setOutput('evidence_ready', 'true');
core.setOutput('evidence_path', memoryPath);
