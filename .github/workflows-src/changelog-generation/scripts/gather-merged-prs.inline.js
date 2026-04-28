const fs = require('fs');
const { execSync } = require('child_process');
//include: ../../lib/changelog-engine-workflow.js

const previousTag = process.env.PREVIOUS_TAG || '';
const compareRange = process.env.COMPARE_RANGE || 'HEAD';
const mode = process.env.MODE || 'unreleased';
const targetVersion = process.env.TARGET_VERSION || '';

const { owner, repo } = context.repo;

const prRecords = await gatherMergedPRRecordsForRange({
  github,
  owner,
  repo,
  compareRange,
  exec: execSync,
  core,
});

const manifest = {
  generated_at: new Date().toISOString(),
  mode,
  target_version: targetVersion,
  previous_tag: previousTag,
  compare_range: compareRange,
  pr_count: prRecords.length,
  pull_requests: prRecords,
};

const artifactDir = '/tmp/changelog-assembly';
const artifactPath = `${artifactDir}/merged-prs.json`;
fs.mkdirSync(artifactDir, { recursive: true });
fs.writeFileSync(artifactPath, JSON.stringify(manifest, null, 2), 'utf8');

core.setOutput('merged_prs_path', artifactPath);
const hasPRs = prRecords.length > 0;
core.setOutput('has_prs', hasPRs ? 'true' : 'false');
core.info(`Merged PR manifest written to ${artifactPath} (${prRecords.length} PRs)`);
