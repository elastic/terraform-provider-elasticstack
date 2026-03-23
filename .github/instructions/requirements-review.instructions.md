---
applyTo:
  - "openspec/specs/**/*.md"
  - "openspec/changes/**/specs/**/*.md"
excludeAgent: "coding-agent"
---

# Requirements Document Review

When reviewing changes to OpenSpec requirement documents under `openspec/specs/` (or delta specs under `openspec/changes/**/specs/`), follow this process:

## 1. Read the requirements-verification skill

Read and apply the workflow in `.agents/skills/requirements-verification/SKILL.md`. That skill defines how to analyze requirements documents for internal consistency, implementation compliance, and test opportunities.

## 2. Review requirements and implementation for consistency

- **Internal consistency**: Check that requirements do not contradict each other or the schema. Use the consistency checks in `.agents/skills/requirements-verification/reference.md` (Identity/Import, Schema vs requirements, Lifecycle, Compatibility, State/Plan, StateUpgrade, API).
- **Implementation compliance**: Resolve the implementation path from the doc (e.g. "Resource implementation: `internal/elasticsearch/security/role`"). For each requirement, verify the implementation meets it. Classify as Met, Not met, or Unclear with evidence.
- **Test opportunities**: Identify requirements not covered by existing tests and suggest concrete unit or acceptance tests.

## 3. Ensure implementation meets documented requirements

Confirm that the Terraform resource or data source implementation (schema, create/read/update/delete, import, state handling) aligns with every documented requirement. Flag any gaps or contradictions in your review.

## 4. Report findings

Explicitly report the outcome of the requirements analysis in your review regardless of the outcome. If there are no reportable issues/improvements simply acknowledge the implementation meets all documented requirements.