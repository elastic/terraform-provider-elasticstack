provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "example" {
  space_id          = "test_space"
  name              = "Test Space"
  description       = "A fresh space for testing visualisations"
  disabled_features = ["ingestManager", "enterpriseSearch"]
  initials          = "ts"
}
