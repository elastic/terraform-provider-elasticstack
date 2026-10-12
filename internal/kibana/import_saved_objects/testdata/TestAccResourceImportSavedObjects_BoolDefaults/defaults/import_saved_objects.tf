variable "object_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# All optional boolean flags are omitted; the resource is expected to leave
# them unset (no schema default) and the object below is unique per test run,
# so the import succeeds without needing overwrite.
resource "elasticstack_kibana_import_saved_objects" "settings" {
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"${var.object_id}","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
