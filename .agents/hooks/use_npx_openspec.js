#!/usr/bin/env node

async function readStdin(stream = process.stdin) {
  const chunks = [];

  for await (const chunk of stream) {
    chunks.push(chunk);
  }

  return Buffer.concat(chunks).toString("utf8");
}

function rewriteOpenSpecCommand(command) {
  if (typeof command !== "string" || command.length === 0) {
    return command;
  }

  return command.replace(/(^|[;&|()])(\s*)openspec(?=\s|$)/g, "$1$2npx openspec");
}

function allow(extra = {}) {
  return {
    permission: "allow",
    ...extra,
  };
}

function buildHookResponse(payload) {
  const toolInput = payload?.tool_input ?? {};
  const command = typeof toolInput.command === "string" ? toolInput.command : "";

  if (payload?.tool_name !== "Shell" || !command) {
    return allow();
  }

  const rewrittenCommand = rewriteOpenSpecCommand(command);

  if (rewrittenCommand === command) {
    return allow();
  }

  return allow({
    updated_input: {
      ...toolInput,
      command: rewrittenCommand,
    },
  });
}

async function main({ input = process.stdin, output = process.stdout } = {}) {
  try {
    const rawInput = await readStdin(input);
    const payload = rawInput ? JSON.parse(rawInput) : {};

    output.write(JSON.stringify(buildHookResponse(payload)));
  } catch (_error) {
    output.write(JSON.stringify(allow()));
  }
}

module.exports = {
  allow,
  buildHookResponse,
  main,
  readStdin,
  rewriteOpenSpecCommand,
};

if (require.main === module) {
  void main();
}
