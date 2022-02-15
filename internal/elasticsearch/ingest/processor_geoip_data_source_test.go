package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorGeoip(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorGeoip,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "field", "ip"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_geoip.test", "json", expectedJsonGeoip),
				),
			},
		},
	})
}

const expectedJsonGeoip = `{
  "geoip": {
		"field": "ip",
		"first_only": true,
		"ignore_missing": false,
		"target_field": "geoip"
	}
}
`

const testAccDataSourceIngestProcessorGeoip = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_geoip" "test" {
  field = "ip"
}
`
