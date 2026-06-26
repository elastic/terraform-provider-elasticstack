variable "suffix" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-osquery-pack-ds-${var.space_id}"
  description = "Kibana space for osquery pack data source acceptance test"
}

resource "elasticstack_kibana_osquery_pack" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "tf-acc-osquery-pack-ds-space-${var.suffix}"
  description = "Terraform data source non-default space acceptance test pack"
  enabled     = true

  queries = {
    find_procs = {
      query    = "SELECT pid, name FROM processes LIMIT 5;"
      platform = ["linux", "darwin"]
      version  = "1.0.0"
      snapshot = false
      removed  = false
      ecs_mapping = {
        "process.name" = {
          field = "name"
        }
        "process.pid" = {
          value = "0"
        }
        "host.name" = {
          values = ["host-a", "host-b"]
        }
      }
    }
  }
}

data "elasticstack_kibana_osquery_pack" "test" {
  space_id = elasticstack_kibana_space.test.space_id
  pack_id  = elasticstack_kibana_osquery_pack.test.pack_id
}
