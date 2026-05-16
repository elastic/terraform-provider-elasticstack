# worktrunk-project-config Specification

## Purpose
TBD - created by archiving change worktrunk-config. Update Purpose after archive.
## Requirements
### Requirement: Worktree receives generated .env on creation
When a new worktree is created, the blocking `pre-start` hook pipeline SHALL generate a `.env` file in the worktree root by copying `.env.template` and appending acceptance-test environment variables after `make setup` completes. The generated `.env` SHALL define `ELASTICSEARCH_PORT`, `KIBANA_PORT`, `ELASTICSEARCH_URL`, `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `KIBANA_ENDPOINT`, and `KIBANA_USERNAME`. The port values SHALL be deterministically derived from the branch name and SHALL remain in the range 10000–19999. To guarantee that Elasticsearch and Kibana ports for the same worktree do not collide, `ELASTICSEARCH_PORT` SHALL be assigned from the subrange 10000–14999 and `KIBANA_PORT` SHALL be assigned from the subrange 15000–19999. The generated `.env` SHALL NOT define `TF_ACC`.

#### Scenario: New worktree gets distinct ES and KB ports
- **WHEN** a new worktree is created for branch `feature-x`
- **THEN** `.env` in that worktree contains an `ELASTICSEARCH_PORT` value in the range 10000–14999 and a `KIBANA_PORT` value in the range 15000–19999, both derived deterministically from `feature-x`

#### Scenario: ES and KB ports cannot collide within a worktree
- **WHEN** a new worktree is created for any branch
- **THEN** the generated `ELASTICSEARCH_PORT` and `KIBANA_PORT` values are different from each other

#### Scenario: .env template static values are preserved
- **WHEN** a new worktree is created
- **THEN** the generated `.env` contains all values from `.env.template` (e.g. `STACK_VERSION`, `ELASTICSEARCH_PASSWORD`, `GOVERSION`) unchanged

#### Scenario: Acceptance test connection variables are ready without manual export edits
- **WHEN** a new worktree is created
- **THEN** the generated `.env` contains `ELASTICSEARCH_ENDPOINTS=http://localhost:<ELASTICSEARCH_PORT>`, `ELASTICSEARCH_USERNAME=elastic`, `KIBANA_ENDPOINT=http://localhost:<KIBANA_PORT>`, and `KIBANA_USERNAME=elastic`
- **AND** a developer can run `TF_ACC=1 go test ...` after exporting `.env` without separately defining those variables

#### Scenario: TF_ACC remains opt-in
- **WHEN** a new worktree is created
- **THEN** the generated `.env` does not define `TF_ACC`

### Requirement: Docker Compose stacks are isolated per worktree
The `docker-compose.yml` file SHALL NOT contain any `container_name:` directives. Docker Compose SHALL derive container names from the project name (the worktree directory name), ensuring containers and volumes are namespaced per worktree.

#### Scenario: Containers do not conflict across worktrees
- **WHEN** `docker compose up` is run in two different worktrees simultaneously
- **THEN** each worktree's containers have distinct names (prefixed with the worktree directory name) and no naming conflicts occur

#### Scenario: Volumes are isolated per worktree
- **WHEN** `docker compose up` is run in two different worktrees
- **THEN** each worktree's named volumes (elasticsearch data, kibana data, fleet data) are distinct Docker volumes

### Requirement: make setup runs automatically on new worktree creation
The `pre-start` hook pipeline SHALL run `make setup` when a new worktree is created. This pipeline is blocking, and the `.env` generation step SHALL run only after `make setup` completes successfully.

#### Scenario: New worktree has Go dependencies and OpenSpec CLI ready
- **WHEN** a new worktree is created
- **THEN** `make setup` completes (Go vendor cache populated, OpenSpec CLI installed) before the worktree becomes usable

### Requirement: make check-lint runs before every commit
The `pre-commit` hook SHALL run `make check-lint`. A non-zero exit code SHALL abort the commit.

#### Scenario: Commit blocked on lint failure
- **WHEN** `wt step commit` or `wt merge` is invoked and `make check-lint` exits non-zero
- **THEN** the commit is aborted and the lint output is shown to the user

#### Scenario: Commit proceeds on lint pass
- **WHEN** `wt step commit` or `wt merge` is invoked and `make check-lint` exits zero
- **THEN** the commit proceeds normally

### Requirement: Default Docker Compose stack is torn down when a worktree is removed
The `pre-remove` hook SHALL run `docker compose down --volumes` in the worktree being removed to tear down the default Docker Compose stack for that worktree. This runs while the worktree directory still exists so Docker Compose can resolve the correct project name from the working directory.

#### Scenario: Default stack stops on worktree removal
- **WHEN** `wt remove` is run in a worktree that has a running default docker compose stack
- **THEN** `docker compose down --volumes` completes before the worktree directory is deleted, stopping the default stack's containers and removing its volumes

#### Scenario: Removal succeeds even if no default stack is running
- **WHEN** `wt remove` is run in a worktree with no running default docker compose stack
- **THEN** `docker compose down --volumes` exits cleanly (Docker Compose reports nothing to stop) and worktree removal proceeds

### Requirement: Makefile port targets use configurable variables
The Makefile SHALL define `ELASTICSEARCH_PORT ?= 9200` and `KIBANA_PORT ?= 5601` with `?=` defaults. All targets that currently hardcode `localhost:9200` or `localhost:5601` SHALL use `$(ELASTICSEARCH_PORT)` and `$(KIBANA_PORT)` instead.

#### Scenario: Default ports apply when .env is not loaded
- **WHEN** a Makefile target that references `$(ELASTICSEARCH_PORT)` is run without exporting `.env`
- **THEN** the target uses port 9200 for Elasticsearch and 5601 for Kibana

#### Scenario: Worktree-specific ports apply when .env is exported
- **WHEN** a developer exports `.env` before running a Makefile target
- **THEN** the target uses the port values from `.env` rather than the defaults

### Requirement: .env.template is committed and contains static configuration
A file `.env.template` SHALL be committed to the repository. It SHALL contain all static docker and test configuration values currently in `.env` (stack version, passwords, Java opts, Go version, feature flags, fleet image). It SHALL NOT contain `container_name` variables or port variables (`ELASTICSEARCH_PORT`, `KIBANA_PORT`, `ELASTICSEARCH_URL`).

#### Scenario: .env.template covers all static values
- **WHEN** a developer inspects `.env.template`
- **THEN** it contains `STACK_VERSION`, `ELASTICSEARCH_PASSWORD`, `KIBANA_PASSWORD`, `KIBANA_ENCRYPTION_KEY`, `ELASTICSEARCH_JAVA_OPTS`, `GOVERSION`, `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL`, and `FLEET_IMAGE`

#### Scenario: .env.template contains no container names or port values
- **WHEN** a developer inspects `.env.template`
- **THEN** it contains no `*_CONTAINER_NAME` variables, no `ELASTICSEARCH_PORT`, no `KIBANA_PORT`, and no `ELASTICSEARCH_URL`

