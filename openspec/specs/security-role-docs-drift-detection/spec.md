# Security Role Docs Drift Detection

## Purpose

Define the canonical requirements for detecting and remediating drift between the provider's documented Kibana feature privilege reference and the live Kibana features API, including the curated features manifest, pre-activation diff script, and self-healing workflow behavior.

## Requirements

### Requirement: Curated features JSON file exists as source of truth

A machine-readable file SHALL exist at `scripts/security-role-docs/kibana-features.json` that records which Kibana feature IDs are documented in the guide (`documented` array) and which have been deliberately excluded (`skip` array). Features in neither array are "unknown" and SHALL trigger the drift detection workflow.

The JSON schema SHALL be:
```json
{
  "documented": ["<feature-id>", ...],
  "skip": ["<feature-id>", ...]
}
```

#### Scenario: File exists with required structure

- **WHEN** `scripts/security-role-docs/kibana-features.json` is read
- **THEN** it is valid JSON containing a `documented` array and a `skip` array of strings
- **THEN** the `documented` array contains at minimum: `discover`, `dashboard`, `siem`, `fleet`, `apm`, `osquery`

#### Scenario: File reflects guide content at time of authoring

- **WHEN** the `documented` array is compared against the feature names in the guide's reference table
- **THEN** every feature listed in the guide's table appears in the `documented` array

---

### Requirement: Go pre-activation script computes drift

A Go pre-activation script SHALL exist at `scripts/security-role-docs/` (invoked as `go run ./scripts/security-role-docs pre-activation`) that:

1. Calls `GET /api/features` against the Kibana instance configured via standard provider env vars (`KIBANA_ENDPOINT`, `KIBANA_USERNAME`, `KIBANA_PASSWORD`)
2. Reads `scripts/security-role-docs/kibana-features.json`
3. Computes the set of feature IDs returned by the API that appear in neither `documented` nor `skip`
4. Writes a drift report JSON to a configurable `--report-path`
5. Outputs a `run_agent` GitHub Actions step output: `true` if unknown features were found or documented features are missing from the API response, `false` otherwise

The script SHALL accept flags: `--features-path`, `--report-path`, `--issue-cap`.

#### Scenario: Script detects net-new features

- **WHEN** the Kibana API returns a feature ID not present in `documented` or `skip`
- **THEN** the script writes a report with `unknown_features` containing that ID
- **THEN** `run_agent` output is `true`

#### Scenario: Script is quiet when no drift

- **WHEN** all API-returned feature IDs are in either `documented` or `skip`
- **THEN** `run_agent` output is `false`
- **THEN** no PR is opened

#### Scenario: Script detects documented features removed from the API

- **WHEN** a feature ID in the `documented` array is absent from the `GET /api/features` response
- **THEN** the script includes it in the report as `removed_features`
- **THEN** `run_agent` output is `true`

---

### Requirement: gh-aw workflow opens a self-healing PR on drift

A gh-aw workflow SHALL exist at `.github/workflows/security-role-docs-drift.md` that:

- Triggers on: `workflow_dispatch`, weekly schedule, and `push` to `main` with paths matching `generated/kbapi/**`
- Uses the shared live-stack setup (same pattern as workflows that require a running Elastic stack)
- Runs the pre-activation Go script and gates the agent step on `run_agent == 'true'`
- When the agent runs, it SHALL open a pull request (not an issue) that:
  - Updates `scripts/security-role-docs/kibana-features.json` (adds new features to `documented` or `skip` as appropriate, removes features no longer returned by the API)
  - Updates the feature privilege table in `templates/guides/security-roles.md.tmpl`
  - Regenerates `docs/guides/security-roles.md` via `make docs-generate`

The workflow SHALL NOT open issues. It SHALL use `safe-outputs: create-pr` (or equivalent gh-aw PR output).

#### Scenario: Workflow opens a PR when drift detected

- **WHEN** the pre-activation script outputs `run_agent=true`
- **THEN** the agent step runs
- **THEN** a pull request is opened updating `kibana-features.json` and the guide template

#### Scenario: Workflow is silent when no drift

- **WHEN** the pre-activation script outputs `run_agent=false`
- **THEN** the agent step does not run
- **THEN** no PR or issue is created

#### Scenario: Workflow triggers on kbapi update

- **WHEN** a commit is pushed to `main` that changes files under `generated/kbapi/**`
- **THEN** the workflow is triggered

#### Scenario: Workflow triggers on weekly schedule

- **WHEN** the scheduled cron fires
- **THEN** the pre-activation step runs regardless of recent kbapi changes

---

### Requirement: PR content distinguishes unknown from removed features

When the agent opens a PR, it SHALL categorise drift clearly:

- **Unknown features** (in API response but not in JSON): added to the guide table with their privilege strings populated from the API response; added to `documented` in the JSON
- **Unknown features the team decides to skip**: the PR description SHALL instruct reviewers to move entries from `documented` to `skip` in `kibana-features.json` if the feature should not appear in the guide
- **Removed features** (in `documented` but absent from API response): removed from the guide table and moved from `documented` to a `removed` array (or deleted entirely) in the JSON

#### Scenario: PR adds new feature to the guide table

- **WHEN** the API returns a previously unknown feature `newFeature` with privileges `["all", "read"]`
- **THEN** the PR adds a row for `newFeature` to the guide's feature table
- **THEN** the PR adds `newFeature` to the `documented` array in `kibana-features.json`

#### Scenario: PR removes a feature no longer in the API

- **WHEN** a feature in `documented` is absent from the `GET /api/features` response
- **THEN** the PR removes that feature's row from the guide table
- **THEN** the PR removes the feature ID from `documented` in `kibana-features.json`
