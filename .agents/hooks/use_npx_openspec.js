#!/usr/bin/env node

main();

async function main() {
  const chunks = [];

  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }

  const rawInput = Buffer.concat(chunks).toString("utf8");

  try {
    const payload = rawInput ? JSON.parse(rawInput) : {};
    const toolInput = payload.tool_input ?? {};
    const command = typeof toolInput.command === "string" ? toolInput.command : "";

    if (payload.tool_name !== "Shell" || !command) {
      allow();
      return;
    }

    const rewrittenCommand = command.replace(/(^|[;&|()])(\s*)openspec(?=\s|$)/g, "$1$2npx openspec");

    if (rewrittenCommand === command) {
      allow();
      return;
    }

    allow({
      updated_input: {
        ...toolInput,
        command: rewrittenCommand,
      },
    });
  } catch (_error) {
    allow();
  }
}

function allow(extra = {}) {
  process.stdout.write(
    JSON.stringify({
      permission: "allow",
      ...extra,
    }),
  );
}
