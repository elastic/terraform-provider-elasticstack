package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorNetworkDirection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorNetworkDirection,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_network_direction.test", "json", expectedJsonNetworkDirection),
				),
			},
		},
	})
}

const expectedJsonNetworkDirection = `{
	"network_direction": {
		"ignore_failure": false,
		"ignore_missing": true,
		"internal_networks": [
			"private"
		]
	}
}`

const testAccDataSourceIngestProcessorNetworkDirection = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_network_direction" "test" {
  internal_networks = ["private"]
}
`
