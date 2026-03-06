---
name: schema-coverage
description: Analyzes a Terraform resource schema and compares it to attributes used in the acceptance test suite (configs + assertions). Produces a prioritized report of missing and poor coverage (set-only assertions, single-value coverage, missing unset/empty cases, missing update coverage). Use when the user asks about schema coverage, test coverage gaps, or improving Terraform acceptance tests for a resource.
---

# Schema Coverage (Terraform resource vs acceptance tests)

## Goal

Given a Terraform resource, compare its **schema attributes/blocks** to what the acceptance tests **configure** and **assert**, then highlight opportunities to improve coverage.

Report findings in this order:
1. **Attributes with no coverage**
2. **Attributes with poor coverage**

## Inputs (infer if not provided)

- Resource name (e.g. `elasticstack_elasticsearch_security_api_key`)
- Or the schema file / acceptance test file path
- Or the Go package directory containing the resource and `*_acc_test.go`

If the user has an acceptance test open, infer the resource under test from:
- `resource.Test(...)` names like `resourceName := "elasticstack_..."`
- `resource.ParallelTest(...)` steps referencing a single resource
- config builder function names like `testAcc...Resource...`

## Workflow

### 1) Locate and parse the schema

Find the resource schema definition and capture **all schema keys**:
- Top-level attributes in `Schema: map[string]*schema.Schema{ ... }`
- Nested block schemas in:
  - `Elem: &schema.Resource{ Schema: ... }`
  - lists/sets/maps of resources and objects
  - nested blocks referenced via helpers

For each attribute/block, record:
- **Path**: terraform attribute path (top-level `foo`, nested `block.0.bar`, etc.)
- **Schema metadata**: `Required`/`Optional`/`Computed`, `ForceNew`, type (`TypeString`, `TypeList`, etc.), `MaxItems/MinItems`

### 2) Locate acceptance tests and collect attribute usage

Scan the acceptance tests for two independent signals:

- **Configured attributes**: attributes explicitly set in HCL test configs (including nested blocks).
- **Asserted attributes**: attributes referenced in checks, including:
  - value assertions (e.g. `TestCheckResourceAttr`)
  - set-only assertions (e.g. `TestCheckResourceAttrSet`)
  - absence assertions (e.g. `TestCheckNoResourceAttr`)
  - collection assertions (e.g. type-set element checks, list length checks)

Also detect **update coverage** by identifying multiple `resource.TestStep{ Config: ... }` steps for the same resource and whether attribute values change between steps.

### 3) Build a coverage matrix

For each schema attribute path, compute:
- **Configured?** (ever set in any test config)
- **Asserted?** (ever referenced by any check)
- **Assertion quality**:
  - **value-specific**: asserts the exact value (preferred)
  - **set-only**: only checks “is set” (weaker)
  - **absence**: checks unset / removed (good for optional fields)
- **Value diversity**: count distinct values across steps/configs (e.g. `name="a"` vs `name="b"`)
- **Optional-unset coverage**: optional field has a test step where it is deliberately omitted, plus an assertion that it is absent (or defaults appropriately)
- **Empty-collection coverage**: collection has a step where it is empty (or omitted if optional), plus assertions validating the empty state
- **Update coverage**: attribute value changes across steps, and a post-update assertion verifies the new value

### 4) Produce the report (strict ordering)

Use the template below.

## Report template

```markdown
## Schema coverage report: <resource>

### Scope
- **Schema**: <file(s) or function(s)>
- **Acceptance tests**: <test file(s)>

### 1) Attributes with no coverage
These schema attributes/blocks are not referenced in acceptance tests (neither configured nor asserted):
- `<attr_path>`: <Required/Optional/Computed>, <type>. **Gap**: not configured, not asserted.

### 2) Attributes with poor coverage
These attributes appear in tests but the coverage is weak:
- `<attr_path>`: <schema flags/type>
  - **Observed**: <how it’s currently used (configured/asserted), example values>
  - **Gaps**:
    - <one or more of: set-only assertion, single value only, no unset coverage, no empty collection coverage, no update coverage>
  - **Suggested improvements**:
    - <concrete test step or assertion to add>

### Suggested next steps (smallest diffs first)
1. Add value-specific assertions for set-only checks
2. Add an “unset optional” step + `TestCheckNoResourceAttr`
3. Add an “empty collection” step + collection assertions
4. Add an “update” step changing the attribute + post-update assertions
```

## Rules of thumb (for prioritization)

- Prefer **value-specific assertions** over “is set”.
- For **Optional** attributes, include at least one step where the attribute is **omitted**, and assert absence/default behavior.
- For **collections** (list/set/map), add a case for **empty** (or omitted if optional), and assert the expected empty state.
- For **Update** behavior, ensure there is at least one multi-step test where the attribute changes and the test asserts the new state.
- For **Computed-only** attributes, it’s acceptable to use set-only assertions when exact values are not deterministic, but prefer deterministic assertions when possible.

## Notes / limitations

- Some schemas are built via helpers; follow helper references until all attribute keys are accounted for.
- Attribute paths in checks often use `block.0.attr` indexing; normalize consistently when matching to schema blocks.
- Don’t mark an attribute as “covered” solely because it appears in raw HCL—prefer it being **asserted**. Treat “configured but never asserted” as poor coverage, not good coverage.
