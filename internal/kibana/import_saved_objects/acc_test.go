package import_saved_objects_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceImportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceImportSavedObjects(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
				),
			},
		},
	})
}

func testAccResourceImportSavedObjects() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "settings" {
  overwrite = true
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
	`
}
