variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The write index name"
  type        = string
}

variable "index_name2" {
  description = "The read index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "index1" {
  name                = var.index_name
  deletion_protection = false
  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index" "index2" {
  name                = var.index_name2
  deletion_protection = false
  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  write_index = {
    name           = elasticstack_elasticsearch_index.index1.name
    routing        = "wr2"
    index_routing  = "wir2"
    search_routing = "wsr2"
  }

  read_indices = [{
    name           = elasticstack_elasticsearch_index.index2.name
    routing        = "rr2"
    index_routing  = "rir2"
    search_routing = "rsr2"
  }]
}
