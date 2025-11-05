package import_saved_objects_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceImportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
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
			{
				Config: testAccResourceImportSavedObjectsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
				),
			},
			{
				// Ensure a partially successful import doesn't throw a provider error
				Config: testAccResourceImportSavedObjectsMissingRef(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "1"),
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

func testAccResourceImportSavedObjectsUpdate() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "settings" {
  overwrite = true
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":false},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
	`
}

func testAccResourceImportSavedObjectsMissingRef() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "settings" {
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"attributes":{"description":"","kibanaSavedObjectMeta":{"searchSourceJSON":"{\"query\":{\"query\":\"\",\"language\":\"kuery\"},\"filter\":[{\"$state\":{\"store\":\"appState\"},\"meta\":{\"alias\":null,\"disabled\":false,\"key\":\"testName\",\"negate\":false,\"params\":{\"query\":\"perf-features\"},\"type\":\"phrase\",\"indexRefName\":\"kibanaSavedObjectMeta.searchSourceJSON.filter[0].meta.index\"},\"query\":{\"match_phrase\":{\"testName\":\"perf-features\"}}},{\"$state\":{\"store\":\"appState\"},\"meta\":{\"alias\":null,\"disabled\":false,\"key\":\"status\",\"negate\":false,\"params\":{\"query\":\"failed\"},\"type\":\"phrase\",\"indexRefName\":\"kibanaSavedObjectMeta.searchSourceJSON.filter[1].meta.index\"},\"query\":{\"match_phrase\":{\"status\":\"failed\"}}},{\"$state\":{\"store\":\"appState\"},\"meta\":{\"alias\":null,\"disabled\":false,\"key\":\"targetEnvironment\",\"negate\":false,\"params\":{\"query\":\"dev\"},\"type\":\"phrase\",\"indexRefName\":\"kibanaSavedObjectMeta.searchSourceJSON.filter[2].meta.index\"},\"query\":{\"match_phrase\":{\"targetEnvironment\":\"dev\"}}}],\"indexRefName\":\"kibanaSavedObjectMeta.searchSourceJSON.index\"}"},"title":"healthchecks","uiStateJSON":"{}","version":1,"visState":"{\"title\":\"healthchecks\",\"type\":\"histogram\",\"aggs\":[{\"id\":\"1\",\"enabled\":true,\"type\":\"count\",\"params\":{},\"schema\":\"metric\"},{\"id\":\"2\",\"enabled\":true,\"type\":\"date_histogram\",\"params\":{\"field\":\"@timestamp\",\"timeRange\":{\"from\":\"now/d\",\"to\":\"now/d\"},\"useNormalizedEsInterval\":true,\"scaleMetricValues\":false,\"interval\":\"30m\",\"drop_partials\":false,\"min_doc_count\":1,\"extended_bounds\":{}},\"schema\":\"segment\"},{\"id\":\"3\",\"enabled\":true,\"type\":\"terms\",\"params\":{\"field\":\"testName\",\"orderBy\":\"1\",\"order\":\"desc\",\"size\":5,\"otherBucket\":false,\"otherBucketLabel\":\"Other\",\"missingBucket\":false,\"missingBucketLabel\":\"Missing\"},\"schema\":\"group\"}],\"params\":{\"type\":\"histogram\",\"grid\":{\"categoryLines\":false},\"categoryAxes\":[{\"id\":\"CategoryAxis-1\",\"type\":\"category\",\"position\":\"bottom\",\"show\":true,\"style\":{},\"scale\":{\"type\":\"linear\"},\"labels\":{\"show\":true,\"filter\":true,\"truncate\":100},\"title\":{}}],\"valueAxes\":[{\"id\":\"ValueAxis-1\",\"name\":\"LeftAxis-1\",\"type\":\"value\",\"position\":\"left\",\"show\":true,\"style\":{},\"scale\":{\"type\":\"linear\",\"mode\":\"normal\"},\"labels\":{\"show\":true,\"rotate\":0,\"filter\":false,\"truncate\":100},\"title\":{\"text\":\"Count\"}}],\"seriesParams\":[{\"show\":true,\"type\":\"histogram\",\"mode\":\"normal\",\"data\":{\"label\":\"Count\",\"id\":\"1\"},\"valueAxis\":\"ValueAxis-1\",\"drawLinesBetweenPoints\":true,\"lineWidth\":2,\"showCircles\":true}],\"addTooltip\":true,\"addLegend\":true,\"legendPosition\":\"right\",\"times\":[],\"addTimeMarker\":true,\"labels\":{\"show\":false},\"thresholdLine\":{\"show\":false,\"value\":10,\"width\":1,\"style\":\"full\",\"color\":\"#E7664C\"},\"palette\":{\"type\":\"palette\",\"name\":\"kibana_palette\"},\"isVislibVis\":true,\"detailedTooltip\":true,\"legendSize\":\"auto\"}}"},"coreMigrationVersion":"8.8.0","id":"bacba650-0c8e-11eb-977b-cd2574857abe","managed":false,"references":[{"id":"5ba9aac0-0c7e-11ea-b151-954d48d0eae6","name":"kibanaSavedObjectMeta.searchSourceJSON.index","type":"index-pattern"},{"id":"5ba9aac0-0c7e-11ea-b151-954d48d0eae6","name":"kibanaSavedObjectMeta.searchSourceJSON.filter[0].meta.index","type":"index-pattern"},{"id":"5ba9aac0-0c7e-11ea-b151-954d48d0eae6","name":"kibanaSavedObjectMeta.searchSourceJSON.filter[1].meta.index","type":"index-pattern"},{"id":"5ba9aac0-0c7e-11ea-b151-954d48d0eae6","name":"kibanaSavedObjectMeta.searchSourceJSON.filter[2].meta.index","type":"index-pattern"}],"type":"visualization","typeMigrationVersion":"8.5.0","updated_at":"2020-10-12T13:27:52.373Z","version":"WzE2NjgsMl0="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":22,"missingRefCount":1,"missingReferences":[{"id":"5ba9aac0-0c7e-11ea-b151-954d48d0eae6","type":"index-pattern"}]}
EOT
  overwrite     = true
}`
}

func TestAccResourceImportSavedObjectsWriteOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceImportSavedObjectsWriteOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_test", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_test", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_test", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_test", "errors.#", "0"),
				),
			},
		},
	})
}

func testAccResourceImportSavedObjectsWriteOnly() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "wo_test" {
  overwrite = true
  file_contents_wo = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
	`
}

func TestAccResourceImportSavedObjectsWriteOnlyWithVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceImportSavedObjectsWriteOnlyWithVersion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_version_test", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_version_test", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_version_test", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.wo_version_test", "errors.#", "0"),
				),
			},
		},
	})
}

func testAccResourceImportSavedObjectsWriteOnlyWithVersion() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "wo_version_test" {
  overwrite = true
  file_contents_wo = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
  file_contents_wo_version = "v1.0.0"
}
	`
}

func TestAccResourceImportSavedObjectsConflictValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceImportSavedObjectsConflict(),
				ExpectError: regexp.MustCompile(`cannot be specified when`),
			},
		},
	})
}

func testAccResourceImportSavedObjectsConflict() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "conflict_test" {
  overwrite = true
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
  file_contents_wo = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":false},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
}
	`
}

func TestAccResourceImportSavedObjectsDependencyValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceImportSavedObjectsDependency(),
				ExpectError: regexp.MustCompile(`Attribute "file_contents_wo" must be specified when\s+"file_contents_wo_version" is specified`),
			},
		},
	})
}

func testAccResourceImportSavedObjectsDependency() string {
	return `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_import_saved_objects" "dependency_test" {
  overwrite = true
  file_contents = <<-EOT
{"attributes":{"buildNum":42747,"defaultIndex":"metricbeat-*","theme:darkMode":true},"coreMigrationVersion":"7.0.0","id":"7.14.0","managed":false,"references":[],"type":"config","typeMigrationVersion":"7.0.0","updated_at":"2021-08-04T02:04:43.306Z","version":"WzY1MiwyXQ=="}
{"excludedObjects":[],"excludedObjectsCount":0,"exportedCount":1,"missingRefCount":0,"missingReferences":[]}
EOT
  file_contents_wo_version = "v1.0.0"
}
	`
}
