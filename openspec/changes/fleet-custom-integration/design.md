## Context

The existing `elasticstack_fleet_integration` resource installs packages from the Elastic package registry using `POST /api/fleet/epm/packages/{name}/{version}`. The Fleet API also supports installing a custom package by uploading a binary zip/gzip archive to `POST /api/fleet/epm/packages` (no path params, binary body). This endpoint is already present in the generated kbapi client as `PostFleetEpmPackagesWithBodyWithResponse` but is not yet used by any Terraform resource.

The generated response type (`PostFleetEpmPackagesResponse`) has only `Body []byte` — no typed JSON200 field — so the response must be manually unmarshalled. The response body can identify the package name (and sometimes version), but post-upload verification still requires Fleet read APIs: `GetPackages` as the primary source and `GetPackage` as a secondary exact-version check when a concrete version is known.

## Goals / Non-Goals

**Goals:**
- New resource `elasticstack_fleet_custom_integration` that uploads a local zip/gzip package file to Fleet and manages its lifecycle
- File-content change detection: re-upload automatically when the file content changes
- Computed `package_name` and `package_version` attributes for downstream resource references
- Space-awareness and `skip_destroy` parity with `elasticstack_fleet_integration`

**Non-Goals:**
- Building or validating the zip format (that is the user's responsibility)
- Supporting directory input (only prebuilt zip/gzip files)
- Diff/patching of individual assets inside the package
- Upgrading via the registry path — this resource is upload-only

## Decisions

### 1. File-content change detection via computed `checksum`

**Decision**: Store a SHA256 hash of the uploaded file as a computed `checksum` attribute. A plan modifier on `package_path` reads the file at plan time, computes the hash, and if it differs from the state checksum, marks `checksum`, `package_name`, and `package_version` as Unknown to signal pending changes.

**Rationale**: This is the standard Terraform pattern (analogous to `aws_lambda_function.source_code_hash`). It detects content changes regardless of filename and avoids re-uploading when only the path is renamed. Computing the hash at plan time gives users a preview of changes before apply.

**Alternative considered**: Require users to supply the hash explicitly (like `source_code_hash` on AWS Lambda). Rejected — it adds friction and is error-prone. Computing automatically is strictly better here.

### 2. Version extraction and post-upload verification

**Decision**: After upload, extract the package identity from the response body when present, using the archive manifest only as a fallback when the response omits the name or version. Then call `GetPackages` and filter by name plus installed status to resolve the canonical installed version. If `GetPackages` does not return a matching installed entry but a concrete version is known, call `GetPackage` for that exact name/version pair as a secondary verification step. Persist state only after one of those Fleet read APIs confirms the package.

**Rationale**: `items[].version` in the upload response is marked optional in the API spec, so relying on response data alone is fragile. `GetPackages` is the best source for the installed version that should be tracked in state, while `GetPackage` is still useful as a narrower exact-version check when Fleet has enough identity information but the global packages view lags or omits the package.

**Alternative considered**: Trust the version from the upload response or `manifest.yml` whenever Fleet read APIs cannot confirm the package. Rejected — it weakens post-upload verification and can hide installation or visibility problems behind guessed metadata.

### 3. Update path handles package replacement without `RequiresReplace`

**Decision**: `package_name` is computed and does NOT use `RequiresReplace`. On update, the provider re-uploads the package when file content or upload query parameters change. If the resulting package name or version differs from state, the update handler uninstalls the old package after the new one has been uploaded, then waits until the replacement package is readable as installed before completing.

**Rationale**: Making `package_name` ForceNew would mean any file content change would force a destroy+create cycle, which could cascade destructively to integration policies that reference the package. Handling replacement inside Update is safer, while the post-cleanup wait reduces transient refresh drift after old-version removal.

**Trade-off**: If the old package uninstall fails, the new package may still be uploaded, leaving the system in a partially updated state. This is mitigated by checking errors at each step and reporting diagnostics without further state mutation.

### 4. New resource package, minimal fleet.go additions

**Decision**: New Go package `internal/fleet/customintegration/`. Add a single `UploadPackage` wrapper to `internal/clients/fleet/fleet.go`. No other shared code changes.

**Rationale**: Follows the established per-resource package pattern. Keeps the fleet.go wrapper thin — business logic stays in the resource CRUD files.

### 5. Content-type determined from file extension

**Decision**: Determine `Content-Type` from the file extension: `.zip` → `application/zip`, `.gz` / `.tar.gz` → `application/gzip`. Fall back to `application/zip` for unknown extensions.

**Rationale**: The Fleet API accepts both. Extension-based detection is simple and sufficient; users building packages with `elastic-package` always produce standard zip files.

### 6. Upload response parsing and manifest fallback

**Decision**: `UploadPackage` treats the upload response as the first source of package identity and uses the archive manifest only as a fallback when the response omits the package name or version. The upload operation then verifies the result by querying the packages list API, and if that does not yield a match but a concrete version is known, it checks the package info API for that exact name/version pair. The provider returns an error if neither verification path confirms the package.

**Rationale**: On supported Kibana versions (8.2+), the provider should only persist state after Fleet exposes the uploaded package through a verifiable Fleet read API. The packages list is the primary source, but the exact package info API is still useful as a narrower secondary check when the upload response or manifest already identified a concrete version. Falling back to guessed identity/version data without either read API confirming the package weakens verification and can mask installation or visibility problems.

**Zip manifest parsing**: The name and version are parsed from `manifest.yml` inside the archive as a fallback when the upload response body does not contain one or both fields.

**Response body parsing across versions**: The upload response field for name and version changed across Kibana versions:
- `_meta.name` / `_meta.version` — Kibana 8.8+
- `items[0].name` / `items[0].version` — Kibana 8.0–8.7

The implementation tries both paths in order. If neither yields a package name, it falls back to parsing the archive manifest directly. If the manifest parse also fails, an error is returned. If a name was obtained from the response but no version was found, only the version is filled from the manifest before post-upload verification via the packages list and, when possible, the exact package info API.

### 7. Minimum Kibana version: 8.2.0

**Decision**: Create, Read, and Update gate on Kibana >= 8.2.0 using `EnforceMinVersion` and return a clear error diagnostic if the connected version is older.

**Rationale**: On Kibana < 8.2, `GET /api/fleet/epm/packages/{name}/{version}` returns HTTP 400 (7.17.x) or HTTP 404 (8.0.x–8.1.x) for custom-uploaded packages regardless of installation status, making drift detection impossible. A resource that silently loses drift detection provides a broken user experience. An explicit version gate with a clear error message is better than silent misbehaviour.

**Consequence**: Users on Kibana < 8.2 cannot use this resource. Kibana < 8.2 is EOL, so this is an acceptable constraint.

## Risks / Trade-offs

- **Upload is not idempotent for different versions**: Uploading a package with the same name but a different version installs both versions. The resource tracks only one version; the previous version must be explicitly uninstalled during update. The update path handles this.
- **GetPackages may be slow for large registries**: The post-upload GetPackages call lists all packages. This is a single API call and is consistent with existing patterns, so acceptable.
- **Plan modifier reads the file on every plan**: For large package files this adds latency at plan time. Users who want to avoid this can pin the file to a path where content changes are intentional.
- **Package replacement can temporarily lag Fleet reads**: After upload and old-package cleanup, Fleet may expose the new package through different read APIs at slightly different times. The implementation mitigates this by verifying uploads through `GetPackages`/`GetPackage`, using a read fallback to the packages list, and waiting for replacement packages to become readable before completing certain updates.
- **Package name changes mid-lifecycle**: If the embedded package name changes between versions, the update path uninstalls the old name. If the old package is referenced by an integration policy, the uninstall will fail in Fleet. Users must remove policy references first — this mirrors the behavior of `skip_destroy = false` on the integration resource.
- **`space_id` acceptance test**: The `space_id` feature is implemented but is not covered by the basic acceptance tests because it requires a pre-existing Kibana space. The existing `elasticstack_kibana_space` resource can be used to create one in practice.

## Migration Plan

No migration needed. This is an additive new resource with no impact on existing resources or state.

## Open Questions

None — design is fully specified based on API exploration and established codebase patterns.
