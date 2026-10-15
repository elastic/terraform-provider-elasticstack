provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "orders" {
  name           = "orders"
  index_patterns = ["orders-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_shards   = 1
      number_of_replicas = 1
    })
  }
}

resource "elasticstack_elasticsearch_rollover_index_series" "orders" {
  prefix         = "orders"
  rollover_alias = "orders-write"

  # Defaults to true. A prior, separate apply must set this to false before destroy
  # or replacement. Setting it false in the same operation that replaces the resource
  # still uses the previously applied protection, so destroy deletes nothing.
  # After that apply, destroy deletes every current alias member by concrete name.
  # Quiesce producers and let in-flight replacements finish first. Membership
  # rechecks do not see a transformation target that is not yet attached.
  # Searchable snapshots and detached indices need separate cleanup, and other
  # series can still block deletion of a shared ILM policy.
  deletion_protection = false

  # Changing prefix or rollover_alias replaces the resource. Terraform Core orders
  # that replacement. Keep destroy-before-create. Do not set create_before_destroy
  # here or on a dependency that propagates it: the old instance can then delete
  # members the replacement just adopted.

  depends_on = [elasticstack_elasticsearch_index_template.orders]
}
