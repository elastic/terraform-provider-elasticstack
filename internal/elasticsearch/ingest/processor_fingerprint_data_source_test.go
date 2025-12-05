package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorFingerprint(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorFingerprint,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_fingerprint.test", "json", expectedJsonFingerprint),
				),
			},
		},
	})
}

const expectedJsonFingerprint = `{
  "fingerprint": {
		"fields": [
			"user"
		],
		"ignore_failure": false,
		"ignore_missing": false,
		"method": "SHA-1",
		"target_field": "fingerprint"
	}
}
`

const testAccDataSourceIngestProcessorFingerprint = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_fingerprint" "test" {
  fields = ["user"]
}
`
