package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorUserAgent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorUserAgent,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "field", "agent"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_user_agent.test", "json", expectedJsonUserAgent),
				),
			},
		},
	})
}

const expectedJsonUserAgent = `{
	"user_agent": {
		"field": "agent",
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorUserAgent = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_user_agent" "test" {
  field = "agent"
}
`
