//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/change-factory-issue.gh.js

const { owner, repo } = context.repo;
const sender = context.payload.sender?.login ?? '';

if (!sender) {
  const missing = actorTrustWhenSenderMissing();
  core.setOutput('actor_trusted', 'false');
  core.setOutput('actor_trusted_reason', missing.actor_trusted_reason);
  core.info('Actor not trusted: sender login is missing from the event payload.');
} else {
  let permission = null;
  if (sender !== 'github-actions[bot]') {
    const { data } = await github.rest.repos.getCollaboratorPermissionLevel({
      owner,
      repo,
      username: sender,
    });
    permission = data.permission;
  }

  const result = checkActorTrust({ sender, permission });

  core.setOutput('actor_trusted', result.actor_trusted ? 'true' : 'false');
  core.setOutput('actor_trusted_reason', result.actor_trusted_reason);

  if (result.actor_trusted) {
    core.info(`Actor trusted: ${result.actor_trusted_reason}`);
  } else {
    core.info(`Actor not trusted: ${result.actor_trusted_reason}`);
  }
}
