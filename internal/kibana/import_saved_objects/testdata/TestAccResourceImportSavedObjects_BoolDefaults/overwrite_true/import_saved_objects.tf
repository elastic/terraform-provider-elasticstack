variable "object_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Sets overwrite = true for the same object_id used in the "defaults" step,
# exercising update coverage for a boolean flag.
resource "elasticstack_kibana_import_saved_objects" "settings" {
  overwrite = true

  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"${var.object_id}","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
