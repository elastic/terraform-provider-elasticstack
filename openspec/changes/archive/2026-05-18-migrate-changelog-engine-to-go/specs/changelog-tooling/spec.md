## ADDED Requirements

### Requirement: Single Go CLI under `scripts/changelog/`

The repository SHALL provide a Go command-line tool at `scripts/changelog/` that subsumes all behaviour currently implemented by `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/changelog/*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, and `.github/scripts/workflows/pr-changelog-check/*.js`. The tool MUST be invokable as `go run ./scripts/changelog <subcommand>` from any workflow step that has previously run `actions/setup-go` and `actions/checkout`.

#### Scenario: Tool installs no extra build tooling

- **WHEN** a workflow job already runs `actions/setup-go` and `actions/checkout`
- **THEN** the changelog tool runs via `go run ./scripts/changelog <subcommand>` with no additional setup step required

#### Scenario: Unknown subcommand exits non-zero

- **WHEN** the tool is invoked with a subcommand name that is not one of the documented subcommands
- **THEN** the tool writes a usage message to stderr and exits with a non-zero status code

### Requirement: Subcommand surface

The tool SHALL expose exactly the following subcommands, each preserving the behaviour of the JavaScript module(s) it replaces:

- `gather-evidence` — replaces `changelog/gather-pr-evidence.js` and `lib/changelog-pr-evidence.js`.
- `run-engine` — replaces `changelog/run-changelog-engine.js`, `lib/changelog-engine.js`, `lib/changelog-engine-factory.js`, `lib/changelog-engine-workflow.js`, `lib/changelog-renderer.js`, `lib/changelog-rewriter.js`, and `lib/changelog-release-context.js`.
- `manage-unreleased-pr` — replaces `changelog/manage-unreleased-pr.js` and the relevant pieces of `lib/changelog-pr-management.js`.
- `refresh-release-pr` — replaces `changelog/refresh-release-pr.js` and the relevant pieces of `lib/changelog-pr-management.js`.
- `validate-pr-section` — replaces `pr-changelog-check/check.js`, `lib/pr-changelog-check.js`, and `lib/pr-changelog-parser.js`.

Each subcommand MUST accept its parameters via subcommand-scoped flags parsed with the standard library `flag` package and/or via documented environment variables (e.g. `GITHUB_TOKEN`, `GITHUB_REPOSITORY`, `GITHUB_EVENT_PATH`, `GITHUB_OUTPUT`).

#### Scenario: Each documented subcommand resolves to a JS module behaviour

- **WHEN** a documented subcommand is invoked with valid inputs that previously triggered a behaviour in its JavaScript predecessor
- **THEN** the Go subcommand produces the same externally observable output (CHANGELOG.md mutation, PR comment payload, `$GITHUB_OUTPUT` keys/values, and exit code) as the predecessor on the same inputs

#### Scenario: Subcommand exit codes follow standard convention

- **WHEN** a subcommand encounters an unrecoverable error (invalid inputs, GitHub API failure, malformed PR body when expected to fail, etc.)
- **THEN** the subcommand prints a diagnostic to stderr and exits with a non-zero status code so the calling workflow step fails

### Requirement: Single canonical `## Changelog` section parser

The tool SHALL implement the `## Changelog` PR-body parser exactly once, in an internal package consumed by every subcommand that needs to read PR-body changelog content (at least `gather-evidence`, `run-engine`, and `validate-pr-section`). No subcommand may carry its own parallel parser implementation.

The parser MUST preserve all parsing semantics encoded in `lib/pr-changelog-parser.js` at the time of migration, including:

- Recognising `## Changelog` as the section start and terminating on the next `## ` heading outside any fenced code block.
- Honouring both `` ``` `` and `~~~` fenced code blocks: `## `-prefixed lines inside a fence MUST NOT terminate the section.
- Extracting `Customer impact: <value>` where `<value>` is restricted to the set `{none, fix, enhancement, breaking}`.
- Requiring a non-empty `Summary:` field when `Customer impact` is anything other than `none`.
- Extracting the optional `### Breaking changes` subsection verbatim when present.

#### Scenario: Two subcommands parse the same body identically

- **WHEN** `validate-pr-section` and `gather-evidence` both receive a PR body containing a `## Changelog` section
- **THEN** they produce the same parsed representation of customer impact, summary, and breaking-changes content

#### Scenario: Fenced code block masking is preserved

- **WHEN** a PR body contains a fenced code block inside the `## Changelog` section whose contents begin with `## ` on a line
- **THEN** the parser does NOT treat that line as a section terminator and continues collecting changelog content until either a real `## ` heading outside a fence or end-of-input

#### Scenario: Customer impact non-none requires summary

- **WHEN** a PR body declares `Customer impact: fix|enhancement|breaking` but omits a `Summary:` field or supplies an empty `Summary:`
- **THEN** the parser reports a validation error identifying the missing/empty `Summary:` requirement

### Requirement: Internal package layout enforces separation of concerns

The tool SHALL place all parsing, rewriting, engine, evidence, PR-check, and GitHub-client code inside `scripts/changelog/internal/` so that no caller outside the `scripts/changelog/` module can import these packages. The minimum partitioning is:

- `internal/section` — `## Changelog` PR-body parser and `CHANGELOG.md` entry renderer.
- `internal/rewriter` — `CHANGELOG.md` section and link-table mutation.
- `internal/engine` — unreleased and release mode orchestration.
- `internal/evidence` — evidence gathering and manifest format.
- `internal/prcheck` — PR-body validation pipeline (uses `internal/section`).
- `internal/githubx` — `go-github` client construction, `$GITHUB_EVENT_PATH` parsing, and `$GITHUB_OUTPUT` writing.

#### Scenario: External callers cannot import internal packages

- **WHEN** any Go package outside `scripts/changelog/` attempts to import `scripts/changelog/internal/...`
- **THEN** the Go toolchain rejects the build with the standard `internal` package access error

#### Scenario: Only `internal/githubx` touches env vars and the network

- **WHEN** reviewing the implementation of `internal/section`, `internal/rewriter`, `internal/engine`, `internal/evidence`, or `internal/prcheck`
- **THEN** none of these packages reads `os.Getenv`, calls the GitHub API, or shells out to `git` directly — they accept typed inputs and return typed results

### Requirement: GitHub client convention matches existing `scripts/*` tools

The tool SHALL use `github.com/google/go-github/v86/github` for GitHub REST interactions, authenticated by reading `GITHUB_TOKEN` from the environment and wrapping it via `golang.org/x/oauth2` — matching the pattern established by `scripts/auto-approve/main.go`. The repository slug MUST come from `GITHUB_REPOSITORY`. Event payload context MUST come from `$GITHUB_EVENT_PATH`.

#### Scenario: Missing GITHUB_TOKEN fails fast

- **WHEN** any subcommand that requires GitHub API access is invoked without `GITHUB_TOKEN` set
- **THEN** the subcommand writes a diagnostic identifying the missing variable to stderr and exits with non-zero status

#### Scenario: go-github major version is shared

- **WHEN** the tool is added to the repository
- **THEN** its `go.mod` direct dependency on `github.com/google/go-github` uses the same major version (`v86` at time of writing) as `scripts/auto-approve/`

### Requirement: Workflow outputs are written via `$GITHUB_OUTPUT` heredoc framing

When a subcommand needs to emit a value for a downstream workflow step (replacing JS `core.setOutput(...)` calls), it SHALL write to the file named by the `GITHUB_OUTPUT` environment variable using GitHub's documented multi-line heredoc framing for any value that may contain newlines, and the simple `<name>=<value>` form otherwise. Output key names MUST match the names previously emitted by the JavaScript implementations so existing workflow YAML expressions (`steps.X.outputs.<name>`) continue to resolve.

#### Scenario: Multi-line output uses heredoc framing

- **WHEN** a subcommand emits a multi-line value (e.g. an assembled PR comment body or JSON evidence blob)
- **THEN** the output is written using the `<name>«EOF»\n<content>\n«EOF»\n` form with a unique end-of-file delimiter

#### Scenario: Output key names are preserved

- **WHEN** an existing workflow YAML expression references `steps.<id>.outputs.<key>` that was previously emitted by a JavaScript module
- **THEN** the replacement Go subcommand emits the same `<key>` so the YAML expression continues to resolve

### Requirement: Tests live alongside the Go source under `scripts/changelog/`

Every `.test.mjs` file that previously covered the JavaScript clusters being replaced SHALL have an equivalent Go test covering the same input/output pairs, located in `scripts/changelog/...` or `scripts/changelog/internal/.../*_test.go`. The full Go test suite MUST be exercisable via `go test ./scripts/changelog/...`.

#### Scenario: `go test` exercises every migrated behaviour

- **WHEN** `go test ./scripts/changelog/...` is run on a clean checkout
- **THEN** the tests cover (at minimum) parser corner cases, mode validation, semver tag selection, section rendering, CHANGELOG.md rewriting idempotence, evidence manifest format, and PR-check verdict shape — at parity with the deleted `.test.mjs` cases

#### Scenario: Workflow CI runs the Go tests for this tool

- **WHEN** the `Workflows CI` job runs
- **THEN** `go test ./scripts/changelog/...` is part of the executed test set (covered by the existing `go test ./scripts/...` patterns or an explicit invocation)

### Requirement: Migration leaves no JS changelog modules behind

After the migration commit lands, the repository MUST NOT contain any of: `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/lib/changelog-*.test.mjs`, `.github/scripts/workflows/changelog/*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, `.github/scripts/workflows/lib/pr-changelog-*.test.mjs`, or `.github/scripts/workflows/pr-changelog-check/*.js`. The `Makefile`'s `workflow-test` target MUST NOT reference these files. The agentic `.lock.yml` workflows MUST be regenerated so they no longer inline references to the deleted JS modules.

#### Scenario: No orphaned JS modules remain

- **WHEN** `git ls-files` is run on the working tree after the migration commit
- **THEN** none of the listed JS or `.test.mjs` paths above appear in the output

#### Scenario: `make workflow-test` runs cleanly

- **WHEN** `make workflow-test` is invoked after the migration
- **THEN** it exits zero and does not attempt to enumerate any of the deleted `.test.mjs` files
