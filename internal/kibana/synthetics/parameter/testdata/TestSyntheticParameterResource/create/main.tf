provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_parameter" "test" {
  key         = "test-key"
  value       = "test-value"
  description = "Test description"
  tags        = ["a", "b"]
}
