# Excluded from TestAccExamples_planOnly (planOnlySkippedEmbedPaths): agent policy
# and enrollment token reads require a Fleet agent policy that exists in Kibana/Fleet;
# example names and UUIDs vary by stack. Substitute name or policy_id from your Fleet
# agent policy before apply.

provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_agent_policy" "pernode" {
  name     = "k8s-metrics-pernode"
  space_id = "default"
}

data "elasticstack_fleet_enrollment_tokens" "pernode" {
  policy_id = data.elasticstack_fleet_agent_policy.pernode.policy_id
}

data "elasticstack_fleet_agent_policy" "by_id" {
  policy_id = "223b1bf8-240f-463f-8466-5062670d0754"
}
