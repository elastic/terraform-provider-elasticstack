# Excluded from TestAccExamples_planOnly (planOnlySkippedEmbedPaths): enrollment token
# reads require a Fleet agent policy_id that exists in Kibana/Fleet; example UUIDs vary by stack.
# Substitute policy_id from your Fleet agent policy before apply.

provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = "223b1bf8-240f-463f-8466-5062670d0754"
}
