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

test("non-PR event → fixed shards [0,1], has_packages=true, no selector invoked", () => {
  const stubs = makeStubs();
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "push", PR_BASE_SHA: "" },
    });
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.shards, "[0,1]");
    assert.equal(outputs.packages_0, "");
    assert.equal(outputs.packages_1, "");
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

test("PR event with successful base fetch → selector invoked once with --base, no shard index", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: ["./internal/acctest/a", "./internal/acctest/b"],
      shards: [["./internal/acctest/a", "./internal/acctest/b"]],
      rationale: ["diff against the resolved baseline changed 2 files"],
    }),
  });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef" },
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
    ]);
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.shards, "[0]");
    assert.equal(
      outputs.packages_0,
      "./internal/acctest/a ./internal/acctest/b",
    );
    assert.equal(outputs.packages_1, "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event with failed base fetch → selector invoked WITHOUT --base", () => {
  const stubs = makeStubs({
    gitFetchOk: false,
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: ["./pkg/a"],
      shards: [["./pkg/a"]],
      rationale: [],
    }),
  });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef" },
    });
    const goArgs = readFileSync(
      path.join(stubs.root, "log", "go.log"),
      "utf8",
    )
      .split("\n")
      .filter(Boolean);
    assert.deepEqual(goArgs, [
      "run",
      "./scripts/targeted-testacc/...",
      "--total-shards=2",
    ]);
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.shards, "[0]");
    assert.equal(outputs.packages_0, "./pkg/a");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, selector failure → exit non-zero, ::error:: on stderr, no outputs (fail-closed)", () => {
  const stubs = makeStubs({ goExit: 1, goOutput: "" });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    let stderr = "";
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch (err) {
      threw = true;
      stderr = err.stderr;
    }
    assert.equal(threw, true, "script must exit non-zero when the selector fails");
    assert.match(
      stderr,
      /::error::targeted-testacc failed; refusing to skip tests/,
    );
    // has_packages must not be set to a skipping value.
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, zero-package plan → has_packages=false, one no-op shard", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: false,
      selected_packages: [],
      shards: [[]],
      rationale: ["no changed files map to a Go package; zero packages selected"],
    }),
  });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef" },
    });
    assert.equal(outputs.has_packages, "false");
    assert.equal(outputs.shards, "[0]");
    assert.equal(outputs.packages_0, "");
    assert.equal(outputs.packages_1, "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, selector emits malformed JSON → exit non-zero, no outputs (fail-closed)", () => {
  const stubs = makeStubs({ goOutput: "not json at all" });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    let stderr = "";
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch (err) {
      threw = true;
      stderr = err.stderr;
    }
    assert.equal(threw, true, "script must exit non-zero on malformed plan JSON");
    assert.match(
      stderr,
      /invalid targeted-testacc shard plan; refusing to skip tests/,
    );
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, has_packages=false with populated selection → exit non-zero (fail-closed)", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: false,
      selected_packages: ["./pkg/a"],
      shards: [["./pkg/a"]],
      rationale: [],
    }),
  });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch {
      threw = true;
    }
    assert.equal(threw, true, "plan must be rejected when has_packages disagrees with the selection");
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, selected package missing from shards → exit non-zero (fail-closed)", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: ["./pkg/a", "./pkg/b"],
      shards: [["./pkg/a"]],
      rationale: [],
    }),
  });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch {
      threw = true;
    }
    assert.equal(threw, true, "plan must be rejected when a selected package is not assigned to any shard");
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, package duplicated across shards → exit non-zero (fail-closed)", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: ["./pkg/a", "./pkg/b"],
      shards: [["./pkg/a", "./pkg/a"], ["./pkg/b"]],
      rationale: [],
    }),
  });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch {
      threw = true;
    }
    assert.equal(threw, true, "plan must be rejected when a package is assigned to more than one shard");
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, empty shard in nonempty plan → exit non-zero (fail-closed)", () => {
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: ["./pkg/a", "./pkg/b"],
      shards: [["./pkg/a", "./pkg/b"], []],
      rationale: [],
    }),
  });
  try {
    const ghOutput = path.join(stubs.root, "github-output.txt");
    writeFileSync(ghOutput, "");
    let threw;
    try {
      execFileSync(scriptPath, {
        env: {
          ...process.env,
          PATH: `${stubs.stubDir}${path.delimiter}${process.env.PATH}`,
          GITHUB_OUTPUT: ghOutput,
          EVENT_NAME: "pull_request",
          PR_BASE_SHA: "deadbeef",
        },
        encoding: "utf8",
      });
      threw = false;
    } catch {
      threw = true;
    }
    assert.equal(threw, true, "plan must be rejected when a nonempty plan contains an empty shard");
    assert.equal(readFileSync(ghOutput, "utf8"), "");
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});

test("PR event, 30-package plan → two populated shards covering the selection exactly", () => {
  const selected = Array.from(
    { length: 30 },
    (_, i) => `./internal/acctest/p${String(i).padStart(2, "0")}`,
  );
  const shards = [0, 1].map((idx) =>
    selected.filter((_, i) => i % 2 === idx),
  );
  const stubs = makeStubs({
    goOutput: JSON.stringify({
      has_packages: true,
      selected_packages: selected,
      shards,
      rationale: ["diff against the resolved baseline changed 30 files"],
    }),
  });
  try {
    const { outputs } = runScript({
      ...stubs,
      env: { EVENT_NAME: "pull_request", PR_BASE_SHA: "deadbeef" },
    });
    const goArgs = readFileSync(path.join(stubs.root, "log", "go.log"), "utf8")
      .split("\n")
      .filter(Boolean);
    assert.deepEqual(goArgs, [
      "run",
      "./scripts/targeted-testacc/...",
      "--base=refs/remotes/origin/pr-base",
      "--total-shards=2",
    ]);
    assert.equal(outputs.has_packages, "true");
    assert.equal(outputs.shards, "[0,1]");
    const shard0 = shards[0].join(" ");
    const shard1 = shards[1].join(" ");
    assert.equal(outputs.packages_0, shard0);
    assert.equal(outputs.packages_1, shard1);
    const union = outputs.packages_0
      .split(" ")
      .filter(Boolean)
      .concat(outputs.packages_1.split(" ").filter(Boolean))
      .sort();
    assert.deepEqual(union, [...selected].sort());
  } finally {
    rmSync(stubs.root, { recursive: true, force: true });
  }
});
