## ADDED Requirements

### Requirement: Provider-wide documentation for `internal/utils/writeonlyhash` (REQ-DOC-001)

The provider repo SHALL contain a dedicated documentation file at `dev-docs/high-level/writeonly-secret-hashing.md` that explains how to use the shared `internal/utils/writeonlyhash` helper to detect drift on write-only secret attributes. The document SHALL be the canonical reference for any provider resource that exposes write-only secret material.

The document SHALL cover:

1. **Why hash-in-private-state over `_wo_version` companions.** Brief comparison: companion-attribute pattern requires user discipline (bumping a version) and is invisible to silent in-config edits; hash-in-private-state catches edits automatically.
2. **Threat model.** State files can leak; a fast hash (SHA-256) enables offline brute-force of low-entropy secrets. bcrypt with a per-resource-type salt is the conservative choice.
3. **`ModifyPlan` contract.** When to compute, compare, and emit warning diagnostics; when NOT to (Read must not touch private state).
4. **Private-state key convention.** `secret_hash:<attribute_path>` where `<attribute_path>` matches the Terraform attribute path (e.g. `aws.external_id`, `configuration_values["password"].secret_value`). Map elements use the bracketed key form in their path.
5. **Post-import behaviour.** First refresh after `terraform import` produces no drift; first subsequent apply baselines the hash. Matches `random_password.bcrypt_hash`.
6. **Worked example.** A complete adoption walkthrough for a resource with one write-only attribute, showing: helper construction, `ModifyPlan` integration, post-Create/Update hash write, post-Delete cleanup.
7. **Anti-patterns.** Logging the value, including the value in diagnostic messages, using the helper from Read, sharing salts across resource types.

The document SHALL link to the helper package Godoc.

#### Scenario: Document exists at the canonical path

- **WHEN** a contributor looks for write-only-secret guidance
- **THEN** `dev-docs/high-level/writeonly-secret-hashing.md` SHALL exist
- **AND** SHALL contain all sections listed above

#### Scenario: Coding standards links to the document

- **WHEN** a contributor reads `dev-docs/high-level/coding-standards.md`
- **THEN** the document SHALL include a reference linking to `writeonly-secret-hashing.md` under a "Write-only secret attributes" sub-heading

#### Scenario: Worked example matches the helper API

- **GIVEN** the worked example in `writeonly-secret-hashing.md`
- **WHEN** a reviewer compares the example to the actual `internal/utils/writeonlyhash` package
- **THEN** every helper call in the example SHALL match the package's exported API

### Requirement: Documentation lands regardless of helper-implementation ownership (REQ-DOC-002)

The documentation deliverable SHALL be implementable independently of whether this change or the in-flight [fleet-cloud-connector change (PR #3415)](https://github.com/elastic/terraform-provider-elasticstack/pull/3415) ships the helper implementation.

Tasks for this change SHALL include the documentation deliverable. Tasks for the helper implementation SHALL be conditional on the helper not already being present in `internal/utils/writeonlyhash/` at implementation time.

#### Scenario: Docs deliverable is independent of merge order

- **WHEN** the `elasticsearch-content-connector` change is implemented
- **AND** the `writeonlyhash` helper is already present (built by an earlier change)
- **THEN** the documentation deliverable SHALL still be completed by this change

#### Scenario: Docs deliverable is independent when helper is built here

- **WHEN** the `elasticsearch-content-connector` change is implemented
- **AND** the `writeonlyhash` helper is not yet present
- **THEN** this change SHALL build the helper AND the documentation deliverable
- **AND** the documentation SHALL be authored to match the helper API the change ships
