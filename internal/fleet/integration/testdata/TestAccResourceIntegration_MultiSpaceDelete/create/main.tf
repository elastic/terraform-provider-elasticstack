variable "space_a" {
  type = string
}

variable "space_b" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "space_a" {
  space_id = var.space_a
  name     = var.space_a
}

resource "elasticstack_kibana_space" "space_b" {
  space_id = var.space_b
  name     = var.space_b
}

resource "elasticstack_fleet_integration" "test_a" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = false
  space_id     = elasticstack_kibana_space.space_a.space_id
}

resource "elasticstack_fleet_integration" "test_b" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = false
  space_id     = elasticstack_kibana_space.space_b.space_id

  # Serialize the two space-scoped kibana_assets installs of the same
  # package. Installing the same package into two spaces concurrently has
  # been observed to leave one install stuck in Fleet indefinitely (see
  # CI timeouts on TestAccResourceIntegration_MultiSpaceInstall), so force
  # test_a to complete first to keep this test deterministic.
  depends_on = [elasticstack_fleet_integration.test_a]
}
