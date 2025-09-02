package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorUrldecode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorUrldecode,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_urldecode.test", "field", "my_url_to_decode"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_urldecode.test", "json", expectedJsonUrldecode),
				),
			},
		},
	})
}

const expectedJsonUrldecode = `{
	"urldecode": {
		"field": "my_url_to_decode",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorUrldecode = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_urldecode" "test" {
  field = "my_url_to_decode"
}
`
