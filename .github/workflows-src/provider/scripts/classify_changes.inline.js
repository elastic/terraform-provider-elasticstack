//include: ../../lib/classify-changes.js

let changedFiles = [];

if (context.eventName === 'pull_request') {
  try {
    const files = await github.paginate(github.rest.pulls.listFiles, {
      owner: context.repo.owner,
      repo: context.repo.repo,
      pull_number: context.payload.pull_request.number,
      per_page: 100,
    });
    changedFiles = files.map((f) => f.filename);
  } catch (err) {
    core.warning(`Failed to list PR files: ${err.message}; defaulting to provider_changes=true.`);
    core.setOutput('provider_changes', 'true');
    return;
  }
} else if (context.eventName === 'push') {
  core.setOutput('provider_changes', 'true');
  core.info('Push event: unconditionally running provider CI.');
  return;
} else {
  // workflow_dispatch or other events: conservative default
  core.setOutput('provider_changes', 'true');
  core.info(`${context.eventName}: defaulting to provider_changes=true.`);
  return;
}

const result = classifyChanges(changedFiles);
core.setOutput('provider_changes', result.providerChanges);
