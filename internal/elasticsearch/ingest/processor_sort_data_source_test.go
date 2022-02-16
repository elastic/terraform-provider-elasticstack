package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorSort(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorSort,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_sort.test", "field", "array_field_to_sort"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_sort.test", "json", expectedJsonSort),
				),
			},
		},
	})
}

const expectedJsonSort = `{
	"sort": {
		"field": "array_field_to_sort",
		"ignore_failure": false,
		"order": "desc"
	}
}`

const testAccDataSourceIngestProcessorSort = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_sort" "test" {
  field = "array_field_to_sort"
  order = "desc"
}
`
