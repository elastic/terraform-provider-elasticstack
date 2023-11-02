package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorAppend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorAppend,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_append.test", "field", "tags"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_append.test", "json", expectedJsonAppend),
				),
			},
		},
	})
}

const expectedJsonAppend = `{
	"append": {
		"field": "tags", 
		"value": ["production", "{{{app}}}", "{{{owner}}}"], 
		"allow_duplicates": true,
		"description": "Append tags to the doc", 
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorAppend = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "test" {
  description      = "Append tags to the doc"
  field            = "tags"
  value            = ["production", "{{{app}}}", "{{{owner}}}"]
  allow_duplicates = true
}
`
