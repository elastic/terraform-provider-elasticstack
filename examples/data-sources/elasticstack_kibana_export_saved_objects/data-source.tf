provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Minimal connector so the export references a saved object that exists on first plan
# (dashboard IDs are environment-specific; action connectors created here are deterministic).
resource "elasticstack_kibana_action_connector" "export_demo" {
  name              = "examples-export-demo-connector"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://example.com"
  })
}

data "elasticstack_kibana_export_saved_objects" "example" {
  space_id                = "default"
  exclude_export_details  = true
  include_references_deep = true
  objects = [
    {
      type = "action"
      id   = elasticstack_kibana_action_connector.export_demo.connector_id
    }
  ]
}

output "saved_objects" {
  value = data.elasticstack_kibana_export_saved_objects.example.exported_objects
}
