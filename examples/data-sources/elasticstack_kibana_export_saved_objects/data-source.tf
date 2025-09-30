provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_export_saved_objects" "example" {
  exclude_export_details  = true
  include_references_deep = true
  objects = [
    {
      type = "dashboard",
      id   = "7c5f07ee-7e41-4d50-ae1f-dfe54cc87209"
    }
  ]
}

output "saved_objects" {
  value = data.elasticstack_kibana_export_saved_objects.example.exported_objects
}
