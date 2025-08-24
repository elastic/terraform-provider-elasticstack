package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorDissect(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorDissect,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "field", "message"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_dissect.test", "json", expectedJsonDissect),
				),
			},
		},
	})
}

const expectedJsonDissect = `{
  "dissect": {
		"append_separator": "",
		"field": "message",
		"ignore_failure": false,
		"ignore_missing": false,
		"pattern": "%{clientip} %{ident} %{auth} [%{@timestamp}] \"%{verb} %{request} HTTP/%{httpversion}\" %{status} %{size}"
	}
}
`

const testAccDataSourceIngestProcessorDissect = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dissect" "test" {
  field   = "message"
  pattern = "%%{clientip} %%{ident} %%{auth} [%%{@timestamp}] \"%%{verb} %%{request} HTTP/%%{httpversion}\" %%{status} %%{size}"
}
`
