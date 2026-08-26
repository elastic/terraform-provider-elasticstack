provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# An empty objects list is rejected by the SizeAtLeast(1) validator.
data "elasticstack_kibana_export_saved_objects" "test" {
  space_id = "default"
  objects  = []
}
