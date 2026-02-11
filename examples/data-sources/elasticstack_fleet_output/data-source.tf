provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_output" "example" {
  output_id = "my-fleet-output"
}
