# Development workflow

For full contributor guidance (setup, PR expectations), see [`contributing.md`](./contributing.md).

## Typical change loop

1. Prepare an OpenSpec proposal before writing provider code.
   Start with `openspec-explore` when the problem or scope is still fuzzy and you want to investigate the codebase, compare approaches, or clarify requirements without implementing yet.
   Example: "Use `openspec-explore` to think through how dashboard schema alignment should work before we formalize the change."
   Use `openspec-propose` when you are ready to generate an implementation-ready change in one pass. It creates the change plus the key artifacts such as `proposal.md`, `design.md`, and `tasks.md`.
   Example: "Use `openspec-propose` to create `dashboard-api-schema-alignment` with proposal, design, and tasks."
   Use `openspec-new-change` when you want to scaffold the change first and then create artifacts incrementally.
   Example: "Use `openspec-new-change` to start `dashboard-api-schema-alignment` and show me the first artifact template."
   The goal of this step is an approved change under `openspec/changes/<change-name>/` that is ready to review.

2. Open a proposal PR.
   Send the OpenSpec artifacts for review before implementation. In most cases this PR should contain only the proposal artifacts under `openspec/changes/<change-name>/`.
   This is the point to resolve scope, requirements, and design questions before code lands.

3. Implement the approved proposal.
   Use `openspec-apply-change` to read the change context, work through the task list, make the code changes, and update task checkboxes as work completes.
   Example: "Use `openspec-apply-change` for `dashboard-api-schema-alignment` and implement the remaining tasks."
   Use `openspec-continue-change` if the change is not fully apply-ready yet, or if review/implementation feedback means you need to create the next artifact before continuing.
   Example: "Use `openspec-continue-change` for `dashboard-api-schema-alignment` and create the next required artifact."
   Use `openspec-implementation-loop` when you want a more automated end-to-end loop around a single approved change, including implementation, local review, push, and optional PR handling.
   Example: "Use `openspec-implementation-loop` for `dashboard-api-schema-alignment` in PR mode."
   During implementation, add or update acceptance tests for new behavior and bug fixes. For bugs, verify the new test fails first so it reproduces the original issue.
   Make small, reviewable changes. Keep generated artifacts up to date (docs and generated clients when applicable). Run the narrowest tests that prove correctness, then broaden as appropriate.
   The System User resource (see `internal/elasticsearch/security/system_user` referenced from [`coding-standards.md`](./coding-standards.md)) is the canonical example for new resources. Follow it.

4. Verify the implementation against the spec.
   Run `openspec-verify-change` to check completeness, correctness, and coherence against the approved change artifacts.
   Example: "Use `openspec-verify-change` for `dashboard-api-schema-alignment` and report any gaps before we open the implementation PR."
   Address any verification findings before moving on.

5. Open the implementation PR.
   Once the change is implemented and verified, open a separate PR for the provider code and any related generated artifacts.
   Link back to the approved proposal change so reviewers can compare the implementation with the agreed requirements.

## Common make targets

The canonical list is the root `Makefile`, but the usual ones are:

- `make lint`
- `make test`
- `make testacc` (requires Docker and `TF_ACC=1`)
- `make docs-generate`

## Parallel development with worktrunk

[worktrunk](https://github.com/elastic/worktrunk) manages feature worktrees for this repository so multiple branches can be developed in parallel without switching the main working tree.

### Shell integration

Install the shell hook once to get the `wt` alias and tab completion:

```bash
wt config shell install
```

After reloading your shell profile, you can use `wt <branch>` to create or switch to a feature worktree, and `wt` commands will have tab completion.

### User configuration

Keep worktrees inside the bare repo by setting the worktree path template in `~/.config/worktrunk/config.toml`:

```toml
worktree-path = "{{ repo_path }}/{{ branch | sanitize }}"
```

Each feature worktree becomes a subdirectory of the bare repo directory, keeping related branches discoverable and avoiding scattered worktrees across the filesystem.

### Environment in a feature worktree

When a new worktree is created the blocking `pre-start` hook pipeline (`.config/wt.toml`) runs `make setup` and then generates a `.env` from `.env.template` with per-worktree port variables derived deterministically from the branch name, plus the acceptance-test connection variables (`ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `KIBANA_ENDPOINT`, `KIBANA_USERNAME`). `TF_ACC` is intentionally not written so acceptance mode remains opt-in.

The main checkout's `.env` may not contain port variables if it predates the worktrunk setup; port variables are generated only in worktrees created via `wt switch --create`.

Before running Makefile targets that talk directly to Elasticsearch or Kibana on `localhost`, or before running acceptance tests directly with `go test`, export the worktree's `.env` so the generated connection variables are visible in your shell:

```bash
set -a; . ./.env; set +a
# Then run port-dependent targets, for example:
make testacc-vs-docker
make set-kibana-password
make setup-synthetics
make create-es-api-key
make create-es-bearer-token
make setup-kibana-fleet

# Or run acceptance tests directly:
TF_ACC=1 go test -v ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1
```

Targets that use `docker compose` (for example `make docker-elasticsearch`, `make docker-kibana`, and `make docker-fleet`) automatically read `.env` from the current directory, so they do not require the export step.

Alternatively, pass the variables explicitly on the command line:

```bash
make testacc-vs-docker ELASTICSEARCH_PORT=12345 KIBANA_PORT=16789
```

### Cleanup

When a worktree is removed (`wt remove`), the `pre-remove` hook (`docker compose down --volumes`) automatically tears down the Docker Compose stack for that worktree.

## Example snippets (`examples/resources`, `examples/data-sources`)

These trees hold copy-paste-ready Terraform for this provider. Snippets **may** be surfaced on generated reference pages (`docs/resources/`, `docs/data-sources/`), in docs templates, or in guides—not every `.tf` is shown on every page, but **each covered file** participates in validation below.

Regardless of how a file is surfaced, contributions must satisfy both of the following:

- **Self-contained modules:** A file must not depend on declarations that exist only in a sibling `.tf` in the same directory (locals, variables, resources, data sources copied from another file).
- **Plan-only acceptance coverage:** `TestAccExamples_planOnly` in `internal/acctest/` plans every covered example in isolation against the provider (with `TF_ACC=1` and the usual Elasticsearch/Kibana environment variables used elsewhere in acceptance tests).

If you touch or add snippets, run the harness targeted at your change—for example:

`TF_ACC=1 go test ./internal/acctest -run '^TestAccExamples_planOnly$' -count=1`

Some paths are intentionally skipped in the harness (documented beside the harness); those remain rare exceptions.
