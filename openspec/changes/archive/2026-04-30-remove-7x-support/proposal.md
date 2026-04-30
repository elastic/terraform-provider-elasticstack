## Why

Elastic Stack 7.x is no longer a supported target for this provider. Keeping 7.x in the documented support floor and acceptance test matrix adds CI cost and preserves compatibility branches that are only relevant to unsupported stack versions.

## What Changes

- **BREAKING**: Update the documented minimum supported Elastic Stack version from 7.x to 8.0.
- Remove Elastic Stack 7.x from the acceptance test matrix.
- Remove 7.x-only compatibility checks and test skips where the behavior is only needed for stack versions below 8.0.
- Keep existing version gates for 8.x and 9.x feature boundaries.
- Do not add a global runtime block for 7.x; the provider may continue to work incidentally where APIs remain compatible, but 7.x is no longer documented or intentionally tested.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `ci-build-lint-test`: The matrix acceptance test coverage no longer includes Elastic Stack 7.x.
- `makefile-workflows`: Fleet Docker image selection no longer needs a 7.17-specific fallback.
- `elasticsearch-transform`: Transform behavior no longer needs compatibility gates for Elasticsearch versions below 8.0.
- `elasticsearch-index-lifecycle`: ILM behavior no longer needs compatibility gates or documentation for Elasticsearch 7.16-only support boundaries.

## Impact

- `README.md` and generated Terraform docs will advertise Elastic Stack 8.0 or higher.
- `.github/workflows-src/test/workflow.yml.tmpl` and generated `.github/workflows/test.yml` will drop the 7.17 acceptance matrix entry.
- `Makefile` will keep Docker Hub Fleet image fallback only for 8.0 and 8.1 stack lines.
- Resource and data source code may be simplified where pre-8.0 version checks only guarded unsupported 7.x behavior.
- OpenSpec specs and generated docs will be updated to match the 8.0+ support floor, including stale 7.x references in processor descriptions.
