variable "package_path" {
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
  name        = "Test Space for Custom Integration"
  description = "Acceptance test space — managed by TestAccFleetCustomIntegration_SpaceID"
}

resource "elasticstack_fleet_custom_integration" "test" {
  package_path = var.package_path
  space_id     = elasticstack_kibana_space.test.space_id

  depends_on = [elasticstack_kibana_space.test]
}
