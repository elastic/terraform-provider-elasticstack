# `elasticstack_fleet_managed_integration` examples

Each `.tf` file in this directory is an **independent**, copy-paste-ready Terraform module. Files are planned in isolation by the provider’s example acceptance harness; **do not combine** multiple snippet files into one root module (duplicate `provider` blocks and resource addresses will conflict).

Set `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` before running Terraform against these configurations.

- `resource.tf` — full CSPM AWS create example.
- `package_version_update.tf` — same resource address (`cspm_aws`) after an in-place `package.version` bump (`3.4.0` → `3.5.0`), not a second integration.
- `import.sh` — import by managed integration ID (`policy_id` in state).
