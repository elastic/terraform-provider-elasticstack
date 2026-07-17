## 1. Document serverless delete behavior

- [x] 1.1 Add an OpenSpec delta that defines serverless lifecycle deletion as a warning-only, no-API-request operation that removes the Terraform state entry.
- [x] 1.2 Sync the delta into the canonical `elasticsearch-data-stream-lifecycle` specification.

## 2. Validate the change

- [x] 2.1 Run focused unit tests for `internal/clients/elasticsearch`.
- [x] 2.2 Validate canonical and active-change OpenSpec specifications.
