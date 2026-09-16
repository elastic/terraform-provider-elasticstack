provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# kibana_connection omitted: falls back to the provider-level Kibana client.
resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.17.0"
  force        = true
  skip_destroy = true
}
