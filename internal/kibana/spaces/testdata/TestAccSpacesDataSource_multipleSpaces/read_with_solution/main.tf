variable "space_id" {
  description = "The space ID for the custom test space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id          = var.space_id
  name              = "Test Coverage Space"
  description       = "Test space for data source coverage"
  initials          = "TC"
  color             = "#E8478B"
  solution          = "classic"
  disabled_features = ["ingestManager"]
}

data "elasticstack_kibana_spaces" "all_spaces" {
  depends_on = [elasticstack_kibana_space.test]
}
