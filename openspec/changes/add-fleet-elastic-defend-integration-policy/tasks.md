# Tasks: Add Fleet Elastic Defend Integration Policy

## 1. Define the new capability

- [x] 1.1 Add the `fleet-elastic-defend-integration-policy` delta spec describing the dedicated Elastic Defend resource, its focused schema, and its Fleet package-policy lifecycle.
- [x] 1.2 Confirm the delta keeps `fleet-integration-policy` unchanged as the generic simplified package-policy capability.

## 2. Implement the resource shell

- [x] 2.1 Add `internal/fleet/elastic_defend_integration_policy` with resource metadata, schema, configure, and import support.
- [x] 2.2 Register `elasticstack_fleet_elastic_defend_integration_policy` in the provider and wire generated documentation for the new resource.
- [x] 2.3 Add `space_ids` handling consistent with the existing Fleet integration policy resource so read, update, and delete operate in the correct Kibana space.

## 3. Extend shared Fleet package policy client support

- [x] 3.1 Update `generated/kbapi` and `generated/kbapi/transform_schema.go` so Fleet package policies support both mapped and typed input encodings.
- [x] 3.2 Ensure the shared package policy client also preserves typed input `type`, typed input `config`, and the top-level package policy `version` required for Defend updates.
- [x] 3.3 Extend `internal/clients/fleet` package policy helpers so mapped and typed workflows can choose the correct Fleet query-format behavior.
- [x] 3.4 Add or update shared Fleet package-policy helpers, including secret handling as needed, so both encodings are supported correctly where they are actually used.
- [x] 3.5 Keep `internal/fleet/integration_policy` on the mapped-input path only, matching its existing schema and behavior.

## 4. Implement Defend-specific request and response mapping

- [x] 4.1 Implement create using the documented Defend bootstrap flow with the bootstrap `_config` preset path, then finalize creation with the typed policy payload using the persisted `integration_config`, `artifact_manifest`, and top-level `version`.
- [x] 4.2 Implement read and import mapping from the typed Defend response into Terraform state, preserving only modeled schema fields and validating that the imported/read package name is `endpoint`.
- [x] 4.3 Implement update and delete using the Fleet package-policy APIs while preserving opaque server-managed Defend payloads required for updates.

## 5. Preserve internal server-managed data safely

- [x] 5.1 Store the latest server-managed Defend payloads required for update, such as `artifact_manifest` and the package policy `version`, outside the public schema.
- [x] 5.2 Ensure read, import, and update refresh that internal data from the API response so subsequent updates remain valid.

## 6. Verify behavior

- [ ] 6.1 Add unit coverage for shared `kbapi` mapped-versus-typed package policy handling, including typed `config`, input `type`, and request/response `version`.
- [ ] 6.2 Add unit coverage for Fleet helper query-format selection so the mapped and typed paths cannot accidentally converge.
- [ ] 6.3 Add regression coverage proving `elasticstack_fleet_integration_policy` remains mapped-only after the shared client changes.
- [ ] 6.4 Add unit coverage for typed Defend request construction, typed response parsing, package-name validation on import/read, and preservation of opaque server-managed Defend data.
- [ ] 6.5 Add acceptance coverage for create, update, import, refresh after out-of-band delete, and delete of `elasticstack_fleet_elastic_defend_integration_policy`.
- [ ] 6.6 Run the relevant OpenSpec validation and targeted provider test commands for the new resource and the shared Fleet package-policy client changes.
