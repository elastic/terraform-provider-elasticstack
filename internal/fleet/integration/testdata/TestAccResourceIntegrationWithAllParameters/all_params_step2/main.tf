provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration_all_params" {
  name                         = "tcp"
  version                      = "1.16.0"
  prerelease                   = true
  force                        = true
  ignore_mapping_update_errors = true
  skip_data_stream_rollover    = true
  ignore_constraints           = true
  skip_destroy                 = true
}
