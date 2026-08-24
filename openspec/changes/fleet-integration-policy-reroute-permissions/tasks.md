## 1. Version gate and capability bit

- [x] 1.1 Add `MinVersionAdditionalDatastreamsPermissions` (9.1.0) alongside `MinVersionPolicyIDs` and `MinVersionOutputID` in `internal/fleet/integration_policy/resource.go`; verify by `go build ./...`
- [x] 1.2 Add `SupportsAdditionalDatastreamsPermissions` to `integrationPolicyFeatures` and resolve it via `client.EnforceMinVersion` in `resolveIntegrationPolicyFeatures` (`capabilities.go`); verify existing package unit tests still pass with `go test ./internal/fleet/integration_policy/...`
- [x] 1.3 Report an attribute-level version requirement from `integrationPolicyModel.GetVersionRequirements` when `additional_datastreams_permissions` is known, following the `output_id` branch; verify with a unit test asserting the requirement path and 9.1.0 floor, and a second asserting no requirement is reported when the attribute is null

## 2. Schema and model

- [x] 2.1 Add `attrAdditionalDatastreamsPermissions = "additional_datastreams_permissions"` to `constants.go`; verify by `go build ./...`
- [x] 2.2 Add the attribute to `getSchemaV3` in `schema.go` as an `Optional` `ListAttribute` of `types.StringType` with a `listvalidator.SizeAtLeast(1)` validator and a description that names the Kibana UI label ("Add a reroute processor permission"), the 9.1.0 minimum, that Kibana validates entries against the space's allowed namespace prefixes, and that the attribute must be removed rather than set to `[]` to revoke permissions; verify with a schema unit test asserting the attribute type, optionality, and validator presence
- [x] 2.3 Add `AdditionalDatastreamsPermissions types.List` with the matching `tfsdk` tag to `integrationPolicyModel` in `models.go`; verify by `go build ./...` and confirm no schema/model mismatch panic via `go test ./internal/fleet/integration_policy/...`
- [x] 2.4 Confirm no state upgrader or schema version bump is needed by checking that `getSchemaV3` still reports `Version: 3` and that the existing v0/v1/v2 upgrade tests pass unchanged with `go test ./internal/fleet/integration_policy/... -run Upgrade`

## 3. Request body conversion

- [x] 3.1 In `toAPIModel` (`models.go`), set `mappedBody.AdditionalDatastreamsPermissions` from the attribute when it is known and non-null, and to an explicit empty slice when it is null and `feat.SupportsAdditionalDatastreamsPermissions` is true, leaving it nil otherwise — mirroring the existing `PolicyIds` branch; verify by `go build ./...`
- [x] 3.2 Add a unit test asserting the create body carries the configured values in order when the attribute is set; verify with `go test ./internal/fleet/integration_policy/... -run TestIntegrationPolicy`
- [x] 3.3 Add a unit test asserting the body carries an empty array when the attribute is null and the capability bit is true, and omits the field entirely when the capability bit is false

## 4. State population

- [x] 4.1 In `populateFromAPI` (`models.go`), set `AdditionalDatastreamsPermissions` from `data.AdditionalDatastreamsPermissions` when the API returns a non-empty list, and to `types.ListNull(types.StringType)` when the field is absent or empty; verify by `go build ./...`
- [x] 4.2 Add a unit test asserting a two-element API response populates state in API order, and a second asserting an absent or empty API list yields a null attribute

## 5. Acceptance coverage

- [x] 5.1 Add a `TestAccResourceIntegrationPolicyAdditionalDatastreamsPermissions` acceptance test with testdata configs for create (one permission), update (two permissions), and clear (attribute removed), skipping below Kibana 9.1.0 using the package's existing version-skip helper; verify by running the test against a 9.1.0+ stack with `make docker-testacc TESTARGS='-run ^TestAccResourceIntegrationPolicyAdditionalDatastreamsPermissions$'`
- [x] 5.2 Assert the API-side policy after each step via a Fleet GET check function (permissions present after create, both present after update, absent or empty after clear), so a silent server-side behaviour change fails the test rather than passing quietly
- [x] 5.3 Add an `ImportState` step to the acceptance test verifying an existing policy's permissions are populated into state on import

## 6. Documentation and validation

- [x] 6.1 Add `additional_datastreams_permissions` to the example in `examples/resources/elasticstack_fleet_integration_policy/` with a comment tying it to a reroute processor destination; verify the example parses with `make lint`
- [x] 6.2 Regenerate provider docs with `make docs-generate` and confirm `docs/resources/fleet_integration_policy.md` lists the new attribute with its description
- [x] 6.3 Run `openspec validate --changes` and `make check-openspec` and confirm both pass
- [x] 6.4 Run `make lint` and `make test` and confirm both pass with no new findings
