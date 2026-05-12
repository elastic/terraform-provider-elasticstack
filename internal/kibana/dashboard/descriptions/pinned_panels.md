Ordered dashboard-level pinned controls (Kibana’s control bar above the grid). Each element uses the same typed `*_control_config` shapes as `panels[]` for these control kinds, without a `grid` block.

When omitted from configuration and Kibana returns an empty list, Terraform keeps this attribute unset (see dashboard resource unset-vs-empty semantics). When set, order is preserved for API requests and read back in API order.
