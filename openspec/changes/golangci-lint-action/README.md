# golangci-lint-action

Split the provider CI lint job: run golangci-lint via golangci/golangci-lint-action in a dedicated parallel job, and run remaining lint checks (openspec, fmt, gen, docs) in a separate job
