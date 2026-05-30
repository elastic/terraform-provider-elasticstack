variable "space_id" {
  description = "The space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id          = var.space_id
  name              = format("Initials %s", var.space_id)
  description       = "Single initial test"
  initials          = "Z"
  disabled_features = []
}
