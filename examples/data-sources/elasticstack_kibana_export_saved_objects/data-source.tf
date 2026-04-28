provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_export_saved_objects" "example" {
  exclude_export_details  = true
  include_references_deep = true
  objects = [
    {
      type = "dashboard"
      id   = "elastic_agent-02117980-6082-11f0-89d2-bb7ceae5af7f"
    }
  ]
}

output "saved_objects" {
  value = data.elasticstack_kibana_export_saved_objects.example.exported_objects
}
