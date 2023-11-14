package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorKV(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorKV,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_kv.test", "json", expectedJsonKV),
				),
			},
		},
	})
}

const expectedJsonKV = `{
  "kv": {
		"exclude_keys": [
			"tags"
		],
		"field": "message",
		"field_split": " ",
		"ignore_failure": false,
		"ignore_missing": false,
		"prefix": "setting_",
		"strip_brackets": false,
		"value_split": "="
	}
}
`

const testAccDataSourceIngestProcessorKV = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_kv" "test" {
  field       = "message"
  field_split = " "
  value_split = "="

  exclude_keys = ["tags"]
  prefix       = "setting_"
}
`
