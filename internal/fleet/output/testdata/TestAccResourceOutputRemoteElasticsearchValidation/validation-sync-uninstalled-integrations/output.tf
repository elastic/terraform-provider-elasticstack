provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                          = "Validation Output ${var.policy_name}"
  output_id                     = "${var.policy_name}-validation-output"
  type                          = "elasticsearch"
  sync_uninstalled_integrations = true
  hosts                         = ["https://elasticsearch:9200"]
}
