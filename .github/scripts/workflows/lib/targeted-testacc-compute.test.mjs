import assert from "node:assert/strict";
import test from "node:test";
import { execFileSync } from "node:child_process";
import {
  chmodSync,
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const scriptPath = path.resolve(
  __dirname,
  "..",
  "..",
  "targeted-testacc-compute.sh",
);

// Creates a temp dir containing fake `git` and `go` executables that log their
// argv to <dir>/git.log / <dir>/go.log and whose behavior is controlled by the
// returned knobs. Returns { stubDir, logDir, controls } and a cleanup fn.
function makeStubs({ gitFetchOk = true, goExit = 0, goOutput = "" } = {}) {
  const root = mkdtempSync(path.join(tmpdir(), "targeted-testacc-compute-"));
  const stubDir = path.join(root, "bin");
  const logDir = path.join(root, "log");
  mkdirSync(stubDir);
  mkdirSync(logDir);

  const gitLog = path.join(logDir, "git.log");
  const goLog = path.join(logDir, "go.log");
  writeFileSync(gitLog, "");
  writeFileSync(goLog, "");
  const exitCodeDir = path.join(root, "exit");
  mkdirSync(exitCodeDir);
  const goExitFile = path.join(exitCodeDir, "go-exit");
  writeFileSync(goExitFile, `${goExit}\n`);

  writeFileSync(
    path.join(stubDir, "git"),
    `#!/bin/sh
printf '%s\\n' "$@" >> '${gitLog}'
exit ${gitFetchOk ? 0 : 1}
`,
  );
  writeFileSync(
    path.join(stubDir, "go"),
    `#!/bin/sh
printf '%s\\n' "$@" >> '${goLog}'
code=$(cat '${goExitFile}' | tr -d '[:space:]')
if [ "$code" != "0" ]; then
  exit "$code"
fi
cat '${path.join(exitCodeDir, "go-output")}'
`,
  );
  writeFileSync(path.join(exitCodeDir, "go-output"), goOutput);
  for (const name of ["git", "go"]) {
    chmodSync(path.join(stubDir, name), 0o755);
  }

  return { root, stubDir };
}

function runScript({ root, stubDir, env }) {
  const ghOutput = path.join(root, "github-output.txt");
  writeFileSync(ghOutput, "");
  const stdout = execFileSync(scriptPath, {
    env: {
      ...process.env,
      PATH: `${stubDir}${path.delimiter}${process.env.PATH}`,
      GITHUB_OUTPUT: ghOutput,
      ...env,
    },
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
  const outputs = {};
  for (const line of readFileSync(ghOutput, "utf8").split("\n")) {
    if (!line) continue;
    const eq = line.indexOf("=");
    outputs[line.slice(0, eq)] = line.slice(eq + 1);
  }
  return { stdout, outputs };
}

test("non-PR event → has_packages=true, empty targeted_pkgs, no git/go invoked", () => {
  const stubs = makeStubs();
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "push", PR_BASE_SHA: "", SHARD: "0" },
    });
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.targeted_pkgs, "");
    // Neither git nor go should have been invoked for non-PR events.
    assert.equal(
      readFileSync(path.join(stubs.root, "log", "git.log"), "utf8"),
      "",
    );
    assert.equal(
      readFileSync(path.join(stubs.root, "log", "go.log"), "utf8"),
      "",
    );
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event with successful base fetch → tool invoked with --base=refs/remotes/origin/pr-base", () => {
  const stubs = makeStubs({
    goOutput: "./internal/acctest/a\n./internal/acctest/b\n",
  });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef", SHARD: "1" },
    });
    const gitArgs = readFileSync(
      path.join(stubs.root, "log", "git.log"),
      "utf8",
    )
      .split("\n")
      .filter(Boolean);
    assert.deepEqual(gitArgs, [
      "fetch",
      "origin",
      "+deadbeef:refs/remotes/origin/pr-base",
      "--depth=1",
    ]);
    const goArgs = readFileSync(path.join(stubs.root, "log", "go.log"), "utf8")
      .split("\n")
      .filter(Boolean);
    assert.deepEqual(goArgs, [
      "run",
      "./scripts/targeted-testacc/...",
      "--base=refs/remotes/origin/pr-base",
      "--total-shards=2",
      "--shard-index=1",
    ]);
    assert.equal(outputs.has_packages, "true");
    assert.equal(
      outputs.targeted_pkgs,
      "./internal/acctest/a ./internal/acctest/b",
    );
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event with failed base fetch → tool invoked WITHOUT --base", () => {
  const stubs = makeStubs({ gitFetchOk: false, goOutput: "./pkg/a\n" });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef", SHARD: "0" },
    });
    const goArgs = readFileSync(path.join(stubs.root, "log", "go.log"), "utf8")
      .split("\n")
      .filter(Boolean);
    assert.deepEqual(goArgs, [
      "run",
      "./scripts/targeted-testacc/...",
      "--total-shards=2",
      "--shard-index=0",
    ]);
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.targeted_pkgs, "./pkg/a");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, tool failure → exit 1 and ::error:: line (fail-closed)", () => {
  const stubs = makeStubs({ goExit: 1, goOutput: "" });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    let stdout = "";
    try {
      stdout = execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
          SHARD: "0",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch (err) {
      threw = true;
      stdout = err.stdout;
    }
    assert.equal(threw, true, "script must exit non-zero when the tool fails");
    assert.match(
      stdout,
      /::error::targeted-testacc failed; refusing to skip tests/,
    );
    // has_packages must not be set to a skipping value.
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, empty tool output → has_packages=false", () => {
  const stubs = makeStubs({ goOutput: "\n" });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef", SHARD: "0" },
    });
    assert.equal(outputs.has_packages, "false");
    assert.equal(outputs.targeted_pkgs, undefined);
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});
