variable "aliases" {
  type = list(object({
    name = string

    write_index = optional(object({
      name      = string
      is_hidden = optional(bool)
    }), null)

    read_indices = optional(set(object({
      name      = string
      is_hidden = optional(bool)
    })), [])
  }))
}

provider "elasticstack" {
  elasticsearch {}
}

locals {
  # This is intentionally unknown during plan (index IDs are computed),
  # but will be known during apply. It mirrors real-world scenarios where
  # alias inputs come from upstream resources/modules.
  enable_apply_time_values = elasticstack_elasticsearch_index.write[0].id != ""

  write_index_names = sort(distinct([
    for alias in var.aliases : alias.write_index.name
    if alias.write_index != null
  ]))

  read_index_names = sort(distinct(flatten([
    for alias in var.aliases : [
      for read_index in alias.read_indices : read_index.name
    ]
  ])))
}

resource "elasticstack_elasticsearch_index" "write" {
  count               = length(local.write_index_names)
  name                = local.write_index_names[count.index]
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index" "read" {
  count               = length(local.read_index_names)
  name                = local.read_index_names[count.index]
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "this" {
  count = length(var.aliases)
  name  = var.aliases[count.index].name

  write_index = local.enable_apply_time_values ? (
    var.aliases[count.index].write_index == null ? null : var.aliases[count.index].write_index
  ) : null

  read_indices = local.enable_apply_time_values ? (
    var.aliases[count.index].read_indices == null ? null : var.aliases[count.index].read_indices
  ) : null

  depends_on = [
    elasticstack_elasticsearch_index.write,
    elasticstack_elasticsearch_index.read,
  ]
}

