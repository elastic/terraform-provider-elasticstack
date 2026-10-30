## 1. Full-suite CI runs dashboards on an isolated Stack

- [x] 1.1 Prepare explicit acceptance matrix assignments that retain two general shards per configured version and add one dedicated dashboard suite only for non-PR versions supporting dashboards, including snapshots, without changing targeted PR plans or version-derived environment flags.
- [x] 1.2 Deliver deterministic full-suite package selection that runs the measured dashboard hotspot packages separately on supported versions, covers every original package exactly once, preserves older-version and local all-package defaults, and rejects invalid suite or shard assignments.
- [x] 1.3 Route the prepared suites through the existing acceptance job lifecycle with distinguishable job names, unchanged targeted PR execution and no-op behavior, and unchanged snapshot, diagnostics, teardown, and Provider Gate handling.
- [x] 1.4 Document isolated-suite operation and timing comparison for contributors, and provide regression coverage demonstrating version boundaries, exact package coverage, path boundaries, invalid input, and preserved workflow behavior.
