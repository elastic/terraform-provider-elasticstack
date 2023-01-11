package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorRegisteredDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorRegisteredDomain,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "field", "fqdn"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "target_field", "url"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_registered_domain.test", "json", expectedJsonRegisteredDomain),
				),
			},
		},
	})
}

const expectedJsonRegisteredDomain = `{
	"registered_domain": {
		"field": "fqdn",
		"target_field": "url",
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorRegisteredDomain = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_registered_domain" "test" {
  field        = "fqdn"
  target_field = "url"
}
`
