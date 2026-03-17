provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  name     = var.space_name
  space_id = "test"
}

resource "elasticstack_fleet_output" "test" {
  name      = "test"
  type      = "elasticsearch"
  hosts     = ["https://elasticsearch:9200"]
  space_ids = ["default", "space1"]
}
