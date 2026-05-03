## ADDED Requirements

### Requirement: Typed client implementation for snapshot repository CRUD
The resource and data source SHALL use the go-elasticsearch Typed API for all snapshot repository operations. `GetSnapshotRepository` SHALL use `Snapshot.GetRepository().Do(ctx)`, `PutSnapshotRepository` SHALL use `Snapshot.CreateRepository().Do(ctx)`, and `DeleteSnapshotRepository` SHALL use `Snapshot.DeleteRepository().Do(ctx)`. Manual JSON marshaling and unmarshaling SHALL be eliminated.

#### Scenario: Typed API read with union type handling
- GIVEN a successful Get Snapshot Repository API response
- WHEN the provider processes the response
- THEN the typed API response (`getrepository.Response`) SHALL be type-switched over `types.Repository` union variants
- AND each known repository type (`fs`, `url`, `gcs`, `azure`, `s3`, `hdfs`, `source`) SHALL be mapped to its corresponding schema block

#### Scenario: Unknown repository type error
- GIVEN the API returns a repository type not covered by the `types.Repository` union handling
- WHEN read runs
- THEN the provider SHALL return an error diagnostic and SHALL NOT panic

#### Scenario: Typed API write without manual marshal
- GIVEN a snapshot repository to create or update
- WHEN the provider calls the Put API
- THEN the request body SHALL be constructed using typed API request builders
- AND manual `json.Marshal` into an intermediate `models.SnapshotRepository` SHALL NOT occur
