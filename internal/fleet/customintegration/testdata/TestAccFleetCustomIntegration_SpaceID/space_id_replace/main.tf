variable "package_path" {
  type = string
}

variable "space_id" {
  type = string
}

variable "new_space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Keep the space resource stable (same space_id as the previous step).
resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "Test Space for Custom Integration"
  description = "Acceptance test space — managed by TestAccFleetCustomIntegration_SpaceID"
}

# Change only the custom integration's space_id to a different value, which must
# trigger a replacement plan via RequiresReplace(), while the space above stays unchanged.
resource "elasticstack_fleet_custom_integration" "test" {
  package_path = var.package_path
  space_id     = var.new_space_id

  depends_on = [elasticstack_kibana_space.test]
}
