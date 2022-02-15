package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorCircle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorCircle,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_circle.test", "field", "circle"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_circle.test", "json", expectedJsonCircle),
				),
			},
		},
	})
}

const expectedJsonCircle = `{
	"circle": {
		"field": "circle",
		"error_distance": 28.1,
		"shape_type": "geo_shape",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorCircle = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_circle" "test" {
  field          = "circle"
  error_distance = 28.1
  shape_type     = "geo_shape"
}
`
