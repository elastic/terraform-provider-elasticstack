## 1. Watch Helper Migration

- [x] 1.1 Convert `internal/clients/elasticsearch/watch.go` to return Terraform Plugin Framework diagnostics
- [x] 1.2 Update watch callers so the migrated resource can use the helper layer directly without SDK diagnostic conversion glue

## 2. Plugin Framework Resource Implementation

- [x] 2.1 Replace the SDK watch resource implementation with a Plugin Framework package layout that preserves schema, defaults, IDs, import behavior, and connection overrides
- [x] 2.2 Register the Plugin Framework watch resource in `provider/plugin_framework.go` and remove the SDK registration from `provider/provider.go`
- [x] 2.3 Preserve semantically equivalent JSON handling for watch JSON string attributes in the new Framework schema and read/write paths

## 3. Upgrade And Acceptance Coverage

- [x] 3.1 Move or adapt the watch acceptance suite to the migrated resource package
- [x] 3.2 Add a `FromSDK` acceptance test pinned to provider version `0.14.3`
- [x] 3.3 Add a Framework state upgrader only if the SDK-to-Framework upgrade test proves one is required

## 4. Verification

- [x] 4.1 Validate the OpenSpec change artifacts for `migrate-elasticsearch-watch-to-plugin-framework`
- [x] 4.2 Run build plus focused watch tests, including the SDK-to-Framework upgrade path
