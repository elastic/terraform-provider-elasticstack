## MODIFIED Requirements

### Requirement: Delete — close before delete (REQ-021–REQ-022)

On delete, the resource SHALL first attempt to close the job by calling the Close Anomaly Detection Job API with `force=true` and `allow_no_match=true`. If the close call fails, the resource SHALL log a warning and continue. After the Close Job API call returns (whether it succeeded or failed), the resource SHALL poll the job's state via the Get Job Stats API until the job reports `closed` state or is no longer found, before calling the Delete Job API. This polling SHALL be bounded by the Terraform operation context (i.e. the delete timeout). The resource SHALL then call the Delete Anomaly Detection Job API. A non-success response from the Delete API SHALL be surfaced as an error diagnostic.

#### Scenario: Close succeeds before delete

- **WHEN** delete is called on an open anomaly detection job
- **THEN** the resource SHALL call Close Job, then poll until the job is `closed`, then call Delete Job

#### Scenario: Close returns error; polling confirms job is already closed

- **WHEN** the Close Job API call returns an error
- **THEN** the resource SHALL log a warning, poll until the job reports `closed` state, and then call Delete Job

#### Scenario: Polling confirms job not found before delete

- **WHEN** polling the job stats returns no result (job not found)
- **THEN** the resource SHALL treat the job as closed and proceed to Delete Job

#### Scenario: Delete is called after polling confirms closed state

- **WHEN** the job reaches `closed` state and Delete Job is called
- **THEN** the resource SHALL call Delete Job only after observing `closed` state, and SHALL not fail with a version conflict caused by the close/delete race
