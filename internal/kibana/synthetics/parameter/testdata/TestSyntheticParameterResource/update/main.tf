provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_parameter" "test" {
  key         = "test-key-2"
  value       = "test-value-2"
  description = "Test description 2"
  tags        = ["c", "d", "e"]
}
