provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_integration" "full_options_integration" {
  name                         = "tcp"
  version                      = "1.16.0"
  force                        = true
  prerelease                   = false
  ignore_mapping_update_errors = true
  skip_data_stream_rollover    = false
  ignore_constraints           = false
}
