provider "elasticstack" {
  fleet {}
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = "223b1bf8-240f-463f-8466-5062670d0754"
}
