provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  name     = var.space_name
  space_id = "test"
}

resource "elasticstack_fleet_output" "test" {
  name                   = "test"
  type                   = "elasticsearch"
  hosts                  = ["https://elasticsearch:9200"]
  space_ids              = ["default", "space1"]
  ca_sha256              = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
  ca_trusted_fingerprint = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
}

data "elasticstack_fleet_output" "test" {
  space_id = "space"
}
