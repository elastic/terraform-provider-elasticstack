variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "TF ACC Import Saved Objects Test"
  description = "Kibana space for import_saved_objects acceptance test"
}

resource "elasticstack_kibana_import_saved_objects" "settings" {
  space_id  = elasticstack_kibana_space.test.space_id
  overwrite = true

  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
