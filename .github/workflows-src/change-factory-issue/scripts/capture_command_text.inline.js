//include: ../intake-constants.js
//include: ../../lib/factory-issue-shared.js
//include: ../../lib/factory-issue-module.gh.js

const eventName = context.eventName;

if (eventName === 'issue_comment') {
  const body = context.payload.comment?.body ?? '';
  // Strip the leading /change-factory token and surrounding whitespace
  const humanDirection = body.replace(/^\s*\/change-factory\s*/, '').trim();
  core.setOutput('human_direction', humanDirection);
  core.info(`Captured human direction from slash command: "${humanDirection}"`);
} else {
  core.setOutput('human_direction', '');
  core.info('Not an issue_comment event; human_direction is empty.');
}