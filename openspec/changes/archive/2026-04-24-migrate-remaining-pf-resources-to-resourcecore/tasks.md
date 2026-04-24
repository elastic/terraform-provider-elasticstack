## 1. Audit rollout scope

- [x] 1.1 Inventory the remaining Plugin Framework resources that still duplicate canonical `client` / `Configure` / `Metadata` wiring and classify them as compatible or out-of-scope based on current `resourcecore` semantics.
- [x] 1.2 Record the component namespace, literal resource-name suffix, and import shape for each compatible resource so the migration preserves existing Terraform type names and import behavior.

## 2. Migrate compatible resources

- [x] 2.1 Convert the compatible Fleet resources to embed `*resourcecore.Core`, initialize the core in their constructors, and replace direct client-field access with `Client()`.
- [x] 2.2 Convert the compatible Kibana resources to embed `*resourcecore.Core`, initialize the core in their constructors, and replace direct client-field access with `Client()`.
- [x] 2.3 Convert any additional audited compatible Plugin Framework resources outside Fleet and Kibana, or explicitly leave audited outliers for follow-up if they still differ from canonical `resourcecore` behavior. *(Includes the audited Elasticsearch PF resources in `resource-inventory.md`, including `elasticstack_elasticsearch_security_api_key` after removing its unused package-level slice.)*

## 3. Verify the rollout

- [x] 3.1 Add a provider-package unit test that iterates the resources registered in `provider/plugin_framework.go` and asserts the registered Plugin Framework resources embed `*resourcecore.Core`.
- [x] 3.2 Add or update targeted tests and compile-time assertions that cover representative migrated resources with passthrough import, custom import, and no import support.
- [x] 3.3 Run targeted `go test` coverage for `./internal/resourcecore/...`, the provider package, and representative migrated packages, then run `make build`.
