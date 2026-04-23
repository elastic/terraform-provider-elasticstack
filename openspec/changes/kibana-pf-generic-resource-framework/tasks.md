## 1. Create the shared Kibana Plugin Framework resource foundation

- [x] 1.1 Add a shared package under `internal/kibana/` for Plugin Framework resource orchestration, including provider configuration/metadata helpers, Kibana scoped-client resolution, version enforcement, and composite-ID helpers for space-aware resources
- [x] 1.2 Define the generic resource assembly surface plus the typed model and API contracts that separate Terraform lifecycle orchestration, model/request mapping, and Kibana transport operations
- [x] 1.3 Add unit coverage for the shared framework behavior, especially version enforcement, composite-ID handling, read-after-write flow, and remote-not-found read behavior

## 2. Introduce focused Agent Builder transport APIs

- [ ] 2.1 Reorganize the Agent Builder transport helpers under `internal/clients/kibanaoapi/` into focused APIs that implement the new generic resource transport contracts for agents, tools, and workflows
- [ ] 2.2 Preserve existing request/response semantics and error handling while reducing duplicated endpoint helper wiring
- [ ] 2.3 Add or update tests for the focused transport APIs as needed to keep endpoint behavior covered during the migration

## 3. Migrate Agent Builder resources to the shared framework

- [ ] 3.1 Migrate `internal/kibana/agentbuilderagent/` to the shared framework while preserving schema, import behavior, version-gate messaging, and acceptance semantics
- [ ] 3.2 Migrate `internal/kibana/agentbuildertool/` to the shared framework while preserving normalized JSON configuration handling, space-aware state population, and acceptance semantics
- [ ] 3.3 Migrate `internal/kibana/agentbuilderworkflow/` to the shared framework while preserving normalized YAML handling, optional computed workflow IDs, invalid-workflow diagnostics, and acceptance semantics

## 4. Capture and verify the new reusable capability

- [ ] 4.1 Add delta specs for the new generic-resource capability and the modified Agent Builder capabilities
- [ ] 4.2 Verify the migrated resources with targeted unit and acceptance tests and validate the OpenSpec change artifacts
- [ ] 4.3 Document any follow-on adoption considerations discovered during the Agent Builder migration without expanding the implementation scope in this change
