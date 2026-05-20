## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run
      `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-integration-policy-input-type-defaults --type change`
      (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into
      `openspec/specs/fleet-integration-policy/spec.md` or **archive** the change
      per project workflow.

## 2. Implementation

- [ ] 2.1 Extend `apiPolicyTemplate` in
      `internal/fleet/integration_policy/models_defaults.go` with two new fields:
      `Input string \`json:"input"\`` and `Vars apiVars \`json:"vars"\``.

- [ ] 2.2 Update `apiPolicyTemplates.defaults()` to handle both shapes:
      - If `len(policyTemplate.Inputs) > 0` → existing integration-type path
        (unchanged).
      - If `policyTemplate.Input != ""` → input-type path: call
        `policyTemplate.Vars.defaults()` and key the result as
        `"{policyTemplate.Name}-{policyTemplate.Input}"`.

- [ ] 2.3 Add `internal/fleet/integration_policy/testdata/integration_gcp_pubsub.json`:
      a compact but valid `PackageInfo` JSON for `gcp_pubsub`. Must include:
      - At least one `policyTemplates` entry with `"input": "gcp-pubsub"` and a
        `vars` array.
      - At least one var with a non-null `default` (e.g. `subscription_type:
        "shared"`) and at least one without a default (e.g. `project_id`).
      - At least one `dataStreams` entry that references `"input": "gcp-pubsub"`,
        to exercise the downstream stream-defaults pairing.

- [ ] 2.4 Add `TestPackageInfoToDefaults_GCPPubSub` in
      `internal/fleet/integration_policy/models_defaults_test.go`:
      - Load `testdata/integration_gcp_pubsub.json` (embed it the same way
        `integration_kafka.json` is embedded).
      - Call `packageInfoToDefaults(pkg)` and assert:
        - The expected input ID (e.g. `"gcp_pubsub-gcp-pubsub"`) is present in
          the result map.
        - At least one defaulted var (e.g. `subscription_type`) appears in the
          extracted vars JSON.
        - At least one non-defaulted var (e.g. `project_id`) is absent from the
          extracted vars JSON.

## 3. Testing

- [ ] 3.1 Add `TestAccResourceIntegrationPolicyGCPPubSub` in
      `internal/fleet/integration_policy/acc_test.go`:
      - Version-gate with `minVersionIntegrationPolicy` (≥ 8.10.0), consistent
        with other acceptance tests in this package.
      - Provide a Terraform config fixture under
        `testdata/TestAccResourceIntegrationPolicyGCPPubSub/create/` that:
        - Creates an agent policy.
        - Installs `gcp_pubsub` (or uses a pre-installed version).
        - Creates an `elasticstack_fleet_integration_policy` targeting
          `gcp_pubsub` with only the required user-visible vars set
          (e.g. `project_id`, `subscription_name`, `topic`).
      - Apply the configuration and assert no inconsistency error (`no_plan_diff`
        / `PlanOnly: false`).
      - Assert the resource state contains the expected computed stream vars
        (including the defaulted `subscription_type`).
      - Run a plan after the initial apply and assert it produces no further
        changes (`resource.TestCheckNoResourceAttr` / `PlanOnly` with
        `ExpectNonEmptyPlan: false`).
