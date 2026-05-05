## MODIFIED Requirements

### Requirement: Delete — close before delete (REQ-021–REQ-022)

On delete, the resource SHALL first attempt to close the job by calling the Close Anomaly Detection Job API with `force=true` and `allow_no_match=true`. If the close call fails, the resource SHALL log a warning and continue. After the Close Job API call returns (whether it succeeded or failed), the resource SHALL poll the job's state via the Get Job Stats API until the job reports `closed` state or is no longer found, before calling the Delete Job API. This polling SHALL be bounded by the Terraform operation context (i.e. the delete timeout). If polling fails, the resource SHALL log a warning and continue to deletion.

The resource SHALL then call the Delete Anomaly Detection Job API. If the first delete attempt fails, the resource SHALL retry once with `force=true`. If the retry also fails, the error SHALL be surfaced as a Terraform diagnostic.

#### Scenario: Normal delete succeeds

- **WHEN** the job is `closed` (or not found) before delete is called
- **THEN** the resource SHALL call Delete Job and it SHALL succeed without `force`

#### Scenario: First delete fails — retry with force succeeds

- **WHEN** the initial Delete Job call fails (e.g. job is still open due to polling timeout)
- **THEN** the resource SHALL retry Delete Job with `force=true` and treat a success response as the job being deleted

#### Scenario: Both delete attempts fail

- **WHEN** both the initial Delete Job and the `force=true` retry fail
- **THEN** the resource SHALL surface the retry error as a Terraform diagnostic

#### Scenario: Delete called regardless of polling outcome

- **WHEN** the polling wait fails for any reason
- **THEN** the resource SHALL still call the Delete Job API (not skip it)
