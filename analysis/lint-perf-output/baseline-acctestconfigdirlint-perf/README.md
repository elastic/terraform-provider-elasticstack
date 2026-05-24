# acctestconfigdirlint perf baseline

Reference snapshot for the `acctestconfigdirlint-perf` OpenSpec change. Regenerate
locally with `make lint-perf` and compare against these files when reviewing
analyzer performance regressions.

## Headline numbers (Apple M4 Pro, darwin/arm64)

| Metric                                 | Pre-change (proposal) | Post-change |
| -------------------------------------- | --------------------: | ----------: |
| `golangci-lint --concurrency=1 ./...`  |                ~29.5s |        6.5s |
| `BenchmarkAnalyzer_Compliant`          |              ~1.6s/op |    ~1.2s/op |
| `BenchmarkAnalyzer_Violations`         |              ~1.6s/op |    ~1.0s/op |

`BenchmarkAnalyzer_LargePackage` is new in this change; it runs both compliant
and violations testdata packages in a single iteration and lands around
~1.3s/op.

The benchmark numbers above were sampled with `go test
./analysis/acctestconfigdirlint/... -bench=. -benchmem -benchtime=3s -count=3`;
individual benchmark iterations vary by ±25% on this hardware due to
`analysistest` startup variance, so the wall-clock golangci number is the
authoritative regression signal.

## Files

- `acctestconfigdirlint-lint.txt` — `time` output for the isolated golangci run
- `acctestconfigdirlint-bench.txt` — `go test -bench` summary
- `acctestconfigdirlint-golangci-cpu.prof`, `acctestconfigdirlint-golangci-mem.prof` — golangci-level pprof samples
- `acctestconfigdirlint-cpu.prof`, `acctestconfigdirlint-mem.prof` — benchmark-level pprof samples

Trace files (`*-trace.out`) are excluded from the committed baseline because
they exceed 10 MB each; regenerate them locally with `make lint-perf` when
needed.
