variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Osquery Pack Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-osquery-pack-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = "Osquery Pack Agent Policy ${var.suffix}"
  namespace          = "testacc"
  description        = "Osquery pack acceptance test agent policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-${var.suffix}"
  description = "Terraform acceptance test pack"
  enabled     = true

  policy_ids = [elasticstack_fleet_agent_policy.test_policy.policy_id]
  shards = {
    (elasticstack_fleet_agent_policy.test_policy.policy_id) = 100
  }

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
