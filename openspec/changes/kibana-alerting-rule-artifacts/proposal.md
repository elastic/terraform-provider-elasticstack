## Why

Practitioners cannot attach **artifacts** (dashboard links and investigation guides) to Kibana alerting rules via Terraform today ([terraform-provider-elasticstack#1408](https://github.com/elastic/terraform-provider-elasticstack/issues/1408)). The Kibana alerting rule API supports an `artifacts` object on create and update requests from **Kibana 8.19.0** / **9.1.0** onward; public GET/find responses include `artifacts` only on stacks with the [kibana#247279](https://github.com/elastic/kibana/pull/247279) fix (9.5.0+ and planned backports). The provider's generated `kbapi` models the field; the Terraform resource does not yet surface or map it.

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

| Operation | Minimum Kibana version | Notes |
|-----------|------------------------|-------|
| Create/update with `artifacts` | **8.19.0** (8.x) or **9.1.0** (9.x) | Dashboards: [kibana#216292](https://github.com/elastic/kibana/pull/216292) / [kibana#218920](https://github.com/elastic/kibana/pull/218920). Investigation guide: [kibana#216377](https://github.com/elastic/kibana/pull/216377) / [kibana#219943](https://github.com/elastic/kibana/pull/219943). **9.0.x is unsupported.** |
| Read round-trip via public GET/find | **9.5.0** (fix); backports TBD | Public GET/find omitted `artifacts` until [kibana#247279](https://github.com/elastic/kibana/pull/247279) ([kibana#242792](https://github.com/elastic/kibana/issues/242792)). Write still works on 8.19+; provider preserves state when GET omits the field. |

The provider SHALL enforce the write minimums at create/update (REQ-053). Resource documentation SHALL mention the GET limitation on older stacks.

### Acceptance tests

- Tests that **set** `artifacts` MUST be skipped when the stack is strictly below **8.19.0** (8.x) or strictly below **9.1.0** (9.x).
- Tests that assert **read round-trip** of `artifacts` from the API (dashboard IDs or inline `content` after refresh) SHOULD be skipped unless the stack is **9.5.0** or newer, unless CI targets a stack known to include the GET fix on an earlier minor.
- Separate test steps (or tests) SHOULD cover: dashboards only, investigation_guide with `content` only, investigation_guide with `content_path` + checksum drift detection, and clearing artifacts.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Add optional `artifacts` block (dashboards list + investigation guide), including validation rules (mutually exclusive `content`/`content_path`), version gates (REQ-053), create/update/read mapping with GET-limitation handling (REQ-054), checksum drift-detection for file-based content, and acceptance-test expectations (REQ-045–REQ-055).

## Impact

- **Specs**: Delta under `openspec/changes/kibana-alerting-rule-artifacts/specs/kibana-alerting-rule/spec.md` until merged into canonical spec.
- **Implementation** (future): `internal/kibana/alertingrule` (schema, model, plan modifier), `internal/models` (add `Artifacts` to `AlertingRule`), `internal/clients/kibanaoapi` (request/response mapping), docs/descriptions, acceptance tests.
