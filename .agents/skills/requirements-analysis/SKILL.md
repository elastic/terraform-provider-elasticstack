---
name: requirements-analysis
description: Analyzes a Terraform entity requirements document for internal consistency, implementation compliance, and test opportunities. Outputs consistency findings, a requirement-by-requirement implementation check, and suggested unit/acceptance tests. Use when reviewing requirements docs, verifying implementation against requirements, or identifying test gaps.
---

# Requirements Document Analysis

Analyze a requirements document (from `dev-docs/requirements/`) and produce three outputs:

1. **Internal consistency** — whether requirements contradict each other or the schema.
2. **Implementation compliance** — whether the implementation meets each requirement.
3. **Test opportunities** — concrete unit or acceptance tests that would verify requirements programmatically.

## Input

- **Requirements document**: User specifies the path (e.g. `dev-docs/requirements/elasticsearch/security/role.md`) or the entity name/implementation path. Resolve to the single `.md` file under `dev-docs/requirements/`.
- **Implementation**: From the doc’s “Resource implementation” or “Data source implementation” line (e.g. `internal/elasticsearch/security/role`). Use that package for compliance and test analysis.

## Workflow

### 1. Parse the requirements document

- Read the doc and extract: **title/type name**, **implementation path**, **Schema** (HCL block: attributes/blocks, required/optional/computed, notes).
- List every requirement with **id** (REQ-NNN), **category**, and **text** (the “shall” statement). Normalize references (e.g. “id format”, “cluster_uuid/name”) for consistency checks.

### 2. Internal consistency

Apply the checks in [reference.md](reference.md) (Consistency checks):

- **Identity/Import**: Same id format in Identity and Import requirements; no conflicting formats.
- **Schema vs requirements**: Required/optional/computed in schema matches what requirements imply (e.g. “when X is configured” implies X is optional or optional+computed).
- **Lifecycle**: No requirement that X triggers replace and another that X is updated in place.
- **Compatibility**: Version numbers and feature names aligned across requirements; no conflicting minimum versions for the same feature.
- **State/Plan**: No requirement that the resource “preserve null” and another that it “store empty list” for the same field.
- **StateUpgrade**: Referenced prior schema version exists; upgrade steps are ordered and non-conflicting.

Output: **Consistent** or list **Inconsistencies** with requirement ids and short explanation.

### 3. Implementation compliance

- Resolve implementation package from the doc (e.g. `internal/elasticsearch/security/role`). Locate `resource.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `models.go`, state upgrade code, and any validators/plan modifiers.
- For **each requirement**, determine where the behavior would be implemented (see [reference.md](reference.md) “Requirement → implementation mapping”). Search or read that code and verify the behavior.
- Classify each requirement: **Met** (code clearly implements it), **Not met** (code contradicts or omits it), **Unclear** (cannot determine from code or tests).
- Output: Table or list: REQ-ID, Category, Status (Met / Not met / Unclear), Evidence (file/function or “not found”).

### 4. Test opportunities

- Locate **unit tests**: `*_test.go` in the implementation package (e.g. `resource_test.go`, `state_upgrade_test.go`) and **acceptance tests**: `acc_test.go` (or `*_acc_test.go`) for the resource.
- For each requirement, decide whether it is **verifiable by unit test** (e.g. state upgrade logic, id parsing, validation) or **verifiable by acceptance test** (e.g. create/read/update/delete, import, error diagnostics). See [reference.md](reference.md) “Test opportunity patterns”.
- Identify requirements that are **not** covered by existing tests (or only weakly covered). Suggest **concrete test cases**: unit test name + scenario, or acceptance test step (config + checks) that would assert the requirement.
- Output: List of **Suggested tests**, each with: requirement id(s), type (unit / acceptance), description, and how it would verify the requirement.

### 5. Report

Produce a single report with three sections:

1. **Internal consistency**: Result + any inconsistencies.
2. **Implementation compliance**: Summary (e.g. X/Y met, Z unclear) + per-requirement status table.
3. **Test opportunities**: List of suggested tests with requirement ids and verification approach.

## Output format

```markdown
# Requirements analysis: <entity name>

**Document**: `dev-docs/reqs/.../...md`  
**Implementation**: `internal/...`

## 1. Internal consistency

- **Result**: Consistent | Inconsistent
- **Inconsistencies** (if any): [REQ-xxx] vs [REQ-yyy]: ...

## 2. Implementation compliance

| REQ-ID | Category   | Status   | Evidence |
|--------|------------|----------|----------|
| REQ-001| API        | Met      | create.go, update.go call PutRole |
...

**Summary**: X met, Y not met, Z unclear.

## 3. Test opportunities

| REQ-ID(s) | Type       | Suggested test | Verifies |
|-----------|------------|----------------|----------|
| REQ-008   | Acceptance | Import with invalid id; expect error diagnostic | Import validation |
...
```

## Reference

- Requirement categories and implementation mapping: [reference.md](reference.md)
- Existing entity code-path checklist (for locating implementation): `.cursor/skills/existing-entity-requirements/reference.md`
- Schema/acceptance test coverage (for test patterns): `.cursor/skills/schema-coverage/` if analyzing attribute-level coverage alongside requirements.
