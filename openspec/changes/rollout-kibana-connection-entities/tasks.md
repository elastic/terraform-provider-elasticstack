## 1. Add entity-local schema support

- [ ] 1.1 Add `kibana_connection` to the in-scope Kibana and Fleet entity schemas using the shared provider schema helper for each implementation style.
- [ ] 1.2 Add or update entity models so resource and data source state carries `kibana_connection` where required.

## 2. Adopt effective scoped clients

- [ ] 2.1 Update Kibana entities to resolve an effective client from `kibana_connection` and use the scoped Kibana client surfaces when configured.
- [ ] 2.2 Update Fleet entities to resolve an effective client from `kibana_connection` and use the scoped Fleet client when configured.
- [ ] 2.3 Ensure version checks, space-aware operations, and read-after-write flows continue to run against the effective client for each adopted entity.

## 3. Finish rollout docs and regression checks

- [ ] 3.1 Update the affected entity documentation and examples so `kibana_connection` appears where the rollout adds support.
- [ ] 3.2 Add or update focused entity tests that confirm default provider behavior still works and scoped `kibana_connection` is plumbed through the adopted code paths.
