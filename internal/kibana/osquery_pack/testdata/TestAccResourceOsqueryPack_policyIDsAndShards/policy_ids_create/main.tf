variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name         = "tf-acc-osquery-policy-${var.suffix}"
  namespace    = "default"
  skip_destroy = true
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name       = "tf-acc-osquery-pack-policy-${var.suffix}"
  enabled    = true
  policy_ids = [elasticstack_fleet_agent_policy.test.policy_id]
  shards = {
    (elasticstack_fleet_agent_policy.test.policy_id) = 50
  }

  queries = {
    find_procs = {
      query = "SELECT pid, name FROM processes LIMIT 5;"
    }
  }
}
