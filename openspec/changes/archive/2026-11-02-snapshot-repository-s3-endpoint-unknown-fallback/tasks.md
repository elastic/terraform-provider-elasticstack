## 1. S3 repositories without endpoint apply cleanly

- [x] 1.1 S3 read-back in `internal/elasticsearch/snapshot/repository/read.go` inherits a prior `endpoint` or `path_style_access` only when it is known (REQ-017): an unknown prior `endpoint` with no `endpoint` in the GET response yields null, an unknown prior `path_style_access` yields the schema default, a known prior value is still preserved, and a value returned by the API always wins
- [x] 1.2 An S3 repository without `endpoint` gets through create, an unchanged re-plan, an update of another S3 setting, and removal of a previously set `endpoint` together with another change against a live stack: every apply succeeds, `s3.endpoint` is null whenever it is not configured, and the re-plan is empty
