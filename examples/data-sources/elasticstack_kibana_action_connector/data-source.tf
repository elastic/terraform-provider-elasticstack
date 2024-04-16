
provider "elasticstack" {
  elasticsearch {}
  kibana {}

}

data "elasticstack_kibana_action_connector" "example" {
  name           = "myslackconnector"
  space_id       = "default"
  connector_type = ".slack"
}

output "connector_id" {
  value = data.elasticstack_kibana_action_connector.example.connector_id
}
