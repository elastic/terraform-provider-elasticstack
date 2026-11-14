# Dashboard acceptance baseline

## Method

Baseline data comes from three successful `Provider CI` push runs on `main`, before this change. All used the checked-in defaults of six concurrent Go packages (`-p 6`) and ten concurrent tests per package (`-parallel 10`). GitHub job and step timestamps measure elapsed wall-clock time; `gotestsum` test durations are summed separately and are not wall-clock durations because tests run concurrently.

The table reports `TF acceptance tests` step time followed by total job time.

| Run | Stack | Shard 0 step/job | Shard 1 step/job |
| --- | --- | ---: | ---: |
| [40473094948](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/40473094948) (2026-10-27) | 9.4.8 | 38:21 / 42:53 | 23:39 / 27:45 |
|  | 9.5.5 | 32:41 / 36:29 | 38:44 / 42:33 |
|  | 9.6.0-SNAPSHOT | 45:13 / 50:41 | 38:05 / 42:54 |
| [40686687883](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/40686687883) (2026-10-28) | 9.4.8 | 42:12 / 46:50 | 32:51 / 37:59 |
|  | 9.5.5 | 44:44 / 49:29 | 37:16 / 41:04 |
|  | 9.6.0-SNAPSHOT | 45:13 / 51:18 | 26:54 / 31:12 |
| [40723614922](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/40723614922) (2026-10-28) | 9.4.8 | 41:52 / 46:29 | 25:49 / 30:10 |
|  | 9.5.5 | 33:32 / 37:27 | 33:04 / 38:08 |
|  | 9.6.0-SNAPSHOT | 45:12 / 50:44 | 34:16 / 41:06 |

The two hotspot packages both landed on shard 0. Their 9.5.5 `gotestsum` durations were:

| Run | `internal/kibana/dashboard` | `panel/lensxy` |
| --- | ---: | ---: |
| 40473094948 | 45 tests, 169.5 summed min, 18.5 min max | 16 tests, 102.8 summed min, 11.6 min max |
| 40686687883 | 45 tests, 239.2 summed min, 26.4 min max | 16 tests, 144.9 summed min, 16.5 min max |
| 40723614922 | 45 tests, 177.5 summed min, 19.4 min max | 16 tests, 96.9 summed min, 11.7 min max |

In run 40723614922, all dashboard descendants contributed 414.6 summed test minutes on shard 0 and 137.1 minutes on shard 1. Moving the complete 43-package subtree to one Stack would therefore consolidate work currently spread across two Stacks. The selected suite instead isolates only the two measured hotspots and leaves the other descendants in the general pool.

## Slow-test investigation

The longest root-package case was `TestAccResourceDashboardPanels_multi_sections_multi_panels_each`, at 18.5–26.4 minutes across the sampled 9.5.5 runs. It and the related section/panel cases use `resource.ParallelTest`, contain one ordinary acceptance-test step, and do not contain explicit sleeps, polling loops, or retries. The provider's dashboard create, read, and delete paths call the Kibana API directly; no resource-specific wait loop explains these durations.

Dashboard resources inherit generic operation deadlines of 20 minutes for create, update, and delete and five minutes for read. The observed 18.5–26.4 minute whole-test range makes a create or delete request approaching its deadline plausible, but a passing whole-test duration cannot prove which phase consumed the time.

The logs therefore support shared Kibana/API contention as the leading explanation, but they do not identify which create/read/destroy request is slow because request-level elapsed timings are not emitted. The increasing durations for larger panel payloads and the large gap between summed test minutes and wall-clock time are consistent with concurrent pressure, not proof of a provider CPU bottleneck.

If isolation does not reduce the individual test durations, the focused follow-up is request-phase timing around dashboard create, read, update, and delete calls during a selected-package CI run. No provider optimization is justified from the current logs alone.

## Local isolation check

A clean local Stack 9.5.5 run on Apple Silicon used the same in-package concurrency limit (`-parallel 10`) but is not a substitute for hosted-runner measurements.

| Run | Root slow test | Lens XY lifecycle test | Total wall time |
| --- | ---: | ---: | ---: |
| Each test alone | 37.11s | 26.75s | run separately |
| Both complete hotspot packages together (`-p 2`) | 90.63s | 64.45s | 2m24.75s |

The complete concurrent package run passed: the root package took 2m20.06s and Lens XY took 2m19.02s. Concurrency increased the two representative test durations by roughly 2.4 times, confirming contention sensitivity, but both remained far below their 9.5.5 CI ranges of 18.5–26.4 and 11.6–16.5 minutes. This rules out the test definitions being intrinsically 10–26 minutes in isolation and strengthens the case for separating these packages from the full-suite Stack load.

## After-change measurement

Post-change results require a committed branch run and are intentionally pending. Compare the same metrics and concurrency against the baseline above:

- critical-path acceptance step and total job time per Stack version;
- summed runner minutes for all acceptance jobs;
- dedicated hotspot-suite elapsed time;
- the same root dashboard and Lens XY test durations.

The change should be retained only if the slowest 9.4+ job gains useful headroom without disproportionate runner cost.
