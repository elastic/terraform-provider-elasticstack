## 1. Older stateful stacks reject affected dashboard writes with an actionable diagnostic

- [x] 1.1 Dashboard create and update writes detect affected Lens XY `y[]` metrics in top-level and
      section panels and reject unsupported stateful versions before issuing the dashboard API
      request, with a diagnostic identifying the offending panel, layer, metric, and required
      version.
- [x] 1.2 When status retrieval or version parsing fails during preflight, propagate the returned
      diagnostic and abort the dashboard request without substituting the minimum-version message;
      cover both failures in write-path tests.
- [x] 1.3 Document the 9.5.2 minimum version for the affected `y[].config_json` format in the Lens XY
      chart attribute documentation.
- [x] 1.4 Verify the older-version diagnostic and aborted create/update requests for both top-level
      panels and panels nested in sections.

## 2. Supported stateful and Serverless dashboard writes proceed without gate errors

- [x] 2.1 Dashboard creates and updates with the affected format proceed without a version-gate
      diagnostic on Kibana 9.5.2 or later, and acceptance coverage confirms the configuration has no
      follow-up plan drift.
- [x] 2.2 Dashboard preflight/write-path coverage uses a Serverless version result and verifies that
      the affected configuration proceeds without a gate diagnostic.
- [x] 2.3 Detection recognizes `type` or legacy `id` duration formats with `decimals` only,
      `compact` only, both properties, or either property nested in `format.params`; missing/null
      formats, unaffected `number`/`percent` formats, and dashboards without a match continue without
      a version check or gate diagnostic.
