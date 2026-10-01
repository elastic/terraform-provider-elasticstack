## Why

Practitioners cannot attach **artifacts** (dashboard links and investigation guides) to Kibana alerting rules via Terraform today ([terraform-provider-elasticstack#1408](https://github.com/elastic/terraform-provider-elasticstack/issues/1408)). The Kibana alerting rule API already supports an `artifacts` object on create and update requests, and returns it on read (GET). The provider's generated `kbapi` models the `artifacts` field; the Terraform resource does not yet surface or map it.

Teams managing large alert rule inventories across environments (Kubernetes, AWS) want to link dashboards and attach investigation guides alongside rule definitions, so that runbooks and context are co-located with the alert configuration as code.

## What Changes

- Add OpenSpec requirements (delta) for `elasticstack_kibana_alerting_rule`: optional **`artifacts`** configuration block, including **`dashboards`** (list of dashboard IDs) and **`investigation_guide`** (inline content or file-path-based content with drift-detecting checksum).
- **Out of scope for this proposal artifact**: editing `openspec/specs/kibana-alerting-rule/spec.md` directly; that happens when the change is synced or archived.

### Schema sketch (to merge into canonical `## Schema` on sync)

Add an optional single nested block at rule level:

```hcl
  artifacts {
    dashboards {
      id = "<dashboard_id>"
    }
    investigation_guide {
      content      = "<inline_markdown>"  # mutually exclusive with content_path
      content_path = "/path/to/guide.md" # mutually exclusive with content
      checksum     = "<computed>"         # computed SHA-256 of file at content_path
    }
  }
```

- **`artifacts.dashboards`**: optional list block; each entry has a required `id` string (the Kibana dashboard saved-object id). Maps to `artifacts.dashboards[].id` in the API.
- **`artifacts.investigation_guide`**: optional single nested block; holds investigation guide text. Maps to `artifacts.investigation_guide.blob` in the API.
  - **`content`** (optional string): inline text/Markdown sent directly as `blob`.
  - **`content_path`** (optional string): path to a local file whose content is read and sent as `blob`. At plan time the provider computes a SHA-256 of the file and stores it as `checksum` to detect when the file changes.
  - **`checksum`** (computed string): SHA-256 hex digest of the file at `content_path`. Not user-settable; used for drift detection. Irrelevant when `content` is used.
  - Exactly one of `content` or `content_path` MUST be set when `investigation_guide` is present.

### Version rules

The minimum Kibana version that supports `artifacts` on alerting rules has not been independently confirmed from release notes. The kbapi generated client already models the field, and it is present in the current API. An exact minimum version MUST be determined during implementation; if no server-side version check can be confirmed, the provider SHOULD omit the version gate and rely on the Kibana API to reject the field on incompatible versions. This is an open question recorded in `design.md`.

### Acceptance tests

- Any acceptance test that sets `artifacts` MUST be skipped if the minimum Kibana version for this feature cannot be met by the test environment.
- Separate test steps (or tests) SHOULD cover: dashboards only, investigation_guide with `content` only, investigation_guide with `content_path` + checksum drift detection, and clearing artifacts.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Add optional `artifacts` block (dashboards list + investigation guide), including validation rules (mutually exclusive `content`/`content_path`), create/update/read mapping, checksum drift-detection for file-based content, and acceptance-test expectations (REQ-045–REQ-050).

## Impact

- **Specs**: Delta under `openspec/changes/kibana-alerting-rule-artifacts/specs/kibana-alerting-rule/spec.md` until merged into canonical spec.
- **Implementation** (future): `internal/kibana/alertingrule` (schema, model, plan modifier), `internal/models` (add `Artifacts` to `AlertingRule`), `internal/clients/kibanaoapi` (request/response mapping), docs/descriptions, acceptance tests.
