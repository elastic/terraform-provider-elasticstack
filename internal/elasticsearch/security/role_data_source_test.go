package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSecurityRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
				),
			},
		},
	})
}

const testAccDataSourceSecurityRole = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "data_source_test"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
`
