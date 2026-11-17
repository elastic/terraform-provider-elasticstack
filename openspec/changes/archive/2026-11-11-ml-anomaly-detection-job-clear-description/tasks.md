## 1. Removing description clears it

- [x] 1.1 `BuildFromPlan` sends `description` as an empty string when the prior state has a value and the plan is null; unknown planned values are omitted
- [x] 1.2 Unit coverage for `BuildFromPlan` and for update followed by read against a fake Elasticsearch
- [x] 1.3 Acceptance test that creates a job with `description` and then removes it
