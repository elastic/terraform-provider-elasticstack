## Why

`elasticstack_fleet_integration_policy` consistently fails with
`"inconsistent values for sensitive attribute"` on `.inputs` after apply when
the managed package is an **input-type** package (e.g. `gcp_pubsub`,
`aws_s3`).

The provider's semantic-equality logic reconciles the Terraform plan (only
what the user set) against what Kibana returns after apply (which also carries
the package's variable defaults). It does this by extracting variable defaults
from `PackageInfo` and folding them into both sides of the comparison.

That extraction, located in
`internal/fleet/integration_policy/models_defaults.go`, was written only for
**integration-type** packages. Integration-type packages declare nested
`policyTemplates[].inputs[].vars` in the Fleet package manifest. **Input-type**
packages have a different manifest shape: each policy template declares a single
`input` field (not an `inputs` array) and a flat `vars` list at the template
level, and Kibana materialises those as an implicit stream on the resulting
policy.

Because the extractor did not recognise the input-type shape, the variable
defaults for input-type packages were effectively empty. Any default-valued
variable the user omitted (e.g. `data_stream.dataset`, `subscription_type`,
`credential_json`) was present in the post-apply Kibana response but absent from
the plan. Semantic equality therefore reported a difference, the planned value
was not substituted back into state, and Terraform Core raised the inconsistency
error.

This explains why the bug was invisible until recently: every existing
acceptance test exercises integration-type packages (`tcp`, `kafka`, `sql`,
`azure_metrics`, `aws_logs`, `gcp_vertexai`). Input-type packages were simply
not covered.

## What Changes

Teach the defaults extractor about the input-type package shape so that for
those packages it produces the same stream-level defaults that the
integration-type code path already produces for nested inputs.

No user-facing schema changes are required. The fix is purely in how the
provider reads the package manifest.

### Changes

- **`internal/fleet/integration_policy/models_defaults.go`**: Extend
  `apiPolicyTemplate` with `Input string` and `Vars apiVars` fields (the
  input-type fields). Update `apiPolicyTemplates.defaults()` to handle both the
  existing integration-type shape (`inputs` array) and the input-type shape
  (single `input` + top-level `vars`).

- **`internal/fleet/integration_policy/testdata/integration_gcp_pubsub.json`**:
  Add a representative `PackageInfo` fixture for the `gcp_pubsub` input-type
  package.

- **`internal/fleet/integration_policy/models_defaults_test.go`**: Add unit test
  `TestPackageInfoToDefaults_GCPPubSub` that exercises the new input-type
  extraction path with the fixture above, mirroring the existing
  `TestPackageInfoToDefaults_Kafka` test.

- **`internal/fleet/integration_policy/acc_test.go`** (and associated
  `testdata/`): Add acceptance test `TestAccResourceIntegrationPolicyGCPPubSub`
  that deploys a `gcp_pubsub` integration policy with a subset of variables,
  applies it, and asserts that Terraform produces no inconsistency error and that
  state round-trips cleanly.

## Capabilities

### Modified Capabilities

- `fleet-integration-policy`: Extend defaults extraction to handle input-type
  packages, fixing post-apply inconsistency errors and adding acceptance-test
  coverage for input-type packages (REQ-NEW-INPUT-TYPE).

## Impact

- **Specs**: Delta under
  `openspec/changes/fleet-integration-policy-input-type-defaults/specs/fleet-integration-policy/spec.md`
  until merged into canonical spec.
- **Implementation**: `internal/fleet/integration_policy/models_defaults.go`,
  new test fixture, new unit test, new acceptance test.
- **No provider schema changes.**
