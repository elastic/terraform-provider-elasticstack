# Proposal: Fleet Agent Download Source — OpenSpec requirements

## Why

Functional requirements for `elasticstack_fleet_agent_download_source` were authored under `dev-docs/requirements/`. The project now keeps canonical behavior in OpenSpec (`openspec/specs/` and change deltas under `openspec/changes/`). This change captures those requirements as a proper OpenSpec proposal so reviewers and agents can trace schema and behavior against versioned specs.

## What Changes

- **Introduce** a delta spec for the `fleet-agent-download-source` capability covering schema, API usage, identity, spaces, import, errors, updates, and v1 scope (no `auth`/`secrets` attributes).
- **Add** `proposal.md`, `design.md`, and `tasks.md` for this capability so implementation and future sync/archive follow the standard workflow.
- **Replace** the standalone dev-docs requirements file with a pointer to this change (and to the canonical spec path after sync).

## Capabilities

### New Capabilities

- `fleet-agent-download-source`: Terraform resource `elasticstack_fleet_agent_download_source` — Agent Binary Download Sources via Kibana Fleet APIs, including space-aware operations and import.

### Modified Capabilities

- (none)

## Impact

- **Documentation**: `dev-docs/requirements/kibana/fleet_agent_download_source.md` becomes a short redirect; detailed requirements live in the change delta spec.
- **Implementation**: Existing code under `internal/fleet/agentdownloadsource` and `internal/clients/fleet` should align with the delta spec; gaps are tracked via tasks.
- **No breaking change** to Terraform schema or behavior by this documentation-only migration; the spec documents intended behavior.
