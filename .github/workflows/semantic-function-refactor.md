---
imports: 
  - shared/setup-dev.md
  - shared/go-source-analysis.md
  - shared/dispatch-code-factory.md
on:
  schedule:
  - cron: daily
  steps:
  - name: Checkout repository
    uses: actions/checkout@v7.0.0
    with:
      fetch-depth: 1
      persist-credentials: false
  - env:
      ISSUE_SLOTS_CAP: "3"
      ISSUE_SLOTS_LABEL: semantic-refactor
    id: compute_issue_slots
    name: Compute issue slots
    uses: actions/github-script@v9.0.0
    with:
      github-token: ${{ secrets.GITHUB_TOKEN }}
      script: |
        const fn = require('${{ github.workspace }}/.github/scripts/workflows/issue-slots/compute.js');
        await fn({ github, context, core });
  workflow_dispatch: null
permissions:
  actions: read
  contents: read
  issues: read
  pull-requests: read
if: needs.pre_activation.outputs.issue_slots_available != '0'
network:
  allowed:
  - defaults
  - go
  - elastic.litellm-prod.ai
safe-outputs:
  create-issue:
    labels:
    - semantic-refactor
    - refactoring
    - code-quality
    - automated-analysis
    - triaged
    max: 3
    title-prefix: "[semantic-refactor] "
checkout:
  fetch-depth: 0
description: Analyzes Go source organization and identifies actionable semantic refactoring opportunities
engine:
  args:
  - --effort
  - high
  env:
    ANTHROPIC_API_KEY: ${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}
    ANTHROPIC_BASE_URL: https://elastic.litellm-prod.ai/
  id: claude
  model: llm-gateway/claude-sonnet-4-6
jobs:
  pre-activation:
    outputs:
      gate_reason: ${{ steps.compute_issue_slots.outputs.gate_reason }}
      issue_slots_available: ${{ steps.compute_issue_slots.outputs.issue_slots_available }}
      open_issues: ${{ steps.compute_issue_slots.outputs.open_issues }}
name: Semantic Function Refactor
timeout-minutes: 35
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [default, issues]
---
# Semantic Function Refactor

Analyze Go source organization to identify actionable semantic refactoring opportunities: misplaced functions, duplicate or near-duplicate functions, scattered helpers, and extraction opportunities.

Upstream baseline: `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md`

## Pre-activation context

A deterministic pre-activation step has already computed semantic-refactor issue capacity for this run. Do **not** query GitHub issue counts yourself; use only the values below.

- **Open semantic-refactor issues**: `${{ needs.pre_activation.outputs.open_issues }}`
- **Issue slots available**: `${{ needs.pre_activation.outputs.issue_slots_available }}`
- **Gate reason**: ${{ needs.pre_activation.outputs.gate_reason }}

The workflow reached this point only because `issue_slots_available` is non-zero. You may open up to `${{ needs.pre_activation.outputs.issue_slots_available }}` new issues in this run.

## Task

Detect and report semantic refactoring opportunities by:

1. **Activating Serena** for semantic code analysis
2. **Discovering Go Source Files** across the repository
3. **Collecting function inventories** using Serena's symbol overview
4. **Clustering functions semantically** by name and purpose
5. **Identifying outliers** (functions in wrong files) using naming and symbolic analysis
6. **Detecting duplicates** using Serena's semantic duplicate detection
7. **Reporting findings** as actionable issues, capped by available slots

## Important Constraints

1. **Only analyze `.go` files** — Ignore all other file types
2. **Skip test files** — Never analyze files ending in `_test.go` or in test directories
3. **Use Serena for semantic analysis** — Leverage the MCP server's capabilities for symbol navigation and duplicate detection
4. **One file per feature rule** — Files should be named after their primary purpose/feature
5. **Read-only analysis** — Do not modify any files during analysis

## Serena Configuration

The Serena MCP server is configured for this workspace:

- **Workspace**: ${{ github.workspace }}
- **Context**: codex
- **Language service**: Go (gopls)

## Analysis Workflow

### 1. Activate Serena Project

Activate the project in Serena to enable semantic analysis:

```
Tool: activate_project
Args: { "path": "${{ github.workspace }}" }
```

### 2. Discover Go Source Files

Find all non-test Go files in the repository:

```bash
find . -name "*.go" ! -name "*_test.go" -type f | sort
```

Group files by package/directory to understand the organization. Exclude:

- **Test files** (`*_test.go`, `test/`, `tests/`, `__tests__/`, `spec/` directories)
- **Generated files** and build artifacts
- **Workflow files** (`.github/workflows/*`)
- **Vendored dependencies** (`vendor/`, module cache)
- **Non-Go files**

### 3. Collect Function Inventory Per File

For each discovered Go file:

1. Use Serena's `get_symbols_overview` to get all symbols (functions, methods, types)
2. Use Serena's `read_file` to understand context for ambiguous symbols
3. Create a structured inventory of:
   - File path and package name
   - All function names with signatures
   - All method names with receiver types
   - Brief purpose inferred from naming and content

Example structure:

```
File: pkg/workflow/compiler.go
Package: workflow
Functions:
  - CompileWorkflow(path string) error
  - compileFile(data []byte) (*Workflow, error)
  - validateFrontmatter(fm map[string]interface{}) error
```

### 4. Semantic Clustering Analysis

Analyze the collected functions to identify patterns:

**Clustering by Naming Patterns:**

- Group functions with similar prefixes (`create*`, `parse*`, `validate*`)
- Group functions with similar suffixes (`*Helper`, `*Config`, `*Step`)
- Identify functions operating on the same data types
- Identify functions sharing common functionality

**File Organization Rules:**
According to Go best practices, files should be organized by feature:

- `compiler.go` — compilation-related functions
- `parser.go` — parsing-related functions
- `validator.go` — validation-related functions
- `create_*.go` — creation/construction functions for specific entities

**Identify Outliers:**
Look for functions that don't match their file's primary purpose:

- Validation functions in a compiler file
- Parser functions in a network file
- Helper functions scattered across multiple files
- Generic utility functions not in a dedicated utils file

### 5. Serena Semantic Duplicate Detection

For each cluster of similar functions, use Serena to detect duplicates:

1. Use `find_symbol` to locate functions with similar names across files
2. Use `search_for_pattern` to find similar code patterns
3. Use `find_referencing_symbols` to understand usage patterns
4. Compare function implementations to identify:
   - **Exact duplicates** — identical implementations
   - **Near duplicates** — similar logic with variations
   - **Functional duplicates** — different implementations, same purpose

Example Serena tool usage:

```
Tool: find_symbol
Args: { "symbol_name": "processData", "workspace": "${{ github.workspace }}" }
```

### 6. Deep Reasoning Analysis

Apply deep reasoning to identify refactoring opportunities:

**Duplicate Detection Criteria:**

- Functions with >80% code similarity
- Functions with identical logic but different variable names
- Functions performing the same operation on different types (candidates for generics)
- Helper functions repeated across multiple files

**Refactoring Patterns to Suggest:**

- **Extract Common Function** — When 2+ functions share significant code
- **Move to Appropriate File** — When a function is in the wrong file
- **Create Utility File** — When helper functions are scattered
- **Use Generics** — When similar functions differ only by type
- **Extract Interface** — When similar methods are defined on different types

### 7. Issue Reporting

Create separate issues for each distinct actionable refactoring opportunity, up to `${{ needs.pre_activation.outputs.issue_slots_available }}` opportunities this run.

**When to Create Issues:**

- Only create issues if significant refactoring opportunities are found
- **Create one issue per distinct opportunity** — do NOT bundle multiple patterns
- Limit to the top `${{ needs.pre_activation.outputs.issue_slots_available }}` most significant opportunities
- Use the `create_issue` tool from safe-outputs MCP **once for each opportunity**

**Issue Contents for Each Opportunity:**

- **Executive Summary**: Brief description of this specific opportunity
- **Concrete Evidence**: Specific file paths, function names, signatures, line numbers
- **Impact Analysis**: How this affects code organization, maintainability, duplication
- **Refactoring Recommendations**: Suggested approaches with estimated effort
- **Code Examples**: Concrete examples showing current state and proposed structure

## Detection Scope

### Report These Issues

- Functions clearly in the wrong file based on their semantic purpose
- Duplicate or near-duplicate implementations across files
- Scattered helper functions that should be centralized
- Repeated logic blocks that could be extracted to utilities
- Opportunities to use Go generics for type-specific duplicates
- Packages or files that have grown too large and could be split by feature

### Skip These Patterns

- Standard boilerplate code (imports, exports, package declarations)
- Go idioms intentionally repeated (constructors, standard patterns)
- **All test files** (`*_test.go`, `test/`, `tests/`, `__tests__/`, `spec/`)
- **All workflow files** (`.github/workflows/*`)
- **Generated code** and build artifacts
- **Vendored dependencies** (`vendor/`)
- Small code snippets (<5 lines) unless highly repetitive
- Configuration files with similar structure
- Single-occurrence patterns without clear organizational impact

### Analysis Depth

- **Primary Focus**: Go source files across the repository (excluding tests, generated, vendored, workflow files)
- **Secondary Analysis**: Cross-package relationship and helper reuse patterns
- **Cross-Reference**: Look for duplication across the entire repository
- **Historical Context**: Consider whether issues are new trends or long-standing technical debt

## Issue Template

For each distinct refactoring opportunity, create a separate issue using this structure:

````markdown
# 🔧 Semantic Refactor: [Opportunity Name]

*Analysis of repository: ${{ github.repository }}*

## Summary

[Brief overview of this specific refactoring opportunity]

## Concrete Evidence

### Opportunity: [Description]
- **Severity**: High/Medium/Low
- **Type**: [misplaced-function | duplicate-function | scattered-helper | extraction-opportunity | generics-candidate]
- **Locations**:
  - `path/to/file1.go` — function `FuncName(...)` (lines X-Y)
  - `path/to/file2.go` — function `OtherFunc(...)` (lines A-B)
- **Code Sample**:
  ````go
  [Example of the code pattern]
  ````

## Impact Analysis

- **Maintainability**: [How this affects code maintenance]
- **Organization**: [How file/package structure is impacted]
- **Duplication Risk**: [Potential for drift between copies]

## Refactoring Recommendations

1. **[Recommendation 1]**
   - Target: `suggested/path/file.go`
   - Action: [move | extract | consolidate | introduce-generic]
   - Estimated effort: [hours/complexity]
   - Benefits: [specific improvements]

2. **[Recommendation 2]**
   [... additional recommendations ...]

## Implementation Checklist

- [ ] Review refactoring findings
- [ ] Prioritize refactoring tasks
- [ ] Create refactoring plan
- [ ] Implement changes
- [ ] Update tests as needed
- [ ] Verify no functionality broken

## Analysis Metadata

- **Analyzed Files**: [count]
- **Total Functions Cataloged**: [count]
- **Function Clusters Identified**: [count]
- **Outliers Found**: [count]
- **Duplicates Detected**: [count]
- **Detection Method**: Serena semantic code analysis + naming pattern analysis
- **Analysis Date**: [timestamp]

````

## Serena Tool Usage Guide

### Project Activation
```
Tool: activate_project
Args: { "path": "${{ github.workspace }}" }
```

### Symbol Overview
```
Tool: get_symbols_overview
Args: { "file_path": "pkg/workflow/compiler.go" }
```

### Find Similar Symbols
```
Tool: find_symbol
Args: { "symbol_name": "parseConfig", "workspace": "${{ github.workspace }}" }
```

### Search for Patterns
```
Tool: search_for_pattern
Args: { "pattern": "func.*Config.*error", "workspace": "${{ github.workspace }}" }
```

### Find References
```
Tool: find_referencing_symbols
Args: { "symbol_name": "CompileWorkflow", "file_path": "pkg/workflow/compiler.go" }
```

### Read File Content
```
Tool: read_file
Args: { "file_path": "pkg/workflow/compiler.go" }
```

## Operational Guidelines

### Security
- Never execute untrusted code or commands
- Only use read-only analysis tools
- Do not modify files during analysis (read-only mode)

### Efficiency
- Use Serena's semantic analysis capabilities effectively
- Balance thoroughness with timeout constraints
- Focus on meaningful patterns, not trivial similarities

### Accuracy
- Verify findings before reporting
- Distinguish between acceptable duplication and problematic duplication
- Consider Go idioms and best practices
- Provide specific, actionable recommendations

### Issue Creation
- Create **one issue per distinct refactoring opportunity** — do NOT bundle multiple patterns in a single issue
- Never create more than `${{ needs.pre_activation.outputs.issue_slots_available }}` issues in this run
- Only create issues if actionable opportunities are found
- Include sufficient detail for coding agents to understand and act on findings
- Provide concrete examples with file paths, function names, and line numbers
- Suggest practical refactoring approaches
- Use descriptive titles that clearly identify the specific opportunity (e.g., "Semantic Refactor: Scattered validation helpers across provider package")

#### Issue title length guardrail

GitHub issue titles are limited to **256 characters total**, including the
`title-prefix` that `create-issue` prepends automatically.

- This workflow's prefix is `"[semantic-refactor] "` (20 characters),
  leaving **236** characters for the title you provide.
- Before calling `create-issue`, verify that
  `len("[semantic-refactor] ") + len(your title)` is **≤ 256**.
- Keep titles concise. Move full file paths, function signatures, attribute
  lists, failure excerpts, and detailed descriptions into the issue body.
- If the natural title would exceed the limit, shorten the variable portion.
  Prefer a short opportunity label over long descriptive phrases (e.g.,
  "scattered validation helpers in provider package").
- Do not include markdown heading markers (`#`), emoji, or the prefix label
  redundantly in the title. The title field is plain text.

### Dispatch
After creating all issues for this run (or if no issues were created), call the `dispatch_code_factory` safe output tool once to dispatch the `code-factory` workflow for each created issue.

**Objective**: Improve code organization and reduce duplication by identifying actionable semantic refactoring opportunities through Serena semantic function clustering and duplicate detection. Focus on high-impact, actionable findings that enable automated or manual refactoring.
