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
  name        = "acc-osquery-pack-${var.space_id}"
  description = "Kibana space for osquery pack acceptance test"
}

resource "elasticstack_kibana_osquery_pack" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "tf-acc-osquery-pack-space-${var.suffix}"
  description = "Terraform non-default space acceptance test pack"
  enabled     = true

  queries = {
    find_procs = {
      query    = "SELECT pid, name FROM processes LIMIT 5;"
      platform = ["linux", "darwin"]
      ecs_mapping = {
        "process.name" = {
          field = "name"
        }
      }
    }
  }
}
