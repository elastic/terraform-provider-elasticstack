package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorSetSecurityUser(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorSetSecurityUser,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "field", "user"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_set_security_user.test", "json", expectedJsonSetSecurityUser),
				),
			},
		},
	})
}

const expectedJsonSetSecurityUser = `{
	"set_security_user": {
		"field": "user",
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorSetSecurityUser = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set_security_user" "test" {
  field = "user"
}
`
