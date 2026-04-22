## Why

`elasticstack_elasticsearch_watch` is still implemented with the Terraform Plugin SDK while most new Elasticsearch resources now use the Terraform Plugin Framework. Migrating it safely requires an explicit compatibility contract so the resource keeps the same schema and upgrade behavior while existing SDK-managed state continues to work.

## What Changes

- Reimplement the watch resource with the Terraform Plugin Framework while preserving the existing resource type, schema shape, defaults, import format, and composite `id`.
- Convert watch-specific Elasticsearch client helpers to return framework diagnostics and support the Plugin Framework resource implementation directly.
- Add upgrade coverage that creates a watch with the last SDK-backed provider release and verifies the new Plugin Framework resource can manage that state without recreation.
- Update watch requirements to describe framework-based JSON normalization and SDK-to-Framework state compatibility.

## Capabilities

### New Capabilities

### Modified Capabilities

- `elasticsearch-watch`: replace SDK-specific implementation requirements with Plugin Framework requirements while preserving external behavior and upgrade compatibility

## Impact

- `internal/clients/elasticsearch/watch.go`
- `internal/elasticsearch/watcher/`
- `provider/plugin_framework.go`
- `provider/provider.go`
- `openspec/specs/elasticsearch-watch/spec.md`
