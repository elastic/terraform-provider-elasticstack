## REMOVED Requirements

### Requirement: Workflow runs for Kibana spec-impact analysis
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Deterministic helper emits entity impact evidence
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Agent creates one issue per impacted entity from deterministic evidence
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Workflow suppresses duplicate impact issues
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Issue cap does not suppress never-filed entities
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Baseline advances after successful analysis
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Generated kbapi file lifecycle is tolerantly diffed
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: `memory-record-from-report` issuance contract
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Root internal/kibana entity scan robustness
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Deterministic report is handed to the agent via artifact
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.

### Requirement: Pre-activation repo-memory checkout path
**Reason**: The `kibana-spec-impact` workflow is replaced by the Kibana API gap scanner (`ci-kibana-api-gap-scanner`), which detects newly added Kibana API request fields as new gap candidates, covers fields that were never adopted, and applies LLM triage. Removed or retyped fields that the provider uses are caught by the Go build when the generated client is regenerated.

**Migration**: Newly added Kibana API fields surface through `ci-kibana-api-gap-scanner` issues labelled `kibana-api-gap`. The `kibana-spec-impact` workflow, its helper, and the `memory/kibana-spec-impact` branch are removed; existing open `[kibana-spec-impact]` issues are left as they are.
