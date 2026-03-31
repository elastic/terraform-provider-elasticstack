# Tasks: Fleet Agent Download Source OpenSpec migration

## 1. Spec and docs

- [ ] 1.1 Confirm delta spec `specs/fleet-agent-download-source/spec.md` matches intended behavior and passes `openspec validate fleet-agent-download-source --type change`
- [ ] 1.2 After implementation or on merge policy, sync delta into `openspec/specs/fleet-agent-download-source/spec.md` or archive the change per project workflow

## 2. Implementation alignment

- [ ] 2.1 Verify `internal/fleet/agentdownloadsource` implements CRUD, identity (`id` / `source_id`), `space_ids`, import, error handling (including 404), update vs replace semantics, and post-create/post-update read convergence via the shared read path
- [ ] 2.2 Verify `internal/clients/fleet` uses `generated/kbapi` for all agent download source calls with correct space routing
- [ ] 2.3 Confirm minimum Kibana version guard exists and matches product docs once version is chosen

## 3. Testing

- [ ] 3.1 Run or add acceptance tests covering create, update, destroy, import, and space-scoped behavior as required by the spec
- [ ] 3.2 Add or verify unit tests for API mapping (`name`, `host`, `default`/`is_default`, `proxy_id`)
- [ ] 3.3 Add or verify tests that create/update state is derived from the follow-up read path rather than mutation response payloads
