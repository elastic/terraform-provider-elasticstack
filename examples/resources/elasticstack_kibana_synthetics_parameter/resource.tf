provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_parameter" "example" {
  key         = "example_key"
  value       = "example_value"
  description = "Example description"
  tags        = ["tag-a", "tag-b"]
}
