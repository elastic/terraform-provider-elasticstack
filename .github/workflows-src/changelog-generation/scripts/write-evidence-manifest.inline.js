const fs = require('fs');
//include: ../../lib/changelog-evidence-manifest.js

const evidenceJson = resolveEvidenceJsonInput({
  envEvidenceJson: process.env.EVIDENCE_JSON,
});

let writePlan;
try {
  writePlan = buildEvidenceManifestWrite({ evidenceJson });
} catch (err) {
  core.setFailed(err.message);
  process.exit(1);
}

// Ensure the target directory exists
fs.mkdirSync(writePlan.directory, { recursive: true });

// Write the evidence manifest to workflow memory
fs.writeFileSync(writePlan.memoryPath, writePlan.formattedJson, 'utf8');

core.info(`Evidence manifest written to ${writePlan.memoryPath} (${writePlan.prCount} PRs)`);
core.setOutput('evidence_ready', 'true');
core.setOutput('evidence_path', writePlan.memoryPath);
