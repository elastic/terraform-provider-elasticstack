## Why

Users need to manage custom Fleet integration packages (built with `elastic-package`) via Terraform, but the provider only supports installing packages from the Elastic package registry by name+version. There is no resource for uploading a locally-built package zip to Fleet, forcing users to use raw API calls or manual Kibana UI steps.

## What Changes

- **New resource** `elasticstack_fleet_custom_integration`: accepts a path to a local `.zip` or `.tar.gz` custom integration package, uploads it to Fleet via the EPM upload API, and manages its lifecycle (re-upload on content change, uninstall on destroy).
- The resource exposes computed `package_name` and `package_version` attributes populated from the upload response so downstream resources (e.g. `elasticstack_fleet_integration_policy`) can reference them without hard-coding values.
- File-content-based change detection: the resource tracks a SHA256 checksum of the uploaded file and re-uploads automatically when the content changes.

## Capabilities

### New Capabilities

- `fleet-custom-integration`: Terraform resource for uploading and managing a locally-built Fleet custom integration package via the EPM binary upload API (`POST /api/fleet/epm/packages`). Covers create (upload), read (verify installed), update (re-upload on content change), and delete (uninstall).

### Modified Capabilities

## Impact

- New package: `internal/fleet/customintegration/`
- New fleet client wrapper: `UploadPackage` added to `internal/clients/fleet/fleet.go`
- Provider registration: `provider/plugin_framework.go` (new import + resource entry)
- No changes to existing resource behavior; the shared Fleet client is extended with a new `UploadPackage` method
- Depends on generated kbapi function `PostFleetEpmPackagesWithBodyWithResponse` (already present in `generated/kbapi/kibana.gen.go`)
