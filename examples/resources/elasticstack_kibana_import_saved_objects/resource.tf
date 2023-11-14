provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "settings" {
  overwrite     = true
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
