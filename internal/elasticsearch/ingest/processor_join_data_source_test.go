package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorJoin(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorJoin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_join.test", "field", "joined_array_field"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_join.test", "json", expectedJsonJoin),
				),
			},
		},
	})
}

const expectedJsonJoin = `{
	"join": {
		"field": "joined_array_field",
		"ignore_failure": false,
		"separator": "-"
	}
}`

const testAccDataSourceIngestProcessorJoin = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_join" "test" {
  field     = "joined_array_field"
  separator = "-"
}
`
