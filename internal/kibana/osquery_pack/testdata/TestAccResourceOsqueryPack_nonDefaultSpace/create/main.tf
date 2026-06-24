variable "suffix" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-osquery-pack-${var.space_id}"
  description = "Kibana space for osquery pack acceptance test"
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Osquery Pack Space Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-osquery-pack-space-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = [elasticstack_kibana_space.test.space_id]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = "Osquery Pack Space Agent Policy ${var.suffix}"
  namespace          = replace(var.space_id, "-", "_")
  description        = "Osquery pack non-default space acceptance test agent policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  space_ids          = [elasticstack_kibana_space.test.space_id]
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_kibana_osquery_pack" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  name        = "tf-acc-osquery-pack-space-${var.suffix}"
  description = "Terraform non-default space acceptance test pack"
  enabled     = true

  policy_ids = [elasticstack_fleet_agent_policy.test_policy.policy_id]
  shards = {
    (elasticstack_fleet_agent_policy.test_policy.policy_id) = 100
  }

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
