package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorGrok(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorGrok,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_grok.test", "field", "message"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_grok.test", "json", expectedJsonGrok),
				),
			},
		},
	})
}

const expectedJsonGrok = `{
  "grok": {
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern_definitions": {
			"FAVORITE_CAT": "burmese",
			"FAVORITE_DOG": "beagle"
		},
		"patterns": [
			"%{FAVORITE_DOG:pet}",
			"%{FAVORITE_CAT:pet}"
		],
		"trace_match": false
	}
}
`

const testAccDataSourceIngestProcessorGrok = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  field    = "message"
  patterns = ["%%{FAVORITE_DOG:pet}", "%%{FAVORITE_CAT:pet}"]
  pattern_definitions = {
    FAVORITE_DOG = "beagle"
    FAVORITE_CAT = "burmese"
  }
}
`
