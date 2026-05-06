## 1. Envelope Core

- [ ] 1.1 Define `KibanaResourceModel` interface (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`) in `internal/entitycore/`
- [ ] 1.2 Define `KibanaCreateFunc[T]`, `KibanaUpdateFunc[T]`, `kibanaReadFunc[T]`, `kibanaDeleteFunc[T]` callback function types
- [ ] 1.3 Implement `KibanaResource[T]` struct embedding `*ResourceBase` with schema factory and four callback fields
- [ ] 1.4 Implement `NewKibanaResource[T]` constructor
- [ ] 1.5 Implement `Schema` method — inject `kibana_connection` block (parallel to `ElasticsearchResource.Schema`)
- [ ] 1.6 Implement the composite-ID-or-fallback helper: call `CompositeIDFromStr(GetID())` (or the `Fw` variant) and inspect the returned `*CompositeID` — discard any returned diagnostics (parse failure is a non-error "not composite" signal, same pattern as `getMaintenanceWindowIDAndSpaceID`); fall back to `GetResourceID()` + `GetSpaceID()` when the result is nil
- [ ] 1.7 Implement `Create` — decode plan, validate `spaceID` non-empty, resolve `KibanaClient`, invoke `createFunc(ctx, client, spaceID, plan)`, persist state
- [ ] 1.8 Implement `Read` — decode state, resolve identity via composite-ID-or-fallback, validate `resourceID` non-empty, resolve `KibanaClient`, invoke `readFunc(ctx, client, resourceID, spaceID, model)`, found/not-found branching
- [ ] 1.9 Implement `Update` — decode plan and prior state, resolve identity via composite-ID-or-fallback on plan model, validate `resourceID` non-empty, resolve `KibanaClient`, invoke `updateFunc(ctx, client, resourceID, spaceID, plan, prior)`, persist state
- [ ] 1.10 Implement `Delete` — decode state, resolve identity via composite-ID-or-fallback, validate `resourceID` non-empty, resolve `KibanaClient`, invoke `deleteFunc(ctx, client, resourceID, spaceID, model)`
- [ ] 1.11 Implement `PlaceholderKibanaWriteCallbacks[T]()` returning error-surfacing `KibanaCreateFunc[T]` and `KibanaUpdateFunc[T]`
- [ ] 1.12 Add `var _ resource.Resource = (*KibanaResource[KibanaResourceModel])(nil)` compile-time assertion
- [ ] 1.13 Update `entitycore/doc.go` to document the Kibana resource envelope pattern alongside the existing Elasticsearch and data-source patterns

## 2. Envelope Unit Tests

- [ ] 2.1 Add `testKibanaResourceModel` satisfying `KibanaResourceModel` for envelope tests (user-ID variant: `GetResourceID()` returns `m.Name`)
- [ ] 2.2 Test `NewKibanaResource` type assertions: satisfies `resource.Resource`, `resource.ResourceWithConfigure`, does NOT satisfy `resource.ResourceWithImportState`
- [ ] 2.3 Test `Metadata` produces the correct Terraform type name
- [ ] 2.4 Test `Schema` injects `kibana_connection` block and does not mutate the factory return value
- [ ] 2.5 Test `Configure` — nil provider data, valid factory, invalid provider data
- [ ] 2.6 Test `Create` happy path — model decoded, spaceID validated, client resolved, callback invoked, state persisted
- [ ] 2.7 Test `Create` short-circuits: empty spaceID, client resolution failure, callback error
- [ ] 2.8 Test `Create` with nil and placeholder write callbacks
- [ ] 2.9 Test `Read` happy path (found) — composite ID parse path (user-ID resource)
- [ ] 2.10 Test `Read` happy path (found) — fallback path (plain-UUID resource)
- [ ] 2.11 Test `Read` not-found removes resource from state
- [ ] 2.12 Test `Read` short-circuits: state decode error, empty resourceID, client resolution failure, read callback error
- [ ] 2.13 Test `Update` happy path — both plan and prior state decoded, callback receives both
- [ ] 2.14 Test `Update` short-circuits: empty resourceID, client resolution failure, callback error
- [ ] 2.15 Test `Update` with nil and placeholder write callbacks
- [ ] 2.16 Test `Delete` happy path
- [ ] 2.17 Test `Delete` short-circuits: state decode error, empty resourceID, client resolution failure, delete callback error

## 3. Streams Migration (User-ID POC)

- [ ] 3.1 Add `GetResourceID() types.String` (returns `m.Name`), `GetSpaceID() types.String`, and `GetKibanaConnection() types.List` value-receiver methods to `streamModel`
- [ ] 3.2 Change `Resource` struct to embed `*entitycore.KibanaResource[streamModel]` instead of `*entitycore.ResourceBase`
- [ ] 3.3 Update `newResource()` to call `entitycore.NewKibanaResource[streamModel]` with the schema factory and extracted callback functions
- [ ] 3.4 Extract the Create body into a `KibanaCreateFunc[streamModel]` — receives `(ctx, client, spaceID, plan)`; extracts name from `plan.GetResourceID()`
- [ ] 3.5 Extract the Read body into a `kibanaReadFunc[streamModel]` — receives `(ctx, client, resourceID, spaceID, model)`; removes the internal composite ID parse (envelope handles it)
- [ ] 3.6 Extract the Update body into a `KibanaUpdateFunc[streamModel]` — receives `(ctx, client, resourceID, spaceID, plan, prior)`
- [ ] 3.7 Extract the Delete body into a `kibanaDeleteFunc[streamModel]` — receives `(ctx, client, resourceID, spaceID, model)`
- [ ] 3.8 Remove now-redundant `resource.go` CRUD method stubs and inline boilerplate
- [ ] 3.9 Verify `make build` passes and existing streams acceptance tests pass

## 4. Maintenance Window Migration (API-UUID POC)

- [ ] 4.1 Add `GetResourceID() types.String` (returns `m.ID`), `GetSpaceID() types.String`, and `GetKibanaConnection() types.List` value-receiver methods to maintenance window `Model`
- [ ] 4.2 Change `Resource` struct to embed `*entitycore.KibanaResource[Model]` instead of `*entitycore.ResourceBase`
- [ ] 4.3 Update `newResource()` to call `entitycore.NewKibanaResource[Model]` with the schema factory and extracted callback functions
- [ ] 4.4 Extract the Create body into a `KibanaCreateFunc[Model]` — receives `(ctx, client, spaceID, plan)`; UUID from API response is set on the returned model
- [ ] 4.5 Extract the Read body into a `kibanaReadFunc[Model]` — receives `(ctx, client, resourceID, spaceID, model)`; removes `getMaintenanceWindowIDAndSpaceID()` call (envelope handles composite-ID-or-fallback)
- [ ] 4.6 Extract the Update body into a `KibanaUpdateFunc[Model]` — receives `(ctx, client, resourceID, spaceID, plan, prior)`
- [ ] 4.7 Extract the Delete body into a `kibanaDeleteFunc[Model]` — receives `(ctx, client, resourceID, spaceID, model)`; removes `getMaintenanceWindowIDAndSpaceID()` call
- [ ] 4.8 Remove `getMaintenanceWindowIDAndSpaceID()` helper if no longer needed
- [ ] 4.9 Remove now-redundant `resource.go` CRUD method stubs and inline boilerplate
- [ ] 4.10 Verify `make build` passes and existing maintenance window acceptance tests pass
