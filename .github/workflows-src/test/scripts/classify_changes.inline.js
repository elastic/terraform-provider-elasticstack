//include: ../../lib/classify-changes.js

let changedFiles = [];

if (context.eventName === 'pull_request') {
  const files = await github.paginate(github.rest.pulls.listFiles, {
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.pull_request.number,
    per_page: 100,
  });
  changedFiles = files.map((f) => f.filename);
} else if (context.eventName === 'push') {
  const commits = context.payload.commits || [];
  for (const commit of commits) {
    changedFiles.push(...(commit.added || []));
    changedFiles.push(...(commit.modified || []));
    changedFiles.push(...(commit.removed || []));
  }
  if (changedFiles.length === 0) {
    core.info('Push event has no file list in commits payload (may be a force-push or API limitation); defaulting to provider_changes=true (conservative).');
  }
} else {
  // workflow_dispatch or other events: conservative default
  core.setOutput('provider_changes', 'true');
  return;
}

const result = classifyChanges(changedFiles);
core.setOutput('provider_changes', result.providerChanges);
